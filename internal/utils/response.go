package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    *Pagination `json:"meta,omitempty"`
}

// Pagination represents pagination metadata
type Pagination struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
}

// SendSuccess sends a success response with data (200 OK)
func SendSuccess(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendCreated sends a created response with data (201 Created)
func SendCreated(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendPaginated sends a success response with pagination metadata (200 OK)
func SendPaginated(c *gin.Context, message string, data interface{}, page, limit, total int) {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + limit - 1) / limit)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta: &Pagination{
			CurrentPage: page,
			PerPage:     limit,
			Total:       total,
			TotalPages:  totalPages,
		},
	})
}

// SendError sends an error response with a specific status code
func SendError(c *gin.Context, code int, message string, err error) {
	var errDetails interface{}
	if err != nil {
		errDetails = err.Error()
		c.Error(err)
	}

	c.AbortWithStatusJSON(code, Response{
		Success: false,
		Message: message,
		Error:   errDetails,
	})
}

// SendValidationError sends a 400 Bad Request error
func SendValidationError(c *gin.Context, err error) {
	SendError(c, http.StatusBadRequest, "Validation failed", err)
}

// SendInternalError sends a 500 Internal Server Error
func SendInternalError(c *gin.Context, err error) {
	SendError(c, http.StatusInternalServerError, "Internal server error", err)
}
