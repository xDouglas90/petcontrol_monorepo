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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
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

func TestUserHandler_CurrentIncludesSettingsAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewUserService(queries)
	handlerUnderTest := NewUserHandler(serviceUnderTest)

	userID := domainHandlerUUID(t)
	companyID := domainHandlerUUID(t)
	personID := domainHandlerUUID(t)

	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(userID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}))
	mock.ExpectQuery(`(?s)name: GetUserByID`).
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "email_verified", "email_verified_at", "role", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(userID, "admin@petcontrol.local", true, time.Now(), sqlc.UserRoleTypeAdmin, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetUserProfile`).
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"user_id", "person_id", "created_at"}).
			AddRow(userID, personID, time.Now()))
	mock.ExpectQuery(`(?s)name: GetPerson`).
		WithArgs(personID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "kind", "is_active", "has_system_user", "created_at", "updated_at", "full_name", "short_name", "gender_identity", "marital_status", "image_url", "birth_date", "cpf", "identifications_created_at", "identifications_updated_at"}).
			AddRow(personID, sqlc.PersonKindEmployee, true, true, time.Now(), time.Now(), "Maria da Silva", "Maria", nil, nil, nil, nil, nil, time.Now(), time.Now()))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("auth_claims", appjwt.Claims{
			UserID:    uuidToString(userID),
			CompanyID: uuidToString(companyID),
			Role:      "admin",
			Kind:      "owner",
		})
		c.Next()
	})
	router.GET("/users/me", handlerUnderTest.Current)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "\"can_view\":true")
	require.Contains(t, res.Body.String(), "\"can_manage_permissions\":true")
	require.Contains(t, res.Body.String(), "\"editable_permission_codes\":[\"company_settings:edit\"")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanySystemConfigHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewCompanySystemConfigService(queries)
	handlerUnderTest := NewCompanySystemConfigHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)
	scheduleDays := []sqlc.WeekDay{sqlc.WeekDayMonday, sqlc.WeekDayTuesday, sqlc.WeekDayWednesday}

	mock.ExpectExec(`(?s)name: UpdateCompanySystemConfig`).
		WithArgs(
			pgtype.Time{Microseconds: int64(8 * time.Hour / time.Microsecond), Valid: true},
			pgtype.Time{Microseconds: int64(12 * time.Hour / time.Microsecond), Valid: true},
			pgtype.Time{Microseconds: int64(13 * time.Hour / time.Microsecond), Valid: true},
			pgtype.Time{Microseconds: int64(18 * time.Hour / time.Microsecond), Valid: true},
			int16(4),
			int16(10),
			scheduleDays,
			true,
			int16(1),
			int16(2),
			int16(3),
			int16(4),
			true,
			false,
			pgtype.Text{String: "+5511999990001", Valid: true},
			companyID,
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	mock.ExpectQuery(`(?s)name: GetCompanySystemConfig`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{
			"company_id", "schedule_init_time", "schedule_pause_init_time", "schedule_pause_end_time", "schedule_end_time",
			"min_schedules_per_day", "max_schedules_per_day", "schedule_days", "dynamic_cages",
			"total_small_cages", "total_medium_cages", "total_large_cages", "total_giant_cages",
			"whatsapp_notifications", "whatsapp_conversation", "whatsapp_business_phone", "created_at", "updated_at",
		}).AddRow(
			companyID,
			pgtype.Time{Microseconds: int64(8 * time.Hour / time.Microsecond), Valid: true},
			pgtype.Time{Microseconds: int64(12 * time.Hour / time.Microsecond), Valid: true},
			pgtype.Time{Microseconds: int64(13 * time.Hour / time.Microsecond), Valid: true},
			pgtype.Time{Microseconds: int64(18 * time.Hour / time.Microsecond), Valid: true},
			int16(4),
			int16(10),
			scheduleDays,
			true,
			int16(1),
			int16(2),
			int16(3),
			int16(4),
			true,
			false,
			pgtype.Text{String: "+5511999990001", Valid: true},
			time.Now(),
			time.Now(),
		))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Next()
	})
	router.PATCH("/company-system-configs/current", handlerUnderTest.Update)

	body, err := json.Marshal(map[string]any{
		"schedule_init_time":       "08:00",
		"schedule_pause_init_time": "12:00",
		"schedule_pause_end_time":  "13:00",
		"schedule_end_time":        "18:00",
		"min_schedules_per_day":    4,
		"max_schedules_per_day":    10,
		"schedule_days":            []string{"monday", "tuesday", "wednesday"},
		"dynamic_cages":            true,
		"total_small_cages":        1,
		"total_medium_cages":       2,
		"total_large_cages":        3,
		"total_giant_cages":        4,
		"whatsapp_notifications":   true,
		"whatsapp_conversation":    false,
		"whatsapp_business_phone":  "+5511999990001",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/company-system-configs/current", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "\"schedule_init_time\":\"08:00\"")
	require.Contains(t, res.Body.String(), "\"schedule_days\":[\"monday\",\"tuesday\",\"wednesday\"]")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanySystemConfigHandler_UpdateRejectsInvalidPauseWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewCompanySystemConfigService(queries)
	handlerUnderTest := NewCompanySystemConfigHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Next()
	})
	router.PATCH("/company-system-configs/current", handlerUnderTest.Update)

	body, err := json.Marshal(map[string]any{
		"schedule_init_time":       "08:00",
		"schedule_pause_init_time": "07:00",
		"schedule_pause_end_time":  "09:00",
		"schedule_end_time":        "18:00",
		"min_schedules_per_day":    4,
		"max_schedules_per_day":    10,
		"schedule_days":            []string{"monday", "tuesday"},
		"dynamic_cages":            true,
		"total_small_cages":        1,
		"total_medium_cages":       2,
		"total_large_cages":        3,
		"total_giant_cages":        4,
		"whatsapp_notifications":   true,
		"whatsapp_conversation":    false,
		"whatsapp_business_phone":  "+5511999990001",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/company-system-configs/current", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnprocessableEntity, res.Code)
	require.Contains(t, res.Body.String(), "pause window must be inside operational hours")
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
	mock.ExpectQuery(`(?s)name: CreateCompanyUser`).WithArgs(companyID, userID, sqlc.UserKindOwner, true, pgtype.Bool{Bool: true, Valid: true}).WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_owner", "is_active", "created_at", "updated_at", "deleted_at"}).AddRow(createdID, companyID, userID, sqlc.UserKindOwner, true, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).WithArgs(companyID, userID).WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_owner", "is_active", "created_at", "updated_at", "deleted_at"}).AddRow(createdID, companyID, userID, sqlc.UserKindOwner, true, true, joinedAt, nil, nil))
	mock.ExpectExec(`(?s)name: DeactivateCompanyUser`).WithArgs(companyID, userID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).WithArgs(companyID, userID).WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_owner", "is_active", "created_at", "updated_at", "deleted_at"}).AddRow(createdID, companyID, userID, sqlc.UserKindOwner, true, false, joinedAt, time.Now(), nil))

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

func TestCompanyUserHandler_ListPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	companyUserService := service.NewCompanyUserService(queries)
	permissionService := service.NewCompanyUserPermissionService(queries)
	handlerUnderTest := NewCompanyUserHandler(companyUserService, permissionService)

	companyID := domainHandlerUUID(t)
	targetUserID := domainHandlerUUID(t)
	adminUserID := domainHandlerUUID(t)
	grantedAt := time.Now().UTC()

	mock.ExpectQuery(`(?s)name: GetCompanyByID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(companyID, "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, domainHandlerUUID(t), sqlc.ModulePackageStarter, true, time.Now(), nil, nil))

	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, targetUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_owner", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(domainHandlerUUID(t), companyID, targetUserID, sqlc.UserKindEmployee, false, true, time.Now(), nil, nil))

	mock.ExpectQuery(`(?s)name: GetUserByID`).
		WithArgs(targetUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "email_verified", "email_verified_at", "role", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(targetUserID, "system@petcontrol.local", true, time.Now(), sqlc.UserRoleTypeSystem, true, time.Now(), nil, nil))

	mock.ExpectQuery(`(?s)name: ListPermissionsByCodes`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "created_at", "updated_at"}).
			AddRow(domainHandlerUUID(t), "company_settings:edit", "Editar configurações gerais", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin}, time.Now(), nil).
			AddRow(domainHandlerUUID(t), "plan_settings:edit", "Editar configurações de plano", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem}, time.Now(), nil))

	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(targetUserID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}).
			AddRow(domainHandlerUUID(t), "plan_settings:edit", "Editar configurações de plano", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem}, adminUserID, grantedAt, nil, nil))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Set("auth_claims", appjwt.Claims{UserID: uuidToString(adminUserID), Role: "admin"})
		c.Next()
	})
	router.GET("/company-users/:user_id/permissions", handlerUnderTest.ListPermissions)

	req := httptest.NewRequest(http.MethodGet, "/company-users/"+uuidToString(targetUserID)+"/permissions", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "\"code\":\"company_settings:edit\"")
	require.Contains(t, res.Body.String(), "\"code\":\"plan_settings:edit\"")
	require.Contains(t, res.Body.String(), "\"is_active\":true")
	require.Contains(t, res.Body.String(), "\"managed_by\":\""+uuidToString(adminUserID)+"\"")
	require.Contains(t, res.Body.String(), "\"active_package\":\"starter\"")
	require.Contains(t, res.Body.String(), "\"permission_groups\"")
	require.Contains(t, res.Body.String(), "\"module_code\":\"CFG\"")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyUserHandler_UpdatePermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	companyUserService := service.NewCompanyUserService(queries)
	permissionService := service.NewCompanyUserPermissionService(queries)
	handlerUnderTest := NewCompanyUserHandler(companyUserService, permissionService)

	companyID := domainHandlerUUID(t)
	targetUserID := domainHandlerUUID(t)
	adminUserID := domainHandlerUUID(t)
	companyPermissionID := domainHandlerUUID(t)
	planPermissionID := domainHandlerUUID(t)

	mock.ExpectQuery(`(?s)name: GetCompanyByID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(companyID, "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, domainHandlerUUID(t), sqlc.ModulePackageStarter, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, targetUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_owner", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(domainHandlerUUID(t), companyID, targetUserID, sqlc.UserKindEmployee, false, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetUserByID`).
		WithArgs(targetUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "email_verified", "email_verified_at", "role", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(targetUserID, "system@petcontrol.local", true, time.Now(), sqlc.UserRoleTypeSystem, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: ListPermissionsByCodes`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "created_at", "updated_at"}).
			AddRow(companyPermissionID, "company_settings:edit", "Editar configurações gerais", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin}, time.Now(), nil).
			AddRow(planPermissionID, "plan_settings:edit", "Editar configurações de plano", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem}, time.Now(), nil))
	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(targetUserID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}).
			AddRow(companyPermissionID, "company_settings:edit", "Editar configurações gerais", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin}, adminUserID, time.Now(), nil, nil))

	mock.ExpectQuery(`(?s)name: GetCompanyByID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(companyID, "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, domainHandlerUUID(t), sqlc.ModulePackageStarter, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetCompanyByID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(companyID, "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, domainHandlerUUID(t), sqlc.ModulePackageStarter, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, targetUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_owner", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(domainHandlerUUID(t), companyID, targetUserID, sqlc.UserKindEmployee, false, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetUserByID`).
		WithArgs(targetUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "email_verified", "email_verified_at", "role", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(targetUserID, "system@petcontrol.local", true, time.Now(), sqlc.UserRoleTypeSystem, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: ListPermissionsByCodes`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "created_at", "updated_at"}).
			AddRow(companyPermissionID, "company_settings:edit", "Editar configurações gerais", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin}, time.Now(), nil).
			AddRow(planPermissionID, "plan_settings:edit", "Editar configurações de plano", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem}, time.Now(), nil))
	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(targetUserID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}).
			AddRow(companyPermissionID, "company_settings:edit", "Editar configurações gerais", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin}, adminUserID, time.Now(), nil, nil))

	mock.ExpectExec(`(?s)name: DeleteUserPermission`).
		WithArgs(adminUserID, targetUserID, companyPermissionID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(`(?s)name: ReactivateUserPermission`).
		WithArgs(adminUserID, targetUserID, planPermissionID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	mock.ExpectExec(`(?s)name: InsertUserPermission`).
		WithArgs(targetUserID, planPermissionID, adminUserID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectQuery(`(?s)name: GetCompanyByID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(companyID, "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, domainHandlerUUID(t), sqlc.ModulePackageStarter, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, targetUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_owner", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(domainHandlerUUID(t), companyID, targetUserID, sqlc.UserKindEmployee, false, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetUserByID`).
		WithArgs(targetUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "email_verified", "email_verified_at", "role", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(targetUserID, "system@petcontrol.local", true, time.Now(), sqlc.UserRoleTypeSystem, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: ListPermissionsByCodes`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "created_at", "updated_at"}).
			AddRow(companyPermissionID, "company_settings:edit", "Editar configurações gerais", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin}, time.Now(), nil).
			AddRow(planPermissionID, "plan_settings:edit", "Editar configurações de plano", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem}, time.Now(), nil))
	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(targetUserID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}).
			AddRow(planPermissionID, "plan_settings:edit", "Editar configurações de plano", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem}, adminUserID, time.Now(), nil, nil))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Set("auth_claims", appjwt.Claims{UserID: uuidToString(adminUserID), Role: "admin"})
		c.Next()
	})
	router.PATCH("/company-users/:user_id/permissions", handlerUnderTest.UpdatePermissions)

	body, err := json.Marshal(map[string]any{
		"permission_codes": []string{"plan_settings:edit"},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/company-users/"+uuidToString(targetUserID)+"/permissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "\"code\":\"plan_settings:edit\"")
	require.NotContains(t, res.Body.String(), "\"code\":\"company_settings:edit\",\"is_active\":true")
	require.Contains(t, res.Body.String(), "\"permission_groups\"")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyUserHandler_ListPermissionsRejectsNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	companyUserService := service.NewCompanyUserService(queries)
	permissionService := service.NewCompanyUserPermissionService(queries)
	handlerUnderTest := NewCompanyUserHandler(companyUserService, permissionService)

	companyID := domainHandlerUUID(t)
	targetUserID := domainHandlerUUID(t)
	systemUserID := domainHandlerUUID(t)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Set("auth_claims", appjwt.Claims{UserID: uuidToString(systemUserID), Role: "system"})
		c.Next()
	})
	router.GET("/company-users/:user_id/permissions", handlerUnderTest.ListPermissions)

	req := httptest.NewRequest(http.MethodGet, "/company-users/"+uuidToString(targetUserID)+"/permissions", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Contains(t, res.Body.String(), "admin_required")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyUserHandler_UpdatePermissionsRejectsUserOutsideTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	companyUserService := service.NewCompanyUserService(queries)
	permissionService := service.NewCompanyUserPermissionService(queries)
	handlerUnderTest := NewCompanyUserHandler(companyUserService, permissionService)

	companyID := domainHandlerUUID(t)
	targetUserID := domainHandlerUUID(t)
	adminUserID := domainHandlerUUID(t)

	mock.ExpectQuery(`(?s)name: GetCompanyByID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(companyID, "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, domainHandlerUUID(t), sqlc.ModulePackageStarter, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, targetUserID).
		WillReturnError(pgx.ErrNoRows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Set("auth_claims", appjwt.Claims{UserID: uuidToString(adminUserID), Role: "admin"})
		c.Next()
	})
	router.PATCH("/company-users/:user_id/permissions", handlerUnderTest.UpdatePermissions)

	body, err := json.Marshal(map[string]any{
		"permission_codes": []string{"plan_settings:edit"},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/company-users/"+uuidToString(targetUserID)+"/permissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
	require.Contains(t, res.Body.String(), "failed to load company user permissions")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyUserHandler_UpdatePermissionsRejectsPermissionOutsideCompanyPackage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	companyUserService := service.NewCompanyUserService(queries)
	permissionService := service.NewCompanyUserPermissionService(queries)
	handlerUnderTest := NewCompanyUserHandler(companyUserService, permissionService)

	companyID := domainHandlerUUID(t)
	targetUserID := domainHandlerUUID(t)
	adminUserID := domainHandlerUUID(t)
	companyPermissionID := domainHandlerUUID(t)

	mock.ExpectQuery(`(?s)name: GetCompanyByID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(companyID, "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, domainHandlerUUID(t), sqlc.ModulePackageStarter, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, targetUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_owner", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(domainHandlerUUID(t), companyID, targetUserID, sqlc.UserKindEmployee, false, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: GetUserByID`).
		WithArgs(targetUserID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "email_verified", "email_verified_at", "role", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(targetUserID, "system@petcontrol.local", true, time.Now(), sqlc.UserRoleTypeSystem, true, time.Now(), nil, nil))
	mock.ExpectQuery(`(?s)name: ListPermissionsByCodes`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "created_at", "updated_at"}).
			AddRow(companyPermissionID, "company_settings:edit", "Editar configurações gerais", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin}, time.Now(), nil))
	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(targetUserID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}).
			AddRow(companyPermissionID, "company_settings:edit", "Editar configurações gerais", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin}, adminUserID, time.Now(), nil, nil))

	mock.ExpectQuery(`(?s)name: GetCompanyByID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(companyID, "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, domainHandlerUUID(t), sqlc.ModulePackageStarter, true, time.Now(), nil, nil))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Set("auth_claims", appjwt.Claims{UserID: uuidToString(adminUserID), Role: "admin"})
		c.Next()
	})
	router.PATCH("/company-users/:user_id/permissions", handlerUnderTest.UpdatePermissions)

	body, err := json.Marshal(map[string]any{
		"permission_codes": []string{"chat:view"},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/company-users/"+uuidToString(targetUserID)+"/permissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnprocessableEntity, res.Code)
	require.Contains(t, res.Body.String(), "update_company_user_permissions_failed")
	require.Contains(t, res.Body.String(), "failed to update company user permissions")
	require.NoError(t, mock.ExpectationsWereMet())
}
