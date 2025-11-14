package clients

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aioutlet/cart-service/internal/models"
	dapr "github.com/dapr/go-sdk/client"
	"go.uber.org/zap"
)

// ProductClient interface for product service communication
type ProductClient interface {
	GetProduct(ctx context.Context, productID string) (*models.ProductInfo, error)
	GetProducts(ctx context.Context, productIDs []string) ([]models.ProductInfo, error)
}

// productClient implements ProductClient interface using Dapr service invocation
type productClient struct {
	daprClient dapr.Client
	logger     *zap.Logger
}

// NewProductClient creates a new product client using Dapr SDK
func NewProductClient(daprClient dapr.Client, logger *zap.Logger) ProductClient {
	return &productClient{
		daprClient: daprClient,
		logger:     logger,
	}
}

// GetProduct retrieves product information by ID using Dapr service invocation
func (c *productClient) GetProduct(ctx context.Context, productID string) (*models.ProductInfo, error) {
	methodPath := fmt.Sprintf("/api/products/%s", productID)
	
	// Invoke product-service via Dapr
	resp, err := c.daprClient.InvokeMethod(ctx, "product-service", methodPath, "GET")
	if err != nil {
		c.logger.Error("Failed to invoke product service via Dapr", 
			zap.String("productID", productID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to invoke product service: %w", err)
	}

	// Check if product not found (empty response or error structure)
	if len(resp) == 0 {
		return nil, models.ErrProductNotFound
	}

	// Product service returns the product data directly
	var productInfo models.ProductInfo
	if err := json.Unmarshal(resp, &productInfo); err != nil {
		c.logger.Error("Failed to unmarshal product response", 
			zap.String("productID", productID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Validate we got a valid product
	if productInfo.ID == "" {
		return nil, models.ErrProductNotFound
	}

	return &productInfo, nil
}

// GetProducts retrieves multiple products by IDs using Dapr service invocation
func (c *productClient) GetProducts(ctx context.Context, productIDs []string) ([]models.ProductInfo, error) {
	methodPath := "/api/products/batch"
	
	requestBody := map[string][]string{
		"productIds": productIDs,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Invoke product-service via Dapr with POST method
	content := &dapr.DataContent{
		Data:        bodyBytes,
		ContentType: "application/json",
	}
	resp, err := c.daprClient.InvokeMethodWithContent(ctx, "product-service", methodPath, "POST", content)
	if err != nil {
		c.logger.Error("Failed to invoke product service batch via Dapr", zap.Error(err))
		return nil, fmt.Errorf("failed to invoke product service: %w", err)
	}

	var response struct {
		Success bool                  `json:"success"`
		Data    []models.ProductInfo  `json:"data"`
		Message string                `json:"message"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		c.logger.Error("Failed to unmarshal batch products response", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("product service returned error: %s", response.Message)
	}

	return response.Data, nil
}
