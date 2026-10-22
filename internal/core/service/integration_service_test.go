package service

import (
	"context"
	"testing"
	"time"

	"eastwater-backend/internal/adapter/thirdparty"
	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/schema"
)

func TestIntegrationService_TransmitData(t *testing.T) {
	mockPort := thirdparty.NewMockAdapter()
	svc := NewIntegrationService(mockPort)
	ctx := context.Background()

	t.Run("Success - valid payload", func(t *testing.T) {
		req := &schema.TransmitDataRequest{
			StationCode: "STATION-01",
			Parameter:   "WATER_LEVEL",
			Value:       12.5,
			Unit:        "m",
			RecordedAt:  time.Now(),
			Metadata:    map[string]interface{}{"sensor": "ultrasonic"},
		}

		resp, err := svc.TransmitData(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp == nil {
			t.Fatal("expected response, got nil")
		}
		if resp.StationCode != req.StationCode {
			t.Errorf("expected station code %s, got %s", req.StationCode, resp.StationCode)
		}
		if resp.TransactionID == "" {
			t.Error("expected non-empty transaction id")
		}
	})

	t.Run("Fail - missing station code", func(t *testing.T) {
		req := &schema.TransmitDataRequest{
			StationCode: "",
			Parameter:   "WATER_LEVEL",
			Value:       12.5,
		}

		_, err := svc.TransmitData(ctx, req)
		if err != common.ErrBadRequest {
			t.Errorf("expected ErrBadRequest, got %v", err)
		}
	})
}

func TestIntegrationService_GetDataByID(t *testing.T) {
	mockPort := thirdparty.NewMockAdapter()
	svc := NewIntegrationService(mockPort)
	ctx := context.Background()

	t.Run("Success - existing record", func(t *testing.T) {
		resp, err := svc.GetDataByID(ctx, "rec-001")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.ID != "rec-001" {
			t.Errorf("expected id rec-001, got %s", resp.ID)
		}
		if resp.StationCode != "STATION-EW-01" {
			t.Errorf("expected STATION-EW-01, got %s", resp.StationCode)
		}
	})

	t.Run("Fail - not found", func(t *testing.T) {
		_, err := svc.GetDataByID(ctx, "non-existent-id")
		if err != common.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
