package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type ClientService struct {
	db      clientTxStarter
	queries *sqlc.Queries
}

type UpdateClientInput struct {
	CompanyID      pgtype.UUID
	ClientID       pgtype.UUID
	FullName       *string
	ShortName      *string
	GenderIdentity *sqlc.GenderIdentity
	MaritalStatus  *sqlc.MaritalStatus
	BirthDate      *pgtype.Date
	CPF            *string
	Email          *string
	Phone          *string
	Cellphone      *string
	HasWhatsapp    *bool
	ClientSince    *pgtype.Date
	Notes          *string
}

type CreateClientInput struct {
	CompanyID      pgtype.UUID
	FullName       string
	ShortName      string
	GenderIdentity sqlc.GenderIdentity
	MaritalStatus  sqlc.MaritalStatus
	BirthDate      pgtype.Date
	CPF            string
	Email          string
	Phone          pgtype.Text
	Cellphone      string
	HasWhatsapp    bool
	ClientSince    pgtype.Date
	Notes          pgtype.Text
}

type clientTxStarter interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewClientService(db clientTxStarter, queries *sqlc.Queries) *ClientService {
	return &ClientService{db: db, queries: queries}
}

func (s *ClientService) ListClientsByCompanyID(ctx context.Context, companyID pgtype.UUID) ([]sqlc.ListClientsByCompanyIDRow, error) {
	return s.queries.ListClientsByCompanyID(ctx, companyID)
}

func (s *ClientService) GetClientByID(ctx context.Context, companyID pgtype.UUID, clientID pgtype.UUID) (sqlc.GetClientByIDAndCompanyIDRow, error) {
	client, err := s.queries.GetClientByIDAndCompanyID(ctx, sqlc.GetClientByIDAndCompanyIDParams{
		CompanyID: companyID,
		ID:        clientID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.GetClientByIDAndCompanyIDRow{}, apperror.ErrNotFound
	}
	return client, err
}

func (s *ClientService) CreateClient(ctx context.Context, params CreateClientInput) (sqlc.GetClientByIDAndCompanyIDRow, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	txQueries := s.queries.WithTx(tx)

	person, err := txQueries.InsertClientPerson(ctx)
	if err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, mapClientDBError(err)
	}

	_, err = txQueries.InsertClientIdentification(ctx, sqlc.InsertClientIdentificationParams{
		PersonID:       person.ID,
		FullName:       params.FullName,
		ShortName:      params.ShortName,
		GenderIdentity: params.GenderIdentity,
		MaritalStatus:  params.MaritalStatus,
		BirthDate:      params.BirthDate,
		CPF:            params.CPF,
	})
	if err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, mapClientDBError(err)
	}

	_, err = txQueries.InsertClientPrimaryContact(ctx, sqlc.InsertClientPrimaryContactParams{
		PersonID:    person.ID,
		Email:       params.Email,
		Phone:       params.Phone,
		Cellphone:   params.Cellphone,
		HasWhatsapp: params.HasWhatsapp,
	})
	if err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, mapClientDBError(err)
	}

	client, err := txQueries.InsertClientRecord(ctx, sqlc.InsertClientRecordParams{
		PersonID:    person.ID,
		ClientSince: params.ClientSince,
		Notes:       params.Notes,
	})
	if err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, mapClientDBError(err)
	}

	_, err = txQueries.CreateCompanyClient(ctx, sqlc.CreateCompanyClientParams{
		CompanyID: params.CompanyID,
		ClientID:  client.ID,
	})
	if err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, mapClientDBError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, err
	}
	committed = true

	return s.GetClientByID(ctx, params.CompanyID, client.ID)
}

func (s *ClientService) UpdateClient(ctx context.Context, input UpdateClientInput) (sqlc.GetClientByIDAndCompanyIDRow, error) {
	if _, err := s.GetClientByID(ctx, input.CompanyID, input.ClientID); err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, err
	}

	rows, err := s.queries.UpdateClientIdentification(ctx, sqlc.UpdateClientIdentificationParams{
		FullName:       optionalText(input.FullName),
		ShortName:      optionalText(input.ShortName),
		GenderIdentity: optionalGenderIdentity(input.GenderIdentity),
		MaritalStatus:  optionalMaritalStatus(input.MaritalStatus),
		BirthDate:      optionalDate(input.BirthDate),
		CPF:            optionalText(input.CPF),
		ID:             input.ClientID,
		CompanyID:      input.CompanyID,
	})
	if err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, mapClientDBError(err)
	}
	if rows == 0 {
		return sqlc.GetClientByIDAndCompanyIDRow{}, apperror.ErrNotFound
	}

	_, err = s.queries.UpdateClientPrimaryContact(ctx, sqlc.UpdateClientPrimaryContactParams{
		Email:       optionalText(input.Email),
		Phone:       optionalText(input.Phone),
		Cellphone:   optionalText(input.Cellphone),
		HasWhatsapp: optionalBool(input.HasWhatsapp),
		ID:          input.ClientID,
		CompanyID:   input.CompanyID,
	})
	if err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, mapClientDBError(err)
	}

	_, err = s.queries.UpdateClientRecord(ctx, sqlc.UpdateClientRecordParams{
		ClientSince: optionalDate(input.ClientSince),
		Notes:       optionalText(input.Notes),
		ID:          input.ClientID,
		CompanyID:   input.CompanyID,
	})
	if err != nil {
		return sqlc.GetClientByIDAndCompanyIDRow{}, mapClientDBError(err)
	}

	return s.GetClientByID(ctx, input.CompanyID, input.ClientID)
}

func (s *ClientService) DeactivateClient(ctx context.Context, companyID pgtype.UUID, clientID pgtype.UUID) error {
	rows, err := s.queries.DeactivateClient(ctx, sqlc.DeactivateClientParams{
		CompanyID: companyID,
		ClientID:  clientID,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func optionalText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return toText(*value)
}

func optionalDate(value *pgtype.Date) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return *value
}

func optionalBool(value *bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *value, Valid: true}
}

func optionalGenderIdentity(value *sqlc.GenderIdentity) sqlc.NullGenderIdentity {
	if value == nil {
		return sqlc.NullGenderIdentity{}
	}
	return sqlc.NullGenderIdentity{GenderIdentity: *value, Valid: true}
}

func optionalMaritalStatus(value *sqlc.MaritalStatus) sqlc.NullMaritalStatus {
	if value == nil {
		return sqlc.NullMaritalStatus{}
	}
	return sqlc.NullMaritalStatus{MaritalStatus: *value, Valid: true}
}

func mapClientDBError(err error) error {
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
