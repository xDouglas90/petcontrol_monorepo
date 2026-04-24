package sqlc_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_Pets_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	queries, _, pool := setupQueriesWithPool(t)
	defer pool.Close()

	// 1. Setup Company and Client (required for Pets)
	company := mustCreateCompany(t, queries, pool)

	// Create Person for Client
	person, err := queries.InsertClientPerson(ctx)
	require.NoError(t, err)

	_, err = queries.InsertClientIdentification(ctx, sqlc.InsertClientIdentificationParams{
		PersonID:       person.ID,
		FullName:       "Pet Owner",
		ShortName:      "Owner",
		GenderIdentity: sqlc.GenderIdentityWomanCisgender,
		MaritalStatus:  sqlc.MaritalStatusMarried,
		CPF:            "99988877766",
		BirthDate:      pgtype.Date{Time: time.Now().AddDate(-30, 0, 0), Valid: true},
	})
	require.NoError(t, err)

	client, err := queries.InsertClientRecord(ctx, sqlc.InsertClientRecordParams{
		PersonID: person.ID,
	})
	require.NoError(t, err)

	_, err = queries.CreateCompanyClient(ctx, sqlc.CreateCompanyClientParams{
		CompanyID: company.ID,
		ClientID:  client.ID,
	})
	require.NoError(t, err)

	t.Run("CreateAndGetPet", func(t *testing.T) {
		pet, err := queries.CreatePet(ctx, sqlc.CreatePetParams{
			Name:           "Buddy",
			Race:           "SRD",
			Color:          "Branco",
			Sex:            "M",
			Size:           sqlc.PetSizeSmall,
			Kind:           sqlc.PetKindDog,
			Temperament:    sqlc.PetTemperamentCalm,
			BirthDate:      pgtype.Date{Time: time.Now().AddDate(-1, 0, 0), Valid: true},
			OwnerID:        client.ID,
			IsActive:       pgtype.Bool{Bool: true, Valid: true},
			IsDeceased:     pgtype.Bool{Bool: false, Valid: true},
			IsVaccinated:   pgtype.Bool{Bool: false, Valid: true},
			IsNeutered:     pgtype.Bool{Bool: false, Valid: true},
			IsMicrochipped: pgtype.Bool{Bool: false, Valid: true},
		})
		require.NoError(t, err)
		require.NotEqual(t, pgtype.UUID{}, pet.ID)

		res, err := queries.GetPetByIDAndCompanyID(ctx, sqlc.GetPetByIDAndCompanyIDParams{
			CompanyID: company.ID,
			ID:        pet.ID,
		})
		require.NoError(t, err)
		require.Equal(t, "Buddy", res.Name)
		require.Equal(t, "Pet Owner", res.OwnerName)
	})

	t.Run("ListPetsByCompany", func(t *testing.T) {
		// Buddy already exists from previous test
		list, err := queries.ListPetsByCompanyID(ctx, sqlc.ListPetsByCompanyIDParams{
			CompanyID: company.ID,
			Search:    "",
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		require.NotEmpty(t, list)
		require.Equal(t, "Buddy", list[0].Name)
	})
}
