package hikcentral

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eastwater-backend/internal/config"
	"eastwater-backend/internal/core/schema"
)

func TestHikCentralAdapter_VisitorV2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/artemis/api/visitor/v2/appointment":
			_ = json.NewEncoder(w).Encode(schema.HikApiResponse[schema.HikAppointmentV2Result]{
				Code: "0",
				Msg:  "Success",
				Data: schema.HikAppointmentV2Result{
					AppointRecordID: "rec-test-999",
					VisitorID:       "vis-test-999",
					AppointCode:     "CODE999",
				},
			})

		case "/artemis/api/visitor/v2/appointment/update":
			_ = json.NewEncoder(w).Encode(schema.HikApiResponse[map[string]interface{}]{
				Code: "0",
				Msg:  "Success",
			})

		case "/artemis/api/visitor/v1/appointment/appointmentlist":
			var reqBody schema.HikAppointmentListRequest
			_ = json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody.IdentiCode == "NEW_P999" {
				_ = json.NewEncoder(w).Encode(schema.HikApiResponse[schema.HikAppointmentListResult]{
					Code: "0",
					Msg:  "Success",
					Data: schema.HikAppointmentListResult{
						Total:    0,
						PageNo:   1,
						PageSize: 10,
						List:     []schema.HikAppointmentListItem{},
					},
				})
				return
			}

			_ = json.NewEncoder(w).Encode(schema.HikApiResponse[schema.HikAppointmentListResult]{
				Code: "0",
				Msg:  "Success",
				Data: schema.HikAppointmentListResult{
					Total:    1,
					PageNo:   1,
					PageSize: 10,
					List: []schema.HikAppointmentListItem{
						{
							AppointID:        "rec-test-999",
							AppointCode:      "CODE999",
							AppointStartTime: "2026-05-01T08:00:00+07:00",
							AppointEndTime:   "2026-12-31T18:00:00+07:00",
							VisitorInfo: schema.HikAppointmentItemVisitor{
								VisitorID:         "vis-test-999",
								VisitorGivenName:  "Somchai",
								VisitorFamilyName: "Jaidee",
								PlateNo:           "1กก1234",
								CertificateNo:     "P001",
							},
						},
					},
				},
			})

		case "/artemis/api/visitor/v1/appointment/single/delete":
			_ = json.NewEncoder(w).Encode(schema.HikApiResponse[map[string]interface{}]{
				Code: "0",
				Msg:  "Success",
			})

		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	adapter := NewAdapter(config.ThirdPartyConfig{
		BaseURL: ts.URL,
		APIKey:  "test-app-key",
		Secret:  "test-app-secret",
		Timeout: 5 * time.Second,
	})

	ctx := context.Background()

	t.Run("GetPeople - lists appointments", func(t *testing.T) {
		list, total, err := adapter.GetPeople(ctx, schema.PersonSearchQuery{
			PaginationQuery: schema.PaginationQuery{Page: 1, Limit: 10},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(list) != 1 || list[0].PersonCode != "P001" {
			t.Errorf("expected personCode P001, got %+v", list)
		}
	})

	t.Run("CreatePerson - makes v2 appointment", func(t *testing.T) {
		item, err := adapter.CreatePerson(ctx, &schema.HikCreatePersonPayload{
			PersonCode: "NEW_P999",
			PersonName: "Wichai S.",
			BeginTime:  "2026-06-01T00:00:00+07:00",
			EndTime:    "2026-12-31T23:59:59+07:00",
			Vehicles:   []schema.HikVehicleData{{PlateNo: "5กก5555"}},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if item.AppointRecordID != "rec-test-999" {
			t.Errorf("expected appointRecordId rec-test-999, got %s", item.AppointRecordID)
		}
	})

	t.Run("UpdatePerson - edits v2 appointment", func(t *testing.T) {
		item, err := adapter.UpdatePerson(ctx, &schema.HikCreatePersonPayload{
			AppointRecordID: "rec-test-999",
			PersonCode:      "P001",
			PersonName:      "Somchai Updated",
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if item.AppointRecordID != "rec-test-999" {
			t.Errorf("expected rec-test-999, got %s", item.AppointRecordID)
		}
	})

	t.Run("DeletePeople - calls single/delete", func(t *testing.T) {
		deleted, failed, err := adapter.DeletePeople(ctx, []string{"P001"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(deleted) != 1 || deleted[0] != "P001" {
			t.Errorf("expected P001 deleted, got %v, failed %v", deleted, failed)
		}
	})
}
