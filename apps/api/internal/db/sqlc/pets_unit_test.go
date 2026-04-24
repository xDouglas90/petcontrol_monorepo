package sqlc_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_Pets_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()

	t.Run("CreatePet", func(t *testing.T) {
		arg := sqlc.CreatePetParams{
			Name:           "Rex",
			Race:           "SRD",
			Color:          "Preto",
			Sex:            "M",
			Size:           sqlc.PetSizeMedium,
			Kind:           sqlc.PetKindDog,
			Temperament:    sqlc.PetTemperamentPlayful,
			BirthDate:      pgtype.Date{Time: time.Now().AddDate(-2, 0, 0), Valid: true},
			OwnerID:        uuidValue(),
			IsActive:       pgtype.Bool{Bool: true, Valid: true},
			IsDeceased:     pgtype.Bool{Bool: false, Valid: true},
			IsVaccinated:   pgtype.Bool{Bool: false, Valid: true},
			IsNeutered:     pgtype.Bool{Bool: false, Valid: true},
			IsMicrochipped: pgtype.Bool{Bool: false, Valid: true},
		}

		mock.ExpectQuery(`(?s)INSERT INTO pets`).
			WithArgs(arg.Name, arg.Race, arg.Color, arg.Sex, arg.Size, arg.Kind, arg.Temperament, arg.ImageUrl, arg.BirthDate, arg.OwnerID, arg.IsActive, arg.IsDeceased, arg.IsVaccinated, arg.IsNeutered, arg.IsMicrochipped, arg.MicrochipNumber, arg.MicrochipExpirationDate, arg.Notes).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "race", "color", "sex", "size", "kind", "temperament", "image_url", "birth_date", "owner_id", "is_active", "is_deceased", "is_vaccinated", "is_neutered", "is_microchipped", "microchip_number", "microchip_expiration_date", "notes", "created_at", "updated_at", "deleted_at"}).
				AddRow(uuidValue(), arg.Name, arg.Race, arg.Color, arg.Sex, arg.Size, arg.Kind, arg.Temperament, pgtype.Text{}, arg.BirthDate, arg.OwnerID, true, false, false, false, false, pgtype.Text{}, pgtype.Date{}, pgtype.Text{}, time.Now(), pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		res, err := queries.CreatePet(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Name, res.Name)
	})

	t.Run("GetPetByIDAndCompanyID", func(t *testing.T) {
		arg := sqlc.GetPetByIDAndCompanyIDParams{
			CompanyID: uuidValue(),
			ID:        uuidValue(),
		}

		mock.ExpectQuery(`(?s)SELECT.*?FROM.*?pets p`).
			WithArgs(arg.CompanyID, arg.ID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "owner_id", "company_id", "owner_name", "name", "race", "color", "sex", "size", "kind", "temperament", "image_url", "birth_date", "is_active", "is_deceased", "is_vaccinated", "is_neutered", "is_microchipped", "microchip_number", "microchip_expiration_date", "notes", "created_at", "updated_at", "deleted_at",
			}).AddRow(
				arg.ID, uuidValue(), arg.CompanyID, "John Owner", "Rex", "SRD", "Preto", "M", sqlc.PetSizeMedium, sqlc.PetKindDog, sqlc.PetTemperamentPlayful, pgtype.Text{}, pgtype.Date{Time: time.Now(), Valid: true}, true, false, false, false, false, pgtype.Text{}, pgtype.Date{}, pgtype.Text{}, time.Now(), pgtype.Timestamptz{}, pgtype.Timestamptz{},
			))

		res, err := queries.GetPetByIDAndCompanyID(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, "Rex", res.Name)
		require.Equal(t, "John Owner", res.OwnerName)
	})
}
