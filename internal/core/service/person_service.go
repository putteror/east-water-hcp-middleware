package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"strings"
	"time"

	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/port"
	"eastwater-backend/internal/core/schema"
)

// PersonService handles business logic and DTO conversions for VMS Person integration
type PersonService struct {
	hikPort port.HikCentralPort
}

// NewPersonService creates a new instance of PersonService
func NewPersonService(hikPort port.HikCentralPort) *PersonService {
	return &PersonService{
		hikPort: hikPort,
	}
}

// GetPeople retrieves the list of people from HikCentral and returns client-friendly DTOs
func (s *PersonService) GetPeople(ctx context.Context, query schema.PersonSearchQuery) ([]schema.PersonBasicInfo, *schema.PaginationData, error) {
	query.SetDefaults()

	items, total, err := s.hikPort.GetPeople(ctx, query)
	if err != nil {
		return nil, nil, err
	}

	basicList := make([]schema.PersonBasicInfo, 0, len(items))
	for _, item := range items {
		var plates []string
		for _, v := range item.Vehicles {
			if v.PlateNo != "" {
				plates = append(plates, v.PlateNo)
			}
		}

		basicList = append(basicList, schema.PersonBasicInfo{
			PersonCode:        item.PersonCode,
			PersonName:        item.PersonName,
			LicensePlateTexts: plates,
			DeviceNames:       item.DeviceNames,
			AccessLevelIDs:    item.AccessLevelIDs,
			StartDateTime:     item.BeginTime,
			EndDateTime:       item.EndTime,
			HasFaceImage:      item.HasFaceImage || item.GetFaceBase64() != "" || item.FacesCount > 0,
			Status:            item.Status,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(query.Limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	pageData := &schema.PaginationData{
		CurrentPage: query.Page,
		Limit:       query.Limit,
		TotalPages:  totalPages,
		TotalRows:   total,
	}

	return basicList, pageData, nil
}

// GetPersonByCode retrieves details for a single person by personCode
func (s *PersonService) GetPersonByCode(ctx context.Context, personCode string) (*schema.PersonResponse, error) {
	if personCode == "" {
		return nil, common.ErrBadRequest
	}

	item, err := s.hikPort.GetPersonByCode(ctx, personCode)
	if err != nil {
		return nil, err
	}

	var plates []string
	for _, v := range item.Vehicles {
		if v.PlateNo != "" {
			plates = append(plates, v.PlateNo)
		}
	}

	faceBase64 := item.GetFaceBase64()
	hasFace := item.HasFaceImage || faceBase64 != "" || item.FacesCount > 0

	return &schema.PersonResponse{
		PersonCode:        item.PersonCode,
		PersonName:        item.PersonName,
		LicensePlateTexts: plates,
		DeviceNames:       item.DeviceNames,
		AccessLevelIDs:    item.AccessLevelIDs,
		StartDateTime:     item.BeginTime,
		EndDateTime:       item.EndTime,
		HasFaceImage:      hasFace,
		FaceImage:         faceBase64,
		Status:            item.Status,
		ProcessedAt:       time.Now(),
	}, nil
}

// CreatePerson strictly creates a new person. Returns ErrAlreadyExists if personCode is already registered.
func (s *PersonService) CreatePerson(ctx context.Context, req *schema.CreatePersonRequest) (*schema.PersonResponse, error) {
	if req.PersonName == "" || req.PersonCode == "" {
		return nil, common.ErrBadRequest
	}

	faceBase64 := cleanBase64(req.FaceImage)
	if faceBase64 == "" {
		var err error
		faceBase64, err = s.processFaceImage(req.GetFaceImageFile())
		if err != nil {
			return nil, err
		}
	}

	plateTexts := req.GetLicensePlateList()
	vehicles := make([]schema.HikVehicleData, 0, len(plateTexts))
	for _, plate := range plateTexts {
		vehicles = append(vehicles, schema.HikVehicleData{PlateNo: plate})
	}

	deviceNames := schema.MatchDeviceNames(req.GetDeviceNames(), nil)
	accessLevelIDs := req.GetAccessLevelIDs()

	hikPayload := &schema.HikCreatePersonPayload{
		PersonCode:     req.PersonCode,
		PersonName:     req.PersonName,
		BeginTime:      req.StartDateTime,
		EndTime:        req.EndDateTime,
		Vehicles:       vehicles,
		DeviceNames:    deviceNames,
		AccessLevelIDs: accessLevelIDs,
		FaceBase64:     faceBase64,
	}

	result, err := s.hikPort.CreatePerson(ctx, hikPayload)
	if err != nil {
		return nil, err
	}

	resultFace := result.GetFaceBase64()
	if resultFace == "" {
		resultFace = faceBase64
	}

	resultDevices := result.DeviceNames
	if len(resultDevices) == 0 && len(deviceNames) > 0 {
		resultDevices = deviceNames
	}

	resultAccessLevels := result.AccessLevelIDs
	if len(resultAccessLevels) == 0 && len(accessLevelIDs) > 0 {
		resultAccessLevels = accessLevelIDs
	}

	return &schema.PersonResponse{
		Action:            "created",
		PersonCode:        result.PersonCode,
		PersonName:        result.PersonName,
		LicensePlateTexts: plateTexts,
		DeviceNames:       resultDevices,
		AccessLevelIDs:    resultAccessLevels,
		StartDateTime:     result.BeginTime,
		EndDateTime:       result.EndTime,
		HasFaceImage:      resultFace != "",
		FaceImage:         resultFace,
		Status:            result.Status,
		ProcessedAt:       time.Now(),
	}, nil
}

// UpdatePerson strictly updates an existing person. Returns ErrNotFound if personCode does not exist.
func (s *PersonService) UpdatePerson(ctx context.Context, personCode string, req *schema.UpdatePersonRequest) (*schema.PersonResponse, error) {
	if personCode == "" {
		return nil, common.ErrBadRequest
	}

	faceBase64 := cleanBase64(req.FaceImage)
	if faceBase64 == "" {
		var err error
		faceBase64, err = s.processFaceImage(req.GetFaceImageFile())
		if err != nil {
			return nil, err
		}
	}

	plateTexts := req.GetLicensePlateList()
	vehicles := make([]schema.HikVehicleData, 0, len(plateTexts))
	for _, plate := range plateTexts {
		vehicles = append(vehicles, schema.HikVehicleData{PlateNo: plate})
	}

	var deviceNames []string
	if rawDevices := req.GetDeviceNames(); len(rawDevices) > 0 {
		deviceNames = schema.MatchDeviceNames(rawDevices, nil)
	}
	accessLevelIDs := req.GetAccessLevelIDs()

	hikPayload := &schema.HikCreatePersonPayload{
		PersonCode:     personCode,
		PersonName:     req.PersonName,
		BeginTime:      req.StartDateTime,
		EndTime:        req.EndDateTime,
		Vehicles:       vehicles,
		DeviceNames:    deviceNames,
		AccessLevelIDs: accessLevelIDs,
		FaceBase64:     faceBase64,
	}

	result, err := s.hikPort.UpdatePerson(ctx, hikPayload)
	if err != nil {
		return nil, err
	}

	resultFace := result.GetFaceBase64()
	if resultFace == "" && faceBase64 != "" {
		resultFace = faceBase64
	}

	resultDevices := result.DeviceNames
	if len(resultDevices) == 0 && len(deviceNames) > 0 {
		resultDevices = deviceNames
	}

	resultAccessLevels := result.AccessLevelIDs
	if len(resultAccessLevels) == 0 && len(accessLevelIDs) > 0 {
		resultAccessLevels = accessLevelIDs
	}

	return &schema.PersonResponse{
		Action:            "updated",
		PersonCode:        result.PersonCode,
		PersonName:        result.PersonName,
		LicensePlateTexts: plateTexts,
		DeviceNames:       resultDevices,
		AccessLevelIDs:    resultAccessLevels,
		StartDateTime:     result.BeginTime,
		EndDateTime:       result.EndTime,
		HasFaceImage:      result.HasFaceImage || resultFace != "",
		FaceImage:         resultFace,
		Status:            result.Status,
		ProcessedAt:       time.Now(),
	}, nil
}

// UpsertPerson creates the person if non-existent, or updates them if already existing in HikCentral.
func (s *PersonService) UpsertPerson(ctx context.Context, req *schema.CreatePersonRequest) (*schema.PersonResponse, error) {
	if req.PersonCode == "" {
		return nil, common.ErrBadRequest
	}

	faceBase64 := cleanBase64(req.FaceImage)
	if faceBase64 == "" {
		var err error
		faceBase64, err = s.processFaceImage(req.GetFaceImageFile())
		if err != nil {
			return nil, err
		}
	}

	plateTexts := req.GetLicensePlateList()
	vehicles := make([]schema.HikVehicleData, 0, len(plateTexts))
	for _, plate := range plateTexts {
		vehicles = append(vehicles, schema.HikVehicleData{PlateNo: plate})
	}

	deviceNames := schema.MatchDeviceNames(req.GetDeviceNames(), nil)
	accessLevelIDs := req.GetAccessLevelIDs()

	hikPayload := &schema.HikCreatePersonPayload{
		PersonCode:     req.PersonCode,
		PersonName:     req.PersonName,
		BeginTime:      req.StartDateTime,
		EndTime:        req.EndDateTime,
		Vehicles:       vehicles,
		DeviceNames:    deviceNames,
		AccessLevelIDs: accessLevelIDs,
		FaceBase64:     faceBase64,
	}

	result, wasCreated, err := s.hikPort.UpsertPerson(ctx, hikPayload)
	if err != nil {
		return nil, err
	}

	action := "updated"
	if wasCreated {
		action = "created"
	}

	resultFace := result.GetFaceBase64()
	if resultFace == "" && faceBase64 != "" {
		resultFace = faceBase64
	}

	resultDevices := result.DeviceNames
	if len(resultDevices) == 0 && len(deviceNames) > 0 {
		resultDevices = deviceNames
	}

	resultAccessLevels := result.AccessLevelIDs
	if len(resultAccessLevels) == 0 && len(accessLevelIDs) > 0 {
		resultAccessLevels = accessLevelIDs
	}

	return &schema.PersonResponse{
		Action:            action,
		PersonCode:        result.PersonCode,
		PersonName:        result.PersonName,
		LicensePlateTexts: plateTexts,
		DeviceNames:       resultDevices,
		AccessLevelIDs:    resultAccessLevels,
		StartDateTime:     result.BeginTime,
		EndDateTime:       result.EndTime,
		HasFaceImage:      result.HasFaceImage || resultFace != "",
		FaceImage:         resultFace,
		Status:            result.Status,
		ProcessedAt:       time.Now(),
	}, nil
}

// DeletePeople removes persons from HikCentral by their personCodes
func (s *PersonService) DeletePeople(ctx context.Context, req *schema.DeletePeopleRequest) (*schema.PersonDeleteResponse, error) {
	if len(req.PersonCodes) == 0 {
		return nil, common.ErrBadRequest
	}

	deletedCodes, failedCodes, err := s.hikPort.DeletePeople(ctx, req.PersonCodes)
	if err != nil {
		return nil, err
	}

	return &schema.PersonDeleteResponse{
		TotalRequested: len(req.PersonCodes),
		DeletedCount:   len(deletedCodes),
		DeletedCodes:   deletedCodes,
		FailedCodes:    failedCodes,
	}, nil
}

// processFaceImage reads a single uploaded file header and converts to Base64 string
func (s *PersonService) processFaceImage(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", nil
	}
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded face image %s: %w", fileHeader.Filename, err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read uploaded face image %s: %w", fileHeader.Filename, err)
	}

	return base64.StdEncoding.EncodeToString(fileBytes), nil
}

// cleanBase64 strips data URI prefix (e.g. data:image/jpeg;base64,) if provided
func cleanBase64(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, ","); idx != -1 && strings.HasPrefix(s, "data:") {
		return strings.TrimSpace(s[idx+1:])
	}
	return s
}
