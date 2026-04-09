package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestScheduleService_CreateSchedule(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewScheduleService(queries)

	companyID := newDomainUUID(t)
	clientID := newDomainUUID(t)
	petID := newDomainUUID(t)
	creatorID := newDomainUUID(t)
	scheduleID := newDomainUUID(t)
	statusHistoryID := newDomainUUID(t)
	now := time.Now().UTC().Truncate(time.Second)
	end := now.Add(90 * time.Minute)

	mock.ExpectQuery(`(?s)name: ValidateScheduleOwnership`).
		WithArgs(petID, companyID, clientID).
		WillReturnRows(pgxmock.NewRows([]string{"is_valid"}).AddRow(true))

	mock.ExpectQuery(`(?s)name: CreateSchedule`).
		WithArgs(companyID, clientID, petID, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), creatorID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), now, end, "Banho e tosa", creatorID.String(), now, nil, nil))

	mock.ExpectQuery(`(?s)name: InsertScheduleStatusHistory`).
		WithArgs(scheduleID, sqlc.ScheduleStatusWaiting, creatorID, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "schedule_id", "status", "changed_at", "changed_by", "notes"}).
			AddRow(statusHistoryID.String(), scheduleID.String(), sqlc.ScheduleStatusWaiting, now, creatorID.String(), "status inicial"))

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), now, end, "Banho e tosa", creatorID.String(), now, nil, nil, sqlc.ScheduleStatusWaiting))

	result, err := serviceUnderTest.CreateSchedule(context.Background(), CreateScheduleInput{
		CompanyID:    companyID,
		ClientID:     clientID,
		PetID:        petID,
		ScheduledAt:  now,
		EstimatedEnd: &end,
		Notes:        "Banho e tosa",
		CreatedBy:    creatorID,
		Status:       sqlc.ScheduleStatusWaiting,
		StatusNotes:  "status inicial",
	})
	require.NoError(t, err)
	require.Equal(t, scheduleID, result.ID)
	require.Equal(t, sqlc.ScheduleStatusWaiting, result.CurrentStatus)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleService_CreateScheduleRejectsInvalidWindow(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewScheduleService(queries)

	companyID := newDomainUUID(t)
	clientID := newDomainUUID(t)
	petID := newDomainUUID(t)
	creatorID := newDomainUUID(t)
	now := time.Now().UTC()
	end := now.Add(-time.Minute)

	_, err = serviceUnderTest.CreateSchedule(context.Background(), CreateScheduleInput{
		CompanyID:    companyID,
		ClientID:     clientID,
		PetID:        petID,
		ScheduledAt:  now,
		EstimatedEnd: &end,
		CreatedBy:    creatorID,
	})
	require.ErrorIs(t, err, apperror.ErrUnprocessableEntity)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleService_UpdateScheduleRejectsInvalidStatus(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewScheduleService(queries)

	companyID := newDomainUUID(t)
	scheduleID := newDomainUUID(t)
	clientID := newDomainUUID(t)
	petID := newDomainUUID(t)
	now := time.Now().UTC().Truncate(time.Second)
	invalidStatus := sqlc.ScheduleStatus("not-valid")

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), now, nil, "", nil, now, nil, nil, sqlc.ScheduleStatusWaiting))

	mock.ExpectQuery(`(?s)name: ValidateScheduleOwnership`).
		WithArgs(petID, companyID, clientID).
		WillReturnRows(pgxmock.NewRows([]string{"is_valid"}).AddRow(true))

	_, err = serviceUnderTest.UpdateSchedule(context.Background(), UpdateScheduleInput{
		CompanyID:  companyID,
		ScheduleID: scheduleID,
		Status:     &invalidStatus,
		ChangedBy:  pgtype.UUID{},
	})
	require.ErrorIs(t, err, apperror.ErrUnprocessableEntity)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleService_DeleteScheduleNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewScheduleService(queries)

	companyID := newDomainUUID(t)
	scheduleID := newDomainUUID(t)

	mock.ExpectExec(`(?s)name: DeleteSchedule`).
		WithArgs(scheduleID, companyID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err = serviceUnderTest.DeleteSchedule(context.Background(), companyID, scheduleID)
	require.ErrorIs(t, err, apperror.ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScheduleService_ListScheduleStatusHistory(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewScheduleService(queries)

	companyID := newDomainUUID(t)
	scheduleID := newDomainUUID(t)
	clientID := newDomainUUID(t)
	petID := newDomainUUID(t)
	userID := newDomainUUID(t)
	now := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`(?s)name: GetScheduleByIDAndCompanyID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "client_id", "pet_id", "scheduled_at", "estimated_end", "notes", "created_by", "created_at", "updated_at", "deleted_at", "current_status"}).
			AddRow(scheduleID.String(), companyID.String(), clientID.String(), petID.String(), now, nil, "", userID.String(), now, nil, nil, sqlc.ScheduleStatusConfirmed))

	mock.ExpectQuery(`(?s)name: ListScheduleStatusHistoryByScheduleID`).
		WithArgs(scheduleID, companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "schedule_id", "status", "changed_at", "changed_by", "notes"}).
			AddRow(newDomainUUID(t).String(), scheduleID.String(), sqlc.ScheduleStatusWaiting, now.Add(-time.Hour), userID.String(), "created").
			AddRow(newDomainUUID(t).String(), scheduleID.String(), sqlc.ScheduleStatusConfirmed, now, userID.String(), "confirmed"))

	items, err := serviceUnderTest.ListScheduleStatusHistory(context.Background(), companyID, scheduleID)
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, sqlc.ScheduleStatusConfirmed, items[1].Status)
	require.NoError(t, mock.ExpectationsWereMet())
}
