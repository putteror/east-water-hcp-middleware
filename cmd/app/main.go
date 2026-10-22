package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eastwater-backend/internal/adapter/thirdparty"
	"eastwater-backend/internal/adapter/hikcentral"
	"eastwater-backend/internal/config"
	"eastwater-backend/internal/core/port"
	"eastwater-backend/internal/core/service"
	httpRouter "eastwater-backend/internal/interface/http"
	"eastwater-backend/internal/interface/http/handler"
)

// @title East Water - VMS to HikCentral Integration Gateway API
// @version 1.0
// @description Gateway proxy service connecting VMS management software to HikCentral Pro OpenAPI without local database.
// @termsOfService http://swagger.io/terms/

// @contact.name CENT ACCESS Integration Team
// @contact.url https://centaccess.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name x-api-key

func main() {
	log.Println("=== Starting East Water API Integration Service (VMS to HikCentral Pro) ===")

	// 1. Centralized Configuration
	cfg := config.LoadConfig()

	// 2. Outbound Adapters (Ports & Adapters pattern)
	// Database connection is intentionally omitted as this service acts as an API proxy/integration gateway.
	var thirdPartyAdapter port.ThirdPartyPort
	var hikCentralAdapter port.HikCentralPort

	if cfg.ThirdParty.UseMock {
		log.Println("[Init] ⚠️ Running with in-memory Mock 3rd-Party & HikCentral Pro Adapters")
		thirdPartyAdapter = thirdparty.NewMockAdapter()
		hikCentralAdapter = hikcentral.NewMockAdapter()
	} else {
		log.Printf("[Init] 🌐 Connected to 3rd-Party API at: %s", cfg.ThirdParty.BaseURL)
		thirdPartyAdapter = thirdparty.NewAdapter(cfg.ThirdParty)
		hikCentralAdapter = hikcentral.NewAdapter(cfg.ThirdParty)
	}

	// 3. Core Business Services (Dependencies injected via interfaces)
	integrationService := service.NewIntegrationService(thirdPartyAdapter)
	personService := service.NewPersonService(hikCentralAdapter)

	// 4. Primary Delivery / Interface Handlers
	integrationHandler := handler.NewIntegrationHandler(integrationService)
	personHandler := handler.NewPersonHandler(personService)
	healthHandler := handler.NewHealthHandler(integrationService)

	// 5. Router & HTTP Server Setup
	router := httpRouter.NewRouter(httpRouter.RouterDeps{
		Config:             cfg,
		IntegrationHandler: integrationHandler,
		PersonHandler:      personHandler,
		HealthHandler:      healthHandler,
	})

	serverAddr := fmt.Sprintf(":%s", cfg.App.Port)
	srv := &http.Server{
		Addr:              serverAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 6. Asynchronous HTTP Server Start
	go func() {
		log.Printf("[Server] 🚀 Server is listening on http://localhost%s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[Server] Fatal error starting HTTP server: %v", err)
		}
	}()

	// 7. Graceful Shutdown on SIGINT (Ctrl+C) or SIGTERM (Docker/K8s)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Server] 🛑 Shutdown signal received. Closing server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[Server] Server forced to shutdown: %v", err)
	}

	log.Println("[Server] Server exited cleanly.")
}
