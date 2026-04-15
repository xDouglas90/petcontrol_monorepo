package sqlc_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_PeopleIdentifications_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()

	t.Run("InsertPersonIdentifications", func(t *testing.T) {
		arg := sqlc.InsertPersonIdentificationsParams{
			PersonID:       uuidValue(),
			FullName:       "John Doe",
			ShortName:      "John",
			GenderIdentity: sqlc.GenderIdentityManCisgender,
			MaritalStatus:  sqlc.MaritalStatusSingle,
			CPF:            "12345678901",
		}

		mock.ExpectQuery(`(?s)INSERT INTO people_identifications`).
			WithArgs(arg.PersonID, arg.FullName, arg.ShortName, arg.GenderIdentity, arg.MaritalStatus, arg.ImageURL, arg.BirthDate, arg.CPF).
			WillReturnRows(pgxmock.NewRows([]string{"id", "person_id", "full_name", "short_name", "gender_identity", "marital_status", "image_url", "birth_date", "cpf", "created_at", "updated_at"}).
				AddRow(uuidValue(), arg.PersonID, arg.FullName, arg.ShortName, arg.GenderIdentity, arg.MaritalStatus, pgtype.Text{}, pgtype.Date{}, arg.CPF, pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		res, err := queries.InsertPersonIdentifications(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.FullName, res.FullName)
	})

	t.Run("GetPersonIdentifications", func(t *testing.T) {
		personID := uuidValue()
		mock.ExpectQuery(`(?s)SELECT.*?FROM.*?people_identifications`).
			WithArgs(personID).
			WillReturnRows(pgxmock.NewRows([]string{
				"person_id", "full_name", "short_name", "gender_identity", "marital_status", "image_url", "birth_date", "cpf", "created_at", "updated_at",
			}).AddRow(
				personID, "John Doe", "John", sqlc.GenderIdentityManCisgender, sqlc.MaritalStatusSingle, pgtype.Text{}, pgtype.Date{}, "12345678901", pgtype.Timestamptz{}, pgtype.Timestamptz{},
			))

		res, err := queries.GetPersonIdentifications(ctx, personID)
		require.NoError(t, err)
		require.Equal(t, personID, res.PersonID)
		require.Equal(t, "John Doe", res.FullName)
	})

	t.Run("UpdatePersonIdentifications", func(t *testing.T) {
		arg := sqlc.UpdatePersonIdentificationsParams{
			PersonID:  uuidValue(),
			FullName:  pgtype.Text{String: "John Updated", Valid: true},
			ShortName: pgtype.Text{String: "JohnU", Valid: true},
		}

		mock.ExpectExec(`(?s)UPDATE.*?people_identifications`).
			WithArgs(arg.FullName, arg.ShortName, arg.GenderIdentity, arg.MaritalStatus, arg.ImageURL, arg.BirthDate, arg.CPF, arg.PersonID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		rows, err := queries.UpdatePersonIdentifications(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)
	})
}
