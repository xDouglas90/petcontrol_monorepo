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
	moduleID := mustCreateServiceModule(t, ctx, pool)
	mustAttachScheduleModule(t, ctx, pool, tenantA.companyID, moduleID)
	mustAttachScheduleModule(t, ctx, pool, tenantB.companyID, moduleID)
	mustGrantServicePermissions(t, ctx, pool, tenantA.userID, service.PermissionServicesView, service.PermissionServicesCreate)

	serviceB := mustInsertServiceCatalogRecord(t, ctx, pool, tenantB.companyID, "Tosa luxo")
	router := setupServiceRouterForTenant(pool, queries, tenantA)

	body, err := json.Marshal(map[string]any{
		"type_name":     "Banho",
		"title":         "Banho completo",
		"description":   "Banho com secagem",
		"price":         "89.90",
		"discount_rate": "0.00",
		"sub_services": []map[string]any{
			serviceSubServicePayload("Banho médio", "Banho para pets médios", "89.90"),
		},
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

func TestServiceEndpoints_SystemCannotCreateWithViewOnlyPermission(t *testing.T) {
	ctx := context.Background()
	pool := setupScheduleIntegrationPool(t)
	queries := sqlc.New(pool)

	tenant := mustCreateTenantFixture(t, ctx, pool, "tenant-service-system")
	moduleID := mustCreateServiceModule(t, ctx, pool)
	mustAttachScheduleModule(t, ctx, pool, tenant.companyID, moduleID)
	mustGrantServicePermissions(t, ctx, pool, tenant.userID, service.PermissionServicesView)

	router := setupServiceRouterForTenantWithRole(pool, queries, tenant, "system")

	body, err := json.Marshal(map[string]any{
		"type_name":     "Banho",
		"title":         "Banho completo",
		"description":   "Banho com secagem",
		"price":         "89.90",
		"discount_rate": "0.00",
		"sub_services": []map[string]any{
			serviceSubServicePayload("Banho médio", "Banho para pets médios", "89.90"),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/services", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Contains(t, res.Body.String(), "permission_required")
}

func setupServiceRouterForTenant(pool *pgxpool.Pool, queries *sqlc.Queries, tenant integrationTenantFixture) *gin.Engine {
	return setupServiceRouterForTenantWithRole(pool, queries, tenant, "admin")
}

func setupServiceRouterForTenantWithRole(pool *pgxpool.Pool, queries *sqlc.Queries, tenant integrationTenantFixture, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	serviceHandler := handler.NewServiceHandler(service.NewServiceService(pool, queries))

	router := gin.New()
	router.Use(middleware.RequestContext())
	router.Use(func(c *gin.Context) {
		c.Set("company_id", tenant.companyID)
		c.Set("auth_claims", appjwt.Claims{
			UserID:    tenant.userID.String(),
			CompanyID: tenant.companyID.String(),
			Role:      role,
			Kind:      "owner",
		})
		c.Next()
	})
	router.Use(middleware.Audit(queries, nil))

	services := router.Group("/api/v1/services")
	services.Use(middleware.RequireModule(queries, "SVC"))
	services.GET("", middleware.RequirePermission(queries, service.PermissionServicesView), serviceHandler.List)
	services.POST("", middleware.RequirePermission(queries, service.PermissionServicesCreate), serviceHandler.Create)
	services.GET("/:id", middleware.RequirePermission(queries, service.PermissionServicesView), serviceHandler.GetByID)
	services.PUT("/:id", middleware.RequirePermission(queries, service.PermissionServicesUpdate), serviceHandler.Update)
	services.DELETE("/:id", middleware.RequirePermission(queries, service.PermissionServicesDelete), serviceHandler.Delete)

	return router
}

func mustGrantServicePermissions(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID pgtype.UUID, codes ...string) {
	t.Helper()

	for _, code := range codes {
		var permissionID pgtype.UUID
		err := pool.QueryRow(ctx, `
			INSERT INTO permissions (code, description, default_roles)
			VALUES ($1, $2, ARRAY['admin']::user_role_type[])
			ON CONFLICT (code) DO UPDATE SET description = EXCLUDED.description
			RETURNING id
		`, code, code).Scan(&permissionID)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `
			INSERT INTO user_permissions (user_id, permission_id, granted_by)
			VALUES ($1, $2, $1)
			ON CONFLICT (user_id, permission_id) DO UPDATE SET
				is_active = TRUE,
				revoked_at = NULL,
				revoked_by = NULL
		`, userID, permissionID)
		require.NoError(t, err)
	}
}

func mustCreateServiceModule(t *testing.T, ctx context.Context, pool *pgxpool.Pool) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO modules (code, name, description, min_package, is_active)
		VALUES ('SVC', 'Services', 'Services module', 'starter', TRUE)
		RETURNING id
	`).Scan(&id)
	require.NoError(t, err)
	return id
}

func serviceSubServicePayload(title string, description string, price string) map[string]any {
	return map[string]any{
		"title":         title,
		"description":   description,
		"price":         price,
		"discount_rate": "0.00",
		"average_times": []map[string]any{
			{
				"pet_size":             "medium",
				"pet_kind":             "dog",
				"pet_temperament":      "playful",
				"average_time_minutes": 60,
			},
		},
	}
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
