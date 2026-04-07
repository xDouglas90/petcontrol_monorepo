package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func newPGUUID(t *testing.T) pgtype.UUID {
	t.Helper()

	value := uuid.New()
	var out pgtype.UUID
	copy(out.Bytes[:], value[:])
	out.Valid = true
	return out
}

func TestRequireModule_AllowsAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	companyID := newPGUUID(t)
	mock.ExpectQuery(`(?s)name: HasActiveCompanyModuleByCode`).WithArgs(companyID, "SCH").WillReturnRows(pgxmock.NewRows([]string{"has_access"}).AddRow(true))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(companyIDContextKey, companyID)
		c.Next()
	}, RequireModule(sqlc.New(mock), "SCH"))
	router.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequireModule_UsesRouteParamWhenModuleCodeMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	companyID := newPGUUID(t)
	mock.ExpectQuery(`(?s)name: HasActiveCompanyModuleByCode`).WithArgs(companyID, "FIN").WillReturnRows(pgxmock.NewRows([]string{"has_access"}).AddRow(true))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(companyIDContextKey, companyID)
		c.Next()
	})
	router.GET("/companies/:code/private", RequireModule(sqlc.New(mock), ""), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/companies/FIN/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequireModule_DeniesMissingModuleCode(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(RequireModule(nil, ""))
	router.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestRequireModule_DeniesMissingTenantContext(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(RequireModule(nil, "SCH"))
	router.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestRequireModule_DeniesWhenModuleUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	companyID := newPGUUID(t)
	mock.ExpectQuery(`(?s)name: HasActiveCompanyModuleByCode`).WithArgs(companyID, "SCH").WillReturnRows(pgxmock.NewRows([]string{"has_access"}).AddRow(false))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(companyIDContextKey, companyID)
		c.Next()
	}, RequireModule(sqlc.New(mock), "SCH"))
	router.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRequireModule_DeniesWhenQueryFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	companyID := newPGUUID(t)
	mock.ExpectQuery(`(?s)name: HasActiveCompanyModuleByCode`).WithArgs(companyID, "SCH").WillReturnError(context.DeadlineExceeded)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(companyIDContextKey, companyID)
		c.Next()
	}, RequireModule(sqlc.New(mock), "SCH"))
	router.GET("/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
