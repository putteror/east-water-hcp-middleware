package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"eastwater-backend/internal/adapter/hikcentral"
	"eastwater-backend/internal/core/schema"
	"eastwater-backend/internal/core/service"
	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockAdapter := hikcentral.NewMockAdapter()
	personSvc := service.NewPersonService(mockAdapter)
	personHandler := NewPersonHandler(personSvc)

	api := r.Group("/api/people")
	{
		api.GET("", personHandler.GetPeople)
		api.GET("/:personCode", personHandler.GetPersonByCode)
		api.POST("", personHandler.CreatePerson)
		api.PUT("/:personCode", personHandler.UpdatePerson)
		api.POST("/upsert", personHandler.UpsertPerson)
		api.PUT("/upsert", personHandler.UpsertPerson)
		api.DELETE("", personHandler.DeletePeople)
		api.DELETE("/:personCode", personHandler.DeletePersonByCode)
	}

	return r
}

func TestPersonHandler_GetPeople(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/people?page=1&limit=10", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Error("expected success = true")
	}
	if len(resp.Data) < 2 {
		t.Errorf("expected at least 2 people, got %d", len(resp.Data))
	}
}

func TestPersonHandler_CreatePerson_JSON(t *testing.T) {
	r := setupTestRouter()

	t.Run("JSON payload - Success", func(t *testing.T) {
		reqPayload := schema.CreatePersonRequest{
			PersonName:        "Kittisak S.",
			PersonCode:        "P777",
			LicensePlateTexts: []string{"5กก7777"},
			DeviceNames:       []string{"Main Entrance"},
			StartDateTime:     "2026-06-01T00:00:00Z",
			EndDateTime:       "2026-12-31T23:59:59Z",
			FaceImage:         "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...sampleBase64...",
		}
		bodyBytes, _ := json.Marshal(reqPayload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/people", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Success bool `json:"success"`
			Data    struct {
				PersonCode   string   `json:"personCode"`
				HasFaceImage bool     `json:"hasFaceImage"`
				FaceImage    string   `json:"faceImage"`
				DeviceNames  []string `json:"deviceNames"`
				Action       string   `json:"action"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response JSON: %v", err)
		}
		if resp.Data.PersonCode != "P777" {
			t.Errorf("expected personCode P777, got %s", resp.Data.PersonCode)
		}
		if !resp.Data.HasFaceImage {
			t.Error("expected hasFaceImage = true")
		}
		if resp.Data.Action != "created" {
			t.Errorf("expected action created, got %s", resp.Data.Action)
		}
		if len(resp.Data.DeviceNames) == 0 {
			t.Errorf("expected non-empty deviceNames, got %v", resp.Data.DeviceNames)
		}

		// Ensure hikCentralPersonId is NOT in the JSON response
		var rawMap map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &rawMap)
		dataMap := rawMap["data"].(map[string]interface{})
		if _, exists := dataMap["hikCentralPersonId"]; exists {
			t.Errorf("expected hikCentralPersonId to be removed from response, but was found")
		}
	})

	t.Run("Multipart fallback - Success", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("personName", "Multipart Person")
		_ = writer.WriteField("personCode", "P778")
		part, err := writer.CreateFormFile("faceImage", "face.jpg")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		_, _ = part.Write([]byte("fake-jpeg-image-binary-data"))
		_ = writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/people", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
		}
	})
}

func TestPersonHandler_UpdatePerson(t *testing.T) {
	r := setupTestRouter()

	updatePayload := schema.UpdatePersonRequest{
		PersonName:        "Somchai Updated",
		LicensePlateTexts: []string{"8กก8888"},
		FaceImage:         "/9j/4AAQSkZJRg...updatedBase64...",
	}
	bodyBytes, _ := json.Marshal(updatePayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/people/P001", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			PersonName   string `json:"personName"`
			HasFaceImage bool   `json:"hasFaceImage"`
			Action       string `json:"action"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data.PersonName != "Somchai Updated" {
		t.Errorf("expected Somchai Updated, got %s", resp.Data.PersonName)
	}
	if !resp.Data.HasFaceImage {
		t.Error("expected hasFaceImage = true")
	}
	if resp.Data.Action != "updated" {
		t.Errorf("expected action updated, got %s", resp.Data.Action)
	}
}

func TestPersonHandler_UpsertPerson(t *testing.T) {
	r := setupTestRouter()

	// 1. Upsert on existing record (should update)
	existingPayload := schema.CreatePersonRequest{
		PersonCode: "P002",
		PersonName: "Wichai Upserted via HTTP",
		FaceImage:  "/9j/4AAQSkZJRg...upsertBase64...",
	}
	bodyBytes, _ := json.Marshal(existingPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/people/upsert", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2. Upsert on new record (should create)
	newPayload := schema.CreatePersonRequest{
		PersonCode: "P_UPSERT_NEW",
		PersonName: "New Person via Upsert",
	}
	bodyNewBytes, _ := json.Marshal(newPayload)

	wNew := httptest.NewRecorder()
	reqNew, _ := http.NewRequest(http.MethodPost, "/api/people/upsert", bytes.NewBuffer(bodyNewBytes))
	reqNew.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wNew, reqNew)

	if wNew.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d. Body: %s", wNew.Code, wNew.Body.String())
	}
}

func TestPersonHandler_DeletePeople(t *testing.T) {
	r := setupTestRouter()

	payload := map[string]interface{}{
		"personCodes": []string{"P001", "P002"},
	}
	bodyBytes, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/people", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestPersonHandler_GetPersonByCode(t *testing.T) {
	r := setupTestRouter()

	t.Run("Success - existing personCode", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/people/P001", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Success bool `json:"success"`
			Data    struct {
				PersonCode   string `json:"personCode"`
				PersonName   string `json:"personName"`
				HasFaceImage bool   `json:"hasFaceImage"`
				FaceImage    string `json:"faceImage"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Error("expected success = true")
		}
		if resp.Data.PersonCode != "P001" {
			t.Errorf("expected personCode P001, got %s", resp.Data.PersonCode)
		}
		if !resp.Data.HasFaceImage {
			t.Error("expected hasFaceImage = true")
		}
		if resp.Data.FaceImage == "" {
			t.Error("expected FaceImage string to not be empty")
		}
	})

	t.Run("NotFound - non-existent personCode", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/people/NON_EXISTENT_999", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Query by code alias - /api/people?code=P001", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/people?code=P001", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Success bool `json:"success"`
			Data    []struct {
				PersonCode string `json:"personCode"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Data) != 1 || resp.Data[0].PersonCode != "P001" {
			t.Errorf("expected 1 result with personCode P001, got %v", resp.Data)
		}
	})
}
