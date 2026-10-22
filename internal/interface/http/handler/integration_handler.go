package handler

import (
	"errors"
	"net/http"

	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/schema"
	"eastwater-backend/internal/core/service"
	"github.com/gin-gonic/gin"
)

// IntegrationHandler handles HTTP requests for data transmission and querying
type IntegrationHandler struct {
	service *service.IntegrationService
}

// NewIntegrationHandler creates a new instance of IntegrationHandler
func NewIntegrationHandler(service *service.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{
		service: service,
	}
}

// TransmitData handles POST /api/v1/telemetry/data
func (h *IntegrationHandler) TransmitData(c *gin.Context) {
	var req schema.TransmitDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	resp, err := h.service.TransmitData(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	common.SuccessResponse(c, "Telemetry data transmitted successfully", resp)
}

// GetDataList handles GET /api/v1/telemetry/data
func (h *IntegrationHandler) GetDataList(c *gin.Context) {
	var query schema.DataSearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	dataList, pageData, err := h.service.GetDataList(c.Request.Context(), query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	common.GetDataListResponse(c, "Telemetry records retrieved successfully", dataList, pageData)
}

// GetDataByID handles GET /api/v1/telemetry/data/:id
func (h *IntegrationHandler) GetDataByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "Missing record ID")
		return
	}

	resp, err := h.service.GetDataByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	common.SuccessResponse(c, "Telemetry record details retrieved successfully", resp)
}

// handleError maps domain/adapter errors to appropriate HTTP responses
func (h *IntegrationHandler) handleError(c *gin.Context, err error) {
	var tpErr *common.ThirdPartyError
	if errors.As(err, &tpErr) {
		// External 3rd party returned a non-2xx status code
		status := http.StatusBadGateway
		if tpErr.StatusCode == http.StatusBadRequest {
			status = http.StatusBadRequest
		} else if tpErr.StatusCode == http.StatusUnauthorized || tpErr.StatusCode == http.StatusForbidden {
			status = http.StatusBadGateway // From our client's perspective, upstream credentials failed
		}
		common.ErrorResponse(c, status, "Upstream 3rd party error: "+tpErr.Message)
		return
	}

	switch {
	case errors.Is(err, common.ErrBadRequest):
		common.ErrorResponse(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, common.ErrNotFound):
		common.ErrorResponse(c, http.StatusNotFound, "Record not found")
	case errors.Is(err, common.ErrThirdPartyTimeout):
		common.ErrorResponse(c, http.StatusGatewayTimeout, "Upstream 3rd party request timed out")
	case errors.Is(err, common.ErrThirdPartyService), errors.Is(err, common.ErrThirdPartyBadResp):
		common.ErrorResponse(c, http.StatusBadGateway, "Upstream 3rd party service error: "+err.Error())
	default:
		common.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
	}
}
