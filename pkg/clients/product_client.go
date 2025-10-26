package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aioutlet/cart-service/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ProductClient interface for product service communication
type ProductClient interface {
	GetProduct(ctx context.Context, productID string) (*models.ProductInfo, error)
	GetProducts(ctx context.Context, productIDs []string) ([]models.ProductInfo, error)
}

// productClient implements ProductClient interface
type productClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewProductClient creates a new product client
func NewProductClient(baseURL string, logger *zap.Logger) ProductClient {
	return &productClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetProduct retrieves product information by ID
func (c *productClient) GetProduct(ctx context.Context, productID string) (*models.ProductInfo, error) {
	tracer := otel.Tracer("cart-service")
	ctx, span := tracer.Start(ctx, "ProductClient.GetProduct")
	defer span.End()

	span.SetAttributes(
		attribute.String("product.id", productID),
		attribute.String("service", "product-service"),
		attribute.String("operation", "get_product"),
	)

	url := fmt.Sprintf("%s/api/products/%s", c.baseURL, productID)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "failed to create request"))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Inject trace context into the request headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Add correlation ID if present in context
	if correlationID := ctx.Value("correlationID"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			req.Header.Set("X-Correlation-ID", id)
		}
	}

	span.AddEvent("Sending HTTP request", trace.WithAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.url", url),
	))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "HTTP request failed"))
		c.logger.Error("Failed to call product service", 
			zap.String("productID", productID),
			zap.String("traceID", span.SpanContext().TraceID().String()),
			zap.String("spanID", span.SpanContext().SpanID().String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to call product service: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode == http.StatusNotFound {
		span.SetAttributes(attribute.String("error", "product not found"))
		return nil, models.ErrProductNotFound
	}

	if resp.StatusCode != http.StatusOK {
		span.RecordError(fmt.Errorf("product service returned status %d", resp.StatusCode))
		span.SetAttributes(attribute.String("error", "non-200 status code"))
		c.logger.Error("Product service returned error", 
			zap.String("productID", productID),
			zap.Int("statusCode", resp.StatusCode),
			zap.String("traceID", span.SpanContext().TraceID().String()),
			zap.String("spanID", span.SpanContext().SpanID().String()))
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	// Product service returns the product data directly
	var productInfo models.ProductInfo

	if err := json.NewDecoder(resp.Body).Decode(&productInfo); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "failed to decode response"))
		c.logger.Error("Failed to decode product response", 
			zap.String("productID", productID),
			zap.String("traceID", span.SpanContext().TraceID().String()),
			zap.String("spanID", span.SpanContext().SpanID().String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	span.SetAttributes(
		attribute.String("product.name", productInfo.Name),
		attribute.Float64("product.price", productInfo.Price),
		attribute.Bool("product.active", productInfo.IsActive),
	)
	span.AddEvent("Product information retrieved successfully")

	return &productInfo, nil
}

// GetProducts retrieves multiple products by IDs
func (c *productClient) GetProducts(ctx context.Context, productIDs []string) ([]models.ProductInfo, error) {
	url := fmt.Sprintf("%s/api/products/batch", c.baseURL)
	
	requestBody := map[string][]string{
		"productIds": productIDs,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
		c.logger.Error("Failed to call product service for batch", zap.Error(err))
		return nil, fmt.Errorf("failed to call product service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Product service returned error for batch request", 
			zap.Int("statusCode", resp.StatusCode))
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	var response struct {
		Success bool                  `json:"success"`
		Data    []models.ProductInfo  `json:"data"`
		Message string                `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Error("Failed to decode products response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("product service returned error: %s", response.Message)
	}

	return response.Data, nil
}
