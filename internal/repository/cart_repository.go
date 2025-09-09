package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aioutlet/cart-service/internal/models"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	cartKeyPrefix = "cart:"
	cartLockPrefix = "cart_lock:"
)

// CartRepository interface defines cart repository operations
type CartRepository interface {
	GetCart(ctx context.Context, userID string) (*models.Cart, error)
	SaveCart(ctx context.Context, cart *models.Cart) error
	DeleteCart(ctx context.Context, userID string) error
	SetCartTTL(ctx context.Context, userID string, ttl time.Duration) error
	CartExists(ctx context.Context, userID string) (bool, error)
	AcquireLock(ctx context.Context, userID string, ttl time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, userID string) error
	GetAllCartKeys(ctx context.Context) ([]string, error)
	GetCartTTL(ctx context.Context, userID string) (time.Duration, error)
}

// cartRepository implements CartRepository interface
type cartRepository struct {
	client *redis.Client
	logger *zap.Logger
}

// NewCartRepository creates a new cart repository
func NewCartRepository(client *redis.Client, logger *zap.Logger) CartRepository {
	return &cartRepository{
		client: client,
		logger: logger,
	}
}

// GetCart retrieves a cart from Redis
func (r *cartRepository) GetCart(ctx context.Context, userID string) (*models.Cart, error) {
	key := r.getCartKey(userID)
	
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, models.ErrCartNotFound
		}
		r.logger.Error("Failed to get cart from Redis", 
			zap.String("userID", userID), 
			zap.Error(err))
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	var cart models.Cart
	if err := json.Unmarshal([]byte(data), &cart); err != nil {
		r.logger.Error("Failed to unmarshal cart data", 
			zap.String("userID", userID), 
			zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal cart: %w", err)
	}

	// Check if cart has expired
	if cart.IsExpired() {
		r.logger.Info("Cart has expired, deleting", zap.String("userID", userID))
		if err := r.DeleteCart(ctx, userID); err != nil {
			r.logger.Error("Failed to delete expired cart", 
				zap.String("userID", userID), 
				zap.Error(err))
		}
		return nil, models.ErrCartExpired
	}

	return &cart, nil
}

// SaveCart saves a cart to Redis
func (r *cartRepository) SaveCart(ctx context.Context, cart *models.Cart) error {
	key := r.getCartKey(cart.UserID)
	
	data, err := json.Marshal(cart)
	if err != nil {
		r.logger.Error("Failed to marshal cart data", 
			zap.String("userID", cart.UserID), 
			zap.Error(err))
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	// Calculate TTL based on cart expiry
	ttl := time.Until(cart.ExpiresAt)
	if ttl <= 0 {
		ttl = time.Minute // Minimum TTL of 1 minute
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		r.logger.Error("Failed to save cart to Redis", 
			zap.String("userID", cart.UserID), 
			zap.Error(err))
		return fmt.Errorf("failed to save cart: %w", err)
	}

	r.logger.Debug("Cart saved successfully", 
		zap.String("userID", cart.UserID),
		zap.Duration("ttl", ttl))
	
	return nil
}

// DeleteCart deletes a cart from Redis
func (r *cartRepository) DeleteCart(ctx context.Context, userID string) error {
	key := r.getCartKey(userID)
	
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("Failed to delete cart from Redis", 
			zap.String("userID", userID), 
			zap.Error(err))
		return fmt.Errorf("failed to delete cart: %w", err)
	}

	r.logger.Debug("Cart deleted successfully", zap.String("userID", userID))
	return nil
}

// SetCartTTL sets the TTL for a cart
func (r *cartRepository) SetCartTTL(ctx context.Context, userID string, ttl time.Duration) error {
	key := r.getCartKey(userID)
	
	err := r.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		r.logger.Error("Failed to set cart TTL", 
			zap.String("userID", userID),
			zap.Duration("ttl", ttl),
			zap.Error(err))
		return fmt.Errorf("failed to set cart TTL: %w", err)
	}

	return nil
}

// CartExists checks if a cart exists in Redis
func (r *cartRepository) CartExists(ctx context.Context, userID string) (bool, error) {
	key := r.getCartKey(userID)
	
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.logger.Error("Failed to check cart existence", 
			zap.String("userID", userID), 
			zap.Error(err))
		return false, fmt.Errorf("failed to check cart existence: %w", err)
	}

	return exists > 0, nil
}

// AcquireLock acquires a distributed lock for cart operations
func (r *cartRepository) AcquireLock(ctx context.Context, userID string, ttl time.Duration) (bool, error) {
	lockKey := r.getLockKey(userID)
	
	// Use SET with NX (only if not exists) and EX (expiry) options
	result, err := r.client.SetNX(ctx, lockKey, "locked", ttl).Result()
	if err != nil {
		r.logger.Error("Failed to acquire cart lock", 
			zap.String("userID", userID), 
			zap.Error(err))
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if result {
		r.logger.Debug("Cart lock acquired", zap.String("userID", userID))
	}

	return result, nil
}

// ReleaseLock releases a distributed lock
func (r *cartRepository) ReleaseLock(ctx context.Context, userID string) error {
	lockKey := r.getLockKey(userID)
	
	err := r.client.Del(ctx, lockKey).Err()
	if err != nil {
		r.logger.Error("Failed to release cart lock", 
			zap.String("userID", userID), 
			zap.Error(err))
		return fmt.Errorf("failed to release lock: %w", err)
	}

	r.logger.Debug("Cart lock released", zap.String("userID", userID))
	return nil
}

// GetAllCartKeys retrieves all cart keys for cleanup operations
func (r *cartRepository) GetAllCartKeys(ctx context.Context) ([]string, error) {
	pattern := cartKeyPrefix + "*"
	
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.logger.Error("Failed to get cart keys", zap.Error(err))
		return nil, fmt.Errorf("failed to get cart keys: %w", err)
	}

	return keys, nil
}

// GetCartTTL gets the remaining TTL for a cart
func (r *cartRepository) GetCartTTL(ctx context.Context, userID string) (time.Duration, error) {
	key := r.getCartKey(userID)
	
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		r.logger.Error("Failed to get cart TTL", 
			zap.String("userID", userID), 
			zap.Error(err))
		return 0, fmt.Errorf("failed to get cart TTL: %w", err)
	}

	return ttl, nil
}

// Helper methods
func (r *cartRepository) getCartKey(userID string) string {
	return cartKeyPrefix + userID
}

func (r *cartRepository) getLockKey(userID string) string {
	return cartLockPrefix + userID
}
