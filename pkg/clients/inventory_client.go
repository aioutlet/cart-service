package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// InventoryClient interface for inventory service communication
type InventoryClient interface {
	CheckAvailability(ctx context.Context, productID string, quantity int) (bool, error)
	GetAvailableQuantity(ctx context.Context, productID string) (int, error)
	ReserveStock(ctx context.Context, productID string, quantity int) error
	ReleaseStock(ctx context.Context, productID string, quantity int) error
}

// inventoryClient implements InventoryClient interface
type inventoryClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewInventoryClient creates a new inventory client
func NewInventoryClient(baseURL string, logger *zap.Logger) InventoryClient {
	return &inventoryClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// CheckAvailability checks if a product has sufficient stock
func (c *inventoryClient) CheckAvailability(ctx context.Context, productID string, quantity int) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/inventory/%s/check?quantity=%d", c.baseURL, productID, quantity)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Add correlation ID if present in context
	if correlationID := ctx.Value("correlationID"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			req.Header.Set("X-Correlation-ID", id)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to call inventory service", 
			zap.String("productID", productID),
			zap.Int("quantity", quantity),
			zap.Error(err))
		return false, fmt.Errorf("failed to call inventory service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Inventory service returned error", 
			zap.String("productID", productID),
			zap.Int("quantity", quantity),
			zap.Int("statusCode", resp.StatusCode))
		return false, fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}

	var response struct {
		Success   bool `json:"success"`
		Available bool `json:"available"`
		Message   string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Error("Failed to decode inventory response", 
			zap.String("productID", productID), 
			zap.Error(err))
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Success && response.Available, nil
}

// GetAvailableQuantity gets the available quantity for a product
func (c *inventoryClient) GetAvailableQuantity(ctx context.Context, productID string) (int, error) {
	url := fmt.Sprintf("%s/api/v1/inventory/%s", c.baseURL, productID)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Add correlation ID if present in context
	if correlationID := ctx.Value("correlationID"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			req.Header.Set("X-Correlation-ID", id)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to call inventory service for quantity", 
			zap.String("productID", productID),
			zap.Error(err))
		return 0, fmt.Errorf("failed to call inventory service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, nil
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Inventory service returned error for quantity check", 
			zap.String("productID", productID),
			zap.Int("statusCode", resp.StatusCode))
		return 0, fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			ProductID string `json:"productId"`
			Quantity  int    `json:"quantity"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Error("Failed to decode inventory quantity response", 
			zap.String("productID", productID), 
			zap.Error(err))
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return 0, fmt.Errorf("inventory service returned error: %s", response.Message)
	}

	return response.Data.Quantity, nil
}

// ReserveStock reserves stock for a product (used during checkout)
func (c *inventoryClient) ReserveStock(ctx context.Context, productID string, quantity int) error {
	url := fmt.Sprintf("%s/api/v1/inventory/%s/reserve", c.baseURL, productID)
	
	requestBody := map[string]int{
		"quantity": quantity,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Add correlation ID if present in context
	if correlationID := ctx.Value("correlationID"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			req.Header.Set("X-Correlation-ID", id)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to reserve stock", 
			zap.String("productID", productID),
			zap.Int("quantity", quantity),
			zap.Error(err))
		return fmt.Errorf("failed to reserve stock: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Failed to reserve stock", 
			zap.String("productID", productID),
			zap.Int("quantity", quantity),
			zap.Int("statusCode", resp.StatusCode))
		return fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}

	return nil
}

// ReleaseStock releases reserved stock for a product
func (c *inventoryClient) ReleaseStock(ctx context.Context, productID string, quantity int) error {
	url := fmt.Sprintf("%s/api/v1/inventory/%s/release", c.baseURL, productID)
	
	requestBody := map[string]int{
		"quantity": quantity,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Add correlation ID if present in context
	if correlationID := ctx.Value("correlationID"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			req.Header.Set("X-Correlation-ID", id)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to release stock", 
			zap.String("productID", productID),
			zap.Int("quantity", quantity),
			zap.Error(err))
		return fmt.Errorf("failed to release stock: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Failed to release stock", 
			zap.String("productID", productID),
			zap.Int("quantity", quantity),
			zap.Int("statusCode", resp.StatusCode))
		return fmt.Errorf("inventory service returned status %d", resp.StatusCode)
	}

	return nil
}
