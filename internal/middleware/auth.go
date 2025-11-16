package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	
	"github.com/aioutlet/cart-service/pkg/secrets"
)

var (
	jwtSecretCache string
	jwtSecretMutex sync.RWMutex
)

// AuthMiddleware validates JWT tokens and extracts user information
// Loads JWT secret from Dapr Secret Store on first use (lazy loading)
func AuthMiddleware(secretManager *secrets.DaprSecretManager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "JWT token is required",
			})
			c.Abort()
			return
		}

		// Get JWT secret with lazy loading from Dapr
		secretKey, err := getJWTSecret(c.Request.Context(), secretManager, logger)
		if err != nil {
			logger.Error("Failed to get JWT secret", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to load JWT configuration",
			})
			c.Abort()
			return
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid JWT token",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// Check if token is valid and extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Extract user ID from claims - try 'sub' first, then 'id'
			var userID string
			if sub, exists := claims["sub"].(string); exists {
				userID = sub
			} else if id, exists := claims["id"].(string); exists {
				userID = id
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "User ID not found in token",
				})
				c.Abort()
				return
			}

			// Set user information in context
			c.Set("userID", userID)
			
			// Extract additional claims if available
			if email, exists := claims["email"].(string); exists {
				c.Set("userEmail", email)
			}
			
			if role, exists := claims["role"].(string); exists {
				c.Set("userRole", role)
			}
			
			if username, exists := claims["username"].(string); exists {
				c.Set("username", username)
			}

			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid JWT token claims",
			})
			c.Abort()
			return
		}
	}
}

// OptionalAuthMiddleware validates JWT tokens if present but doesn't require them
// Loads JWT secret from Dapr Secret Store on first use (lazy loading)
func OptionalAuthMiddleware(secretManager *secrets.DaprSecretManager, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.Next()
			return
		}

		// Get JWT secret with lazy loading from Dapr
		secretKey, err := getJWTSecret(c.Request.Context(), secretManager, logger)
		if err != nil {
			logger.Warn("Failed to get JWT secret for optional auth", zap.Error(err))
			c.Next()
			return
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			c.Next()
			return
		}

		// Check if token is valid and extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Extract user ID from claims - try 'sub' first, then 'id'
			if userID, exists := claims["sub"].(string); exists {
				c.Set("userID", userID)
			} else if userID, exists := claims["id"].(string); exists {
				c.Set("userID", userID)
			}
			
			// Extract additional claims if available
			if email, exists := claims["email"].(string); exists {
				c.Set("userEmail", email)
			}
			
			if role, exists := claims["role"].(string); exists {
				c.Set("userRole", role)
			}
			
			if username, exists := claims["username"].(string); exists {
				c.Set("username", username)
			}
		}

		c.Next()
	}
}

// getJWTSecret retrieves JWT secret with caching
func getJWTSecret(ctx context.Context, secretManager *secrets.DaprSecretManager, logger *zap.Logger) (string, error) {
	// Check cache first (read lock)
	jwtSecretMutex.RLock()
	if jwtSecretCache != "" {
		defer jwtSecretMutex.RUnlock()
		return jwtSecretCache, nil
	}
	jwtSecretMutex.RUnlock()

	// Load from Dapr (write lock)
	jwtSecretMutex.Lock()
	defer jwtSecretMutex.Unlock()

	// Double-check after acquiring write lock
	if jwtSecretCache != "" {
		return jwtSecretCache, nil
	}

	// Load from Dapr Secret Store
	secret, err := secretManager.GetJWTSecret(ctx)
	if err != nil {
		return "", err
	}

	jwtSecretCache = secret
	return secret, nil
}
