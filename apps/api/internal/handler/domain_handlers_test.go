package handler

import (
	"bytes"
	"encoding/json"
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
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

func domainHandlerUUID(t *testing.T) pgtype.UUID {
	t.Helper()

	value := uuid.New()
	var out pgtype.UUID
	copy(out.Bytes[:], value[:])
	out.Valid = true
	return out
}

func domainServiceWithMock(t *testing.T) (*sqlc.Queries, pgxmock.PgxPoolIface) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	return sqlc.New(mock), mock
}

func TestCompanyHandler_Current(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewCompanyService(queries)
	handlerUnderTest := NewCompanyHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)
	mock.ExpectQuery(`(?s)name: GetCompanyByID`).WithArgs(companyID).WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).AddRow(companyID.String(), "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, domainHandlerUUID(t).String(), sqlc.ModulePackageStarter, true, time.Now(), nil, nil))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Next()
	})
	router.GET("/companies/current", handlerUnderTest.Current)

	req := httptest.NewRequest(http.MethodGet, "/companies/current", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "PetControl")
	require.Contains(t, res.Body.String(), "\"slug\":\"petcontrol\"")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPlanHandler_Current(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewPlanService(queries)
	handlerUnderTest := NewPlanHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)
	mock.ExpectQuery(`(?s)name: GetCurrentPlanByCompanyID`).WithArgs(companyID).WillReturnRows(pgxmock.NewRows([]string{"id", "plan_type_id", "name", "description", "package", "price", "billing_cycle_days", "max_users", "is_active", "image_url", "created_at", "updated_at", "deleted_at"}).AddRow(domainHandlerUUID(t).String(), domainHandlerUUID(t).String(), "Starter", "starter plan", sqlc.ModulePackageStarter, "99.90", int32(30), pgtype.Int4{Int32: 5, Valid: true}, true, nil, time.Now(), nil, nil))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Next()
	})
	router.GET("/plans/current", handlerUnderTest.Current)

	req := httptest.NewRequest(http.MethodGet, "/plans/current", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "Starter")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModuleHandler_Active(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewModuleService(queries)
	handlerUnderTest := NewModuleHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)
	mock.ExpectQuery(`(?s)name: ListActiveModulesByCompanyID`).WithArgs(companyID).WillReturnRows(pgxmock.NewRows([]string{"id", "code", "name", "description", "min_package", "is_active", "created_at", "updated_at", "deleted_at"}).AddRow(domainHandlerUUID(t).String(), "SCH", "Scheduling", "Scheduling", sqlc.ModulePackageStarter, true, time.Now(), nil, nil))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Next()
	})
	router.GET("/modules/active", handlerUnderTest.Active)

	req := httptest.NewRequest(http.MethodGet, "/modules/active", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "SCH")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModuleHandler_Access(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewModuleService(queries)
	handlerUnderTest := NewModuleHandler(serviceUnderTest)

	router := gin.New()
	router.GET("/modules/:code/access", handlerUnderTest.Access)

	req := httptest.NewRequest(http.MethodGet, "/modules/SCH/access", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.JSONEq(t, `{"data":{"allowed":true,"module":"SCH"}}`, res.Body.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyUserHandler_CreateAndDeactivate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewCompanyUserService(queries)
	handlerUnderTest := NewCompanyUserHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)
	userID := domainHandlerUUID(t)
	createdID := domainHandlerUUID(t)
	joinedAt := time.Now().UTC()
	leftAt := joinedAt.Add(5 * time.Minute)
	mock.ExpectQuery(`(?s)name: CreateCompanyUser`).WithArgs(companyID, userID, true, true).WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "is_owner", "is_active", "joined_at", "left_at"}).AddRow(createdID.String(), companyID.String(), userID.String(), true, true, time.Now(), nil))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).WithArgs(companyID, userID).WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "is_owner", "is_active", "joined_at", "left_at"}).AddRow(createdID.String(), companyID.String(), userID.String(), true, true, joinedAt, nil))
	mock.ExpectExec(`(?s)name: DeactivateCompanyUser`).WithArgs(companyID, userID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).WithArgs(companyID, userID).WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "is_owner", "is_active", "joined_at", "left_at"}).AddRow(createdID.String(), companyID.String(), userID.String(), true, false, joinedAt, leftAt))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Next()
	})
	router.POST("/company-users", handlerUnderTest.Create)
	router.DELETE("/company-users/:user_id", handlerUnderTest.Deactivate)

	body, err := json.Marshal(map[string]any{"user_id": userID.String(), "is_owner": true})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/company-users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusCreated, res.Code)
	require.Contains(t, res.Body.String(), createdID.String())

	req = httptest.NewRequest(http.MethodDelete, "/company-users/"+userID.String(), nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyUserHandler_CreateRejectsInvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewCompanyUserService(queries)
	handlerUnderTest := NewCompanyUserHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Next()
	})
	router.POST("/company-users", handlerUnderTest.Create)

	body, err := json.Marshal(map[string]any{"user_id": "invalid", "is_owner": true})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/company-users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnprocessableEntity, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
