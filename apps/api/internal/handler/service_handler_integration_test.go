package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/handler"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

func TestServiceEndpoints_ListAndCreateAreTenantScoped(t *testing.T) {
	ctx := context.Background()
	pool := setupScheduleIntegrationPool(t)
	queries := sqlc.New(pool)

	tenantA := mustCreateTenantFixture(t, ctx, pool, "tenant-service-a")
	tenantB := mustCreateTenantFixture(t, ctx, pool, "tenant-service-b")
	moduleID := mustCreateScheduleModule(t, ctx, pool)
	mustAttachScheduleModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachScheduleModule(t, ctx, pool, tenantB.companyID, moduleID)

	serviceB := mustInsertServiceCatalogRecord(t, ctx, pool, tenantB.companyID, "Tosa luxo")
	router := setupServiceRouterForTenant(pool, queries, tenantA)

	body, err := json.Marshal(map[string]any{
		"type_name":     "Banho",
		"title":         "Banho completo",
		"description":   "Banho com secagem",
		"price":         "89.90",
		"discount_rate": "0.00",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/services", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.Contains(t, res.Body.String(), "Banho completo")

	req = httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "Banho completo")
	require.NotContains(t, res.Body.String(), serviceB.String())
	require.NotContains(t, res.Body.String(), "Tosa luxo")
}

func setupServiceRouterForTenant(pool *pgxpool.Pool, queries *sqlc.Queries, tenant integrationTenantFixture) *gin.Engine {
	gin.SetMode(gin.TestMode)

	serviceHandler := handler.NewServiceHandler(service.NewServiceService(pool, queries))

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

	services := router.Group("/api/v1/services")
	services.Use(middleware.RequireModule(queries, "SCH"))
	services.GET("", serviceHandler.List)
	services.POST("", serviceHandler.Create)
	services.GET("/:id", serviceHandler.GetByID)
	services.PUT("/:id", serviceHandler.Update)
	services.DELETE("/:id", serviceHandler.Delete)

	return router
}

func mustInsertServiceCatalogRecord(t *testing.T, ctx context.Context, pool *pgxpool.Pool, companyID pgtype.UUID, title string) pgtype.UUID {
	t.Helper()

	var typeID pgtype.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO service_types (name)
		VALUES ('Banho')
		RETURNING id
	`).Scan(&typeID)
	require.NoError(t, err)

	var serviceID pgtype.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO services (type_id, title, description, price, discount_rate, is_active)
		VALUES ($1, $2, 'Descrição', 50.00, 0.00, TRUE)
		RETURNING id
	`, typeID, title).Scan(&serviceID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO company_services (company_id, service_id, is_active)
		VALUES ($1, $2, TRUE)
	`, companyID, serviceID)
	require.NoError(t, err)

	return serviceID
}
