package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"maukemana-backend/internal/database"
	"maukemana-backend/internal/utils"
)

// CategoryHandler handles category-related HTTP requests
type CategoryHandler struct {
	db *database.DB
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(db *database.DB) *CategoryHandler {
	return &CategoryHandler{db: db}
}

// GetCategories handles GET /api/v1/categories
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()

	query := `
		SELECT category_id, parent_category_id, name_key, icon, created_at
		FROM categories
		ORDER BY name_key
	`

	type Category struct {
		CategoryID       string  `db:"category_id" json:"category_id"`
		ParentCategoryID *string `db:"parent_category_id" json:"parent_category_id,omitempty"`
		NameKey          string  `db:"name_key" json:"name_key"`
		Icon             *string `db:"icon" json:"icon,omitempty"`
		CreatedAt        string  `db:"created_at" json:"created_at"`
	}

	var categories []Category
	err := h.db.SelectContext(ctx, &categories, query)
	if err != nil {
		utils.SendInternalError(c, err)
		return
	}

	utils.SendSuccess(c, "Categories retrieved", gin.H{"data": categories})
}

// VocabularyHandler handles vocabulary-related HTTP requests
type VocabularyHandler struct {
	db *database.DB
}

// NewVocabularyHandler creates a new vocabulary handler
func NewVocabularyHandler(db *database.DB) *VocabularyHandler {
	return &VocabularyHandler{db: db}
}

// GetVocabularies handles GET /api/v1/vocabularies
func (h *VocabularyHandler) GetVocabularies(c *gin.Context) {
	ctx := c.Request.Context()
	vocabType := c.Query("type")

	query := `
		SELECT vocab_id, vocab_type, key, aliases, icon, is_active
		FROM vocabularies
		WHERE is_active = true
	`
	args := []interface{}{}

	if vocabType != "" {
		query += " AND vocab_type = $1"
		args = append(args, vocabType)
	}
	query += " ORDER BY vocab_type, key"

	rows, err := h.db.QueryxContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var vocabularies []map[string]interface{}
	for rows.Next() {
		vocab := make(map[string]interface{})
		if err := rows.MapScan(vocab); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		vocabularies = append(vocabularies, vocab)
	}

	c.JSON(http.StatusOK, gin.H{"data": vocabularies})
}
