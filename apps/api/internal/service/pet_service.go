package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/pagination"
)

type PetService struct {
	queries sqlc.Querier
}

type CreatePetInput struct {
	CompanyID   pgtype.UUID
	OwnerID     pgtype.UUID
	Name        string
	Size        sqlc.PetSize
	Kind        sqlc.PetKind
	Temperament sqlc.PetTemperament
	ImageURL    pgtype.Text
	BirthDate   pgtype.Date
	Notes       pgtype.Text
}

type UpdatePetInput struct {
	CompanyID   pgtype.UUID
	PetID       pgtype.UUID
	OwnerID     *pgtype.UUID
	Name        *string
	Size        *sqlc.PetSize
	Kind        *sqlc.PetKind
	Temperament *sqlc.PetTemperament
	ImageURL    *string
	BirthDate   *pgtype.Date
	Notes       *string
}

func NewPetService(queries sqlc.Querier) *PetService {
	return &PetService{queries: queries}
}

func (s *PetService) ListPetsByCompanyID(ctx context.Context, companyID pgtype.UUID, p pagination.Params) ([]sqlc.ListPetsByCompanyIDRow, error) {
	return s.queries.ListPetsByCompanyID(ctx, sqlc.ListPetsByCompanyIDParams{
		CompanyID: companyID,
		Search:    p.Search,
		Offset:    int32(p.Offset),
		Limit:     int32(p.Limit),
	})
}

func (s *PetService) GetPetByID(ctx context.Context, companyID pgtype.UUID, petID pgtype.UUID) (sqlc.GetPetByIDAndCompanyIDRow, error) {
	pet, err := s.queries.GetPetByIDAndCompanyID(ctx, sqlc.GetPetByIDAndCompanyIDParams{
		CompanyID: companyID,
		ID:        petID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.GetPetByIDAndCompanyIDRow{}, apperror.ErrNotFound
	}
	return pet, err
}

func (s *PetService) CreatePet(ctx context.Context, input CreatePetInput) (sqlc.GetPetByIDAndCompanyIDRow, error) {
	if err := s.validateOwner(ctx, input.CompanyID, input.OwnerID); err != nil {
		return sqlc.GetPetByIDAndCompanyIDRow{}, err
	}

	created, err := s.queries.CreatePet(ctx, sqlc.CreatePetParams{
		Name:        input.Name,
		Size:        input.Size,
		Kind:        input.Kind,
		Temperament: input.Temperament,
		ImageUrl:    input.ImageURL,
		BirthDate:   input.BirthDate,
		OwnerID:     input.OwnerID,
		Notes:       input.Notes,
	})
	if err != nil {
		return sqlc.GetPetByIDAndCompanyIDRow{}, mapPetDBError(err)
	}

	return s.GetPetByID(ctx, input.CompanyID, created.ID)
}

func (s *PetService) UpdatePet(ctx context.Context, input UpdatePetInput) (sqlc.GetPetByIDAndCompanyIDRow, error) {
	current, err := s.GetPetByID(ctx, input.CompanyID, input.PetID)
	if err != nil {
		return sqlc.GetPetByIDAndCompanyIDRow{}, err
	}

	ownerID := current.OwnerID
	if input.OwnerID != nil {
		ownerID = *input.OwnerID
	}
	if err := s.validateOwner(ctx, input.CompanyID, ownerID); err != nil {
		return sqlc.GetPetByIDAndCompanyIDRow{}, err
	}

	rows, err := s.queries.UpdatePet(ctx, sqlc.UpdatePetParams{
		OwnerID: func() pgtype.UUID {
			if input.OwnerID == nil {
				return pgtype.UUID{}
			}
			return *input.OwnerID
		}(),
		Name:        optionalText(input.Name),
		Size:        optionalPetSize(input.Size),
		Kind:        optionalPetKind(input.Kind),
		Temperament: optionalPetTemperament(input.Temperament),
		ImageUrl:    optionalText(input.ImageURL),
		BirthDate:   optionalDate(input.BirthDate),
		Notes:       optionalText(input.Notes),
		ID:          input.PetID,
	})
	if err != nil {
		return sqlc.GetPetByIDAndCompanyIDRow{}, mapPetDBError(err)
	}
	if rows == 0 {
		return sqlc.GetPetByIDAndCompanyIDRow{}, apperror.ErrNotFound
	}

	return s.GetPetByID(ctx, input.CompanyID, input.PetID)
}

func (s *PetService) DeletePet(ctx context.Context, companyID pgtype.UUID, petID pgtype.UUID) error {
	if _, err := s.GetPetByID(ctx, companyID, petID); err != nil {
		return err
	}

	rows, err := s.queries.DeletePet(ctx, petID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (s *PetService) validateOwner(ctx context.Context, companyID pgtype.UUID, ownerID pgtype.UUID) error {
	isValid, err := s.queries.ValidatePetOwnerByCompany(ctx, sqlc.ValidatePetOwnerByCompanyParams{
		CompanyID: companyID,
		OwnerID:   ownerID,
	})
	if err != nil {
		return err
	}
	if !isValid {
		return apperror.ErrUnprocessableEntity
	}
	return nil
}

func optionalPetSize(value *sqlc.PetSize) sqlc.NullPetSize {
	if value == nil {
		return sqlc.NullPetSize{}
	}
	return sqlc.NullPetSize{PetSize: *value, Valid: true}
}

func optionalPetKind(value *sqlc.PetKind) sqlc.NullPetKind {
	if value == nil {
		return sqlc.NullPetKind{}
	}
	return sqlc.NullPetKind{PetKind: *value, Valid: true}
}

func optionalPetTemperament(value *sqlc.PetTemperament) sqlc.NullPetTemperament {
	if value == nil {
		return sqlc.NullPetTemperament{}
	}
	return sqlc.NullPetTemperament{PetTemperament: *value, Valid: true}
}

func mapPetDBError(err error) error {
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
