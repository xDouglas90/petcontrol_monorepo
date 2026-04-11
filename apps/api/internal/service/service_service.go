package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type ServiceService struct {
	db      clientTxStarter
	queries *sqlc.Queries
}

type CreateServiceInput struct {
	CompanyID    pgtype.UUID
	TypeName     string
	Title        string
	Description  string
	Notes        pgtype.Text
	Price        pgtype.Numeric
	DiscountRate pgtype.Numeric
	ImageURL     pgtype.Text
	IsActive     bool
}

type UpdateServiceInput struct {
	CompanyID    pgtype.UUID
	ServiceID    pgtype.UUID
	TypeName     *string
	Title        *string
	Description  *string
	Notes        *string
	Price        *pgtype.Numeric
	DiscountRate *pgtype.Numeric
	ImageURL     *string
	IsActive     *bool
}

func NewServiceService(db clientTxStarter, queries *sqlc.Queries) *ServiceService {
	return &ServiceService{db: db, queries: queries}
}

func (s *ServiceService) ListServicesByCompanyID(ctx context.Context, companyID pgtype.UUID) ([]sqlc.ListServicesByCompanyIDRow, error) {
	return s.queries.ListServicesByCompanyID(ctx, companyID)
}

func (s *ServiceService) GetServiceByID(ctx context.Context, companyID pgtype.UUID, serviceID pgtype.UUID) (sqlc.GetServiceByIDAndCompanyIDRow, error) {
	item, err := s.queries.GetServiceByIDAndCompanyID(ctx, sqlc.GetServiceByIDAndCompanyIDParams{
		CompanyID: companyID,
		ID:        serviceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.GetServiceByIDAndCompanyIDRow{}, apperror.ErrNotFound
	}
	return item, err
}

func (s *ServiceService) CreateService(ctx context.Context, input CreateServiceInput) (sqlc.GetServiceByIDAndCompanyIDRow, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return sqlc.GetServiceByIDAndCompanyIDRow{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	txQueries := s.queries.WithTx(tx)
	typeID, err := resolveServiceTypeID(ctx, txQueries, input.TypeName)
	if err != nil {
		return sqlc.GetServiceByIDAndCompanyIDRow{}, err
	}

	created, err := txQueries.CreateService(ctx, sqlc.CreateServiceParams{
		TypeID:       typeID,
		Title:        input.Title,
		Description:  input.Description,
		Notes:        input.Notes,
		Price:        input.Price,
		DiscountRate: input.DiscountRate,
		ImageURL:     input.ImageURL,
		IsActive:     input.IsActive,
	})
	if err != nil {
		return sqlc.GetServiceByIDAndCompanyIDRow{}, mapServiceDBError(err)
	}

	_, err = txQueries.CreateCompanyService(ctx, sqlc.CreateCompanyServiceParams{
		CompanyID: input.CompanyID,
		ServiceID: created.ID,
	})
	if err != nil {
		return sqlc.GetServiceByIDAndCompanyIDRow{}, mapServiceDBError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.GetServiceByIDAndCompanyIDRow{}, err
	}
	committed = true

	return s.GetServiceByID(ctx, input.CompanyID, created.ID)
}

func (s *ServiceService) UpdateService(ctx context.Context, input UpdateServiceInput) (sqlc.GetServiceByIDAndCompanyIDRow, error) {
	if _, err := s.GetServiceByID(ctx, input.CompanyID, input.ServiceID); err != nil {
		return sqlc.GetServiceByIDAndCompanyIDRow{}, err
	}

	typeID := pgtype.UUID{}
	if input.TypeName != nil {
		resolvedTypeID, err := resolveServiceTypeID(ctx, s.queries, *input.TypeName)
		if err != nil {
			return sqlc.GetServiceByIDAndCompanyIDRow{}, err
		}
		typeID = resolvedTypeID
	}

	rows, err := s.queries.UpdateServiceByIDAndCompanyID(ctx, sqlc.UpdateServiceByIDAndCompanyIDParams{
		TypeID:       typeID,
		Title:        optionalText(input.Title),
		Description:  optionalText(input.Description),
		Notes:        optionalText(input.Notes),
		Price:        optionalNumeric(input.Price),
		DiscountRate: optionalNumeric(input.DiscountRate),
		ImageURL:     optionalText(input.ImageURL),
		IsActive:     optionalBool(input.IsActive),
		ID:           input.ServiceID,
		CompanyID:    input.CompanyID,
	})
	if err != nil {
		return sqlc.GetServiceByIDAndCompanyIDRow{}, mapServiceDBError(err)
	}
	if rows == 0 {
		return sqlc.GetServiceByIDAndCompanyIDRow{}, apperror.ErrNotFound
	}

	return s.GetServiceByID(ctx, input.CompanyID, input.ServiceID)
}

func (s *ServiceService) DeactivateService(ctx context.Context, companyID pgtype.UUID, serviceID pgtype.UUID) error {
	rows, err := s.queries.DeactivateCompanyService(ctx, sqlc.DeactivateCompanyServiceParams{
		CompanyID: companyID,
		ServiceID: serviceID,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func resolveServiceTypeID(ctx context.Context, queries sqlc.Querier, rawName string) (pgtype.UUID, error) {
	name := strings.TrimSpace(rawName)
	if name == "" {
		return pgtype.UUID{}, apperror.ErrUnprocessableEntity
	}

	serviceType, err := queries.FindServiceTypeByName(ctx, name)
	if err == nil {
		return serviceType.ID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, err
	}

	created, err := queries.CreateServiceType(ctx, sqlc.CreateServiceTypeParams{
		Name:        name,
		Description: pgtype.Text{},
	})
	if err != nil {
		return pgtype.UUID{}, mapServiceDBError(err)
	}

	return created.ID, nil
}

func optionalNumeric(value *pgtype.Numeric) pgtype.Numeric {
	if value == nil {
		return pgtype.Numeric{}
	}
	return *value
}

func mapServiceDBError(err error) error {
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
