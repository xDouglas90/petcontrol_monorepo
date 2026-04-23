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
			Name:        "Rex",
			Size:        sqlc.PetSizeMedium,
			Kind:        sqlc.PetKindDog,
			Temperament: sqlc.PetTemperamentPlayful,
			BirthDate:   pgtype.Date{Time: time.Now().AddDate(-2, 0, 0), Valid: true},
			OwnerID:     uuidValue(),
		}

		mock.ExpectQuery(`(?s)INSERT INTO pets`).
			WithArgs(arg.Name, arg.Size, arg.Kind, arg.Temperament, arg.ImageUrl, arg.BirthDate, arg.OwnerID, arg.Notes).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "size", "kind", "temperament", "image_url", "birth_date", "owner_id", "is_active", "notes", "created_at", "updated_at", "deleted_at"}).
				AddRow(uuidValue(), arg.Name, arg.Size, arg.Kind, arg.Temperament, pgtype.Text{}, arg.BirthDate, arg.OwnerID, true, pgtype.Text{}, time.Now(), pgtype.Timestamptz{}, pgtype.Timestamptz{}))

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
				"id", "owner_id", "company_id", "owner_name", "name", "size", "kind", "temperament", "image_url", "birth_date", "is_active", "notes", "created_at", "updated_at", "deleted_at",
			}).AddRow(
				arg.ID, uuidValue(), arg.CompanyID, "John Owner", "Rex", sqlc.PetSizeMedium, sqlc.PetKindDog, sqlc.PetTemperamentPlayful, pgtype.Text{}, pgtype.Date{Time: time.Now(), Valid: true}, true, pgtype.Text{}, time.Now(), pgtype.Timestamptz{}, pgtype.Timestamptz{},
			))

		res, err := queries.GetPetByIDAndCompanyID(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, "Rex", res.Name)
		require.Equal(t, "John Owner", res.OwnerName)
	})
}
