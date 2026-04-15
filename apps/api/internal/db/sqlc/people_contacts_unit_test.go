package sqlc_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_PeopleContacts_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()

	t.Run("InsertPersonContacts", func(t *testing.T) {
		arg := sqlc.InsertPersonContactsParams{
			PersonID:      uuidValue(),
			Email:         "test@example.com",
			Cellphone:     "11988887777",
			HasWhatsapp:   true,
			IsPrimary:     true,
			Phone:         pgtype.Text{String: "1133334444", Valid: true},
			InstagramUser: pgtype.Text{String: "@test", Valid: true},
		}

		mock.ExpectQuery(`(?s)INSERT INTO people_contacts`).
			WithArgs(arg.PersonID, arg.Email, arg.Phone, arg.Cellphone, arg.HasWhatsapp, arg.InstagramUser, arg.EmergencyContact, arg.EmergencyPhone, arg.IsPrimary).
			WillReturnRows(pgxmock.NewRows([]string{"id", "person_id", "email", "phone", "cellphone", "has_whatsapp", "instagram_user", "emergency_contact", "emergency_phone", "is_primary", "created_at", "updated_at"}).
				AddRow(uuidValue(), arg.PersonID, arg.Email, arg.Phone, arg.Cellphone, arg.HasWhatsapp, arg.InstagramUser, pgtype.Text{}, pgtype.Text{}, true, pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		res, err := queries.InsertPersonContacts(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Email, res.Email)
	})

	t.Run("GetPersonContacts", func(t *testing.T) {
		personID := uuidValue()
		mock.ExpectQuery(`(?s)SELECT.*?FROM.*?people_contacts`).
			WithArgs(personID).
			WillReturnRows(pgxmock.NewRows([]string{
				"person_id", "email", "phone", "cellphone", "has_whatsapp", "instagram_user", "emergency_contact", "emergency_phone", "is_primary", "created_at", "updated_at",
			}).AddRow(
				personID, "test@example.com", pgtype.Text{String: "1133334444", Valid: true}, "11988887777", true, pgtype.Text{String: "@test", Valid: true}, pgtype.Text{}, pgtype.Text{}, true, pgtype.Timestamptz{}, pgtype.Timestamptz{},
			))

		res, err := queries.GetPersonContacts(ctx, personID)
		require.NoError(t, err)
		require.Equal(t, personID, res.PersonID)
		require.Equal(t, "test@example.com", res.Email)
	})

	t.Run("UpdatePersonContacts", func(t *testing.T) {
		arg := sqlc.UpdatePersonContactsParams{
			PersonID:  uuidValue(),
			Email:     pgtype.Text{String: "updated@example.com", Valid: true},
			IsPrimary: pgtype.Bool{Bool: false, Valid: true},
		}

		mock.ExpectExec(`(?s)UPDATE.*?people_contacts`).
			WithArgs(arg.Email, arg.Phone, arg.Cellphone, arg.HasWhatsapp, arg.InstagramUser, arg.EmergencyContact, arg.EmergencyPhone, arg.IsPrimary, arg.PersonID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		rows, err := queries.UpdatePersonContacts(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)
	})
}
