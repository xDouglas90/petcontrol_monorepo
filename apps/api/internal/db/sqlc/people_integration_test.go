package sqlc_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_People_Integration(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)
	defer pool.Close()

	t.Run("CreateAndGetPerson", func(t *testing.T) {
		// 1. Insert Person
		person, err := queries.InsertPerson(ctx, sqlc.InsertPersonParams{
			Kind:          sqlc.PersonKindClient,
			IsActive:      pgtype.Bool{Bool: true, Valid: true},
			HasSystemUser: pgtype.Bool{Bool: false, Valid: true},
		})
		require.NoError(t, err)

		// 2. Insert Identifications
		_, err = queries.InsertPersonIdentifications(ctx, sqlc.InsertPersonIdentificationsParams{
			PersonID:       person.ID,
			FullName:       "Integration Test Person",
			ShortName:      "Test",
			GenderIdentity: sqlc.GenderIdentityNotToExpose,
			MaritalStatus:  sqlc.MaritalStatusSingle,
			CPF:            "99988877766",
			BirthDate:      pgtype.Date{Time: time.Now().AddDate(-30, 0, 0), Valid: true},
		})
		require.NoError(t, err)

		// 3. Insert Contacts
		_, err = queries.InsertPersonContacts(ctx, sqlc.InsertPersonContactsParams{
			PersonID:    person.ID,
			Email:       "integration@test.com",
			Cellphone:   "11912345678",
			HasWhatsapp: true,
			IsPrimary:   true,
		})
		require.NoError(t, err)

		// 4. Verify GetPerson (joins identifications)
		res, err := queries.GetPerson(ctx, person.ID)
		require.NoError(t, err)
		require.Equal(t, person.ID, res.ID)
		require.Equal(t, "Integration Test Person", res.FullName.String)
		require.Equal(t, "99988877766", res.Cpf.String)

		// 5. Verify GetPersonContacts
		contacts, err := queries.GetPersonContacts(ctx, person.ID)
		require.NoError(t, err)
		require.Equal(t, "integration@test.com", contacts.Email)
		require.Equal(t, "11912345678", contacts.Cellphone)
	})

	t.Run("UpdateAndList", func(t *testing.T) {
		// Create a fresh person for this subtest to ensure independence
		person, err := queries.InsertPerson(ctx, sqlc.InsertPersonParams{
			Kind:          sqlc.PersonKindClient,
			IsActive:      pgtype.Bool{Bool: true, Valid: true},
			HasSystemUser: pgtype.Bool{Bool: false, Valid: true},
		})
		require.NoError(t, err)

		// 2. Insert Identification first (so we can update it)
		_, err = queries.InsertPersonIdentifications(ctx, sqlc.InsertPersonIdentificationsParams{
			PersonID:       person.ID,
			FullName:       "Original Name",
			ShortName:      "Original",
			GenderIdentity: sqlc.GenderIdentityNotToExpose,
			MaritalStatus:  sqlc.MaritalStatusSingle,
			CPF:            "00000000000",
			BirthDate:      pgtype.Date{Time: time.Now().AddDate(-30, 0, 0), Valid: true},
		})
		require.NoError(t, err)

		// 3. Update Identification
		_, err = queries.UpdatePersonIdentifications(ctx, sqlc.UpdatePersonIdentificationsParams{
			PersonID: person.ID,
			FullName: pgtype.Text{String: "Updated Name", Valid: true},
		})
		require.NoError(t, err)

		// 4. Verify change
		res, err := queries.GetPerson(ctx, person.ID)
		require.NoError(t, err)
		require.Equal(t, "Updated Name", res.FullName.String)

		// List
		list, err := queries.ListPeople(ctx, sqlc.ListPeopleParams{
			Kind:   sqlc.NullPersonKind{PersonKind: sqlc.PersonKindClient, Valid: true},
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		require.NotEmpty(t, list)
		
		// Find our person in the list
		found := false
		for _, p := range list {
			if p.ID == person.ID {
				require.Equal(t, "Updated Name", p.FullName.String)
				found = true
				break
			}
		}
		require.True(t, found, "person not found in list")
	})
}
