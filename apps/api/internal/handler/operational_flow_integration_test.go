package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/handler"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

func TestOperationalFlow_ClientPetServiceScheduleAndDeactivationImpact(t *testing.T) {
	ctx := context.Background()
	pool := setupScheduleIntegrationPool(t)
	queries := sqlc.New(pool)

	tenant := mustCreateTenantFixture(t, ctx, pool, "operational-flow")
	crmModuleID := mustCreateClientModule(t, ctx, pool)
	scheduleModuleID := mustCreateScheduleModule(t, ctx, pool)
	mustAttachClientModule(t, ctx, pool, tenant.companyID, crmModuleID)
	mustAttachScheduleModule(t, ctx, pool, tenant.companyID, scheduleModuleID)

	router := setupOperationalFlowRouterForTenant(pool, queries, tenant)

	clientRes := performJSONRequest(t, router, http.MethodPost, "/api/v1/clients", map[string]any{
		"full_name":       "Julia Martins",
		"short_name":      "Julia",
		"gender_identity": "woman_cisgender",
		"marital_status":  "single",
		"birth_date":      "1990-03-12",
		"cpf":             "123.456.789-88",
		"email":           "julia.martins@petcontrol.local",
		"phone":           "+551130000001",
		"cellphone":       "+5511999990088",
		"has_whatsapp":    true,
		"client_since":    "2026-04-10",
		"notes":           "Fluxo operacional completo",
	})
	require.Equal(t, http.StatusCreated, clientRes.Code)
	clientID := responseDataString(t, clientRes, "id")

	petRes := performJSONRequest(t, router, http.MethodPost, "/api/v1/pets", map[string]any{
		"owner_id":    clientID,
		"name":        "Nina",
		"size":        "small",
		"kind":        "dog",
		"temperament": "loving",
		"birth_date":  "2021-08-20",
		"notes":       "Pet do fluxo operacional",
	})
	require.Equal(t, http.StatusCreated, petRes.Code)
	require.Contains(t, petRes.Body.String(), "Julia Martins")
	petID := responseDataString(t, petRes, "id")

	serviceRes := performJSONRequest(t, router, http.MethodPost, "/api/v1/services", map[string]any{
		"type_name":     "Banho",
		"title":         "Banho completo operacional",
		"description":   "Banho com secagem para o fluxo completo",
		"price":         "89.90",
		"discount_rate": "0.00",
	})
	require.Equal(t, http.StatusCreated, serviceRes.Code)
	serviceID := responseDataString(t, serviceRes, "id")

	scheduledAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	scheduleRes := performJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", map[string]any{
		"client_id":    clientID,
		"pet_id":       petID,
		"service_ids":  []string{serviceID},
		"scheduled_at": scheduledAt.Format(time.RFC3339),
		"notes":        "Fluxo cliente -> pet -> serviço -> agendamento",
	})
	require.Equal(t, http.StatusCreated, scheduleRes.Code)
	require.Contains(t, scheduleRes.Body.String(), "Julia Martins")
	require.Contains(t, scheduleRes.Body.String(), "Nina")
	require.Contains(t, scheduleRes.Body.String(), "Banho completo operacional")

	listRes := performJSONRequest(t, router, http.MethodGet, "/api/v1/schedules", nil)
	require.Equal(t, http.StatusOK, listRes.Code)
	require.Contains(t, listRes.Body.String(), "Julia Martins")
	require.Contains(t, listRes.Body.String(), "Nina")
	require.Contains(t, listRes.Body.String(), "Banho completo operacional")

	deleteServiceRes := performJSONRequest(t, router, http.MethodDelete, "/api/v1/services/"+serviceID, nil)
	require.Equal(t, http.StatusNoContent, deleteServiceRes.Code)

	scheduleWithInactiveServiceRes := performJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", map[string]any{
		"client_id":    clientID,
		"pet_id":       petID,
		"service_ids":  []string{serviceID},
		"scheduled_at": scheduledAt.Add(2 * time.Hour).Format(time.RFC3339),
		"notes":        "Serviço desativado não deve entrar no agendamento",
	})
	require.Equal(t, http.StatusUnprocessableEntity, scheduleWithInactiveServiceRes.Code)

	deletePetRes := performJSONRequest(t, router, http.MethodDelete, "/api/v1/pets/"+petID, nil)
	require.Equal(t, http.StatusNoContent, deletePetRes.Code)

	scheduleWithDeletedPetRes := performJSONRequest(t, router, http.MethodPost, "/api/v1/schedules", map[string]any{
		"client_id":    clientID,
		"pet_id":       petID,
		"scheduled_at": scheduledAt.Add(3 * time.Hour).Format(time.RFC3339),
		"notes":        "Pet removido não deve entrar no agendamento",
	})
	require.Equal(t, http.StatusUnprocessableEntity, scheduleWithDeletedPetRes.Code)
}

func setupOperationalFlowRouterForTenant(pool *pgxpool.Pool, queries *sqlc.Queries, tenant integrationTenantFixture) *gin.Engine {
	gin.SetMode(gin.TestMode)

	clientHandler := handler.NewClientHandler(service.NewClientService(pool, queries))
	petHandler := handler.NewPetHandler(service.NewPetService(queries))
	serviceHandler := handler.NewServiceHandler(service.NewServiceService(pool, queries))
	scheduleHandler := handler.NewScheduleHandler(service.NewScheduleService(pool, queries))

	router := gin.New()
	router.Use(middleware.RequestContext())
	router.Use(func(c *gin.Context) {
		c.Set("company_id", tenant.companyID)
		c.Set("auth_claims", appjwt.Claims{
			UserID:    tenant.userID.String(),
			CompanyID: tenant.companyID.String(),
			Role:      "admin",
			Kind:      "owner",
		})
		c.Next()
	})
	router.Use(middleware.Audit(queries, nil))

	clients := router.Group("/api/v1/clients")
	clients.Use(middleware.RequireModule(queries, "CRM"))
	clients.GET("", clientHandler.List)
	clients.POST("", clientHandler.Create)
	clients.GET("/:id", clientHandler.GetByID)
	clients.PUT("/:id", clientHandler.Update)
	clients.DELETE("/:id", clientHandler.Delete)

	pets := router.Group("/api/v1/pets")
	pets.Use(middleware.RequireModule(queries, "CRM"))
	pets.GET("", petHandler.List)
	pets.POST("", petHandler.Create)
	pets.GET("/:id", petHandler.GetByID)
	pets.PUT("/:id", petHandler.Update)
	pets.DELETE("/:id", petHandler.Delete)

	services := router.Group("/api/v1/services")
	services.Use(middleware.RequireModule(queries, "SCH"))
	services.GET("", serviceHandler.List)
	services.POST("", serviceHandler.Create)
	services.GET("/:id", serviceHandler.GetByID)
	services.PUT("/:id", serviceHandler.Update)
	services.DELETE("/:id", serviceHandler.Delete)

	schedules := router.Group("/api/v1/schedules")
	schedules.Use(middleware.RequireModule(queries, "SCH"))
	schedules.GET("", scheduleHandler.List)
	schedules.POST("", scheduleHandler.Create)
	schedules.GET("/:id", scheduleHandler.GetByID)
	schedules.GET("/:id/history", scheduleHandler.History)
	schedules.PUT("/:id", scheduleHandler.Update)
	schedules.DELETE("/:id", scheduleHandler.Delete)

	return router
}

func performJSONRequest(t *testing.T, router *gin.Engine, method string, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()

	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func responseDataString(t *testing.T, res *httptest.ResponseRecorder, field string) string {
	t.Helper()

	var body struct {
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
	value, ok := body.Data[field].(string)
	require.True(t, ok, "expected response data field %q to be a string", field)
	require.NotEmpty(t, value)
	return value
}
