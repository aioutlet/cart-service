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
	"github.com/aioutlet/cart-service/pkg/redis"
	"github.com/aioutlet/cart-service/pkg/tracing"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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

// @host localhost:8085
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

	// Initialize distributed tracing
	tracingCfg := tracing.TracingConfig{
		ServiceName:     cfg.Tracing.ServiceName,
		ServiceVersion:  cfg.Tracing.ServiceVersion,
		Environment:     cfg.Environment,
		JaegerEndpoint:  cfg.Tracing.JaegerEndpoint,
		Enabled:         cfg.Tracing.Enabled,
		SampleRate:      cfg.Tracing.SampleRate,
	}
	
	tp, err := tracing.InitTracing(tracingCfg, log)
	if err != nil {
		log.Fatal("Failed to initialize tracing", zap.Error(err))
	}
	defer tracing.Shutdown(context.Background(), tp, log)

	// Initialize Redis client
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis ping failed", zap.Error(err))
	}
	log.Info("Successfully connected to Redis")

	// Initialize repository
	cartRepo := repository.NewCartRepository(redisClient, log)

	// Initialize services
	cartService := services.NewCartService(cartRepo, cfg, log)

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
	router.Use(otelgin.Middleware("cart-service"))
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Correlation-ID", "traceparent", "tracestate"},
		ExposeHeaders:    []string{"X-Correlation-ID", "traceparent", "tracestate"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(middleware.CorrelationID())
	router.Use(middleware.Logger(log))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "cart-service",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Cart routes with authentication middleware
		cartRoutes := v1.Group("/cart")
		cartRoutes.Use(middleware.AuthMiddleware(cfg.JWT.SecretKey))
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

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Cart Service stopped")
}
