package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aioutlet/cart-service/internal/config"
	"github.com/aioutlet/cart-service/internal/models"
	"github.com/aioutlet/cart-service/internal/repository"
	"github.com/aioutlet/cart-service/pkg/clients"
	dapr "github.com/dapr/go-sdk/client"
	"go.uber.org/zap"
)

// CartService interface defines cart service operations
type CartService interface {
	GetCart(ctx context.Context, userID string) (*models.Cart, error)
	AddItem(ctx context.Context, userID string, request models.AddItemRequest) (*models.Cart, error)
	UpdateItem(ctx context.Context, userID string, productID string, request models.UpdateItemRequest) (*models.Cart, error)
	RemoveItem(ctx context.Context, userID string, productID string) (*models.Cart, error)
	ClearCart(ctx context.Context, userID string) error
	TransferCart(ctx context.Context, fromUserID, toUserID string) (*models.Cart, error)
	ValidateCart(ctx context.Context, userID string) (*models.Cart, error)
	GetCartSummary(ctx context.Context, userID string) (*models.CartSummary, error)
}

// cartService implements CartService interface
type cartService struct {
	repo           repository.CartRepository
	productClient  clients.ProductClient
	inventoryClient clients.InventoryClient
	config         *config.Config
	logger         *zap.Logger
}

// NewCartService creates a new cart service
func NewCartService(
	repo repository.CartRepository,
	daprClient dapr.Client,
	cfg *config.Config,
	logger *zap.Logger,
) CartService {
	return &cartService{
		repo:            repo,
		productClient:   clients.NewProductClient(daprClient, logger),
		inventoryClient: clients.NewInventoryClient(daprClient, logger),
		config:          cfg,
		logger:          logger,
	}
}

// NewCartServiceWithClients creates a new cart service with provided clients (for testing)
func NewCartServiceWithClients(
	repo repository.CartRepository,
	productClient clients.ProductClient,
	inventoryClient clients.InventoryClient,
	cfg *config.Config,
	logger *zap.Logger,
) CartService {
	return &cartService{
		repo:            repo,
		productClient:   productClient,
		inventoryClient: inventoryClient,
		config:          cfg,
		logger:          logger,
	}
}

// GetCart retrieves a cart for a user
func (s *cartService) GetCart(ctx context.Context, userID string) (*models.Cart, error) {
	s.logger.Debug("Getting cart", 
		zap.String("userID", userID))

	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		if err == models.ErrCartNotFound {
			// Create a new empty cart
			cart = models.NewCart(userID, s.config.Cart.DefaultTTL)
			if err := s.repo.SaveCart(ctx, cart); err != nil {
				s.logger.Error("Failed to save new cart", 
					zap.String("userID", userID),
					zap.Error(err))
				return nil, fmt.Errorf("failed to create new cart: %w", err)
			}
			s.logger.Info("Created new cart", 
				zap.String("userID", userID))
		} else {
			return nil, err
		}
	}

	return cart, nil
}

// AddItem adds an item to the cart
func (s *cartService) AddItem(ctx context.Context, userID string, request models.AddItemRequest) (*models.Cart, error) {
	s.logger.Debug("Adding item to cart", 
		zap.String("userID", userID),
		zap.String("productID", request.ProductID),
		zap.Int("quantity", request.Quantity))

	// Acquire lock for cart operations
	lockAcquired, err := s.repo.AcquireLock(ctx, userID, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire cart lock: %w", err)
	}
	if !lockAcquired {
		err := fmt.Errorf("cart is currently being modified, please try again")
		return nil, err
	}
	defer s.repo.ReleaseLock(ctx, userID)

	// Get product information
	productInfo, err := s.productClient.GetProduct(ctx, request.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product information: %w", err)
	}

	if !productInfo.IsActive {
		return nil, fmt.Errorf("product is not available")
	}

	// Check inventory
	available, err := s.inventoryClient.CheckAvailability(ctx, request.ProductID, request.Quantity)
	if err != nil {
		s.logger.Warn("Failed to check inventory, allowing operation", 
			zap.String("productID", request.ProductID),
			zap.Error(err))
	} else if !available {
		return nil, models.ErrInsufficientStock
	}

	// Get or create cart
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create cart item
	cartItem := models.CartItem{
		ProductID:   productInfo.ID,
		ProductName: productInfo.Name,
		SKU:         productInfo.SKU,
		Price:       productInfo.Price,
		Quantity:    request.Quantity,
		ImageURL:    productInfo.ImageURL,
		Category:    productInfo.Category,
		AddedAt:     time.Now().UTC(),
	}

	// Add item to cart
	if err := cart.AddItem(cartItem, s.config.Cart.MaxItems, s.config.Cart.MaxItemQty); err != nil {
		return nil, err
	}

	// Save cart
	if err := s.repo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	s.logger.Info("Item added to cart successfully", 
		zap.String("userID", userID),
		zap.String("productID", request.ProductID),
		zap.Int("quantity", request.Quantity))

	return cart, nil
}

// UpdateItem updates an item quantity in the cart
func (s *cartService) UpdateItem(ctx context.Context, userID string, productID string, request models.UpdateItemRequest) (*models.Cart, error) {
	s.logger.Debug("Updating item in cart", 
		zap.String("userID", userID),
		zap.String("productID", productID),
		zap.Int("quantity", request.Quantity))

	// Acquire lock for cart operations
	lockAcquired, err := s.repo.AcquireLock(ctx, userID, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire cart lock: %w", err)
	}
	if !lockAcquired {
		return nil, fmt.Errorf("cart is currently being modified, please try again")
	}
	defer s.repo.ReleaseLock(ctx, userID)

	// Get cart
	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check inventory if increasing quantity
	if request.Quantity > 0 {
		currentItem, err := cart.GetItem(productID)
		if err == nil && request.Quantity > currentItem.Quantity {
			additionalQty := request.Quantity - currentItem.Quantity
			available, err := s.inventoryClient.CheckAvailability(ctx, productID, additionalQty)
			if err != nil {
				s.logger.Warn("Failed to check inventory, allowing operation", 
					zap.String("productID", productID),
					zap.Error(err))
			} else if !available {
				return nil, models.ErrInsufficientStock
			}
		}
	}

	// Update item quantity
	if err := cart.UpdateItemQuantity(productID, request.Quantity, s.config.Cart.MaxItemQty); err != nil {
		return nil, err
	}

	// Save cart
	if err := s.repo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	s.logger.Info("Item updated in cart successfully", 
		zap.String("userID", userID),
		zap.String("productID", productID),
		zap.Int("quantity", request.Quantity))

	return cart, nil
}

// RemoveItem removes an item from the cart
func (s *cartService) RemoveItem(ctx context.Context, userID string, productID string) (*models.Cart, error) {
	s.logger.Debug("Removing item from cart", 
		zap.String("userID", userID),
		zap.String("productID", productID))

	// Acquire lock for cart operations
	lockAcquired, err := s.repo.AcquireLock(ctx, userID, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire cart lock: %w", err)
	}
	if !lockAcquired {
		return nil, fmt.Errorf("cart is currently being modified, please try again")
	}
	defer s.repo.ReleaseLock(ctx, userID)

	// Get cart
	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Remove item
	if err := cart.RemoveItem(productID); err != nil {
		return nil, err
	}

	// Save cart
	if err := s.repo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	s.logger.Info("Item removed from cart successfully", 
		zap.String("userID", userID),
		zap.String("productID", productID))

	return cart, nil
}

// ClearCart removes all items from the cart
func (s *cartService) ClearCart(ctx context.Context, userID string) error {
	s.logger.Debug("Clearing cart", zap.String("userID", userID))

	// Acquire lock for cart operations
	lockAcquired, err := s.repo.AcquireLock(ctx, userID, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire cart lock: %w", err)
	}
	if !lockAcquired {
		return fmt.Errorf("cart is currently being modified, please try again")
	}
	defer s.repo.ReleaseLock(ctx, userID)

	// Delete cart from Redis
	if err := s.repo.DeleteCart(ctx, userID); err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	s.logger.Info("Cart cleared successfully", zap.String("userID", userID))
	return nil
}

// TransferCart transfers items from one cart to another (guest to user)
func (s *cartService) TransferCart(ctx context.Context, fromUserID, toUserID string) (*models.Cart, error) {
	s.logger.Debug("Transferring cart", 
		zap.String("fromUserID", fromUserID),
		zap.String("toUserID", toUserID))

	// Acquire locks for both carts
	fromLockAcquired, err := s.repo.AcquireLock(ctx, fromUserID, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire source cart lock: %w", err)
	}
	if !fromLockAcquired {
		return nil, fmt.Errorf("source cart is currently being modified, please try again")
	}
	defer s.repo.ReleaseLock(ctx, fromUserID)

	toLockAcquired, err := s.repo.AcquireLock(ctx, toUserID, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire target cart lock: %w", err)
	}
	if !toLockAcquired {
		return nil, fmt.Errorf("target cart is currently being modified, please try again")
	}
	defer s.repo.ReleaseLock(ctx, toUserID)

	// Get source cart
	fromCart, err := s.repo.GetCart(ctx, fromUserID)
	if err != nil {
		if err == models.ErrCartNotFound {
			// No cart to transfer, return empty target cart
			return s.GetCart(ctx, toUserID)
		}
		return nil, err
	}

	// Get or create target cart
	toCart, err := s.GetCart(ctx, toUserID)
	if err != nil {
		return nil, err
	}

	// Transfer items
	for _, item := range fromCart.Items {
		if err := toCart.AddItem(item, s.config.Cart.MaxItems, s.config.Cart.MaxItemQty); err != nil {
			s.logger.Warn("Failed to transfer item, skipping", 
				zap.String("productID", item.ProductID),
				zap.Error(err))
			continue
		}
	}

	// Save target cart
	if err := s.repo.SaveCart(ctx, toCart); err != nil {
		return nil, fmt.Errorf("failed to save target cart: %w", err)
	}

	// Delete source cart
	if err := s.repo.DeleteCart(ctx, fromUserID); err != nil {
		s.logger.Error("Failed to delete source cart after transfer", 
			zap.String("fromUserID", fromUserID),
			zap.Error(err))
	}

	s.logger.Info("Cart transferred successfully", 
		zap.String("fromUserID", fromUserID),
		zap.String("toUserID", toUserID),
		zap.Int("itemsTransferred", len(fromCart.Items)))

	return toCart, nil
}

// ValidateCart validates all items in the cart against current product and inventory data
func (s *cartService) ValidateCart(ctx context.Context, userID string) (*models.Cart, error) {
	s.logger.Debug("Validating cart", zap.String("userID", userID))

	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	hasChanges := false
	validItems := make([]models.CartItem, 0, len(cart.Items))

	for _, item := range cart.Items {
		// Check product availability
		productInfo, err := s.productClient.GetProduct(ctx, item.ProductID)
		if err != nil || !productInfo.IsActive {
			s.logger.Info("Removing unavailable product from cart", 
				zap.String("userID", userID),
				zap.String("productID", item.ProductID))
			hasChanges = true
			continue
		}

		// Update price if changed
		if item.Price != productInfo.Price {
			s.logger.Info("Updating product price in cart", 
				zap.String("userID", userID),
				zap.String("productID", item.ProductID),
				zap.Float64("oldPrice", item.Price),
				zap.Float64("newPrice", productInfo.Price))
			item.Price = productInfo.Price
			item.Subtotal = float64(item.Quantity) * productInfo.Price
			hasChanges = true
		}

		// Check inventory availability
		available, err := s.inventoryClient.CheckAvailability(ctx, item.ProductID, item.Quantity)
		if err != nil {
			s.logger.Warn("Failed to check inventory during validation", 
				zap.String("productID", item.ProductID),
				zap.Error(err))
		} else if !available {
			// Get available quantity
			availableQty, err := s.inventoryClient.GetAvailableQuantity(ctx, item.ProductID)
			if err != nil || availableQty <= 0 {
				s.logger.Info("Removing out-of-stock product from cart", 
					zap.String("userID", userID),
					zap.String("productID", item.ProductID))
				hasChanges = true
				continue
			}
			
			// Adjust quantity to available amount
			if availableQty < item.Quantity {
				s.logger.Info("Adjusting quantity to available stock", 
					zap.String("userID", userID),
					zap.String("productID", item.ProductID),
					zap.Int("requestedQty", item.Quantity),
					zap.Int("availableQty", availableQty))
				item.Quantity = availableQty
				item.Subtotal = float64(item.Quantity) * item.Price
				hasChanges = true
			}
		}

		validItems = append(validItems, item)
	}

	if hasChanges {
		cart.Items = validItems
		cart.UpdateTotals()
		
		if err := s.repo.SaveCart(ctx, cart); err != nil {
			return nil, fmt.Errorf("failed to save validated cart: %w", err)
		}

		s.logger.Info("Cart validated and updated", 
			zap.String("userID", userID),
			zap.Int("originalItems", len(cart.Items)),
			zap.Int("validItems", len(validItems)))
	}

	return cart, nil
}

// GetCartSummary returns a summary of the cart for order processing
func (s *cartService) GetCartSummary(ctx context.Context, userID string) (*models.CartSummary, error) {
	cart, err := s.ValidateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.CartSummary{
		UserID:     cart.UserID,
		Items:      cart.Items,
		TotalPrice: cart.TotalPrice,
		TotalItems: cart.TotalItems,
	}, nil
}
