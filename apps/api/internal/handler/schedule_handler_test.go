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

	serviceUnderTest := service.NewScheduleService(mock, queries)
	handlerUnderTest := NewScheduleHandler(serviceUnderTest)

	companyID := domainHandlerUUID(t)
	scheduleID := domainHandlerUUID(t)
	clientID := domainHandlerUUID(t)
	petID := domainHandlerUUID(t)
	now := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`(?s)name: ListSchedulesByCompanyID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "client_name", "pet_name", "service_ids", "service_titles", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), "Maria Silva", "Thor", []string{"service-1"}, []string{"Banho"}, now, nil, "banho", nil, now, nil, nil, sqlc.ScheduleStatusWaiting))

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

	serviceUnderTest := service.NewScheduleService(mock, queries)
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

	mock.ExpectBegin()

	mock.ExpectQuery(`(?s)name: CreateSchedule`).
		WithArgs(companyID, clientID, petID, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), creatorID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), now, end, "banho", creatorRaw, now, nil, nil))

	mock.ExpectQuery(`(?s)name: InsertScheduleStatusHistory`).
		WithArgs(scheduleID, sqlc.ScheduleStatusWaiting, creatorID, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "schedule_id", "status", "changed_at", "changed_by", "notes"}).
			AddRow(uuid.NewString(), scheduleID.String(), sqlc.ScheduleStatusWaiting, now, creatorRaw, ""))

	mock.ExpectExec(`(?s)name: DeleteScheduleServicesByScheduleID`).
		WithArgs(scheduleID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	mock.ExpectCommit()

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "client_name", "pet_name", "service_ids", "service_titles", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), "Maria Silva", "Thor", []string{}, []string{}, now, end, "banho", creatorRaw, now, nil, nil, sqlc.ScheduleStatusWaiting))

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "client_name", "pet_name", "service_ids", "service_titles", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), "Maria Silva", "Thor", []string{}, []string{}, now, end, "banho", creatorRaw, now, nil, nil, sqlc.ScheduleStatusWaiting))

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "client_name", "pet_name", "service_ids", "service_titles", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), "Maria Silva", "Thor", []string{}, []string{}, now, end, "banho", creatorRaw, now, nil, nil, sqlc.ScheduleStatusWaiting))

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

func TestScheduleHandler_UpdatePublishesConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewScheduleService(mock, queries)
	pub := &publisherStub{}
	handlerUnderTest := NewScheduleHandler(serviceUnderTest, pub)

	companyID := domainHandlerUUID(t)
	scheduleID := domainHandlerUUID(t)
	clientID := domainHandlerUUID(t)
	petID := domainHandlerUUID(t)
	serviceID := domainHandlerUUID(t)
	changedByRaw := uuid.NewString()
	changedBy, err := parseUUID(changedByRaw)
	require.NoError(t, err)
	now := time.Now().UTC().Truncate(time.Second)
	confirmedAt := now.Add(30 * time.Minute)

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "client_name", "pet_name", "service_ids", "service_titles", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), "Maria Silva", "Thor", []string{"service-1"}, []string{"Banho"}, now, confirmedAt, "banho", changedByRaw, now, nil, nil, sqlc.ScheduleStatusWaiting))

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "client_name", "pet_name", "service_ids", "service_titles", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), "Maria Silva", "Thor", []string{"service-1"}, []string{"Banho"}, now, confirmedAt, "banho", changedByRaw, now, nil, nil, sqlc.ScheduleStatusWaiting))

	mock.ExpectQuery(`(?s)name: ValidateScheduleOwnership`).
		WithArgs(petID, companyID, clientID).
		WillReturnRows(pgxmock.NewRows([]string{"is_valid"}).AddRow(true))

	mock.ExpectQuery(`(?s)name: ValidateServiceByIDAndCompanyID`).
		WithArgs(companyID, serviceID).
		WillReturnRows(pgxmock.NewRows([]string{"is_valid"}).AddRow(true))

	mock.ExpectBegin()

	mock.ExpectQuery(`(?s)name: InsertScheduleStatusHistory`).
		WithArgs(scheduleID, sqlc.ScheduleStatusConfirmed, changedBy, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "schedule_id", "status", "changed_at", "changed_by", "notes"}).
			AddRow(uuid.NewString(), scheduleID.String(), sqlc.ScheduleStatusConfirmed, now, changedByRaw, "confirmado"))

	mock.ExpectExec(`(?s)name: DeleteScheduleServicesByScheduleID`).
		WithArgs(scheduleID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	mock.ExpectQuery(`(?s)name: InsertScheduleService`).
		WithArgs(scheduleID, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "schedule_id", "service_id", "created_at"}).
			AddRow(uuid.NewString(), scheduleID.String(), uuid.NewString(), now))

	mock.ExpectCommit()

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "client_name", "pet_name", "service_ids", "service_titles", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), "Maria Silva", "Thor", []string{"service-1"}, []string{"Banho"}, now, confirmedAt, "banho", changedByRaw, now, nil, nil, sqlc.ScheduleStatusConfirmed))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("company_id", companyID)
		c.Set("auth_claims", appjwt.Claims{UserID: changedByRaw, CompanyID: companyID.String(), Role: "admin", Kind: "owner"})
		c.Next()
	})
	router.PUT("/schedules/:id", handlerUnderTest.Update)

	body, err := json.Marshal(map[string]any{"status": "confirmed", "status_notes": "confirmado", "service_ids": []string{serviceID.String()}})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/schedules/"+scheduleID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.True(t, pub.scheduleCalled)
	require.Equal(t, scheduleID.String(), pub.scheduleCaptured.ScheduleID)
	require.Equal(t, companyID.String(), pub.scheduleCaptured.CompanyID)
	require.Equal(t, changedByRaw, pub.scheduleCaptured.ChangedBy)
	require.Equal(t, "confirmed", pub.scheduleCaptured.Status)
	require.Equal(t, "confirmado", pub.scheduleCaptured.StatusNotes)
	require.Equal(t, 2, pub.scheduleCaptured.Version)
	require.Equal(t, "Maria Silva", pub.scheduleCaptured.ClientName)
	require.Equal(t, "Thor", pub.scheduleCaptured.PetName)
	require.Equal(t, []string{"Banho"}, pub.scheduleCaptured.ServiceTitles)
	require.NotZero(t, pub.scheduleCaptured.OccurredAt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleHandler_CreateRejectsInvalidDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	queries, mock := domainServiceWithMock(t)
	defer mock.Close()

	serviceUnderTest := service.NewScheduleService(mock, queries)
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
