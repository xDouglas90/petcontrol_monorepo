package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRequestContext_GeneratesCorrelationIDWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestContext())
	router.GET("/ping", func(c *gin.Context) {
		correlationID := GetCorrelationID(c)
		require.NotEmpty(t, correlationID)
		JSONData(c, http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.NotEmpty(t, res.Header().Get("X-Correlation-ID"))
}

func TestRequestContext_UsesProvidedCorrelationID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestContext())
	router.GET("/ping", func(c *gin.Context) {
		JSONError(c, http.StatusForbidden, "forbidden", "blocked")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Correlation-ID", "corr-123")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Equal(t, "corr-123", res.Header().Get("X-Correlation-ID"))
	require.Contains(t, res.Body.String(), "corr-123")
	require.Contains(t, res.Body.String(), "\"code\":\"forbidden\"")
}
