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
	Environment string
	Server      ServerConfig
	Redis       RedisConfig
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

type RedisConfig struct {
	Address  string
	Password string
	DB       int
	PoolSize int
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
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:         getEnv("PORT", "8085"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		Redis: RedisConfig{
			Address:  getEnv("REDIS_ADDRESS", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
			PoolSize: getIntEnv("REDIS_POOL_SIZE", 10),
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
