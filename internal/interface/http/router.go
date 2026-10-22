package http

import (
	_ "eastwater-backend/docs"
	"eastwater-backend/internal/config"
	"eastwater-backend/internal/interface/http/handler"
	"eastwater-backend/internal/interface/http/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RouterDeps holds all handlers required to initialize routes
type RouterDeps struct {
	Config             *config.Config
	IntegrationHandler *handler.IntegrationHandler
	PersonHandler      *handler.PersonHandler
	HealthHandler      *handler.HealthHandler
}

// NewRouter sets up Gin engine, global middlewares, and registers all API routes
func NewRouter(deps RouterDeps) *gin.Engine {
	if deps.Config.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// Global middlewares
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORSMiddleware())

	// Public Health check routes & Swagger UI
	r.GET("/health", deps.HealthHandler.HealthCheck)
	r.GET("/ping", deps.HealthHandler.Ping)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Helper function to register all people endpoints
	registerPeopleRoutes := func(g *gin.RouterGroup) {
		g.GET("", deps.PersonHandler.GetPeople)                              // List people
		g.GET("/:personCode", deps.PersonHandler.GetPersonByCode)            // Get person by code
		g.POST("", deps.PersonHandler.CreatePerson)                           // Create person (strict create, or ?upsert=true)
		g.PUT("/:personCode", deps.PersonHandler.UpdatePerson)               // Update person
		g.POST("/upsert", deps.PersonHandler.UpsertPerson)                   // Upsert person (POST)
		g.PUT("/upsert", deps.PersonHandler.UpsertPerson)                    // Upsert person (PUT)
		g.DELETE("", deps.PersonHandler.DeletePeople)                        // Batch delete people by personCodes JSON
		g.DELETE("/:personCode", deps.PersonHandler.DeletePersonByCode)      // Delete single person by URL param
	}

	// 1. Primary API route group without v1: /api/people
	api := r.Group("/api")
	api.Use(middleware.APIKeyAuthMiddleware(deps.Config.Auth.APIKey))
	{
		registerPeopleRoutes(api.Group("/people"))

		telemetry := api.Group("/telemetry")
		{
			telemetry.POST("/data", deps.IntegrationHandler.TransmitData)
			telemetry.GET("/data", deps.IntegrationHandler.GetDataList)
			telemetry.GET("/data/:id", deps.IntegrationHandler.GetDataByID)
		}
	}

	// 2. Direct route group without /api prefix: /people (for VMS software calling root directly)
	rootGroup := r.Group("")
	rootGroup.Use(middleware.APIKeyAuthMiddleware(deps.Config.Auth.APIKey))
	{
		registerPeopleRoutes(rootGroup.Group("/people"))
	}

	return r
}
