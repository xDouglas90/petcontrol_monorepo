package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	docs "github.com/xdouglas90/petcontrol_monorepo/docs"
)

func TestConfigureSwaggerInfo(t *testing.T) {
	configureSwaggerInfo()
	require.Equal(t, "PetControl API", docs.SwaggerInfo.Title)
	require.Equal(t, "/api/v1", docs.SwaggerInfo.BasePath)
}

func TestRegisterSwaggerRoute_ServesSwaggerDoc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerSwaggerRoute(router)

	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Header().Get("Content-Type"), "application/json")
	jsonBody := res.Body.String()
	require.True(t, strings.Contains(jsonBody, "/auth/login") || strings.Contains(jsonBody, "/api/v1/auth/login"))
	require.True(t, strings.Contains(jsonBody, "/schedules") || strings.Contains(jsonBody, "/api/v1/schedules"))
}
