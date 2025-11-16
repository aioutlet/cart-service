package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aioutlet/cart-service/internal/config"
	"github.com/aioutlet/cart-service/internal/handlers"
	"github.com/aioutlet/cart-service/internal/middleware"
	"github.com/aioutlet/cart-service/internal/repository"
	"github.com/aioutlet/cart-service/internal/services"
	"github.com/aioutlet/cart-service/pkg/logger"
	"github.com/aioutlet/cart-service/pkg/secrets"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Cart Service API
// @version 1.0
// @description A microservice for managing shopping carts
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.aioutlet.com/support
// @contact.email support@aioutlet.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:1008
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter 'Bearer ' followed by your JWT token

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.Environment)
	defer log.Sync()

	// Initialize Dapr client
	daprClient, err := dapr.NewClient()
	if err != nil {
		log.Fatal("Failed to create Dapr client", zap.Error(err))
	}
	defer daprClient.Close()

	log.Info("Successfully connected to Dapr")

	// Initialize Dapr Secret Manager
	secretManager := secrets.NewDaprSecretManager(daprClient, "local-secret-store", log)
	log.Info("Dapr Secret Manager initialized")

	// Initialize repository with Dapr
	cartRepo := repository.NewDaprCartRepository(daprClient, cfg.Dapr.StateStoreName, log)

	// Initialize services with Dapr client for service invocation
	cartService := services.NewCartService(cartRepo, daprClient, cfg, log)

	// Initialize handlers
	cartHandler := handlers.NewCartHandler(cartService, log)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CorrelationID())
	router.Use(middleware.Logger(log))

	// Home/Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     cfg.Name,
			"version":     cfg.Version,
			"environment": cfg.Environment,
			"message":     "Cart Service is running",
			"status":      "operational",
		})
	})

	// Version endpoint
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": cfg.Version,
		})
	})

	// Service info endpoint
	router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     cfg.Name,
			"version":     cfg.Version,
			"environment": cfg.Environment,
			"configuration": gin.H{
				"dapr_http_port": cfg.Dapr.HTTPPort,
				"dapr_grpc_port": cfg.Dapr.GRPCPort,
			},
			"timestamp": time.Now().UTC(),
		})
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   cfg.Name,
			"timestamp": time.Now().UTC(),
			"version":   cfg.Version,
		})
	})

	// Liveness probe
	router.GET("/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	})

	// Readiness probe
	router.GET("/readiness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"service":   cfg.Name,
			"timestamp": time.Now().UTC(),
			"version":   cfg.Version,
		})
	})

	// Metrics endpoint
	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   cfg.Name,
			"timestamp": time.Now().UTC(),
			"version":   cfg.Version,
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Cart routes with authentication middleware (using Dapr secrets)
		cartRoutes := v1.Group("/cart")
		cartRoutes.Use(middleware.AuthMiddleware(secretManager, log))
		{
			cartRoutes.GET("", cartHandler.GetCart)
			cartRoutes.POST("/items", cartHandler.AddItem)
			cartRoutes.PUT("/items/:productId", cartHandler.UpdateItem)
			cartRoutes.DELETE("/items/:productId", cartHandler.RemoveItem)
			cartRoutes.DELETE("", cartHandler.ClearCart)
			cartRoutes.POST("/transfer", cartHandler.TransferCart)
		}

		// Guest cart routes (no authentication required)
		guestRoutes := v1.Group("/guest/cart")
		{
			guestRoutes.GET("/:guestId", cartHandler.GetGuestCart)
			guestRoutes.POST("/:guestId/items", cartHandler.AddGuestItem)
			guestRoutes.PUT("/:guestId/items/:productId", cartHandler.UpdateGuestItem)
			guestRoutes.DELETE("/:guestId/items/:productId", cartHandler.RemoveGuestItem)
			guestRoutes.DELETE("/:guestId", cartHandler.ClearGuestCart)
		}
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Info("Starting Cart Service", 
			zap.String("name", cfg.Name),
			zap.String("version", cfg.Version),
			zap.String("port", cfg.Server.Port),
			zap.String("environment", cfg.Environment))
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Cart Service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Cart Service stopped")
}
