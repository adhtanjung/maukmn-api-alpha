package router

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"maukemana-backend/internal/auth"
	"maukemana-backend/internal/database"
	"maukemana-backend/internal/handlers"
	"maukemana-backend/internal/repositories"
	"maukemana-backend/internal/storage"
)

// Setup creates and configures the Gin router
func Setup(db *database.DB) *gin.Engine {
	// Initialize repositories
	poiRepo := repositories.NewPOIRepository(db)

	// Initialize handlers
	poiHandler := handlers.NewPOIHandler(poiRepo)
	categoryHandler := handlers.NewCategoryHandler(db)
	vocabHandler := handlers.NewVocabularyHandler(db)
	authHandler := handlers.NewAuthHandler(db)

	// Initialize R2 storage (optional - continues without if not configured)
	var uploadHandler *handlers.UploadHandler
	r2Client, err := storage.NewR2Client()
	if err != nil {
		log.Printf("Warning: R2 storage not configured: %v", err)
	} else {
		uploadHandler = handlers.NewUploadHandler(r2Client)
	}

	// Initialize Clerk
	auth.InitClerk()

	// Setup router
	router := setupBaseRouter()

	// Health check endpoint
	router.GET("/health", healthCheck(db))

	// Auth routes
	router.GET("/api/me", handlers.AuthMiddleware(db), authHandler.GetMe)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// POI routes
		pois := v1.Group("/pois")
		{
			pois.GET("", poiHandler.SearchPOIs)
			pois.GET("/nearby", poiHandler.GetNearbyPOIs)
			pois.GET("/filter-options", poiHandler.GetFilterOptions)
			pois.GET("/:id", poiHandler.GetPOI)

			// Protected POI routes (require auth)
			poisAuth := pois.Group("")
			poisAuth.Use(handlers.AuthMiddleware(db))
			{
				poisAuth.POST("", poiHandler.CreatePOI)
				poisAuth.PUT("/:id", poiHandler.UpdatePOI)
				poisAuth.DELETE("/:id", poiHandler.DeletePOI)
				poisAuth.GET("/my-drafts", poiHandler.GetMyDrafts)
				poisAuth.POST("/:id/submit", poiHandler.SubmitPOI)
				poisAuth.POST("/:id/approve", poiHandler.ApprovePOI)
				poisAuth.POST("/:id/reject", poiHandler.RejectPOI)
				poisAuth.GET("/pending", poiHandler.GetPendingPOIs)
			}
		}

		// Upload routes (require auth)
		if uploadHandler != nil {
			uploads := v1.Group("/uploads")
			uploads.Use(handlers.AuthMiddleware(db))
			{
				uploads.POST("/presign", uploadHandler.GetPresignedURL)
				uploads.DELETE("", uploadHandler.DeleteUpload)
			}
		}

		// Category routes
		v1.GET("/categories", categoryHandler.GetCategories)

		// Vocabulary routes
		v1.GET("/vocabularies", vocabHandler.GetVocabularies)
	}

	// API documentation endpoint
	router.GET("/api", apiDocumentation())

	return router
}

func setupBaseRouter() *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:8080",
		"http://localhost:3000",
		"http://localhost:5173",
	}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Authorization",
		"Accept",
		"User-Agent",
		"Cache-Control",
		"Pragma",
		"X-Session-ID",
	}
	config.AllowMethods = []string{
		"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
	}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	return router
}

func healthCheck(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.Health(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "unhealthy",
				"error":     err.Error(),
				"database":  "postgresql",
				"timestamp": time.Now().Unix(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"version":   "2.0",
			"database":  "postgresql",
			"timestamp": time.Now().Unix(),
		})
	}
}

func apiDocumentation() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "Maukemana API",
			"version":     "2.0",
			"description": "Travel discovery and planning API (PostgreSQL + PostGIS)",
			"endpoints": map[string]interface{}{
				"health": "GET /health",
				"pois": map[string]string{
					"list":           "GET /api/v1/pois",
					"get":            "GET /api/v1/pois/:id",
					"create":         "POST /api/v1/pois",
					"update":         "PUT /api/v1/pois/:id",
					"delete":         "DELETE /api/v1/pois/:id",
					"nearby":         "GET /api/v1/pois/nearby?lat=...&lng=...&radius=...",
					"filter_options": "GET /api/v1/pois/filter-options",
				},
				"categories":   "GET /api/v1/categories",
				"vocabularies": "GET /api/v1/vocabularies?type=...",
			},
		})
	}
}
