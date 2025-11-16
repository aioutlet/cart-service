package secrets

import (
	"context"
	"fmt"
	"sync"

	dapr "github.com/dapr/go-sdk/client"
	"go.uber.org/zap"
)

// DaprSecretManager manages secrets retrieval from Dapr Secret Store
type DaprSecretManager struct {
	client        dapr.Client
	secretStore   string
	logger        *zap.Logger
	jwtSecret     string
	jwtSecretOnce sync.Once
}

// NewDaprSecretManager creates a new Dapr secret manager
func NewDaprSecretManager(client dapr.Client, secretStore string, logger *zap.Logger) *DaprSecretManager {
	return &DaprSecretManager{
		client:      client,
		secretStore: secretStore,
		logger:      logger,
	}
}

// GetSecret retrieves a secret from Dapr Secret Store
func (m *DaprSecretManager) GetSecret(ctx context.Context, key string) (string, error) {
	secrets, err := m.client.GetSecret(ctx, m.secretStore, key, nil)
	if err != nil {
		m.logger.Error("Failed to retrieve secret from Dapr",
			zap.String("key", key),
			zap.String("store", m.secretStore),
			zap.Error(err))
		return "", fmt.Errorf("failed to retrieve secret '%s': %w", key, err)
	}

	value, ok := secrets[key]
	if !ok {
		m.logger.Error("Secret key not found in response",
			zap.String("key", key),
			zap.String("store", m.secretStore))
		return "", fmt.Errorf("secret '%s' not found in store '%s'", key, m.secretStore)
	}

	return value, nil
}

// GetJWTSecret retrieves JWT secret with lazy loading and caching
func (m *DaprSecretManager) GetJWTSecret(ctx context.Context) (string, error) {
	var err error
	m.jwtSecretOnce.Do(func() {
		m.logger.Info("Loading JWT secret from Dapr Secret Store")
		m.jwtSecret, err = m.GetSecret(ctx, "JWT_SECRET")
		if err != nil {
			m.logger.Error("Failed to load JWT secret", zap.Error(err))
			return
		}
		m.logger.Info("JWT secret loaded successfully from Dapr")
	})

	if err != nil {
		return "", err
	}

	if m.jwtSecret == "" {
		return "", fmt.Errorf("JWT secret not loaded")
	}

	return m.jwtSecret, nil
}
