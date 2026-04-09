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
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

func TestScheduleHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewScheduleService(queries)
	handlerUnderTest := NewScheduleHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)
	scheduleID := domainHandlerUUID(t)
	clientID := domainHandlerUUID(t)
	petID := domainHandlerUUID(t)
	now := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`(?s)name: ListSchedulesByCompanyID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), now, nil, "banho", nil, now, nil, nil, sqlc.ScheduleStatusWaiting))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Next()
	})
	router.GET("/schedules", handlerUnderTest.List)

	req := httptest.NewRequest(http.MethodGet, "/schedules", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), scheduleID.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleHandler_CreateAndGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewScheduleService(queries)
	handlerUnderTest := NewScheduleHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)
	scheduleID := domainHandlerUUID(t)
	clientID := domainHandlerUUID(t)
	petID := domainHandlerUUID(t)
	creatorRaw := uuid.NewString()
	creatorID, err := parseUUID(creatorRaw)
	require.NoError(t, err)
	now := time.Now().UTC().Truncate(time.Second)
	end := now.Add(60 * time.Minute)

	mock.ExpectQuery(`(?s)name: ValidateScheduleOwnership`).
		WithArgs(petID, companyID, clientID).
		WillReturnRows(pgxmock.NewRows([]string{"is_valid"}).AddRow(true))

	mock.ExpectQuery(`(?s)name: CreateSchedule`).
		WithArgs(companyID, clientID, petID, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), creatorID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), now, end, "banho", creatorRaw, now, nil, nil))

	mock.ExpectQuery(`(?s)name: InsertScheduleStatusHistory`).
		WithArgs(scheduleID, sqlc.ScheduleStatusWaiting, creatorID, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "schedule_id", "status", "changed_at", "changed_by", "notes"}).
			AddRow(uuid.NewString(), scheduleID.String(), sqlc.ScheduleStatusWaiting, now, creatorRaw, ""))

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), now, end, "banho", creatorRaw, now, nil, nil, sqlc.ScheduleStatusWaiting))

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), now, end, "banho", creatorRaw, now, nil, nil, sqlc.ScheduleStatusWaiting))

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), now, end, "banho", creatorRaw, now, nil, nil, sqlc.ScheduleStatusWaiting))

	mock.ExpectQuery(`(?s)name: ListScheduleStatusHistoryByScheduleID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "schedule_id", "status", "changed_at", "changed_by", "notes"}).
			AddRow(uuid.NewString(), scheduleID.String(), sqlc.ScheduleStatusWaiting, now, creatorRaw, ""))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Set("auth_claims", appjwt.Claims{UserID: creatorRaw, CompanyID: companyID.String(), Role: "admin", Kind: "owner"})
		c.Next()
	})
	router.POST("/schedules", handlerUnderTest.Create)
	router.GET("/schedules/:id", handlerUnderTest.GetByID)
	router.GET("/schedules/:id/history", handlerUnderTest.History)

	body, err := json.Marshal(map[string]any{
		"client_id":     clientID.String(),
		"pet_id":        petID.String(),
		"scheduled_at":  now.Format(time.RFC3339),
		"estimated_end": end.Format(time.RFC3339),
		"notes":         "banho",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.Contains(t, res.Body.String(), scheduleID.String())

	req = httptest.NewRequest(http.MethodGet, "/schedules/"+scheduleID.String(), nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), scheduleID.String())

	req = httptest.NewRequest(http.MethodGet, "/schedules/"+scheduleID.String()+"/history", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "waiting")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleHandler_CreateRejectsInvalidDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewScheduleService(queries)
	handlerUnderTest := NewScheduleHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)
	userID := uuid.NewString()
	clientID := domainHandlerUUID(t)
	petID := domainHandlerUUID(t)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Set("auth_claims", appjwt.Claims{UserID: userID, CompanyID: companyID.String(), Role: "admin", Kind: "owner"})
		c.Next()
	})
	router.POST("/schedules", handlerUnderTest.Create)

	body, err := json.Marshal(map[string]any{
		"client_id":    clientID.String(),
		"pet_id":       petID.String(),
		"scheduled_at": "not-a-date",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/schedules", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnprocessableEntity, res.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
