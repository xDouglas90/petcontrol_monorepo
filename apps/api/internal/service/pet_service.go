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
	CompanyID               pgtype.UUID
	OwnerID                 pgtype.UUID
	GuardianIDs             []pgtype.UUID
	Name                    string
	Race                    string
	Color                   string
	Sex                     string
	Size                    sqlc.PetSize
	Kind                    sqlc.PetKind
	Temperament             sqlc.PetTemperament
	ImageURL                pgtype.Text
	BirthDate               pgtype.Date
	IsActive                pgtype.Bool
	IsDeceased              pgtype.Bool
	IsVaccinated            pgtype.Bool
	IsNeutered              pgtype.Bool
	IsMicrochipped          pgtype.Bool
	MicrochipNumber         pgtype.Text
	MicrochipExpirationDate pgtype.Date
	Notes                   pgtype.Text
}

type UpdatePetInput struct {
	CompanyID               pgtype.UUID
	PetID                   pgtype.UUID
	OwnerID                 *pgtype.UUID
	GuardianIDs             *[]pgtype.UUID
	Name                    *string
	Race                    *string
	Color                   *string
	Sex                     *string
	Size                    *sqlc.PetSize
	Kind                    *sqlc.PetKind
	Temperament             *sqlc.PetTemperament
	ImageURL                *string
	BirthDate               *pgtype.Date
	IsActive                *bool
	IsDeceased              *bool
	IsVaccinated            *bool
	IsNeutered              *bool
	IsMicrochipped          *bool
	MicrochipNumber         *string
	MicrochipExpirationDate *pgtype.Date
	Notes                   *string
}

type PetFilters struct {
	Size        *sqlc.PetSize
	Kind        *sqlc.PetKind
	Temperament *sqlc.PetTemperament
	Race        *string
	IsActive    *bool
}

func NewPetService(queries sqlc.Querier) *PetService {
	return &PetService{queries: queries}
}

func (s *PetService) ListPetsByCompanyID(ctx context.Context, companyID pgtype.UUID, p pagination.Params, filters PetFilters) ([]sqlc.ListPetsByCompanyIDRow, error) {
	return s.queries.ListPetsByCompanyID(ctx, sqlc.ListPetsByCompanyIDParams{
		CompanyID:   companyID,
		Search:      p.Search,
		Offset:      int32(p.Offset),
		Limit:       int32(p.Limit),
		Size:        optionalPetSizeFilter(filters.Size),
		Kind:        optionalPetKindFilter(filters.Kind),
		Temperament: optionalPetTemperamentFilter(filters.Temperament),
		Race:        optionalTextFilter(filters.Race),
		IsActive:    optionalBoolFilter(filters.IsActive),
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

func (s *PetService) GetPetDetailByID(ctx context.Context, companyID pgtype.UUID, petID pgtype.UUID) (sqlc.GetPetDetailByIDAndCompanyIDRow, error) {
	pet, err := s.queries.GetPetDetailByIDAndCompanyID(ctx, sqlc.GetPetDetailByIDAndCompanyIDParams{
		CompanyID: companyID,
		ID:        petID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.GetPetDetailByIDAndCompanyIDRow{}, apperror.ErrNotFound
	}
	return pet, err
}

func (s *PetService) GetPetGuardians(ctx context.Context, companyID pgtype.UUID, petID pgtype.UUID) ([]sqlc.ListPetGuardiansByPetIDRow, error) {
	return s.queries.ListPetGuardiansByPetID(ctx, sqlc.ListPetGuardiansByPetIDParams{
		PetID:     petID,
		CompanyID: companyID,
	})
}

func (s *PetService) CreatePet(ctx context.Context, input CreatePetInput) (sqlc.GetPetDetailByIDAndCompanyIDRow, error) {
	if err := s.validateOwner(ctx, input.CompanyID, input.OwnerID); err != nil {
		return sqlc.GetPetDetailByIDAndCompanyIDRow{}, err
	}
	if input.GuardianIDs != nil {
		if err := s.validateGuardians(ctx, input.CompanyID, input.GuardianIDs); err != nil {
			return sqlc.GetPetDetailByIDAndCompanyIDRow{}, err
		}
	}

	created, err := s.queries.CreatePet(ctx, sqlc.CreatePetParams{
		Name:                    input.Name,
		Race:                    input.Race,
		Color:                   input.Color,
		Sex:                     input.Sex,
		Size:                    input.Size,
		Kind:                    input.Kind,
		Temperament:             input.Temperament,
		ImageUrl:                input.ImageURL,
		BirthDate:               input.BirthDate,
		OwnerID:                 input.OwnerID,
		IsActive:                input.IsActive,
		IsDeceased:              input.IsDeceased,
		IsVaccinated:            input.IsVaccinated,
		IsNeutered:              input.IsNeutered,
		IsMicrochipped:          input.IsMicrochipped,
		MicrochipNumber:         input.MicrochipNumber,
		MicrochipExpirationDate: input.MicrochipExpirationDate,
		Notes:                   input.Notes,
	})
	if err != nil {
		return sqlc.GetPetDetailByIDAndCompanyIDRow{}, mapPetDBError(err)
	}
	if input.GuardianIDs != nil {
		if err := s.replaceGuardians(ctx, created.ID, input.GuardianIDs); err != nil {
			return sqlc.GetPetDetailByIDAndCompanyIDRow{}, err
		}
	}

	return s.GetPetDetailByID(ctx, input.CompanyID, created.ID)
}

func (s *PetService) UpdatePet(ctx context.Context, input UpdatePetInput) (sqlc.GetPetDetailByIDAndCompanyIDRow, error) {
	current, err := s.GetPetByID(ctx, input.CompanyID, input.PetID)
	if err != nil {
		return sqlc.GetPetDetailByIDAndCompanyIDRow{}, err
	}

	ownerID := current.OwnerID
	if input.OwnerID != nil {
		ownerID = *input.OwnerID
	}
	if err := s.validateOwner(ctx, input.CompanyID, ownerID); err != nil {
		return sqlc.GetPetDetailByIDAndCompanyIDRow{}, err
	}
	if input.GuardianIDs != nil {
		if err := s.validateGuardians(ctx, input.CompanyID, *input.GuardianIDs); err != nil {
			return sqlc.GetPetDetailByIDAndCompanyIDRow{}, err
		}
	}

	rows, err := s.queries.UpdatePet(ctx, sqlc.UpdatePetParams{
		OwnerID: func() pgtype.UUID {
			if input.OwnerID == nil {
				return pgtype.UUID{}
			}
			return *input.OwnerID
		}(),
		Name:                    optionalText(input.Name),
		Race:                    optionalText(input.Race),
		Color:                   optionalText(input.Color),
		Sex:                     optionalText(input.Sex),
		Size:                    optionalPetSize(input.Size),
		Kind:                    optionalPetKind(input.Kind),
		Temperament:             optionalPetTemperament(input.Temperament),
		ImageUrl:                optionalText(input.ImageURL),
		BirthDate:               optionalDate(input.BirthDate),
		IsActive:                optionalBoolFilter(input.IsActive),
		IsDeceased:              optionalBoolFilter(input.IsDeceased),
		IsVaccinated:            optionalBoolFilter(input.IsVaccinated),
		IsNeutered:              optionalBoolFilter(input.IsNeutered),
		IsMicrochipped:          optionalBoolFilter(input.IsMicrochipped),
		MicrochipNumber:         optionalText(input.MicrochipNumber),
		MicrochipExpirationDate: optionalDate(input.MicrochipExpirationDate),
		Notes:                   optionalText(input.Notes),
		ID:                      input.PetID,
	})
	if err != nil {
		return sqlc.GetPetDetailByIDAndCompanyIDRow{}, mapPetDBError(err)
	}
	if rows == 0 {
		return sqlc.GetPetDetailByIDAndCompanyIDRow{}, apperror.ErrNotFound
	}
	if input.GuardianIDs != nil {
		if err := s.replaceGuardians(ctx, input.PetID, *input.GuardianIDs); err != nil {
			return sqlc.GetPetDetailByIDAndCompanyIDRow{}, err
		}
	}

	return s.GetPetDetailByID(ctx, input.CompanyID, input.PetID)
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

func (s *PetService) validateGuardians(ctx context.Context, companyID pgtype.UUID, guardianIDs []pgtype.UUID) error {
	for _, guardianID := range uniquePetGuardianUUIDs(guardianIDs) {
		isValid, err := s.queries.ValidatePetGuardianByCompany(ctx, sqlc.ValidatePetGuardianByCompanyParams{
			CompanyID:  companyID,
			GuardianID: guardianID,
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

func (s *PetService) replaceGuardians(ctx context.Context, petID pgtype.UUID, guardianIDs []pgtype.UUID) error {
	if _, err := s.queries.DeletePetGuardiansByPetID(ctx, petID); err != nil {
		return err
	}
	for _, guardianID := range uniquePetGuardianUUIDs(guardianIDs) {
		if _, err := s.queries.UpsertPetGuardian(ctx, sqlc.UpsertPetGuardianParams{
			PetID:      petID,
			GuardianID: guardianID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func uniquePetGuardianUUIDs(values []pgtype.UUID) []pgtype.UUID {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[[16]byte]struct{}, len(values))
	result := make([]pgtype.UUID, 0, len(values))
	for _, value := range values {
		if !value.Valid {
			continue
		}
		if _, ok := seen[value.Bytes]; ok {
			continue
		}
		seen[value.Bytes] = struct{}{}
		result = append(result, value)
	}
	return result
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

func optionalPetSizeFilter(value *sqlc.PetSize) sqlc.NullPetSize {
	if value == nil {
		return sqlc.NullPetSize{}
	}
	return sqlc.NullPetSize{PetSize: *value, Valid: true}
}

func optionalPetKindFilter(value *sqlc.PetKind) sqlc.NullPetKind {
	if value == nil {
		return sqlc.NullPetKind{}
	}
	return sqlc.NullPetKind{PetKind: *value, Valid: true}
}

func optionalPetTemperamentFilter(value *sqlc.PetTemperament) sqlc.NullPetTemperament {
	if value == nil {
		return sqlc.NullPetTemperament{}
	}
	return sqlc.NullPetTemperament{PetTemperament: *value, Valid: true}
}

func optionalTextFilter(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalBoolFilter(value *bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *value, Valid: true}
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
