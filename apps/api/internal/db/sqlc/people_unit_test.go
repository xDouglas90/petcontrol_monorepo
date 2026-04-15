package sqlc_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_People_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()

	t.Run("InsertPerson", func(t *testing.T) {
		arg := sqlc.InsertPersonParams{
			Kind:          sqlc.PersonKindClient,
			IsActive:      pgtype.Bool{Bool: true, Valid: true},
			HasSystemUser: pgtype.Bool{Bool: false, Valid: true},
		}

		mock.ExpectQuery(`(?s)INSERT INTO people`).
			WithArgs(arg.Kind, arg.IsActive, arg.HasSystemUser).
			WillReturnRows(pgxmock.NewRows([]string{"id", "kind", "is_active", "has_system_user", "created_at", "updated_at"}).
				AddRow(uuidValue(), arg.Kind, true, false, pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		res, err := queries.InsertPerson(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Kind, res.Kind)
	})

	t.Run("GetPerson", func(t *testing.T) {
		id := uuidValue()
		mock.ExpectQuery(`(?s)SELECT.*?FROM.*?people`).
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "kind", "is_active", "has_system_user", "created_at", "updated_at",
				"full_name", "short_name", "gender_identity", "marital_status", "image_url", "birth_date", "cpf",
				"identifications_created_at", "identifications_updated_at",
			}).AddRow(
				id, sqlc.PersonKindClient, true, false, pgtype.Timestamptz{}, pgtype.Timestamptz{},
				pgtype.Text{String: "John Doe", Valid: true}, pgtype.Text{String: "John", Valid: true},
				sqlc.NullGenderIdentity{GenderIdentity: sqlc.GenderIdentityManCisgender, Valid: true},
				sqlc.NullMaritalStatus{MaritalStatus: sqlc.MaritalStatusSingle, Valid: true},
				pgtype.Text{}, pgtype.Date{}, pgtype.Text{String: "12345678901", Valid: true},
				pgtype.Timestamptz{}, pgtype.Timestamptz{},
			))

		res, err := queries.GetPerson(ctx, id)
		require.NoError(t, err)
		require.Equal(t, id, res.ID)
		require.Equal(t, "John Doe", res.FullName.String)
	})

	t.Run("ListPeople", func(t *testing.T) {
		arg := sqlc.ListPeopleParams{
			Limit:  10,
			Offset: 0,
		}

		mock.ExpectQuery(`(?s)SELECT.*?FROM.*?people`).
			WithArgs(arg.Kind, arg.IsActive, arg.HasSystemUser, arg.Offset, arg.Limit).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "kind", "is_active", "has_system_user", "created_at", "updated_at",
				"full_name", "short_name", "gender_identity", "marital_status", "image_url", "birth_date", "cpf",
				"identifications_created_at", "identifications_updated_at",
			}).AddRow(
				uuidValue(), sqlc.PersonKindClient, true, false, pgtype.Timestamptz{}, pgtype.Timestamptz{},
				pgtype.Text{String: "John Doe", Valid: true}, pgtype.Text{String: "John", Valid: true},
				sqlc.NullGenderIdentity{GenderIdentity: sqlc.GenderIdentityManCisgender, Valid: true},
				sqlc.NullMaritalStatus{MaritalStatus: sqlc.MaritalStatusSingle, Valid: true},
				pgtype.Text{}, pgtype.Date{}, pgtype.Text{String: "12345678901", Valid: true},
				pgtype.Timestamptz{}, pgtype.Timestamptz{},
			))

		res, err := queries.ListPeople(ctx, arg)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("UpdatePerson", func(t *testing.T) {
		arg := sqlc.UpdatePersonParams{
			ID:       uuidValue(),
			Kind:     sqlc.NullPersonKind{PersonKind: sqlc.PersonKindSupplier, Valid: true},
			IsActive: pgtype.Bool{Bool: false, Valid: true},
		}

		mock.ExpectExec(`(?s)UPDATE.*?people`).
			WithArgs(arg.Kind, arg.IsActive, arg.HasSystemUser, arg.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		rows, err := queries.UpdatePerson(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)
	})

	t.Run("DeletePerson", func(t *testing.T) {
		id := uuidValue()
		mock.ExpectExec(`(?s)DELETE FROM people`).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		rows, err := queries.DeletePerson(ctx, id)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)
	})
}
