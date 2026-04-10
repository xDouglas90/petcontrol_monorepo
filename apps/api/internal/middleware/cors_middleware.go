package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	accessControlAllowOriginHeader      = "Access-Control-Allow-Origin"
	accessControlAllowMethodsHeader     = "Access-Control-Allow-Methods"
	accessControlAllowHeadersHeader     = "Access-Control-Allow-Headers"
	accessControlAllowCredentialsHeader = "Access-Control-Allow-Credentials"
	accessControlExposeHeadersHeader    = "Access-Control-Expose-Headers"
	varyHeader                          = "Vary"
)

var defaultAllowedMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodOptions,
}

var defaultAllowedHeaders = []string{
	"Authorization",
	"Content-Type",
	"X-Correlation-ID",
}

var defaultExposedHeaders = []string{
	"X-Correlation-ID",
}

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		allowed[trimmed] = struct{}{}
	}

	allowedMethods := strings.Join(defaultAllowedMethods, ", ")
	allowedHeaders := strings.Join(defaultAllowedHeaders, ", ")
	exposedHeaders := strings.Join(defaultExposedHeaders, ", ")

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin == "" {
			c.Next()
			return
		}

		c.Writer.Header().Add(varyHeader, "Origin")
		c.Writer.Header().Add(varyHeader, "Access-Control-Request-Method")
		c.Writer.Header().Add(varyHeader, "Access-Control-Request-Headers")

		if _, ok := allowed[origin]; !ok {
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
			return
		}

		headers := c.Writer.Header()
		headers.Set(accessControlAllowOriginHeader, origin)
		headers.Set(accessControlAllowCredentialsHeader, "true")
		headers.Set(accessControlExposeHeadersHeader, exposedHeaders)

		if c.Request.Method == http.MethodOptions {
			headers.Set(accessControlAllowMethodsHeader, allowedMethods)
			requestHeaders := strings.TrimSpace(c.GetHeader("Access-Control-Request-Headers"))
			if requestHeaders == "" {
				headers.Set(accessControlAllowHeadersHeader, allowedHeaders)
			} else {
				headers.Set(accessControlAllowHeadersHeader, requestHeaders)
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
