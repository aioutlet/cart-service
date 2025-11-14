package testutils

import (
	"time"

	"github.com/aioutlet/cart-service/internal/config"
	"github.com/aioutlet/cart-service/internal/models"
)

// CreateTestConfig creates a test configuration
func CreateTestConfig() *config.Config {
	return &config.Config{
		Environment: "test",
		Server: config.ServerConfig{
			Port:         "8085",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Dapr: config.DaprConfig{
			HTTPPort:       "3508",
			GRPCPort:       "50008",
			StateStoreName: "statestore",
			AppID:          "cart-service",
			AppPort:        "1008",
		},
		JWT: config.JWTConfig{
			SecretKey: "test-secret-key",
		},
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"*"},
		},
		Cart: config.CartConfig{
			DefaultTTL:      24 * time.Hour,
			GuestTTL:        3 * time.Hour,
			MaxItems:        10,
			MaxItemQty:      5,
			CleanupInterval: 1 * time.Hour,
		},
	}
}

// CreateTestCart creates a test cart
func CreateTestCart(userID string, ttl time.Duration) *models.Cart {
	return models.NewCart(userID, ttl)
}

// CreateTestCartItem creates a test cart item
func CreateTestCartItem(productID, productName string, price float64, quantity int) models.CartItem {
	return models.CartItem{
		ProductID:   productID,
		ProductName: productName,
		SKU:         "SKU-" + productID,
		Price:       price,
		Quantity:    quantity,
		ImageURL:    "https://example.com/image.jpg",
		Category:    "Electronics",
		Subtotal:    price * float64(quantity),
		AddedAt:     time.Now().UTC(),
	}
}

// CreateTestProductInfo creates a test product info
func CreateTestProductInfo(id, name string, price float64, isActive bool) *models.ProductInfo {
	return &models.ProductInfo{
		ID:       id,
		Name:     name,
		SKU:      "SKU-" + id,
		Price:    price,
		ImageURL: "https://example.com/image.jpg",
		Category: "Electronics",
		IsActive: isActive,
		StockQty: 100,
	}
}
