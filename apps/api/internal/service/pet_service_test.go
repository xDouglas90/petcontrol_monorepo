package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestPetService_CreatePet(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewPetService(queries)

	companyID := newDomainUUID(t)
	ownerID := newDomainUUID(t)
	petID := newDomainUUID(t)
	now := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`(?s)name: ValidatePetOwnerByCompany`).
		WithArgs(companyID, ownerID).
		WillReturnRows(pgxmock.NewRows([]string{"is_valid"}).AddRow(true))

	mock.ExpectQuery(`(?s)name: CreatePet`).
		WithArgs("Thor", sqlc.PetSizeMedium, sqlc.PetKindDog, sqlc.PetTemperamentPlayful, pgxmock.AnyArg(), pgxmock.AnyArg(), ownerID, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "size", "kind", "temperament", "image_url", "birth_date", "owner_id", "is_active", "notes", "created_at", "updated_at", "deleted_at"}).
			AddRow(petID.String(), "Thor", sqlc.PetSizeMedium, sqlc.PetKindDog, sqlc.PetTemperamentPlayful, nil, time.Date(2021, 8, 20, 0, 0, 0, 0, time.UTC), ownerID.String(), true, "Gosta de brincar", now, nil, nil))

	mock.ExpectQuery(`(?s)name: GetPetByIDAndCompanyID`).
		WithArgs(companyID, petID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "owner_id", "company_id", "owner_name", "name", "size", "kind", "temperament", "image_url", "birth_date", "is_active", "notes", "created_at", "updated_at", "deleted_at"}).
			AddRow(petID.String(), ownerID.String(), companyID.String(), "Maria Silva", "Thor", sqlc.PetSizeMedium, sqlc.PetKindDog, sqlc.PetTemperamentPlayful, nil, time.Date(2021, 8, 20, 0, 0, 0, 0, time.UTC), true, "Gosta de brincar", now, nil, nil))

	result, err := serviceUnderTest.CreatePet(context.Background(), CreatePetInput{
		CompanyID:   companyID,
		OwnerID:     ownerID,
		Name:        "Thor",
		Size:        sqlc.PetSizeMedium,
		Kind:        sqlc.PetKindDog,
		Temperament: sqlc.PetTemperamentPlayful,
		BirthDate:   pgtype.Date{Time: time.Date(2021, 8, 20, 0, 0, 0, 0, time.UTC), Valid: true},
		Notes:       pgtype.Text{String: "Gosta de brincar", Valid: true},
	})
	require.NoError(t, err)
	require.Equal(t, petID, result.ID)
	require.Equal(t, "Maria Silva", result.OwnerName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPetService_CreatePetRejectsOwnerOutsideTenant(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewPetService(queries)

	companyID := newDomainUUID(t)
	ownerID := newDomainUUID(t)

	mock.ExpectQuery(`(?s)name: ValidatePetOwnerByCompany`).
		WithArgs(companyID, ownerID).
		WillReturnRows(pgxmock.NewRows([]string{"is_valid"}).AddRow(false))

	_, err = serviceUnderTest.CreatePet(context.Background(), CreatePetInput{
		CompanyID:   companyID,
		OwnerID:     ownerID,
		Name:        "Thor",
		Size:        sqlc.PetSizeMedium,
		Kind:        sqlc.PetKindDog,
		Temperament: sqlc.PetTemperamentPlayful,
	})
	require.ErrorIs(t, err, apperror.ErrUnprocessableEntity)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPetService_DeletePetNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewPetService(queries)

	companyID := newDomainUUID(t)
	petID := newDomainUUID(t)

	mock.ExpectQuery(`(?s)name: GetPetByIDAndCompanyID`).
		WithArgs(companyID, petID).
		WillReturnError(errors.New("boom"))

	err = serviceUnderTest.DeletePet(context.Background(), companyID, petID)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
