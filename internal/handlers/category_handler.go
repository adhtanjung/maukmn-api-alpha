package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"maukemana-backend/internal/repositories"
	"maukemana-backend/internal/utils"
)

// CategoryRepository defines the interface for category data access
type CategoryRepository interface {
	GetAll(ctx context.Context) ([]repositories.Category, error)
}

// CategoryHandler handles category-related HTTP requests
type CategoryHandler struct {
	repo CategoryRepository
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(repo CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

// GetCategories handles GET /api/v1/categories
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendSuccess(c, "Categories retrieved", gin.H{"data": categories})
}

// VocabularyRepository defines the interface for vocabulary data access
type VocabularyRepository interface {
	GetActive(ctx context.Context, vocabType string) ([]repositories.Vocabulary, error)
}

// VocabularyHandler handles vocabulary-related HTTP requests
type VocabularyHandler struct {
	repo VocabularyRepository
}

// NewVocabularyHandler creates a new vocabulary handler
func NewVocabularyHandler(repo VocabularyRepository) *VocabularyHandler {
	return &VocabularyHandler{repo: repo}
}

// GetVocabularies handles GET /api/v1/vocabularies
func (h *VocabularyHandler) GetVocabularies(c *gin.Context) {
	vocabType := c.Query("type")

	vocabularies, err := h.repo.GetActive(c.Request.Context(), vocabType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vocabularies})
}
