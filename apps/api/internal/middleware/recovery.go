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
		matchedPath := c.FullPath()
		requestPath := c.Request.URL.Path
		path := matchedPath
		if path == "" {
			path = requestPath
		}

		logger.Error("panic recovered",
			"correlation_id", GetCorrelationID(c),
			"method", c.Request.Method,
			"path", path,
			"matched_path", matchedPath,
			"request_path", requestPath,
			"panic", recovered,
		)

		JSONError(c, 500, "internal_error", "internal server error")
	})
}
