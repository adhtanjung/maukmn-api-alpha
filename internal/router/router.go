package router

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"maukemana-backend/internal/auth"
	"maukemana-backend/internal/config"
	"maukemana-backend/internal/database"
	"maukemana-backend/internal/handlers"
	"maukemana-backend/internal/middleware"
	"maukemana-backend/internal/repositories"
	"maukemana-backend/internal/services"
	"maukemana-backend/internal/storage"
)

// Setup creates and configures the Gin router
func Setup(db *database.DB) *gin.Engine {
	// Initialize repositories
	poiRepo := repositories.NewPOIRepository(db)

	// Initialize services
	// Services
	geocodingService := services.NewMockGeocodingService()

	// Initialize handlers
	poiHandler := handlers.NewPOIHandler(poiRepo, geocodingService)
	savedPOIRepo := repositories.NewSavedPOIRepository(db)
	savedPOIHandler := handlers.NewSavedPOIHandler(savedPOIRepo)

	commentRepo := repositories.NewCommentRepository(db)
	commentHandler := handlers.NewCommentHandler(commentRepo)
	categoryHandler := handlers.NewCategoryHandler(db)
	vocabHandler := handlers.NewVocabularyHandler(db)
	authHandler := handlers.NewAuthHandler(db)

	// Initialize R2 storage (optional - continues without if not configured)
	var uploadHandler *handlers.UploadHandler
	r2Client, err := storage.NewR2Client()
	if err != nil {
		log.Printf("Warning: R2 storage not configured: %v", err)
	} else {
		uploadHandler = handlers.NewUploadHandler(r2Client, db)
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
			pois.GET("/:id/comments", commentHandler.GetCommentsByPOI) // Public read for comments

			// Protected POI routes (require auth)
			poisAuth := pois.Group("")
			poisAuth.Use(handlers.AuthMiddleware(db))
			{
				poisAuth.POST("", poiHandler.CreatePOI)
				poisAuth.GET("/my", poiHandler.GetMyPOIs)
				poisAuth.PUT("/:id", poiHandler.UpdatePOI)
				poisAuth.POST("/:id/save", savedPOIHandler.ToggleSave)
				poisAuth.GET("/saved", savedPOIHandler.GetMySavedPOIs)

				// Comments
				poisAuth.POST("/:id/comments", commentHandler.CreateComment)
				poisAuth.DELETE("/:id", poiHandler.DeletePOI)
				poisAuth.GET("/my-drafts", poiHandler.GetMyDrafts)
				poisAuth.POST("/:id/submit", poiHandler.SubmitPOI)
				poisAuth.POST("/:id/approve", poiHandler.ApprovePOI)
				poisAuth.POST("/:id/reject", poiHandler.RejectPOI)
				poisAuth.GET("/pending", poiHandler.GetPendingPOIs)
				poisAuth.GET("/admin-list", poiHandler.GetAdminPOIs)

				// Debug/Admin routes (if needed)
				// r.GET("/api/v1/pois/:id/saved-users", savedPOIHandler.GetUsersWhoSavedPOI)

				// Section-based editing
				sectionHandler := handlers.NewPOISectionHandler(poiRepo)
				poisAuth.GET("/:id/section/profile", sectionHandler.GetPOIProfile)
				poisAuth.PUT("/:id/section/profile", sectionHandler.UpdatePOIProfile)
				poisAuth.GET("/:id/section/location", sectionHandler.GetPOILocation)
				poisAuth.PUT("/:id/section/location", sectionHandler.UpdatePOILocation)
				poisAuth.GET("/:id/section/operations", sectionHandler.GetPOIOperations)
				poisAuth.PUT("/:id/section/operations", sectionHandler.UpdatePOIOperations)
				poisAuth.GET("/:id/section/social", sectionHandler.GetPOISocial)
				poisAuth.PUT("/:id/section/social", sectionHandler.UpdatePOISocial)
				poisAuth.GET("/:id/section/work-prod", sectionHandler.GetPOIWorkProd)
				poisAuth.PUT("/:id/section/work-prod", sectionHandler.UpdatePOIWorkProd)
				poisAuth.GET("/:id/section/atmosphere", sectionHandler.GetPOIAtmosphere)
				poisAuth.PUT("/:id/section/atmosphere", sectionHandler.UpdatePOIAtmosphere)
				poisAuth.GET("/:id/section/food-drink", sectionHandler.GetPOIFoodDrink)
				poisAuth.PUT("/:id/section/food-drink", sectionHandler.UpdatePOIFoodDrink)
				poisAuth.GET("/:id/section/contact", sectionHandler.GetPOIContact)
				poisAuth.PUT("/:id/section/contact", sectionHandler.UpdatePOIContact)
			}
		}

		// Upload routes (require auth)
		if uploadHandler != nil {
			uploads := v1.Group("/uploads")
			uploads.Use(handlers.AuthMiddleware(db))
			{
				uploads.POST("/presign", uploadHandler.GetPresignedURL)
				uploads.POST("/finalize", uploadHandler.FinalizeUpload)
				uploads.DELETE("", uploadHandler.DeleteUpload)
			}

			// Asset routes (require auth)
			assets := v1.Group("/assets")
			assets.Use(handlers.AuthMiddleware(db))
			{
				assets.GET("/:id", uploadHandler.GetAssetStatus)
			}
		}

		// Category routes
		v1.GET("/categories", categoryHandler.GetCategories)

		// Saved POI list route
		v1.GET("/me/saved-pois", handlers.AuthMiddleware(db), savedPOIHandler.GetMySavedPOIs)

		// Vocabulary routes
		v1.GET("/vocabularies", vocabHandler.GetVocabularies)
	}

	// Public image serving route
	router.GET("/img/:hash/:rendition", uploadHandler.ServeImage)

	// API documentation endpoint
	router.GET("/api", apiDocumentation())

	return router
}

func setupBaseRouter() *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(otelgin.Middleware("maukemana-api"))
	router.Use(middleware.Observability())
	router.Use(middleware.SecurityHeaders()) // Add security headers
	router.Use(middleware.RateLimit())

	// Trusted Proxies Configuration
	// In production, you should set this to the specific IP ranges of your load balancers or reverse proxies.
	// For now, setting it to nil means we don't trust any proxy headers (X-Forwarded-For, etc.)
	// This prevents IP spoofing if not behind a configured proxy.
	router.SetTrustedProxies(nil)

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = config.GetAllowedOrigins()
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Authorization",
		"Accept",
		"User-Agent",
		"Cache-Control",
		"Pragma",
		"X-Session-ID",
	}
	corsConfig.AllowMethods = []string{
		"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
	}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

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
