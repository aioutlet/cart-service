package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT tokens and extracts user information
func AuthMiddleware(secretKey string) gin.HandlerFunc {
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
			// Extract user ID from claims
			userID, exists := claims["sub"].(string)
			if !exists {
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
func OptionalAuthMiddleware(secretKey string) gin.HandlerFunc {
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
			// Extract user ID from claims
			if userID, exists := claims["sub"].(string); exists {
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
