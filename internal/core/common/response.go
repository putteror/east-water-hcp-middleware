package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse defines the standard JSON response envelope
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Page    interface{} `json:"page,omitempty"`
}

// SuccessResponse sends a standard 200 OK or 201 Created response
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// CreatedResponse sends a standard 201 Created response
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// GetDataListResponse sends a standard response containing paginated list data
func GetDataListResponse(c *gin.Context, message string, data interface{}, pageData interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Page:    pageData,
	})
}

// ErrorResponse sends a standard error JSON response with the given HTTP status
func ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
	})
}
