package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Logger middleware provides structured logging for HTTP requests
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		correlationID := ""
		if param.Keys != nil {
			if id, exists := param.Keys["correlationID"]; exists {
				if idStr, ok := id.(string); ok {
					correlationID = idStr
				}
			}
		}

		userID := ""
		if param.Keys != nil {
			if id, exists := param.Keys["userID"]; exists {
				if idStr, ok := id.(string); ok {
					userID = idStr
				}
			}
		}

		// Extract trace and span IDs
		traceID := ""
		spanID := ""
		if span := trace.SpanFromContext(param.Request.Context()); span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
			spanID = span.SpanContext().SpanID().String()
		}

		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("query", param.Request.URL.RawQuery),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("clientIP", param.ClientIP),
			zap.String("userAgent", param.Request.UserAgent()),
			zap.String("correlationID", correlationID),
			zap.String("traceID", traceID),
			zap.String("spanID", spanID),
			zap.String("userID", userID),
			zap.Time("timestamp", param.TimeStamp),
		)

		return ""
	})
}

// ErrorLogger middleware logs errors with correlation, trace, and span IDs
func ErrorLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log any errors that occurred
		if len(c.Errors) > 0 {
			correlationID, _ := c.Get("correlationID")
			userID, _ := c.Get("userID")

			// Extract trace and span IDs
			traceID := ""
			spanID := ""
			if span := trace.SpanFromContext(c.Request.Context()); span.SpanContext().IsValid() {
				traceID = span.SpanContext().TraceID().String()
				spanID = span.SpanContext().SpanID().String()
			}

			for _, err := range c.Errors {
				logger.Error("Request error",
					zap.Error(err.Err),
					zap.String("type", fmt.Sprintf("%d", err.Type)),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("correlationID", correlationID.(string)),
					zap.String("traceID", traceID),
					zap.String("spanID", spanID),
					zap.String("userID", userID.(string)),
				)
			}
		}
	}
}
