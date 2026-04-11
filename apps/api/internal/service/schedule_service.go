package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/pagination"
)

type ScheduleService struct {
	db      clientTxStarter
	queries *sqlc.Queries
}

type CreateScheduleInput struct {
	CompanyID    pgtype.UUID
	ClientID     pgtype.UUID
	PetID        pgtype.UUID
	ServiceIDs   []pgtype.UUID
	ScheduledAt  time.Time
	EstimatedEnd *time.Time
	Notes        string
	CreatedBy    pgtype.UUID
	Status       sqlc.ScheduleStatus
	StatusNotes  string
}

type UpdateScheduleInput struct {
	CompanyID    pgtype.UUID
	ScheduleID   pgtype.UUID
	ClientID     *pgtype.UUID
	PetID        *pgtype.UUID
	ServiceIDs   *[]pgtype.UUID
	ScheduledAt  *time.Time
	EstimatedEnd *time.Time
	Notes        *string
	Status       *sqlc.ScheduleStatus
	StatusNotes  *string
	ChangedBy    pgtype.UUID
}

func NewScheduleService(db clientTxStarter, queries *sqlc.Queries) *ScheduleService {
	return &ScheduleService{db: db, queries: queries}
}

func (s *ScheduleService) ListSchedulesByCompanyID(ctx context.Context, companyID pgtype.UUID, p pagination.Params) ([]sqlc.ListSchedulesByCompanyIDRow, error) {
	return s.queries.ListSchedulesByCompanyID(ctx, sqlc.ListSchedulesByCompanyIDParams{
		CompanyID: companyID,
		Search:    p.Search,
		Offset:    int32(p.Offset),
		Limit:     int32(p.Limit),
	})
}

func (s *ScheduleService) GetScheduleByID(ctx context.Context, companyID pgtype.UUID, scheduleID pgtype.UUID) (sqlc.GetScheduleByIDAndCompanyIDRow, error) {
	schedule, err := s.queries.GetScheduleByIDAndCompanyID(ctx, sqlc.GetScheduleByIDAndCompanyIDParams{
		ID:        scheduleID,
		CompanyID: companyID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, apperror.ErrNotFound
	}
	return schedule, err
}

func (s *ScheduleService) ListScheduleStatusHistory(ctx context.Context, companyID pgtype.UUID, scheduleID pgtype.UUID) ([]sqlc.ScheduleStatusHistory, error) {
	if _, err := s.GetScheduleByID(ctx, companyID, scheduleID); err != nil {
		return nil, err
	}

	return s.queries.ListScheduleStatusHistoryByScheduleID(ctx, sqlc.ListScheduleStatusHistoryByScheduleIDParams{
		ScheduleID: scheduleID,
		CompanyID:  companyID,
	})
}

func (s *ScheduleService) CreateSchedule(ctx context.Context, input CreateScheduleInput) (sqlc.GetScheduleByIDAndCompanyIDRow, error) {
	if err := validateScheduleWindow(input.ScheduledAt, input.EstimatedEnd); err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}

	status := input.Status
	if status == "" {
		status = sqlc.ScheduleStatusWaiting
	}
	if !isValidScheduleStatus(status) {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, apperror.ErrUnprocessableEntity
	}

	if err := s.validateOwnership(ctx, input.CompanyID, input.ClientID, input.PetID); err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}

	if err := s.validateServices(ctx, input.CompanyID, input.ServiceIDs); err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	txQueries := s.queries.WithTx(tx)

	created, err := txQueries.CreateSchedule(ctx, sqlc.CreateScheduleParams{
		CompanyID:   input.CompanyID,
		ClientID:    input.ClientID,
		PetID:       input.PetID,
		ScheduledAt: toTimestamptz(input.ScheduledAt),
		EstimatedEnd: func() pgtype.Timestamptz {
			if input.EstimatedEnd == nil {
				return pgtype.Timestamptz{}
			}
			return toTimestamptz(*input.EstimatedEnd)
		}(),
		Notes:     toText(input.Notes),
		CreatedBy: input.CreatedBy,
	})
	if err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, mapScheduleDBError(err)
	}

	_, err = txQueries.InsertScheduleStatusHistory(ctx, sqlc.InsertScheduleStatusHistoryParams{
		ScheduleID: created.ID,
		Status:     status,
		ChangedBy:  input.CreatedBy,
		Notes:      toText(input.StatusNotes),
	})
	if err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}

	if err := s.replaceScheduleServices(ctx, txQueries, created.ID, input.ServiceIDs); err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}
	committed = true

	return s.GetScheduleByID(ctx, input.CompanyID, created.ID)
}

func (s *ScheduleService) UpdateSchedule(ctx context.Context, input UpdateScheduleInput) (sqlc.GetScheduleByIDAndCompanyIDRow, error) {
	current, err := s.GetScheduleByID(ctx, input.CompanyID, input.ScheduleID)
	if err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}

	clientID := current.ClientID
	if input.ClientID != nil {
		clientID = *input.ClientID
	}

	petID := current.PetID
	if input.PetID != nil {
		petID = *input.PetID
	}

	scheduledAt := current.ScheduledAt.Time
	if input.ScheduledAt != nil {
		scheduledAt = input.ScheduledAt.UTC()
	}

	var estimatedEnd *time.Time
	if current.EstimatedEnd.Valid {
		value := current.EstimatedEnd.Time
		estimatedEnd = &value
	}
	if input.EstimatedEnd != nil {
		value := input.EstimatedEnd.UTC()
		estimatedEnd = &value
	}

	if err := validateScheduleWindow(scheduledAt, estimatedEnd); err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}

	if err := s.validateOwnership(ctx, input.CompanyID, clientID, petID); err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}

	serviceIDs := currentServiceIDs(current.ServiceIds)
	if input.ServiceIDs != nil {
		serviceIDs = uniqueUUIDs(*input.ServiceIDs)
	}
	if err := s.validateServices(ctx, input.CompanyID, serviceIDs); err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}

	statusChanged := input.Status != nil
	if statusChanged && !isValidScheduleStatus(*input.Status) {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, apperror.ErrUnprocessableEntity
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	txQueries := s.queries.WithTx(tx)

	hasDirectUpdate := input.ClientID != nil || input.PetID != nil || input.ScheduledAt != nil || input.EstimatedEnd != nil || input.Notes != nil
	if hasDirectUpdate {
		updatedRows, err := txQueries.UpdateSchedule(ctx, sqlc.UpdateScheduleParams{
			ClientID: func() pgtype.UUID {
				if input.ClientID == nil {
					return pgtype.UUID{}
				}
				return *input.ClientID
			}(),
			PetID: func() pgtype.UUID {
				if input.PetID == nil {
					return pgtype.UUID{}
				}
				return *input.PetID
			}(),
			ScheduledAt: func() pgtype.Timestamptz {
				if input.ScheduledAt == nil {
					return pgtype.Timestamptz{}
				}
				return toTimestamptz(*input.ScheduledAt)
			}(),
			EstimatedEnd: func() pgtype.Timestamptz {
				if input.EstimatedEnd == nil {
					return pgtype.Timestamptz{}
				}
				return toTimestamptz(*input.EstimatedEnd)
			}(),
			Notes: func() pgtype.Text {
				if input.Notes == nil {
					return pgtype.Text{}
				}
				return toText(*input.Notes)
			}(),
			ID:        input.ScheduleID,
			CompanyID: input.CompanyID,
		})
		if err != nil {
			return sqlc.GetScheduleByIDAndCompanyIDRow{}, mapScheduleDBError(err)
		}
		if updatedRows == 0 {
			return sqlc.GetScheduleByIDAndCompanyIDRow{}, apperror.ErrNotFound
		}
	}

	if statusChanged {
		_, err := txQueries.InsertScheduleStatusHistory(ctx, sqlc.InsertScheduleStatusHistoryParams{
			ScheduleID: input.ScheduleID,
			Status:     *input.Status,
			ChangedBy:  input.ChangedBy,
			Notes: func() pgtype.Text {
				if input.StatusNotes == nil {
					return pgtype.Text{}
				}
				return toText(*input.StatusNotes)
			}(),
		})
		if err != nil {
			return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
		}
	}

	if input.ServiceIDs != nil {
		if err := s.replaceScheduleServices(ctx, txQueries, input.ScheduleID, serviceIDs); err != nil {
			return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.GetScheduleByIDAndCompanyIDRow{}, err
	}
	committed = true

	return s.GetScheduleByID(ctx, input.CompanyID, input.ScheduleID)
}

func (s *ScheduleService) DeleteSchedule(ctx context.Context, companyID pgtype.UUID, scheduleID pgtype.UUID) error {
	rows, err := s.queries.DeleteSchedule(ctx, sqlc.DeleteScheduleParams{
		ID:        scheduleID,
		CompanyID: companyID,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (s *ScheduleService) validateOwnership(ctx context.Context, companyID pgtype.UUID, clientID pgtype.UUID, petID pgtype.UUID) error {
	isValid, err := s.queries.ValidateScheduleOwnership(ctx, sqlc.ValidateScheduleOwnershipParams{
		PetID:     petID,
		CompanyID: companyID,
		ClientID:  clientID,
	})
	if err != nil {
		return err
	}
	if !isValid {
		return apperror.ErrUnprocessableEntity
	}
	return nil
}

func (s *ScheduleService) validateServices(ctx context.Context, companyID pgtype.UUID, serviceIDs []pgtype.UUID) error {
	seen := map[string]struct{}{}
	for _, serviceID := range serviceIDs {
		key := serviceID.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		isValid, err := s.queries.ValidateServiceByIDAndCompanyID(ctx, sqlc.ValidateServiceByIDAndCompanyIDParams{
			CompanyID: companyID,
			ServiceID: serviceID,
		})
		if err != nil {
			return err
		}
		if !isValid {
			return apperror.ErrUnprocessableEntity
		}
	}
	return nil
}

func (s *ScheduleService) replaceScheduleServices(ctx context.Context, queries *sqlc.Queries, scheduleID pgtype.UUID, serviceIDs []pgtype.UUID) error {
	if _, err := queries.DeleteScheduleServicesByScheduleID(ctx, scheduleID); err != nil {
		return err
	}

	for _, serviceID := range uniqueUUIDs(serviceIDs) {
		if _, err := queries.InsertScheduleService(ctx, sqlc.InsertScheduleServiceParams{
			ScheduleID: scheduleID,
			ServiceID:  serviceID,
		}); err != nil {
			return mapScheduleDBError(err)
		}
	}

	return nil
}

func uniqueUUIDs(items []pgtype.UUID) []pgtype.UUID {
	if len(items) == 0 {
		return nil
	}

	seen := map[string]struct{}{}
	unique := make([]pgtype.UUID, 0, len(items))
	for _, item := range items {
		key := item.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, item)
	}
	return unique
}

func currentServiceIDs(raw interface{}) []pgtype.UUID {
	values := stringSliceFromAny(raw)
	if len(values) == 0 {
		return nil
	}

	serviceIDs := make([]pgtype.UUID, 0, len(values))
	for _, value := range values {
		var parsed pgtype.UUID
		err := parsed.Scan(value)
		if err != nil {
			continue
		}
		serviceIDs = append(serviceIDs, parsed)
	}
	return serviceIDs
}

func validateScheduleWindow(scheduledAt time.Time, estimatedEnd *time.Time) error {
	if scheduledAt.IsZero() {
		return apperror.ErrUnprocessableEntity
	}
	if estimatedEnd == nil {
		return nil
	}
	if !estimatedEnd.After(scheduledAt) {
		return apperror.ErrUnprocessableEntity
	}
	return nil
}

func isValidScheduleStatus(status sqlc.ScheduleStatus) bool {
	switch status {
	case sqlc.ScheduleStatusWaiting,
		sqlc.ScheduleStatusConfirmed,
		sqlc.ScheduleStatusCanceled,
		sqlc.ScheduleStatusInProgress,
		sqlc.ScheduleStatusFinished,
		sqlc.ScheduleStatusDelivered:
		return true
	default:
		return false
	}
}

func toTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func toText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func mapScheduleDBError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return apperror.ErrUnprocessableEntity
		case "23505":
			return apperror.ErrConflict
		}
	}
	return err
}
