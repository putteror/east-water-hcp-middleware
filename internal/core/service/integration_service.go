package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/port"
	"eastwater-backend/internal/core/schema"
	"github.com/google/uuid"
)

// IntegrationService defines use cases for data transmission and retrieval via 3rd party
type IntegrationService struct {
	thirdPartyPort port.ThirdPartyPort
}

// NewIntegrationService creates a new instance of IntegrationService
func NewIntegrationService(thirdPartyPort port.ThirdPartyPort) *IntegrationService {
	return &IntegrationService{
		thirdPartyPort: thirdPartyPort,
	}
}

// TransmitData validates client payload, converts to 3rd party format, calls outbound adapter, and returns response DTO
func (s *IntegrationService) TransmitData(ctx context.Context, req *schema.TransmitDataRequest) (*schema.TransmitDataResponse, error) {
	if req.StationCode == "" || req.Parameter == "" {
		return nil, common.ErrBadRequest
	}

	// 1. Convert client Request DTO -> 3rd party Request payload
	timestampStr := req.RecordedAt.Format(time.RFC3339)
	if req.RecordedAt.IsZero() {
		timestampStr = time.Now().Format(time.RFC3339)
	}

	thirdPartyPayload := &schema.ThirdPartyIngestRequest{
		DeviceID:   req.StationCode,
		MetricName: req.Parameter,
		MetricVal:  req.Value,
		MetricUnit: req.Unit,
		Timestamp:  timestampStr,
		Tags:       req.Metadata,
	}

	// 2. Call Outbound Adapter Port
	thirdPartyResp, err := s.thirdPartyPort.TransmitData(ctx, thirdPartyPayload)
	if err != nil {
		return nil, err
	}

	// 3. Convert 3rd party response -> Client Response DTO
	txID := uuid.New().String()
	resp := &schema.TransmitDataResponse{
		TransactionID: txID,
		Status:        thirdPartyResp.AckCode,
		StationCode:   req.StationCode,
		ProcessedAt:   time.Now(),
		ReferenceNo:   thirdPartyResp.TicketID,
		ExtraInfo: map[string]interface{}{
			"remote_message": thirdPartyResp.Message,
			"acknowledged_at": thirdPartyResp.Timestamp,
		},
	}

	return resp, nil
}

// GetDataList queries list from 3rd party and converts to client response DTOs
func (s *IntegrationService) GetDataList(ctx context.Context, query schema.DataSearchQuery) ([]schema.DataBasicInfo, *schema.PaginationData, error) {
	query.SetDefaults()

	// 1. Fetch from 3rd party via port
	tpResp, err := s.thirdPartyPort.FetchDataList(ctx, query)
	if err != nil {
		return nil, nil, err
	}

	// 2. Convert 3rd party items to client DTOs
	dataList := make([]schema.DataBasicInfo, 0, len(tpResp.Items))
	for _, item := range tpResp.Items {
		occurredAt, _ := time.Parse(time.RFC3339, item.OccurredAt)
		if occurredAt.IsZero() {
			occurredAt = time.Now()
		}

		dataList = append(dataList, schema.DataBasicInfo{
			ID:          item.RecordID,
			StationCode: item.DeviceID,
			Parameter:   item.MetricName,
			Value:       item.Value,
			Unit:        item.Unit,
			Timestamp:   occurredAt,
			Status:      item.State,
		})
	}

	// 3. Calculate pagination
	totalPages := int(math.Ceil(float64(tpResp.TotalCount) / float64(query.Limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	pageData := &schema.PaginationData{
		CurrentPage: query.Page,
		Limit:       query.Limit,
		TotalPages:  totalPages,
		TotalRows:   tpResp.TotalCount,
	}

	return dataList, pageData, nil
}

// GetDataByID queries single detail record by ID from 3rd party
func (s *IntegrationService) GetDataByID(ctx context.Context, id string) (*schema.DataDetailResponse, error) {
	if id == "" {
		return nil, common.ErrBadRequest
	}

	// 1. Fetch from 3rd party via port
	item, err := s.thirdPartyPort.FetchDataByID(ctx, id)
	if err != nil {
		return nil, err
	}

	occurredAt, _ := time.Parse(time.RFC3339, item.OccurredAt)
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}

	// 2. Map to Client Response DTO
	return &schema.DataDetailResponse{
		ID:          item.RecordID,
		StationCode: item.DeviceID,
		Parameter:   item.MetricName,
		Value:       item.Value,
		Unit:        item.Unit,
		Timestamp:   occurredAt,
		Status:      item.State,
		RawPayload:  item.Attributes,
	}, nil
}

// CheckThirdPartyHealth checks 3rd party connectivity
func (s *IntegrationService) CheckThirdPartyHealth(ctx context.Context) error {
	if err := s.thirdPartyPort.CheckHealth(ctx); err != nil {
		return fmt.Errorf("3rd party service unhealthy: %w", err)
	}
	return nil
}
