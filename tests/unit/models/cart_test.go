package models

import (
	"testing"
	"time"

	"github.com/aioutlet/cart-service/internal/models"
	"github.com/aioutlet/cart-service/tests/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewCart(t *testing.T) {
	userID := "user123"
	ttl := 24 * time.Hour

	cart := models.NewCart(userID, ttl)

	assert.Equal(t, userID, cart.UserID)
	assert.Empty(t, cart.Items)
	assert.Equal(t, float64(0), cart.TotalPrice)
	assert.Equal(t, 0, cart.TotalItems)
	assert.False(t, cart.CreatedAt.IsZero())
	assert.False(t, cart.UpdatedAt.IsZero())
	assert.True(t, cart.ExpiresAt.After(time.Now()))
}

func TestCart_AddItem(t *testing.T) {
	tests := []struct {
		name           string
		cart           *models.Cart
		item           models.CartItem
		maxItems       int
		maxQuantity    int
		expectedError  error
		expectedItems  int
		expectedTotal  float64
	}{
		{
			name:          "Add new item successfully",
			cart:          testutils.CreateTestCart("user123", time.Hour),
			item:          testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2),
			maxItems:      10,
			maxQuantity:   5,
			expectedError: nil,
			expectedItems: 1,
			expectedTotal: 21.98,
		},
		{
			name: "Add item to existing product",
			cart: func() *models.Cart {
				cart := testutils.CreateTestCart("user123", time.Hour)
				item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 1)
				cart.AddItem(item, 10, 5)
				return cart
			}(),
			item:          testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2),
			maxItems:      10,
			maxQuantity:   5,
			expectedError: nil,
			expectedItems: 1,
			expectedTotal: 32.97,
		},
		{
			name:          "Exceed max items",
			cart:          testutils.CreateTestCart("user123", time.Hour),
			item:          testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2),
			maxItems:      0,
			maxQuantity:   5,
			expectedError: models.ErrMaxItemsExceeded,
			expectedItems: 0,
			expectedTotal: 0,
		},
		{
			name:          "Exceed max quantity",
			cart:          testutils.CreateTestCart("user123", time.Hour),
			item:          testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 10),
			maxItems:      10,
			maxQuantity:   5,
			expectedError: models.ErrMaxQuantityExceeded,
			expectedItems: 0,
			expectedTotal: 0,
		},
		{
			name: "Exceed max quantity when adding to existing",
			cart: func() *models.Cart {
				cart := testutils.CreateTestCart("user123", time.Hour)
				item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 3)
				cart.AddItem(item, 10, 5)
				return cart
			}(),
			item:          testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 3),
			maxItems:      10,
			maxQuantity:   5,
			expectedError: models.ErrMaxQuantityExceeded,
			expectedItems: 1,
			expectedTotal: 32.97,
		},
		{
			name: "Cart expired",
			cart: func() *models.Cart {
				cart := testutils.CreateTestCart("user123", -time.Hour) // Expired
				return cart
			}(),
			item:          testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2),
			maxItems:      10,
			maxQuantity:   5,
			expectedError: models.ErrCartExpired,
			expectedItems: 0,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cart.AddItem(tt.item, tt.maxItems, tt.maxQuantity)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedItems, len(tt.cart.Items))
			assert.InDelta(t, tt.expectedTotal, tt.cart.TotalPrice, 0.01)
		})
	}
}

func TestCart_UpdateItemQuantity(t *testing.T) {
	tests := []struct {
		name           string
		cart           *models.Cart
		productID      string
		quantity       int
		maxQuantity    int
		expectedError  error
		expectedItems  int
		expectedTotal  float64
	}{
		{
			name: "Update item quantity successfully",
			cart: func() *models.Cart {
				cart := testutils.CreateTestCart("user123", time.Hour)
				item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
				cart.AddItem(item, 10, 5)
				return cart
			}(),
			productID:     "prod1",
			quantity:      3,
			maxQuantity:   5,
			expectedError: nil,
			expectedItems: 1,
			expectedTotal: 32.97,
		},
		{
			name: "Remove item when quantity is 0",
			cart: func() *models.Cart {
				cart := testutils.CreateTestCart("user123", time.Hour)
				item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
				cart.AddItem(item, 10, 5)
				return cart
			}(),
			productID:     "prod1",
			quantity:      0,
			maxQuantity:   5,
			expectedError: nil,
			expectedItems: 0,
			expectedTotal: 0,
		},
		{
			name: "Item not found",
			cart: testutils.CreateTestCart("user123", time.Hour),
			productID:     "nonexistent",
			quantity:      2,
			maxQuantity:   5,
			expectedError: models.ErrItemNotFound,
			expectedItems: 0,
			expectedTotal: 0,
		},
		{
			name: "Invalid quantity",
			cart: func() *models.Cart {
				cart := testutils.CreateTestCart("user123", time.Hour)
				item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
				cart.AddItem(item, 10, 5)
				return cart
			}(),
			productID:     "prod1",
			quantity:      -1,
			maxQuantity:   5,
			expectedError: models.ErrInvalidQuantity,
			expectedItems: 1,
			expectedTotal: 21.98,
		},
		{
			name: "Exceed max quantity",
			cart: func() *models.Cart {
				cart := testutils.CreateTestCart("user123", time.Hour)
				item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
				cart.AddItem(item, 10, 5)
				return cart
			}(),
			productID:     "prod1",
			quantity:      10,
			maxQuantity:   5,
			expectedError: models.ErrMaxQuantityExceeded,
			expectedItems: 1,
			expectedTotal: 21.98,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cart.UpdateItemQuantity(tt.productID, tt.quantity, tt.maxQuantity)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedItems, len(tt.cart.Items))
			assert.InDelta(t, tt.expectedTotal, tt.cart.TotalPrice, 0.01)
		})
	}
}

func TestCart_RemoveItem(t *testing.T) {
	tests := []struct {
		name           string
		cart           *models.Cart
		productID      string
		expectedError  error
		expectedItems  int
		expectedTotal  float64
	}{
		{
			name: "Remove item successfully",
			cart: func() *models.Cart {
				cart := testutils.CreateTestCart("user123", time.Hour)
				item1 := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
				item2 := testutils.CreateTestCartItem("prod2", "Product 2", 15.99, 1)
				cart.AddItem(item1, 10, 5)
				cart.AddItem(item2, 10, 5)
				return cart
			}(),
			productID:     "prod1",
			expectedError: nil,
			expectedItems: 1,
			expectedTotal: 15.99,
		},
		{
			name:          "Item not found",
			cart:          testutils.CreateTestCart("user123", time.Hour),
			productID:     "nonexistent",
			expectedError: models.ErrItemNotFound,
			expectedItems: 0,
			expectedTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cart.RemoveItem(tt.productID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedItems, len(tt.cart.Items))
			assert.InDelta(t, tt.expectedTotal, tt.cart.TotalPrice, 0.01)
		})
	}
}

func TestCart_Clear(t *testing.T) {
	cart := testutils.CreateTestCart("user123", time.Hour)
	item1 := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
	item2 := testutils.CreateTestCartItem("prod2", "Product 2", 15.99, 1)
	cart.AddItem(item1, 10, 5)
	cart.AddItem(item2, 10, 5)

	assert.Equal(t, 2, len(cart.Items))
	assert.Greater(t, cart.TotalPrice, float64(0))

	cart.Clear()

	assert.Equal(t, 0, len(cart.Items))
	assert.Equal(t, float64(0), cart.TotalPrice)
	assert.Equal(t, 0, cart.TotalItems)
}

func TestCart_IsEmpty(t *testing.T) {
	cart := testutils.CreateTestCart("user123", time.Hour)

	assert.True(t, cart.IsEmpty())

	item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
	cart.AddItem(item, 10, 5)

	assert.False(t, cart.IsEmpty())
}

func TestCart_IsExpired(t *testing.T) {
	// Not expired cart
	cart := testutils.CreateTestCart("user123", time.Hour)
	assert.False(t, cart.IsExpired())

	// Expired cart
	expiredCart := testutils.CreateTestCart("user123", -time.Hour)
	assert.True(t, expiredCart.IsExpired())
}

func TestCart_ExtendExpiry(t *testing.T) {
	cart := testutils.CreateTestCart("user123", time.Hour)
	originalExpiry := cart.ExpiresAt

	cart.ExtendExpiry(2 * time.Hour)

	assert.True(t, cart.ExpiresAt.After(originalExpiry))
}

func TestCart_GetItem(t *testing.T) {
	cart := testutils.CreateTestCart("user123", time.Hour)
	item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
	cart.AddItem(item, 10, 5)

	// Get existing item
	foundItem, err := cart.GetItem("prod1")
	assert.NoError(t, err)
	assert.Equal(t, "prod1", foundItem.ProductID)

	// Get non-existing item
	_, err = cart.GetItem("nonexistent")
	assert.Equal(t, models.ErrItemNotFound, err)
}

func TestCart_HasItem(t *testing.T) {
	cart := testutils.CreateTestCart("user123", time.Hour)
	item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
	cart.AddItem(item, 10, 5)

	assert.True(t, cart.HasItem("prod1"))
	assert.False(t, cart.HasItem("nonexistent"))
}
