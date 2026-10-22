package handler

import (
	"net/http"

	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/service"
	"github.com/gin-gonic/gin"
)

// HealthHandler provides readiness and liveness health checks
type HealthHandler struct {
	integrationService *service.IntegrationService
}

// NewHealthHandler creates a new instance of HealthHandler
func NewHealthHandler(integrationService *service.IntegrationService) *HealthHandler {
	return &HealthHandler{
		integrationService: integrationService,
	}
}

// HealthCheck responds with current application status and 3rd party probe
// @Summary Health Check
// @Description Checks service health and outbound 3rd-party connectivity
// @Tags System
// @Produce json
// @Success 200 {object} common.APIResponse "Service Healthy"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	thirdPartyStatus := "connected"
	if err := h.integrationService.CheckThirdPartyHealth(c.Request.Context()); err != nil {
		thirdPartyStatus = "unreachable"
	}

	common.SuccessResponse(c, "Service is healthy", gin.H{
		"status":      "ok",
		"database":    "none (stateless proxy)",
		"third_party": thirdPartyStatus,
	})
}

// Ping provides a simple ping endpoint
// @Summary Ping
// @Description Simple liveness ping endpoint
// @Tags System
// @Produce json
// @Success 200 {object} map[string]string "pong"
// @Router /ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
