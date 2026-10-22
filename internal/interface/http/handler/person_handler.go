package handler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strings"

	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/schema"
	"eastwater-backend/internal/core/service"
	"github.com/gin-gonic/gin"
)

// PersonHandler handles HTTP requests from VMS software for people management
type PersonHandler struct {
	service *service.PersonService
}

// NewPersonHandler creates a new instance of PersonHandler
func NewPersonHandler(service *service.PersonService) *PersonHandler {
	return &PersonHandler{
		service: service,
	}
}

// GetPeople handles GET /api/v1/people
// Retrieves list of people with search and pagination support
// @Summary List People
// @Description Retrieves list of registered visitors / people from HikCentral Pro with pagination
// @Tags People
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size limit" default(10)
// @Param personCode query string false "Search by personCode"
// @Param code query string false "Search by personCode alias"
// @Param personName query string false "Search by personName"
// @Success 200 {object} common.APIResponse{data=[]schema.PersonBasicInfo,page=schema.PaginationData} "Success"
// @Failure 400 {object} common.APIResponse "Bad Request"
// @Failure 500 {object} common.APIResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/people [get]
func (h *PersonHandler) GetPeople(c *gin.Context) {
	var query schema.PersonSearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	people, pageData, err := h.service.GetPeople(c.Request.Context(), query)
	if err != nil {
		h.handleError(c, err)
		return
	}

	common.GetDataListResponse(c, "People retrieved successfully", people, pageData)
}

// GetPersonByCode handles GET /api/v1/people/:personCode
// Retrieves details of a single person by personCode
// @Summary Get Person by Code
// @Description Retrieves full details of a single person by personCode including face image Base64 and vehicles
// @Tags People
// @Produce json
// @Param personCode path string true "Person Code"
// @Success 200 {object} common.APIResponse{data=schema.PersonResponse} "Success"
// @Failure 400 {object} common.APIResponse "Bad Request"
// @Failure 404 {object} common.APIResponse "Not Found"
// @Failure 500 {object} common.APIResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/people/{personCode} [get]
func (h *PersonHandler) GetPersonByCode(c *gin.Context) {
	personCode := c.Param("personCode")
	if personCode == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "Missing personCode parameter")
		return
	}

	person, err := h.service.GetPersonByCode(c.Request.Context(), personCode)
	if err != nil {
		h.handleError(c, err)
		return
	}

	common.SuccessResponse(c, "Person retrieved successfully", person)
}

// CreatePerson handles POST /api/v1/people
// Strictly creates a new person. If query ?upsert=true is provided, delegates to Upsert.
// @Summary Create Person (Strict)
// @Description Registers a new visitor in HikCentral Pro with Base64 face image and access permissions. To grant access, you can provide either `deviceNames` (supports partial match) OR `accessLevelIds` (HikCentral access levels), or both. Returns 409 if personCode exists.
// @Tags People
// @Accept json
// @Produce json
// @Param upsert query bool false "Enable upsert mode"
// @Param request body schema.CreatePersonRequest true "Person Registration JSON Payload"
// @Success 201 {object} common.APIResponse{data=schema.PersonResponse} "Created"
// @Failure 400 {object} common.APIResponse "Bad Request"
// @Failure 409 {object} common.APIResponse "Conflict - Already exists"
// @Failure 500 {object} common.APIResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/people [post]
func (h *PersonHandler) CreatePerson(c *gin.Context) {
	if c.Query("upsert") == "true" {
		h.UpsertPerson(c)
		return
	}

	var req schema.CreatePersonRequest
	if strings.Contains(c.ContentType(), "multipart/form-data") {
		_ = c.Request.ParseMultipartForm(32 << 20)
		if err := c.ShouldBind(&req); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Validation error: "+err.Error())
			return
		}
		if req.FaceImageFile == nil {
			req.FaceImageFile = h.extractFaceImage(c)
		}
		if req.FaceImage == "" {
			req.FaceImage = c.PostForm("faceImage")
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
			return
		}
	}

	resp, err := h.service.CreatePerson(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	common.CreatedResponse(c, "Person created successfully in HikCentral", resp)
}

// UpdatePerson handles PUT /api/v1/people/:personCode
// Strictly updates an existing person. Returns 404 if personCode not found.
// @Summary Update Person (Strict)
// @Description Updates an existing visitor in HikCentral Pro using JSON payload with optional Base64 faceImage and access permissions (specify either `deviceNames` OR `accessLevelIds`). Returns 404 if not found.
// @Tags People
// @Accept json
// @Produce json
// @Param personCode path string true "Person Code to update"
// @Param request body schema.UpdatePersonRequest true "Person Update JSON Payload"
// @Success 200 {object} common.APIResponse{data=schema.PersonResponse} "Updated"
// @Failure 400 {object} common.APIResponse "Bad Request"
// @Failure 404 {object} common.APIResponse "Not Found"
// @Failure 500 {object} common.APIResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/people/{personCode} [put]
func (h *PersonHandler) UpdatePerson(c *gin.Context) {
	personCode := c.Param("personCode")
	if personCode == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "Missing personCode parameter")
		return
	}

	var req schema.UpdatePersonRequest
	if strings.Contains(c.ContentType(), "multipart/form-data") {
		_ = c.Request.ParseMultipartForm(32 << 20)
		if err := c.ShouldBind(&req); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Validation error: "+err.Error())
			return
		}
		if req.FaceImageFile == nil {
			req.FaceImageFile = h.extractFaceImage(c)
		}
		if req.FaceImage == "" {
			req.FaceImage = c.PostForm("faceImage")
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
			return
		}
	}

	resp, err := h.service.UpdatePerson(c.Request.Context(), personCode, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	common.SuccessResponse(c, "Person updated successfully in HikCentral", resp)
}

// UpsertPerson handles POST /api/v1/people/upsert and PUT /api/v1/people/upsert
// Creates person if non-existent, or updates if already existing.
// @Summary Upsert Person
// @Description Creates a new visitor if not exists, or updates the existing visitor if personCode already exists. To grant access, you can provide either `deviceNames` OR `accessLevelIds`, or both.
// @Tags People
// @Accept json
// @Produce json
// @Param request body schema.CreatePersonRequest true "Person Upsert JSON Payload"
// @Success 200 {object} common.APIResponse{data=schema.PersonResponse} "Updated"
// @Success 201 {object} common.APIResponse{data=schema.PersonResponse} "Created"
// @Failure 400 {object} common.APIResponse "Bad Request"
// @Failure 500 {object} common.APIResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/people/upsert [post]
// @Router /api/people/upsert [put]
func (h *PersonHandler) UpsertPerson(c *gin.Context) {
	var req schema.CreatePersonRequest
	if strings.Contains(c.ContentType(), "multipart/form-data") {
		_ = c.Request.ParseMultipartForm(32 << 20)
		if err := c.ShouldBind(&req); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Validation error: "+err.Error())
			return
		}
		if req.FaceImageFile == nil {
			req.FaceImageFile = h.extractFaceImage(c)
		}
		if req.FaceImage == "" {
			req.FaceImage = c.PostForm("faceImage")
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
			return
		}
	}

	resp, err := h.service.UpsertPerson(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	if resp.Action == "created" {
		common.CreatedResponse(c, "Person created successfully in HikCentral (Upsert)", resp)
	} else {
		common.SuccessResponse(c, "Person updated successfully in HikCentral (Upsert)", resp)
	}
}

// DeletePeople handles DELETE /api/v1/people
// Deletes a list of persons by personCodes JSON payload
// @Summary Batch Delete People
// @Description Deletes one or more persons from HikCentral Pro using a list of personCodes
// @Tags People
// @Accept json
// @Produce json
// @Param request body schema.DeletePeopleRequest true "List of personCodes to delete"
// @Success 200 {object} common.APIResponse{data=schema.PersonDeleteResponse} "Success"
// @Failure 400 {object} common.APIResponse "Bad Request"
// @Failure 500 {object} common.APIResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/people [delete]
func (h *PersonHandler) DeletePeople(c *gin.Context) {
	var req schema.DeletePeopleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, http.StatusBadRequest, "Invalid request payload. 'personCodes' array is required: "+err.Error())
		return
	}

	resp, err := h.service.DeletePeople(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	common.SuccessResponse(c, "People deleted successfully from HikCentral", resp)
}

// DeletePersonByCode handles DELETE /api/v1/people/:personCode
// Deletes a single person by URL path parameter
// @Summary Delete Person by Code
// @Description Deletes a single person from HikCentral Pro by URL personCode
// @Tags People
// @Produce json
// @Param personCode path string true "Person Code to delete"
// @Success 200 {object} common.APIResponse{data=schema.PersonDeleteResponse} "Success"
// @Failure 400 {object} common.APIResponse "Bad Request"
// @Failure 500 {object} common.APIResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /api/people/{personCode} [delete]
func (h *PersonHandler) DeletePersonByCode(c *gin.Context) {
	personCode := c.Param("personCode")
	if personCode == "" {
		common.ErrorResponse(c, http.StatusBadRequest, "Missing personCode parameter")
		return
	}

	req := schema.DeletePeopleRequest{
		PersonCodes: []string{personCode},
	}

	resp, err := h.service.DeletePeople(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	common.SuccessResponse(c, "Person deleted successfully from HikCentral", resp)
}

// extractFaceImage safely extracts a single face image file from multipart form supporting common aliases
func (h *PersonHandler) extractFaceImage(c *gin.Context) *multipart.FileHeader {
	if c.Request.MultipartForm != nil {
		possibleFieldNames := []string{"faceImage", "faceImages", "face_image", "face_images"}
		for _, field := range possibleFieldNames {
			if files, exists := c.Request.MultipartForm.File[field]; exists && len(files) > 0 {
				return files[0]
			}
		}
	}
	return nil
}

func (h *PersonHandler) handleError(c *gin.Context, err error) {
	var tpErr *common.ThirdPartyError
	if errors.As(err, &tpErr) {
		status := http.StatusBadGateway
		if tpErr.StatusCode == http.StatusBadRequest {
			status = http.StatusBadRequest
		}
		common.ErrorResponse(c, status, "HikCentral Pro error: "+tpErr.Message)
		return
	}

	switch {
	case errors.Is(err, common.ErrBadRequest):
		common.ErrorResponse(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, common.ErrNotFound):
		common.ErrorResponse(c, http.StatusNotFound, err.Error())
	case errors.Is(err, common.ErrAlreadyExists):
		common.ErrorResponse(c, http.StatusConflict, err.Error())
	case errors.Is(err, common.ErrThirdPartyTimeout):
		common.ErrorResponse(c, http.StatusGatewayTimeout, "HikCentral Pro request timed out")
	case errors.Is(err, common.ErrThirdPartyService), errors.Is(err, common.ErrThirdPartyBadResp):
		common.ErrorResponse(c, http.StatusBadGateway, "HikCentral Pro service error: "+err.Error())
	default:
		common.ErrorResponse(c, http.StatusInternalServerError, "Internal server error: "+err.Error())
	}
}
