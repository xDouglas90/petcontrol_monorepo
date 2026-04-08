package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
)

func TestRequireCompanyOwner_AllowsOwnerMembership(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	companyID := newPGUUID(t)
	userID := uuid.New()
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, mustParseUUID(t, userID.String())).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "is_owner", "is_active", "joined_at", "left_at"}).
			AddRow(uuid.NewString(), companyID.String(), userID.String(), true, true, time.Now(), nil))

	router := gin.New()
	router.Use(setClaimsAndCompany(t, userID.String(), companyID.String(), "admin"), RequireCompanyOwner(sqlc.New(mock)))
	router.POST("/company-users", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/company-users", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequireCompanyOwner_DeniesNonOwnerMembership(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	companyID := newPGUUID(t)
	userID := uuid.New()
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, mustParseUUID(t, userID.String())).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "is_owner", "is_active", "joined_at", "left_at"}).
			AddRow(uuid.NewString(), companyID.String(), userID.String(), false, true, time.Now(), nil))

	router := gin.New()
	router.Use(setClaimsAndCompany(t, userID.String(), companyID.String(), "admin"), RequireCompanyOwner(sqlc.New(mock)))
	router.POST("/company-users", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/company-users", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequireCompanyOwner_AllowsRootRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	companyID := newPGUUID(t)
	userID := uuid.NewString()

	router := gin.New()
	router.Use(setClaimsAndCompany(t, userID, companyID.String(), "root"), RequireCompanyOwner(sqlc.New(mock)))
	router.DELETE("/company-users/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodDelete, "/company-users/"+uuid.NewString(), nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func setClaimsAndCompany(t *testing.T, userID string, companyID string, role string) gin.HandlerFunc {
	t.Helper()

	return func(c *gin.Context) {
		c.Set(claimsContextKey, appjwt.Claims{
			UserID:    userID,
			CompanyID: companyID,
			Role:      role,
			Kind:      "owner",
		})

		parsedCompanyID, err := parseUUID(companyID)
		require.NoError(t, err)
		c.Set(companyIDContextKey, parsedCompanyID)
		c.Next()
	}
}

func mustParseUUID(t *testing.T, raw string) any {
	t.Helper()

	value, err := parseUUID(raw)
	require.NoError(t, err)
	return value
}

func TestRequireCompanyOwner_DeniesWhenLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	companyID := newPGUUID(t)
	userID := uuid.New()
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, mustParseUUID(t, userID.String())).
		WillReturnError(context.DeadlineExceeded)

	router := gin.New()
	router.Use(setClaimsAndCompany(t, userID.String(), companyID.String(), "admin"), RequireCompanyOwner(sqlc.New(mock)))
	router.POST("/company-users", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/company-users", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
