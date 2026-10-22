package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"eastwater-backend/internal/adapter/hikcentral"
	"eastwater-backend/internal/adapter/thirdparty"
	"eastwater-backend/internal/config"
	"eastwater-backend/internal/core/service"
	"eastwater-backend/internal/interface/http/handler"
)

func TestRouter_SwaggerAndHealth(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Env:  "test",
			Port: "8080",
		},
		Auth: config.AuthConfig{
			APIKey: "test_key",
		},
	}

	thirdPartyAdapter := thirdparty.NewMockAdapter()
	hikCentralAdapter := hikcentral.NewMockAdapter()

	integrationSvc := service.NewIntegrationService(thirdPartyAdapter)
	personSvc := service.NewPersonService(hikCentralAdapter)

	integrationHdl := handler.NewIntegrationHandler(integrationSvc)
	personHdl := handler.NewPersonHandler(personSvc)
	healthHdl := handler.NewHealthHandler(integrationSvc)

	r := NewRouter(RouterDeps{
		Config:             cfg,
		IntegrationHandler: integrationHdl,
		PersonHandler:      personHdl,
		HealthHandler:      healthHdl,
	})

	t.Run("Swagger UI redirects or serves index.html", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/swagger/index.html", nil)
		req.RequestURI = "/swagger/index.html"
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 for swagger index, got %d", w.Code)
		}
	})

	t.Run("Swagger doc.json is accessible", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
		req.RequestURI = "/swagger/doc.json"
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 for swagger doc.json, got %d", w.Code)
		}
	})
}
