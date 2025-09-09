package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CorrelationID middleware adds correlation ID to requests for distributed tracing
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get correlation ID from header or generate a new one
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Set correlation ID in context
		c.Set("correlationID", correlationID)
		
		// Add correlation ID to response header
		c.Header("X-Correlation-ID", correlationID)

		// If we have an active span, add correlation ID as an attribute
		if span := trace.SpanFromContext(c.Request.Context()); span.IsRecording() {
			span.SetAttributes(attribute.String("correlation.id", correlationID))
		}

		c.Next()
	}
}

// GetCorrelationID extracts correlation ID from Gin context
func GetCorrelationID(c *gin.Context) string {
	if correlationID, exists := c.Get("correlationID"); exists {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return ""
}

// GetTraceID extracts trace ID from the current span
func GetTraceID(c *gin.Context) string {
	if span := trace.SpanFromContext(c.Request.Context()); span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID extracts span ID from the current span
func GetSpanID(c *gin.Context) string {
	if span := trace.SpanFromContext(c.Request.Context()); span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}
