package thirdparty

import (
	"context"
	"time"

	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/port"
	"eastwater-backend/internal/core/schema"
	"github.com/google/uuid"
)

// Ensure MockAdapter implements ThirdPartyPort
var _ port.ThirdPartyPort = (*MockAdapter)(nil)

// MockAdapter provides an in-memory mock implementation for testing and development
type MockAdapter struct {
	records map[string]*schema.ThirdPartyItemData
}

// NewMockAdapter initializes a mock adapter with sample data
func NewMockAdapter() *MockAdapter {
	m := &MockAdapter{
		records: make(map[string]*schema.ThirdPartyItemData),
	}

	// Pre-seed some mock records
	sampleID := "rec-001"
	m.records[sampleID] = &schema.ThirdPartyItemData{
		RecordID:   sampleID,
		DeviceID:   "STATION-EW-01",
		MetricName: "WATER_PRESSURE",
		Value:      4.25,
		Unit:       "bar",
		OccurredAt: time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		State:      "NORMAL",
		Attributes: map[string]interface{}{"battery": "98%"},
	}

	return m
}

func (m *MockAdapter) TransmitData(ctx context.Context, payload *schema.ThirdPartyIngestRequest) (*schema.ThirdPartyIngestResponse, error) {
	ticketID := "TKT-" + uuid.New().String()[:8]
	recordID := "rec-" + uuid.New().String()[:8]

	m.records[recordID] = &schema.ThirdPartyItemData{
		RecordID:   recordID,
		DeviceID:   payload.DeviceID,
		MetricName: payload.MetricName,
		Value:      payload.MetricVal,
		Unit:       payload.MetricUnit,
		OccurredAt: payload.Timestamp,
		State:      "ACCEPTED",
		Attributes: payload.Tags,
	}

	return &schema.ThirdPartyIngestResponse{
		AckCode:   "SUCCESS",
		TicketID:  ticketID,
		Message:   "Telemetry data ingested successfully (MOCK)",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

func (m *MockAdapter) FetchDataList(ctx context.Context, query schema.DataSearchQuery) (*schema.ThirdPartyQueryResponse, error) {
	var items []schema.ThirdPartyItemData
	for _, rec := range m.records {
		if query.StationCode != "" && rec.DeviceID != query.StationCode {
			continue
		}
		if query.Parameter != "" && rec.MetricName != query.Parameter {
			continue
		}
		items = append(items, *rec)
	}

	return &schema.ThirdPartyQueryResponse{
		TotalCount: len(items),
		Items:      items,
	}, nil
}

func (m *MockAdapter) FetchDataByID(ctx context.Context, id string) (*schema.ThirdPartyItemData, error) {
	item, ok := m.records[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	return item, nil
}

func (m *MockAdapter) CheckHealth(ctx context.Context) error {
	return nil
}
