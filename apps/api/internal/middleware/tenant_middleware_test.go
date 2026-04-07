package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
)

func TestTenantMiddleware_MissingCompanyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Auth("secret"), Tenant())
	router.GET("/tenant", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	token, err := appjwt.GenerateToken("secret", time.Hour, appjwt.Claims{
		UserID: "user-1",
		Role:   "admin",
		Kind:   "owner",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/tenant", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestTenantMiddleware_InvalidCompanyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Auth("secret"), Tenant())
	router.GET("/tenant", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	token, err := appjwt.GenerateToken("secret", time.Hour, appjwt.Claims{
		UserID:    "user-1",
		CompanyID: "invalid-uuid",
		Role:      "admin",
		Kind:      "owner",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/tenant", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestTenantMiddleware_ValidCompanyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Auth("secret"), Tenant())
	router.GET("/tenant", func(c *gin.Context) {
		companyID, ok := GetCompanyID(c)
		require.True(t, ok)
		require.True(t, companyID.Valid)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	token, err := appjwt.GenerateToken("secret", time.Hour, appjwt.Claims{
		UserID:    "user-1",
		CompanyID: "5c5947bb-a2a1-4d4a-aeb4-d30c8534b17d",
		Role:      "admin",
		Kind:      "owner",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/tenant", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
}
