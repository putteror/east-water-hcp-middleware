package thirdparty

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eastwater-backend/internal/config"
	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/schema"
)

func TestAdapter_TransmitData(t *testing.T) {
	t.Run("Success - 200 OK from 3rd party", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/api/v1/telemetry/ingest" {
				t.Errorf("expected path /api/v1/telemetry/ingest, got %s", r.URL.Path)
			}
			if r.Header.Get("Authorization") != "Bearer test-key" {
				t.Errorf("missing or invalid Authorization header")
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(schema.ThirdPartyIngestResponse{
				AckCode:   "ACCEPTED",
				TicketID:  "TK-12345",
				Message:   "Data received",
				Timestamp: time.Now().Format(time.RFC3339),
			})
		}))
		defer ts.Close()

		adapter := NewAdapter(config.ThirdPartyConfig{
			BaseURL: ts.URL,
			APIKey:  "test-key",
			Timeout: 5 * time.Second,
		})

		resp, err := adapter.TransmitData(context.Background(), &schema.ThirdPartyIngestRequest{
			DeviceID:   "STATION-01",
			MetricName: "PRESSURE",
			MetricVal:  3.5,
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.AckCode != "ACCEPTED" {
			t.Errorf("expected ACCEPTED, got %s", resp.AckCode)
		}
		if resp.TicketID != "TK-12345" {
			t.Errorf("expected TK-12345, got %s", resp.TicketID)
		}
	})

	t.Run("Fail - 400 Bad Request from 3rd party", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message":    "invalid metric name",
				"error_code": "ERR_INVALID_METRIC",
			})
		}))
		defer ts.Close()

		adapter := NewAdapter(config.ThirdPartyConfig{
			BaseURL: ts.URL,
			Timeout: 5 * time.Second,
		})

		_, err := adapter.TransmitData(context.Background(), &schema.ThirdPartyIngestRequest{
			DeviceID: "STATION-01",
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var tpErr *common.ThirdPartyError
		if !errors.As(err, &tpErr) {
			t.Fatalf("expected ThirdPartyError, got %T: %v", err, err)
		}
		if tpErr.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", tpErr.StatusCode)
		}
		if tpErr.ErrorCode != "ERR_INVALID_METRIC" {
			t.Errorf("expected ERR_INVALID_METRIC, got %s", tpErr.ErrorCode)
		}
	})
}
