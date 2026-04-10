package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCORS_AllowsConfiguredOriginPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS([]string{"http://localhost:5173"}))
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		JSONData(c, http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "authorization, content-type")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
	require.Equal(t, "http://localhost:5173", res.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "authorization, content-type", res.Header().Get("Access-Control-Allow-Headers"))
	require.Equal(t, "true", res.Header().Get("Access-Control-Allow-Credentials"))
	require.Contains(t, res.Header().Values("Vary"), "Origin")
}

func TestCORS_AllowsConfiguredOriginOnRegularRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS([]string{"http://localhost:5173"}))
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		JSONData(c, http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "http://localhost:5173", res.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "true", res.Header().Get("Access-Control-Allow-Credentials"))
	require.Equal(t, "X-Correlation-ID", res.Header().Get("Access-Control-Expose-Headers"))
}

func TestCORS_BlocksUnknownOriginPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS([]string{"http://localhost:5173"}))
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		JSONData(c, http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "http://malicious.local")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Empty(t, res.Header().Get("Access-Control-Allow-Origin"))
}
