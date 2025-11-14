package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Name        string
	Version     string
	Environment string
	Server      ServerConfig
	Dapr        DaprConfig
	JWT         JWTConfig
	CORS        CORSConfig
	Cart        CartConfig
	Services    ServicesConfig
	Tracing     TracingConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DaprConfig struct {
	HTTPPort       string
	GRPCPort       string
	StateStoreName string
	AppID          string
	AppPort        string
}

type JWTConfig struct {
	SecretKey string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type CartConfig struct {
	DefaultTTL     time.Duration
	GuestTTL       time.Duration
	MaxItems       int
	MaxItemQty     int
	CleanupInterval time.Duration
}

type ServicesConfig struct {
	// NOTE: These URLs are no longer used by the application.
	// Service invocation is now handled via Dapr SDK using app-id names:
	//   - "product-service" instead of ProductServiceURL
	//   - "inventory-service" instead of InventoryServiceURL
	// These fields are kept for backward compatibility only.
	ProductServiceURL   string
	InventoryServiceURL string
	OrderServiceURL     string
	UserServiceURL      string
}

type TracingConfig struct {
	Enabled        bool
	ServiceName    string
	ServiceVersion string
	JaegerEndpoint string
	SampleRate     float64
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Name:        getEnv("NAME", "cart-service"),
		Version:     getEnv("VERSION", "1.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:         getEnv("PORT", "1008"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		Dapr: DaprConfig{
			HTTPPort:       getEnv("DAPR_HTTP_PORT", "3508"),
			GRPCPort:       getEnv("DAPR_GRPC_PORT", "50008"),
			StateStoreName: getEnv("DAPR_STATE_STORE", "statestore"),
			AppID:          getEnv("DAPR_APP_ID", "cart-service"),
			AppPort:        getEnv("DAPR_APP_PORT", "1008"),
		},
		JWT: JWTConfig{
			SecretKey: getEnv("JWT_SECRET", "your-256-bit-secret"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getSliceEnv("CORS_ALLOWED_ORIGINS", []string{"*"}),
		},
		Cart: CartConfig{
			DefaultTTL:      getDurationEnv("CART_DEFAULT_TTL", 30*24*time.Hour), // 30 days
			GuestTTL:        getDurationEnv("CART_GUEST_TTL", 3*24*time.Hour),    // 3 days
			MaxItems:        getIntEnv("CART_MAX_ITEMS", 100),
			MaxItemQty:      getIntEnv("CART_MAX_ITEM_QTY", 10),
			CleanupInterval: getDurationEnv("CART_CLEANUP_INTERVAL", 1*time.Hour),
		},
		Services: ServicesConfig{
			ProductServiceURL:   getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
			InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://localhost:8082"),
			OrderServiceURL:     getEnv("ORDER_SERVICE_URL", "http://localhost:8083"),
			UserServiceURL:      getEnv("USER_SERVICE_URL", "http://localhost:8084"),
		},
		Tracing: TracingConfig{
			Enabled:        getBoolEnv("TRACING_ENABLED", true),
			ServiceName:    getEnv("TRACING_SERVICE_NAME", "cart-service"),
			ServiceVersion: getEnv("TRACING_SERVICE_VERSION", "1.0.0"),
			JaegerEndpoint: getEnv("TRACING_JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			SampleRate:     getFloatEnv("TRACING_SAMPLE_RATE", 1.0),
		},
	}
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
