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
	"github.com/xdouglas90/petcontrol_monorepo/test/integration"
)

func TestCompanyUserEndpoints_AuditLogsOnCreateAndDeactivate(t *testing.T) {
	ctx := context.Background()
	setup := integration.SetupPostgresWithMigrations(t)
	pool := setup.Pool
	queries := sqlc.New(pool)

	tenant := mustCreateTenantFixture(t, ctx, pool, "tenant-company-user-audit")
	targetUser := mustInsertUser(t, ctx, pool, "tenant-company-user-target")

	router := setupCompanyUserRouterForTenant(queries, tenant)

	body, err := json.Marshal(map[string]any{
		"user_id":  targetUser.String(),
		"is_owner": false,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/company-users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusCreated, res.Code)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/company-users/"+targetUser.String(), nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)

	var createdCount int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE entity_table = 'company_users' AND action = 'create' AND company_id = $1`, tenant.companyID).Scan(&createdCount)
	require.NoError(t, err)
	require.Equal(t, 1, createdCount)

	var deactivatedCount int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM audit_logs WHERE entity_table = 'company_users' AND action = 'deactivate' AND company_id = $1`, tenant.companyID).Scan(&deactivatedCount)
	require.NoError(t, err)
	require.Equal(t, 1, deactivatedCount)
}

func setupCompanyUserRouterForTenant(queries sqlc.Querier, tenant integrationTenantFixture) *gin.Engine {
	gin.SetMode(gin.TestMode)

	companyUserService := service.NewCompanyUserService(queries)
	companyUserHandler := handler.NewCompanyUserHandler(companyUserService)

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

	router.POST("/api/v1/company-users", middleware.RequireCompanyOwner(queries), companyUserHandler.Create)
	router.DELETE("/api/v1/company-users/:user_id", middleware.RequireCompanyOwner(queries), companyUserHandler.Deactivate)

	return router
}

func mustInsertCompanyUserMembership(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	companyID pgtype.UUID,
	userID pgtype.UUID,
	kind string,
	isOwner bool,
	isActive bool,
) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO company_users (company_id, user_id, kind, is_owner, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, companyID, userID, kind, isOwner, isActive)
	require.NoError(t, err)
}
