package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	correlationIDHeader  = "X-Correlation-ID"
	correlationIDContext = "correlation_id"
)

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader(correlationIDHeader)
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		c.Set(correlationIDContext, correlationID)
		c.Writer.Header().Set(correlationIDHeader, correlationID)
		c.Next()
	}
}

func GetCorrelationID(c *gin.Context) string {
	if value, ok := c.Get(correlationIDContext); ok {
		if correlationID, castOK := value.(string); castOK {
			return correlationID
		}
	}
	return ""
}

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		fields := []any{
			"correlation_id", GetCorrelationID(c),
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		}

		if claims, ok := GetClaims(c); ok {
			fields = append(fields,
				"user_id", claims.UserID,
				"company_id", claims.CompanyID,
			)
		}

		if c.Writer.Status() >= 500 {
			logger.Error("api request completed", fields...)
			return
		}
		logger.Info("api request completed", fields...)
	}
}
