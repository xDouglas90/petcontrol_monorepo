package middleware

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func Recovery(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered",
			"correlation_id", GetCorrelationID(c),
			"method", c.Request.Method,
			"path", c.FullPath(),
			"panic", recovered,
		)

		JSONError(c, 500, "internal_error", "internal server error")
	})
}
