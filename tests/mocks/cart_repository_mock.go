package mocks

import (
	"context"
	"time"

	"github.com/aioutlet/cart-service/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockCartRepository is a mock implementation of CartRepository
type MockCartRepository struct {
	mock.Mock
}

func (m *MockCartRepository) GetCart(ctx context.Context, userID string) (*models.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (m *MockCartRepository) SaveCart(ctx context.Context, cart *models.Cart) error {
	args := m.Called(ctx, cart)
	return args.Error(0)
}

func (m *MockCartRepository) DeleteCart(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCartRepository) SetCartTTL(ctx context.Context, userID string, ttl time.Duration) error {
	args := m.Called(ctx, userID, ttl)
	return args.Error(0)
}

func (m *MockCartRepository) CartExists(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCartRepository) AcquireLock(ctx context.Context, userID string, ttl time.Duration) (bool, error) {
	args := m.Called(ctx, userID, ttl)
	return args.Bool(0), args.Error(1)
}

func (m *MockCartRepository) ReleaseLock(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCartRepository) GetAllCartKeys(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCartRepository) GetCartTTL(ctx context.Context, userID string) (time.Duration, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(time.Duration), args.Error(1)
}
