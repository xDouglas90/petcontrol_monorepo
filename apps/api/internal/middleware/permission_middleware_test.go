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
)

func permissionUUID(t *testing.T) pgtype.UUID {
	t.Helper()

	value := uuid.New()
	var out pgtype.UUID
	copy(out.Bytes[:], value[:])
	out.Valid = true
	return out
}

func TestRequirePermission_AllowsWhenUserHasPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := permissionUUID(t)
	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).WithArgs(userID, int32(0), int32(1000)).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}).
			AddRow(permissionUUID(t), "people:view", pgtype.Text{String: "Visualizar pessoa", Valid: true}, []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin}, permissionUUID(t), pgtype.Timestamptz{Time: time.Now(), Valid: true}, pgtype.UUID{}, pgtype.Timestamptz{}),
	)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(claimsContextKey, appjwt.Claims{
			UserID: userID.String(),
		})
		c.Next()
	}, RequirePermission(sqlc.New(mock), "people:view"))
	router.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequirePermission_DeniesWhenUserDoesNotHavePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := permissionUUID(t)
	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).WithArgs(userID, int32(0), int32(1000)).WillReturnRows(
		pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}).
			AddRow(permissionUUID(t), "people:view", pgtype.Text{String: "Visualizar pessoa", Valid: true}, []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin}, permissionUUID(t), pgtype.Timestamptz{Time: time.Now(), Valid: true}, pgtype.UUID{}, pgtype.Timestamptz{}),
	)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(claimsContextKey, appjwt.Claims{
			UserID: userID.String(),
		})
		c.Next()
	}, RequirePermission(sqlc.New(mock), "people:update"))
	router.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Contains(t, res.Body.String(), "permission_required")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequirePermission_DeniesMissingPermissionCode(t *testing.T) {
	router := gin.New()
	router.Use(RequirePermission(nil, ""))
	router.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestRequirePermission_DeniesMissingClaims(t *testing.T) {
	router := gin.New()
	router.Use(RequirePermission(nil, "people:view"))
	router.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestRequirePermission_DeniesWhenLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := permissionUUID(t)
	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).WithArgs(userID, int32(0), int32(1000)).WillReturnError(context.DeadlineExceeded)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(claimsContextKey, appjwt.Claims{
			UserID: userID.String(),
		})
		c.Next()
	}, RequirePermission(sqlc.New(mock), "people:view"))
	router.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Contains(t, res.Body.String(), "permission_verification_failed")
	require.NoError(t, mock.ExpectationsWereMet())
}
