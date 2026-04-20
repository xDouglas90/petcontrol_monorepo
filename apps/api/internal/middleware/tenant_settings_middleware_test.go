package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

func tenantSettingsUUID(t *testing.T) pgtype.UUID {
	t.Helper()

	value := uuid.New()
	var out pgtype.UUID
	copy(out.Bytes[:], value[:])
	out.Valid = true
	return out
}

func TestRequireTenantSettingsPermission_AllowsAdminWithoutExplicitPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := tenantSettingsUUID(t)

	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(userID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(claimsContextKey, appjwt.Claims{UserID: userID.String(), Role: "admin"})
		c.Next()
	}, RequireTenantSettingsPermission(sqlc.New(mock), service.TenantSettingsPermissionCompanyEdit))
	router.PATCH("/settings", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPatch, "/settings", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequireTenantSettingsPermission_AllowsSystemWithMatchingPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := tenantSettingsUUID(t)

	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(userID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}).
			AddRow(tenantSettingsUUID(t), service.TenantSettingsPermissionCompanyEdit, "Editar configurações gerais", []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin}, tenantSettingsUUID(t), time.Now(), nil, nil))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(claimsContextKey, appjwt.Claims{UserID: userID.String(), Role: "system"})
		c.Next()
	}, RequireTenantSettingsPermission(sqlc.New(mock), service.TenantSettingsPermissionCompanyEdit))
	router.PATCH("/settings", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPatch, "/settings", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequireTenantSettingsPermission_DeniesSystemWithoutMatchingPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := tenantSettingsUUID(t)

	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(userID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}).
			AddRow(tenantSettingsUUID(t), service.TenantSettingsPermissionPlanEdit, "Editar configurações de plano", []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem}, tenantSettingsUUID(t), time.Now(), nil, nil))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(claimsContextKey, appjwt.Claims{UserID: userID.String(), Role: "system"})
		c.Next()
	}, RequireTenantSettingsPermission(sqlc.New(mock), service.TenantSettingsPermissionCompanyEdit))
	router.PATCH("/settings", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPatch, "/settings", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Contains(t, res.Body.String(), "tenant_settings_permission_required")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequireTenantSettingsPermission_DeniesCommonUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := tenantSettingsUUID(t)

	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(userID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(claimsContextKey, appjwt.Claims{UserID: userID.String(), Role: "common"})
		c.Next()
	}, RequireTenantSettingsPermission(sqlc.New(mock), service.TenantSettingsPermissionCompanyEdit))
	router.PATCH("/settings", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPatch, "/settings", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Contains(t, res.Body.String(), "tenant_settings_access_required")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequireTenantSettingsPermission_DeniesWhenPermissionLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := tenantSettingsUUID(t)

	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(userID, int32(0), int32(1000)).
		WillReturnError(context.DeadlineExceeded)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(claimsContextKey, appjwt.Claims{UserID: userID.String(), Role: "system"})
		c.Next()
	}, RequireTenantSettingsPermission(sqlc.New(mock), service.TenantSettingsPermissionCompanyEdit))
	router.PATCH("/settings", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPatch, "/settings", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
