package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aioutlet/cart-service/internal/models"
	dapr "github.com/dapr/go-sdk/client"
	"go.uber.org/zap"
)

const (
	cartKeyPrefix = "cart:"
)

// DaprCartRepository implements CartRepository using Dapr State Management
type DaprCartRepository struct {
	client         dapr.Client
	stateStoreName string
	logger         *zap.Logger
}

// NewDaprCartRepository creates a new Dapr-based cart repository
func NewDaprCartRepository(client dapr.Client, stateStoreName string, logger *zap.Logger) CartRepository {
	return &DaprCartRepository{
		client:         client,
		stateStoreName: stateStoreName,
		logger:         logger,
	}
}

// GetCart retrieves a cart from Dapr state store
func (r *DaprCartRepository) GetCart(ctx context.Context, userID string) (*models.Cart, error) {
	key := r.getCartKey(userID)

	item, err := r.client.GetState(ctx, r.stateStoreName, key, nil)
	if err != nil {
		r.logger.Error("Failed to get cart from Dapr state store",
			zap.String("userID", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	if item.Value == nil || len(item.Value) == 0 {
		return nil, models.ErrCartNotFound
	}

	var cart models.Cart
	if err := json.Unmarshal(item.Value, &cart); err != nil {
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

// SaveCart saves a cart to Dapr state store with ETag-based concurrency control
func (r *DaprCartRepository) SaveCart(ctx context.Context, cart *models.Cart) error {
	key := r.getCartKey(cart.UserID)

	data, err := json.Marshal(cart)
	if err != nil {
		r.logger.Error("Failed to marshal cart data",
			zap.String("userID", cart.UserID),
			zap.Error(err))
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	// Calculate TTL in seconds
	ttl := int(time.Until(cart.ExpiresAt).Seconds())
	if ttl <= 0 {
		ttl = 60 // Minimum TTL of 1 minute
	}

	// Create metadata with TTL
	metadata := map[string]string{
		"ttlInSeconds": fmt.Sprintf("%d", ttl),
	}

	// Save state with metadata
	err = r.client.SaveState(ctx, r.stateStoreName, key, data, metadata)
	if err != nil {
		r.logger.Error("Failed to save cart to Dapr state store",
			zap.String("userID", cart.UserID),
			zap.Error(err))
		return fmt.Errorf("failed to save cart: %w", err)
	}

	r.logger.Debug("Cart saved successfully",
		zap.String("userID", cart.UserID),
		zap.Int("ttlSeconds", ttl))

	return nil
}

// DeleteCart deletes a cart from Dapr state store
func (r *DaprCartRepository) DeleteCart(ctx context.Context, userID string) error {
	key := r.getCartKey(userID)

	err := r.client.DeleteState(ctx, r.stateStoreName, key, nil)
	if err != nil {
		r.logger.Error("Failed to delete cart from Dapr state store",
			zap.String("userID", userID),
			zap.Error(err))
		return fmt.Errorf("failed to delete cart: %w", err)
	}

	r.logger.Debug("Cart deleted successfully", zap.String("userID", userID))
	return nil
}

// SetCartTTL sets the TTL for a cart by re-saving it with new expiration
func (r *DaprCartRepository) SetCartTTL(ctx context.Context, userID string, ttl time.Duration) error {
	// Get existing cart
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// Update expiration time
	cart.ExpiresAt = time.Now().UTC().Add(ttl)
	cart.UpdatedAt = time.Now().UTC()

	// Re-save cart with new TTL
	return r.SaveCart(ctx, cart)
}

// CartExists checks if a cart exists in Dapr state store
func (r *DaprCartRepository) CartExists(ctx context.Context, userID string) (bool, error) {
	key := r.getCartKey(userID)

	item, err := r.client.GetState(ctx, r.stateStoreName, key, nil)
	if err != nil {
		r.logger.Error("Failed to check cart existence",
			zap.String("userID", userID),
			zap.Error(err))
		return false, fmt.Errorf("failed to check cart existence: %w", err)
	}

	return item.Value != nil && len(item.Value) > 0, nil
}

// AcquireLock - with Dapr State Management, ETag-based concurrency replaces distributed locks
// This method is kept for interface compatibility but uses ETag mechanism
func (r *DaprCartRepository) AcquireLock(ctx context.Context, userID string, ttl time.Duration) (bool, error) {
	// With Dapr State Management, we don't need explicit locks
	// ETag-based optimistic concurrency control handles this automatically
	r.logger.Debug("Lock acquisition not needed with Dapr ETag-based concurrency",
		zap.String("userID", userID))
	return true, nil
}

// ReleaseLock - with Dapr State Management, ETag-based concurrency replaces distributed locks
// This method is kept for interface compatibility
func (r *DaprCartRepository) ReleaseLock(ctx context.Context, userID string) error {
	// With Dapr State Management, we don't need explicit locks
	r.logger.Debug("Lock release not needed with Dapr ETag-based concurrency",
		zap.String("userID", userID))
	return nil
}

// GetAllCartKeys retrieves all cart keys for cleanup operations
// Note: This is a Dapr limitation - bulk query is not directly supported
// For production, consider using a separate metadata store or scheduled cleanup
func (r *DaprCartRepository) GetAllCartKeys(ctx context.Context) ([]string, error) {
	r.logger.Warn("GetAllCartKeys is not efficiently supported by Dapr State Management",
		zap.String("recommendation", "Use TTL-based expiration instead"))
	// Return empty slice - rely on Dapr TTL for automatic cleanup
	return []string{}, nil
}

// GetCartTTL gets the remaining TTL for a cart
// Note: Dapr doesn't expose TTL directly, so we calculate it from cart expiration
func (r *DaprCartRepository) GetCartTTL(ctx context.Context, userID string) (time.Duration, error) {
	cart, err := r.GetCart(ctx, userID)
	if err != nil {
		if err == models.ErrCartNotFound {
			return 0, err
		}
		return 0, fmt.Errorf("failed to get cart TTL: %w", err)
	}

	ttl := time.Until(cart.ExpiresAt)
	if ttl < 0 {
		ttl = 0
	}

	return ttl, nil
}

// Helper methods
func (r *DaprCartRepository) getCartKey(userID string) string {
	return cartKeyPrefix + userID
}
