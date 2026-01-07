package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationQuery represents standard pagination query parameters
type PaginationQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

// GetPagination extracts page and limit from the query string with defaults
// Default: Page 1, Limit 10
func GetPagination(c *gin.Context) (page, limit int) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Max limit cap (optional, safe default)
	if limit > 100 {
		limit = 100
	}

	return page, limit
}

// GetOffset calculates the database offset based on page and limit
func GetOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * limit
}
