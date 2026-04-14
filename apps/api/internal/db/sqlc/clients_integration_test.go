package sqlc_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_Clients_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	queries, _, pool := setupQueriesWithPool(t)
	defer pool.Close()

	// 1. Setup Company (required for Clients)
	company := mustCreateCompany(t, queries, pool)

	t.Run("CreateAndGetClient", func(t *testing.T) {
		// A. Create Person for Client
		person, err := queries.InsertClientPerson(ctx)
		require.NoError(t, err)

		// B. Setup Identity
		_, err = queries.InsertClientIdentification(ctx, sqlc.InsertClientIdentificationParams{
			PersonID:       person.ID,
			FullName:       "Client Name",
			ShortName:      "Client",
			GenderIdentity: sqlc.GenderIdentityNonBinary,
			MaritalStatus:  sqlc.MaritalStatusSingle,
			BirthDate:      pgtype.Date{Time: time.Now().AddDate(-25, 0, 0), Valid: true},
			CPF:            "11122233344",
		})
		require.NoError(t, err)

		// C. Setup Contact
		_, err = queries.InsertClientPrimaryContact(ctx, sqlc.InsertClientPrimaryContactParams{
			PersonID:    person.ID,
			Email:       "client@integration.com",
			Cellphone:   "11988112233",
			HasWhatsapp: true,
		})
		require.NoError(t, err)

		// D. Create Client Record
		client, err := queries.InsertClientRecord(ctx, sqlc.InsertClientRecordParams{
			PersonID: person.ID,
		})
		require.NoError(t, err)

		// E. Link to Company
		_, err = queries.CreateCompanyClient(ctx, sqlc.CreateCompanyClientParams{
			CompanyID: company.ID,
			ClientID:  client.ID,
		})
		require.NoError(t, err)

		// F. Verify Retrieval
		res, err := queries.GetClientByIDAndCompanyID(ctx, sqlc.GetClientByIDAndCompanyIDParams{
			CompanyID: company.ID,
			ID:        client.ID,
		})
		require.NoError(t, err)
		require.Equal(t, client.ID, res.ID)
		require.Equal(t, "Client Name", res.FullName)
		require.Equal(t, "client@integration.com", res.Email)
	})

	t.Run("ListAndSearchClients", func(t *testing.T) {
		// Existing client from previous run or create a new one to be sure
		p, _ := queries.InsertClientPerson(ctx)
		_, _ = queries.InsertClientIdentification(ctx, sqlc.InsertClientIdentificationParams{
			PersonID:       p.ID,
			FullName:       "Searchable John",
			ShortName:      "SJ",
			GenderIdentity: sqlc.GenderIdentityManCisgender,
			MaritalStatus:  sqlc.MaritalStatusSingle,
			CPF:            "00011122233",
			BirthDate:      pgtype.Date{Time: time.Now(), Valid: true},
		})
		_, _ = queries.InsertClientPrimaryContact(ctx, sqlc.InsertClientPrimaryContactParams{
			PersonID:    p.ID,
			Email:       "searchable@test.com",
			Cellphone:   "11988887777",
			HasWhatsapp: true,
		})
		c, _ := queries.InsertClientRecord(ctx, sqlc.InsertClientRecordParams{PersonID: p.ID})
		_, _ = queries.CreateCompanyClient(ctx, sqlc.CreateCompanyClientParams{CompanyID: company.ID, ClientID: c.ID})

		// Search by Name
		list, err := queries.ListClientsByCompanyID(ctx, sqlc.ListClientsByCompanyIDParams{
			CompanyID: company.ID,
			Search:    "Searchable",
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		require.NotEmpty(t, list)
		require.Contains(t, list[0].FullName, "Searchable")

		// Search by CPF
		list2, err := queries.ListClientsByCompanyID(ctx, sqlc.ListClientsByCompanyIDParams{
			CompanyID: company.ID,
			Search:    "00011122233",
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		require.NotEmpty(t, list2)
	})
}
