package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	dapr "github.com/dapr/go-sdk/client"
	"go.uber.org/zap"
)

// InventoryClient interface for inventory service communication
type InventoryClient interface {
	CheckAvailability(ctx context.Context, sku string, quantity int) (bool, error)
	GetAvailableQuantity(ctx context.Context, sku string) (int, error)
	ReserveStock(ctx context.Context, sku string, quantity int) error
	ReleaseStock(ctx context.Context, sku string, quantity int) error
}

// inventoryClient implements InventoryClient interface using Dapr service invocation
type inventoryClient struct {
	daprClient dapr.Client
	logger     *zap.Logger
}

// NewInventoryClient creates a new inventory client using Dapr SDK
func NewInventoryClient(daprClient dapr.Client, logger *zap.Logger) InventoryClient {
	return &inventoryClient{
		daprClient: daprClient,
		logger:     logger,
	}
}

// CheckAvailability checks if a SKU has sufficient stock using Dapr service invocation
func (c *inventoryClient) CheckAvailability(ctx context.Context, sku string, quantity int) (bool, error) {
	// URL encode the SKU to handle special characters like & in H&M
	encodedSKU := url.PathEscape(sku)
	methodPath := fmt.Sprintf("/api/v1/inventory/%s/check?quantity=%d", encodedSKU, quantity)
	
	// Invoke inventory-service via Dapr
	resp, err := c.daprClient.InvokeMethod(ctx, "inventory-service", methodPath, "GET")
	if err != nil {
		c.logger.Error("Failed to invoke inventory service via Dapr", 
			zap.String("sku", sku),
			zap.Int("quantity", quantity),
			zap.Error(err))
		return false, fmt.Errorf("failed to invoke inventory service: %w", err)
	}

	// Check for empty response (not found)
	if len(resp) == 0 {
		return false, nil
	}

	var response struct {
		Success   bool `json:"success"`
		Available bool `json:"available"`
		Message   string `json:"message"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		c.logger.Error("Failed to unmarshal inventory response", 
			zap.String("sku", sku), 
			zap.Error(err))
		return false, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Success && response.Available, nil
}

// GetAvailableQuantity gets the available quantity for a SKU using Dapr service invocation
func (c *inventoryClient) GetAvailableQuantity(ctx context.Context, sku string) (int, error) {
	// URL encode the SKU to handle special characters
	encodedSKU := url.PathEscape(sku)
	methodPath := fmt.Sprintf("/api/v1/inventory/%s", encodedSKU)
	
	// Invoke inventory-service via Dapr
	resp, err := c.daprClient.InvokeMethod(ctx, "inventory-service", methodPath, "GET")
	if err != nil {
		c.logger.Error("Failed to invoke inventory service for quantity via Dapr", 
			zap.String("sku", sku),
			zap.Error(err))
		return 0, fmt.Errorf("failed to invoke inventory service: %w", err)
	}

	// Check for empty response (not found)
	if len(resp) == 0 {
		return 0, nil
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			ProductID string `json:"productId"`
			Quantity  int    `json:"quantity"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		c.logger.Error("Failed to unmarshal inventory quantity response", 
			zap.String("sku", sku), 
			zap.Error(err))
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return 0, fmt.Errorf("inventory service returned error: %s", response.Message)
	}

	return response.Data.Quantity, nil
}

// ReserveStock reserves stock for a SKU (used during checkout) using Dapr service invocation
func (c *inventoryClient) ReserveStock(ctx context.Context, sku string, quantity int) error {
	// URL encode the SKU to handle special characters
	encodedSKU := url.PathEscape(sku)
	methodPath := fmt.Sprintf("/api/v1/inventory/%s/reserve", encodedSKU)
	
	requestBody := map[string]int{
		"quantity": quantity,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Invoke inventory-service via Dapr with POST method
	content := &dapr.DataContent{
		Data:        bodyBytes,
		ContentType: "application/json",
	}
	_, err = c.daprClient.InvokeMethodWithContent(ctx, "inventory-service", methodPath, "POST", content)
	if err != nil {
		c.logger.Error("Failed to reserve stock via Dapr", 
			zap.String("sku", sku),
			zap.Int("quantity", quantity),
			zap.Error(err))
		return fmt.Errorf("failed to reserve stock: %w", err)
	}

	return nil
}

// ReleaseStock releases reserved stock for a SKU using Dapr service invocation
func (c *inventoryClient) ReleaseStock(ctx context.Context, sku string, quantity int) error {
	// URL encode the SKU to handle special characters
	encodedSKU := url.PathEscape(sku)
	methodPath := fmt.Sprintf("/api/v1/inventory/%s/release", encodedSKU)
	
	requestBody := map[string]int{
		"quantity": quantity,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Invoke inventory-service via Dapr with POST method
	content := &dapr.DataContent{
		Data:        bodyBytes,
		ContentType: "application/json",
	}
	_, err = c.daprClient.InvokeMethodWithContent(ctx, "inventory-service", methodPath, "POST", content)
	if err != nil {
		c.logger.Error("Failed to release stock via Dapr", 
			zap.String("sku", sku),
			zap.Int("quantity", quantity),
			zap.Error(err))
		return fmt.Errorf("failed to release stock: %w", err)
	}

	return nil
}
