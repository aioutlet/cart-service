package models

import (
	"errors"
	"time"
)

// Cart represents a shopping cart
type Cart struct {
	UserID     string     `json:"userId" redis:"user_id"`
	Items      []CartItem `json:"items" redis:"items"`
	TotalPrice float64    `json:"totalPrice" redis:"total_price"`
	TotalItems int        `json:"totalItems" redis:"total_items"`
	CreatedAt  time.Time  `json:"createdAt" redis:"created_at"`
	UpdatedAt  time.Time  `json:"updatedAt" redis:"updated_at"`
	ExpiresAt  time.Time  `json:"expiresAt" redis:"expires_at"`
}

// CartItem represents an item in the cart
type CartItem struct {
	ProductID   string  `json:"productId" redis:"product_id"`
	ProductName string  `json:"productName" redis:"product_name"`
	SKU         string  `json:"sku" redis:"sku"`
	Price       float64 `json:"price" redis:"price"`
	Quantity    int     `json:"quantity" redis:"quantity"`
	ImageURL    string  `json:"imageUrl" redis:"image_url"`
	Category    string  `json:"category" redis:"category"`
	Subtotal    float64 `json:"subtotal" redis:"subtotal"`
	AddedAt     time.Time `json:"addedAt" redis:"added_at"`
}

// AddItemRequest represents a request to add an item to cart
type AddItemRequest struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// UpdateItemRequest represents a request to update an item in cart
type UpdateItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=0"`
}

// TransferCartRequest represents a request to transfer guest cart to user cart
type TransferCartRequest struct {
	GuestID string `json:"guestId" binding:"required"`
}

// ProductInfo represents product information from product service
type ProductInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	SKU         string  `json:"sku"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"imageUrl"`
	Category    string  `json:"category"`
	IsActive    bool    `json:"isActive"`
	StockQty    int     `json:"stockQty"`
}

// CartResponse represents the response format for cart operations
type CartResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *Cart  `json:"data,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// CartSummary represents a summary of cart for order processing
type CartSummary struct {
	UserID     string     `json:"userId"`
	Items      []CartItem `json:"items"`
	TotalPrice float64    `json:"totalPrice"`
	TotalItems int        `json:"totalItems"`
}

// Custom errors
var (
	ErrCartNotFound     = errors.New("cart not found")
	ErrItemNotFound     = errors.New("item not found in cart")
	ErrProductNotFound  = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrMaxItemsExceeded = errors.New("maximum number of items exceeded")
	ErrMaxQuantityExceeded = errors.New("maximum quantity per item exceeded")
	ErrInvalidQuantity  = errors.New("invalid quantity")
	ErrCartExpired      = errors.New("cart has expired")
)

// NewCart creates a new cart for a user
func NewCart(userID string, ttl time.Duration) *Cart {
	now := time.Now().UTC()
	return &Cart{
		UserID:     userID,
		Items:      make([]CartItem, 0),
		TotalPrice: 0,
		TotalItems: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
		ExpiresAt:  now.Add(ttl),
	}
}

// AddItem adds an item to the cart or updates quantity if it exists
func (c *Cart) AddItem(item CartItem, maxItems, maxQuantity int) error {
	// Check if cart has expired
	if time.Now().UTC().After(c.ExpiresAt) {
		return ErrCartExpired
	}

	// Check for existing item
	for i, existingItem := range c.Items {
		if existingItem.ProductID == item.ProductID {
			newQty := existingItem.Quantity + item.Quantity
			if newQty > maxQuantity {
				return ErrMaxQuantityExceeded
			}
			c.Items[i].Quantity = newQty
			c.Items[i].Subtotal = float64(newQty) * existingItem.Price
			c.UpdateTotals()
			c.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	// Check max items limit
	if len(c.Items) >= maxItems {
		return ErrMaxItemsExceeded
	}

	// Check quantity limit
	if item.Quantity > maxQuantity {
		return ErrMaxQuantityExceeded
	}

	// Add new item
	item.Subtotal = float64(item.Quantity) * item.Price
	item.AddedAt = time.Now().UTC()
	c.Items = append(c.Items, item)
	c.UpdateTotals()
	c.UpdatedAt = time.Now().UTC()
	
	return nil
}

// UpdateItemQuantity updates the quantity of an item in the cart
func (c *Cart) UpdateItemQuantity(productID string, quantity int, maxQuantity int) error {
	// Check if cart has expired
	if time.Now().UTC().After(c.ExpiresAt) {
		return ErrCartExpired
	}

	if quantity < 0 {
		return ErrInvalidQuantity
	}

	if quantity > maxQuantity {
		return ErrMaxQuantityExceeded
	}

	for i, item := range c.Items {
		if item.ProductID == productID {
			if quantity == 0 {
				// Remove item if quantity is 0
				c.Items = append(c.Items[:i], c.Items[i+1:]...)
			} else {
				c.Items[i].Quantity = quantity
				c.Items[i].Subtotal = float64(quantity) * item.Price
			}
			c.UpdateTotals()
			c.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return ErrItemNotFound
}

// RemoveItem removes an item from the cart
func (c *Cart) RemoveItem(productID string) error {
	// Check if cart has expired
	if time.Now().UTC().After(c.ExpiresAt) {
		return ErrCartExpired
	}

	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			c.UpdateTotals()
			c.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return ErrItemNotFound
}

// Clear removes all items from the cart
func (c *Cart) Clear() {
	c.Items = make([]CartItem, 0)
	c.TotalPrice = 0
	c.TotalItems = 0
	c.UpdatedAt = time.Now().UTC()
}

// IsEmpty checks if the cart is empty
func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

// IsExpired checks if the cart has expired
func (c *Cart) IsExpired() bool {
	return time.Now().UTC().After(c.ExpiresAt)
}

// ExtendExpiry extends the cart expiry time
func (c *Cart) ExtendExpiry(ttl time.Duration) {
	c.ExpiresAt = time.Now().UTC().Add(ttl)
	c.UpdatedAt = time.Now().UTC()
}

// UpdateTotals recalculates total price and total items
func (c *Cart) UpdateTotals() {
	c.TotalPrice = 0
	c.TotalItems = 0
	
	for _, item := range c.Items {
		c.TotalPrice += item.Subtotal
		c.TotalItems += item.Quantity
	}
}

// GetItem returns a specific item from the cart
func (c *Cart) GetItem(productID string) (*CartItem, error) {
	for _, item := range c.Items {
		if item.ProductID == productID {
			return &item, nil
		}
	}
	return nil, ErrItemNotFound
}

// HasItem checks if the cart contains a specific product
func (c *Cart) HasItem(productID string) bool {
	for _, item := range c.Items {
		if item.ProductID == productID {
			return true
		}
	}
	return false
}
