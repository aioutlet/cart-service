package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aioutlet/cart-service/internal/models"
	"github.com/aioutlet/cart-service/internal/services"
	"github.com/aioutlet/cart-service/tests/mocks"
	"github.com/aioutlet/cart-service/tests/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func setupCartService() (services.CartService, *mocks.MockCartRepository, *mocks.MockProductClient, *mocks.MockInventoryClient) {
	mockRepo := &mocks.MockCartRepository{}
	mockProductClient := &mocks.MockProductClient{}
	mockInventoryClient := &mocks.MockInventoryClient{}
	config := testutils.CreateTestConfig()
	logger := zap.NewNop()

	// Create a cart service with mock clients using the test constructor
	cartService := services.NewCartServiceWithClients(
		mockRepo,
		mockProductClient,
		mockInventoryClient,
		config,
		logger,
	)

	return cartService, mockRepo, mockProductClient, mockInventoryClient
}

func TestCartService_GetCart(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMocks     func(*mocks.MockCartRepository)
		expectedError  error
		expectNewCart  bool
	}{
		{
			name:   "Get existing cart",
			userID: "user123",
			setupMocks: func(repo *mocks.MockCartRepository) {
				cart := testutils.CreateTestCart("user123", time.Hour)
				repo.On("GetCart", mock.Anything, "user123").Return(cart, nil)
			},
			expectedError: nil,
			expectNewCart: false,
		},
		{
			name:   "Create new cart when not found",
			userID: "user123",
			setupMocks: func(repo *mocks.MockCartRepository) {
				repo.On("GetCart", mock.Anything, "user123").Return(nil, models.ErrCartNotFound)
				repo.On("SaveCart", mock.Anything, mock.AnythingOfType("*models.Cart")).Return(nil)
			},
			expectedError: nil,
			expectNewCart: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartService, mockRepo, _, _ := setupCartService()
			tt.setupMocks(mockRepo)

			cart, err := cartService.GetCart(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, cart)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cart)
				assert.Equal(t, tt.userID, cart.UserID)
				
				if tt.expectNewCart {
					assert.True(t, cart.IsEmpty())
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCartService_AddItem(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		request       models.AddItemRequest
		setupMocks    func(*mocks.MockCartRepository, *mocks.MockProductClient, *mocks.MockInventoryClient)
		expectedError error
	}{
		{
			name:   "Add item successfully",
			userID: "user123",
			request: models.AddItemRequest{
				ProductID: "prod1",
				Quantity:  2,
			},
			setupMocks: func(repo *mocks.MockCartRepository, productClient *mocks.MockProductClient, inventoryClient *mocks.MockInventoryClient) {
				// Mock lock operations
				repo.On("AcquireLock", mock.Anything, "user123", 30*time.Second).Return(true, nil)
				repo.On("ReleaseLock", mock.Anything, "user123").Return(nil)

				// Mock product service
				productInfo := testutils.CreateTestProductInfo("prod1", "Product 1", 10.99, true)
				productClient.On("GetProduct", mock.Anything, "prod1").Return(productInfo, nil)

				// Mock inventory service
				inventoryClient.On("CheckAvailability", mock.Anything, "prod1", 2).Return(true, nil)

				// Mock repository operations
				cart := testutils.CreateTestCart("user123", time.Hour)
				repo.On("GetCart", mock.Anything, "user123").Return(cart, nil)
				repo.On("SaveCart", mock.Anything, mock.AnythingOfType("*models.Cart")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "Product not found",
			userID: "user123",
			request: models.AddItemRequest{
				ProductID: "nonexistent",
				Quantity:  2,
			},
			setupMocks: func(repo *mocks.MockCartRepository, productClient *mocks.MockProductClient, inventoryClient *mocks.MockInventoryClient) {
				repo.On("AcquireLock", mock.Anything, "user123", 30*time.Second).Return(true, nil)
				repo.On("ReleaseLock", mock.Anything, "user123").Return(nil)
				productClient.On("GetProduct", mock.Anything, "nonexistent").Return(nil, models.ErrProductNotFound)
			},
			expectedError: models.ErrProductNotFound,
		},
		{
			name:   "Insufficient stock",
			userID: "user123",
			request: models.AddItemRequest{
				ProductID: "prod1",
				Quantity:  10,
			},
			setupMocks: func(repo *mocks.MockCartRepository, productClient *mocks.MockProductClient, inventoryClient *mocks.MockInventoryClient) {
				repo.On("AcquireLock", mock.Anything, "user123", 30*time.Second).Return(true, nil)
				repo.On("ReleaseLock", mock.Anything, "user123").Return(nil)

				productInfo := testutils.CreateTestProductInfo("prod1", "Product 1", 10.99, true)
				productClient.On("GetProduct", mock.Anything, "prod1").Return(productInfo, nil)
				inventoryClient.On("CheckAvailability", mock.Anything, "prod1", 10).Return(false, nil)
			},
			expectedError: models.ErrInsufficientStock,
		},
		{
			name:   "Product not active",
			userID: "user123",
			request: models.AddItemRequest{
				ProductID: "prod1",
				Quantity:  2,
			},
			setupMocks: func(repo *mocks.MockCartRepository, productClient *mocks.MockProductClient, inventoryClient *mocks.MockInventoryClient) {
				repo.On("AcquireLock", mock.Anything, "user123", 30*time.Second).Return(true, nil)
				repo.On("ReleaseLock", mock.Anything, "user123").Return(nil)

				productInfo := testutils.CreateTestProductInfo("prod1", "Product 1", 10.99, false)
				productClient.On("GetProduct", mock.Anything, "prod1").Return(productInfo, nil)
			},
			expectedError: fmt.Errorf("product is not available"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartService, mockRepo, mockProductClient, mockInventoryClient := setupCartService()
			tt.setupMocks(mockRepo, mockProductClient, mockInventoryClient)

			cart, err := cartService.AddItem(context.Background(), tt.userID, tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Contains(t, err.Error(), tt.expectedError.Error())
				}
				assert.Nil(t, cart)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cart)
				assert.True(t, cart.HasItem(tt.request.ProductID))
			}

			mockRepo.AssertExpectations(t)
			mockProductClient.AssertExpectations(t)
			mockInventoryClient.AssertExpectations(t)
		})
	}
}

func TestCartService_UpdateItem(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		productID     string
		request       models.UpdateItemRequest
		setupMocks    func(*mocks.MockCartRepository, *mocks.MockInventoryClient)
		expectedError error
	}{
		{
			name:      "Update item quantity successfully",
			userID:    "user123",
			productID: "prod1",
			request: models.UpdateItemRequest{
				Quantity: 3,
			},
			setupMocks: func(repo *mocks.MockCartRepository, inventoryClient *mocks.MockInventoryClient) {
				repo.On("AcquireLock", mock.Anything, "user123", 30*time.Second).Return(true, nil)
				repo.On("ReleaseLock", mock.Anything, "user123").Return(nil)

				cart := testutils.CreateTestCart("user123", time.Hour)
				item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
				cart.AddItem(item, 10, 5)
				
				repo.On("GetCart", mock.Anything, "user123").Return(cart, nil)
				inventoryClient.On("CheckAvailability", mock.Anything, "prod1", 1).Return(true, nil) // Additional quantity check
				repo.On("SaveCart", mock.Anything, mock.AnythingOfType("*models.Cart")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "Remove item when quantity is 0",
			userID:    "user123",
			productID: "prod1",
			request: models.UpdateItemRequest{
				Quantity: 0,
			},
			setupMocks: func(repo *mocks.MockCartRepository, inventoryClient *mocks.MockInventoryClient) {
				repo.On("AcquireLock", mock.Anything, "user123", 30*time.Second).Return(true, nil)
				repo.On("ReleaseLock", mock.Anything, "user123").Return(nil)

				cart := testutils.CreateTestCart("user123", time.Hour)
				item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
				cart.AddItem(item, 10, 5)
				
				repo.On("GetCart", mock.Anything, "user123").Return(cart, nil)
				repo.On("SaveCart", mock.Anything, mock.AnythingOfType("*models.Cart")).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cartService, mockRepo, _, mockInventoryClient := setupCartService()
			tt.setupMocks(mockRepo, mockInventoryClient)

			cart, err := cartService.UpdateItem(context.Background(), tt.userID, tt.productID, tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, cart)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cart)
				
				if tt.request.Quantity > 0 {
					assert.True(t, cart.HasItem(tt.productID))
					item, _ := cart.GetItem(tt.productID)
					assert.Equal(t, tt.request.Quantity, item.Quantity)
				} else {
					assert.False(t, cart.HasItem(tt.productID))
				}
			}

			mockRepo.AssertExpectations(t)
			mockInventoryClient.AssertExpectations(t)
		})
	}
}

func TestCartService_RemoveItem(t *testing.T) {
	cartService, mockRepo, _, _ := setupCartService()

	userID := "user123"
	productID := "prod1"

	// Setup mocks
	mockRepo.On("AcquireLock", mock.Anything, userID, 30*time.Second).Return(true, nil)
	mockRepo.On("ReleaseLock", mock.Anything, userID).Return(nil)

	cart := testutils.CreateTestCart(userID, time.Hour)
	item := testutils.CreateTestCartItem(productID, "Product 1", 10.99, 2)
	cart.AddItem(item, 10, 5)

	mockRepo.On("GetCart", mock.Anything, userID).Return(cart, nil)
	mockRepo.On("SaveCart", mock.Anything, mock.AnythingOfType("*models.Cart")).Return(nil)

	// Execute
	result, err := cartService.RemoveItem(context.Background(), userID, productID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.HasItem(productID))

	mockRepo.AssertExpectations(t)
}

func TestCartService_ClearCart(t *testing.T) {
	cartService, mockRepo, _, _ := setupCartService()

	userID := "user123"

	// Setup mocks
	mockRepo.On("AcquireLock", mock.Anything, userID, 30*time.Second).Return(true, nil)
	mockRepo.On("ReleaseLock", mock.Anything, userID).Return(nil)
	mockRepo.On("DeleteCart", mock.Anything, userID).Return(nil)

	// Execute
	err := cartService.ClearCart(context.Background(), userID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCartService_TransferCart(t *testing.T) {
	cartService, mockRepo, _, _ := setupCartService()

	fromUserID := "guest123"
	toUserID := "user123"

	// Setup mocks
	mockRepo.On("AcquireLock", mock.Anything, fromUserID, 30*time.Second).Return(true, nil)
	mockRepo.On("ReleaseLock", mock.Anything, fromUserID).Return(nil)
	mockRepo.On("AcquireLock", mock.Anything, toUserID, 30*time.Second).Return(true, nil)
	mockRepo.On("ReleaseLock", mock.Anything, toUserID).Return(nil)

	// Create source cart with items
	sourceCart := testutils.CreateTestCart(fromUserID, time.Hour)
	item := testutils.CreateTestCartItem("prod1", "Product 1", 10.99, 2)
	sourceCart.AddItem(item, 10, 5)

	// Create empty target cart
	targetCart := testutils.CreateTestCart(toUserID, time.Hour)

	mockRepo.On("GetCart", mock.Anything, fromUserID).Return(sourceCart, nil)
	mockRepo.On("GetCart", mock.Anything, toUserID).Return(targetCart, nil)
	mockRepo.On("SaveCart", mock.Anything, mock.AnythingOfType("*models.Cart")).Return(nil)
	mockRepo.On("DeleteCart", mock.Anything, fromUserID).Return(nil)

	// Execute
	result, err := cartService.TransferCart(context.Background(), fromUserID, toUserID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, toUserID, result.UserID)
	assert.True(t, result.HasItem("prod1"))

	mockRepo.AssertExpectations(t)
}
