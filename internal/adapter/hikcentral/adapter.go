package hikcentral

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eastwater-backend/internal/config"
	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/port"
	"eastwater-backend/internal/core/schema"
)

var _ port.HikCentralPort = (*Adapter)(nil)

// Adapter implements HikCentralPort using HikCentral Pro Visitor V2 Artemis OpenAPI
type Adapter struct {
	cfg        config.ThirdPartyConfig
	httpClient *http.Client
}

// NewAdapter creates a new HikCentral Pro HTTP adapter
func NewAdapter(cfg config.ThirdPartyConfig) *Adapter {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &Adapter{
		cfg: cfg,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
}

// CreatePerson calls POST /artemis/api/visitor/v2/appointment to make a visitor reservation
func (a *Adapter) CreatePerson(ctx context.Context, payload *schema.HikCreatePersonPayload) (*schema.HikPersonItem, error) {
	// First check if already exists to enforce strict Create
	if _, err := a.GetPersonByCode(ctx, payload.PersonCode); err == nil {
		return nil, fmt.Errorf("%w: personCode %s already has an active appointment in HikCentral", common.ErrAlreadyExists, payload.PersonCode)
	}

	endpoint := fmt.Sprintf("%s/artemis/api/visitor/v2/appointment", strings.TrimRight(a.cfg.BaseURL, "/"))
	reqBodyObj := a.buildAppointmentV2Payload("", payload)

	bodyBytes, err := json.Marshal(reqBodyObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal appointment v2 payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	a.setHeaders(req, "/artemis/api/visitor/v2/appointment", bodyBytes)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrThirdPartyService, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var hikResp schema.HikApiResponse[schema.HikAppointmentV2Result]
	if err := json.Unmarshal(respBody, &hikResp); err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrThirdPartyBadResp, err)
	}

	if hikResp.Code != "0" {
		return nil, &common.ThirdPartyError{
			StatusCode: resp.StatusCode,
			ErrorCode:  hikResp.Code,
			Message:    hikResp.Msg,
			RawBody:    string(respBody),
		}
	}

	faceBase64 := payload.GetFaceBase64()
	return &schema.HikPersonItem{
		AppointRecordID: hikResp.Data.AppointRecordID,
		PersonID:        hikResp.Data.VisitorID,
		PersonCode:      payload.PersonCode,
		PersonName:      payload.PersonName,
		BeginTime:       reqBodyObj.AppointStartTime,
		EndTime:         reqBodyObj.AppointEndTime,
		Vehicles:        payload.Vehicles,
		DeviceNames:     payload.DeviceNames,
		AccessLevelIDs:  payload.AccessLevelIDs,
		HasFaceImage:    faceBase64 != "",
		FaceBase64:      faceBase64,
		FacesCount:      1,
		Status:          "ACTIVE",
	}, nil
}

// UpdatePerson calls POST /artemis/api/visitor/v2/appointment/update to edit reservation record
func (a *Adapter) UpdatePerson(ctx context.Context, payload *schema.HikCreatePersonPayload) (*schema.HikPersonItem, error) {
	// 1. Resolve appointRecordId if not supplied
	appointRecordID := payload.AppointRecordID
	if appointRecordID == "" {
		existing, err := a.GetPersonByCode(ctx, payload.PersonCode)
		if err != nil {
			return nil, fmt.Errorf("%w: cannot update non-existent personCode %s", common.ErrNotFound, payload.PersonCode)
		}
		appointRecordID = existing.AppointRecordID
	}

	endpoint := fmt.Sprintf("%s/artemis/api/visitor/v2/appointment/update", strings.TrimRight(a.cfg.BaseURL, "/"))
	reqBodyObj := a.buildAppointmentV2Payload(appointRecordID, payload)

	bodyBytes, err := json.Marshal(reqBodyObj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal appointment update payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	a.setHeaders(req, "/artemis/api/visitor/v2/appointment/update", bodyBytes)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrThirdPartyService, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var hikResp schema.HikApiResponse[map[string]interface{}]
	if err := json.Unmarshal(respBody, &hikResp); err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrThirdPartyBadResp, err)
	}

	if hikResp.Code != "0" {
		return nil, &common.ThirdPartyError{
			StatusCode: resp.StatusCode,
			ErrorCode:  hikResp.Code,
			Message:    hikResp.Msg,
			RawBody:    string(respBody),
		}
	}

	faceBase64 := payload.GetFaceBase64()
	return &schema.HikPersonItem{
		AppointRecordID: appointRecordID,
		PersonCode:      payload.PersonCode,
		PersonName:      payload.PersonName,
		BeginTime:       reqBodyObj.AppointStartTime,
		EndTime:         reqBodyObj.AppointEndTime,
		Vehicles:        payload.Vehicles,
		DeviceNames:     payload.DeviceNames,
		AccessLevelIDs:  payload.AccessLevelIDs,
		HasFaceImage:    faceBase64 != "",
		FaceBase64:      faceBase64,
		FacesCount:      1,
		Status:          "ACTIVE",
	}, nil
}

// UpsertPerson creates or updates a visitor reservation record
func (a *Adapter) UpsertPerson(ctx context.Context, payload *schema.HikCreatePersonPayload) (*schema.HikPersonItem, bool, error) {
	existing, err := a.GetPersonByCode(ctx, payload.PersonCode)
	if err == nil && existing != nil {
		payload.AppointRecordID = existing.AppointRecordID
		updated, updateErr := a.UpdatePerson(ctx, payload)
		return updated, false, updateErr
	}

	created, createErr := a.CreatePerson(ctx, payload)
	return created, true, createErr
}

// GetPeople queries reservation records using POST /artemis/api/visitor/v1/appointment/appointmentlist
func (a *Adapter) GetPeople(ctx context.Context, query schema.PersonSearchQuery) ([]schema.HikPersonItem, int, error) {
	endpoint := fmt.Sprintf("%s/artemis/api/visitor/v1/appointment/appointmentlist", strings.TrimRight(a.cfg.BaseURL, "/"))

	startTime := time.Now().Add(-30 * 24 * time.Hour).Format("2006-01-02T15:04:05+07:00")
	endTime := time.Now().Add(365 * 24 * time.Hour).Format("2006-01-02T15:04:05+07:00")

	reqBodyObj := schema.HikAppointmentListRequest{
		PageNo:           query.Page,
		PageSize:         query.Limit,
		AppointStartTime: startTime,
		AppointEndTime:   endTime,
		VisitorName:      query.PersonName,
		IdentiCode:       query.PersonCode,
		AppointState:     -1, // All
	}

	bodyBytes, err := json.Marshal(reqBodyObj)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, 0, err
	}

	a.setHeaders(req, "/artemis/api/visitor/v1/appointment/appointmentlist", bodyBytes)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", common.ErrThirdPartyService, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var hikResp schema.HikApiResponse[schema.HikAppointmentListResult]
	if err := json.Unmarshal(respBody, &hikResp); err != nil {
		return nil, 0, fmt.Errorf("%w: %v", common.ErrThirdPartyBadResp, err)
	}

	if hikResp.Code != "0" {
		return nil, 0, &common.ThirdPartyError{
			StatusCode: resp.StatusCode,
			ErrorCode:  hikResp.Code,
			Message:    hikResp.Msg,
			RawBody:    string(respBody),
		}
	}

	var results []schema.HikPersonItem
	for _, item := range hikResp.Data.List {
		fullName := strings.TrimSpace(item.VisitorInfo.VisitorGivenName + " " + item.VisitorInfo.VisitorFamilyName)
		var vehicles []schema.HikVehicleData
		if item.VisitorInfo.PlateNo != "" {
			vehicles = append(vehicles, schema.HikVehicleData{PlateNo: item.VisitorInfo.PlateNo})
		}

		results = append(results, schema.HikPersonItem{
			AppointRecordID: item.AppointID,
			PersonID:        item.VisitorInfo.VisitorID,
			PersonCode:      item.VisitorInfo.CertificateNo,
			PersonName:      fullName,
			BeginTime:       item.AppointStartTime,
			EndTime:         item.AppointEndTime,
			Vehicles:        vehicles,
			FacesCount:      item.VisitorInfo.HaveVisitorHeadPic,
			Status:          "ACTIVE",
		})
	}

	return results, hikResp.Data.Total, nil
}

// GetPersonByCode retrieves single reservation record matching personCode
func (a *Adapter) GetPersonByCode(ctx context.Context, personCode string) (*schema.HikPersonItem, error) {
	list, total, err := a.GetPeople(ctx, schema.PersonSearchQuery{
		PaginationQuery: schema.PaginationQuery{Page: 1, Limit: 10},
		PersonCode:      personCode,
	})
	if err != nil {
		return nil, err
	}
	if total == 0 || len(list) == 0 {
		return nil, common.ErrNotFound
	}

	for _, item := range list {
		if item.PersonCode == personCode {
			return &item, nil
		}
	}

	return nil, common.ErrNotFound
}

// DeletePeople deletes reservations using POST /artemis/api/visitor/v1/appointment/single/delete
func (a *Adapter) DeletePeople(ctx context.Context, personCodes []string) ([]string, []string, error) {
	var deleted []string
	var failed []string

	endpoint := fmt.Sprintf("%s/artemis/api/visitor/v1/appointment/single/delete", strings.TrimRight(a.cfg.BaseURL, "/"))

	for _, code := range personCodes {
		trimmed := strings.TrimSpace(code)
		if trimmed == "" {
			continue
		}

		// Look up appointRecordId
		person, err := a.GetPersonByCode(ctx, trimmed)
		if err != nil || person.AppointRecordID == "" {
			failed = append(failed, trimmed)
			continue
		}

		reqBodyBytes, _ := json.Marshal(map[string]string{
			"appointRecordId": person.AppointRecordID,
		})

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(reqBodyBytes))
		if err != nil {
			failed = append(failed, trimmed)
			continue
		}

		a.setHeaders(req, "/artemis/api/visitor/v1/appointment/single/delete", reqBodyBytes)

		resp, err := a.httpClient.Do(req)
		if err != nil || resp.StatusCode >= 300 {
			failed = append(failed, trimmed)
			if resp != nil {
				_ = resp.Body.Close()
			}
			continue
		}
		_ = resp.Body.Close()
		deleted = append(deleted, trimmed)
	}

	if len(deleted) == 0 && len(failed) > 0 {
		return nil, failed, fmt.Errorf("%w: failed to delete requested personCodes", common.ErrNotFound)
	}

	return deleted, failed, nil
}

// CheckHealth verifies connection to HikCentral Pro
func (a *Adapter) CheckHealth(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s/artemis/api/common/v1/version", strings.TrimRight(a.cfg.BaseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	a.setHeaders(req, "/artemis/api/common/v1/version", nil)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("hikcentral server returned status %d", resp.StatusCode)
	}
	return nil
}

// buildAppointmentV2Payload constructs the HikVisitorAppointmentV2Request payload
func (a *Adapter) buildAppointmentV2Payload(appointRecordID string, payload *schema.HikCreatePersonPayload) *schema.HikVisitorAppointmentV2Request {
	startTime := payload.BeginTime
	if startTime == "" {
		startTime = time.Now().Format("2006-01-02T15:04:05+07:00")
	}

	endTime := payload.EndTime
	if endTime == "" {
		endTime = time.Now().Add(365 * 24 * time.Hour).Format("2006-01-02T15:04:05+07:00")
	}

	givenName, familyName := splitName(payload.PersonName)

	plateNo := ""
	if len(payload.Vehicles) > 0 {
		plateNo = payload.Vehicles[0].PlateNo
	}

	var faces []schema.HikFaceDataItem
	if faceStr := payload.GetFaceBase64(); faceStr != "" {
		faces = append(faces, schema.HikFaceDataItem{FaceData: faceStr})
	}

	var accessInfo *schema.HikAppointmentAccessInfo
	if len(payload.AccessLevelIDs) > 0 {
		var list []schema.HikAppointmentAccessLevelWrapper
		for _, idStr := range payload.AccessLevelIDs {
			if idInt, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				list = append(list, schema.HikAppointmentAccessLevelWrapper{
					AccessLevel: schema.HikAppointmentAccessLevel{
						ID: idInt,
					},
				})
			}
		}
		if len(list) > 0 {
			accessInfo = &schema.HikAppointmentAccessInfo{
				AccessLevelList: list,
			}
		}
	}

	return &schema.HikVisitorAppointmentV2Request{
		AppointRecordID:  appointRecordID,
		AppointStartTime: startTime,
		AppointEndTime:   endTime,
		VisitReasonType:  0, // 0-business
		AccessInfo:       accessInfo,
		VisitorInfoList: []schema.HikVisitorInfoWrapper{
			{
				VisitorInfo: schema.HikVisitorInfo{
					VisitorGivenName:  givenName,
					VisitorFamilyName: familyName,
					Gender:            0, // 0-unknown
					PlateNo:           plateNo,
					CertificateType:   111,
					CertificateNo:     payload.PersonCode,
					Faces:             faces,
				},
			},
		},
	}
}

func (a *Adapter) setHeaders(req *http.Request, path string, body []byte) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("userId", "admin")

	if a.cfg.APIKey != "" {
		req.Header.Set("x-ca-key", a.cfg.APIKey)
		if a.cfg.Secret != "" {
			signature := a.generateArtemisSignature(req.Method, path, a.cfg.Secret)
			req.Header.Set("x-ca-signature", signature)
		}
	}
}

func (a *Adapter) generateArtemisSignature(method, path, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(method + "\n" + path))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func splitName(fullName string) (string, string) {
	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return "Unknown", "."
	}
	if len(parts) == 1 {
		return parts[0], "."
	}
	return parts[0], strings.Join(parts[1:], " ")
}
