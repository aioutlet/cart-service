package mocks

import (
	"context"

	"github.com/aioutlet/cart-service/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockProductClient is a mock implementation of ProductClient
type MockProductClient struct {
	mock.Mock
}

func (m *MockProductClient) GetProduct(ctx context.Context, productID string) (*models.ProductInfo, error) {
	args := m.Called(ctx, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProductInfo), args.Error(1)
}

func (m *MockProductClient) GetProducts(ctx context.Context, productIDs []string) ([]models.ProductInfo, error) {
	args := m.Called(ctx, productIDs)
	return args.Get(0).([]models.ProductInfo), args.Error(1)
}

// MockInventoryClient is a mock implementation of InventoryClient
type MockInventoryClient struct {
	mock.Mock
}

func (m *MockInventoryClient) CheckAvailability(ctx context.Context, productID string, quantity int) (bool, error) {
	args := m.Called(ctx, productID, quantity)
	return args.Bool(0), args.Error(1)
}

func (m *MockInventoryClient) GetAvailableQuantity(ctx context.Context, productID string) (int, error) {
	args := m.Called(ctx, productID)
	return args.Int(0), args.Error(1)
}

func (m *MockInventoryClient) ReserveStock(ctx context.Context, productID string, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}

func (m *MockInventoryClient) ReleaseStock(ctx context.Context, productID string, quantity int) error {
	args := m.Called(ctx, productID, quantity)
	return args.Error(0)
}
