package sqlc_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_ClientPlans_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	queries, _, pool := setupQueriesWithPool(t)
	defer pool.Close()

	// 1. Setup Dependencies
	company := mustCreateCompany(t, queries, pool)
	person, _ := queries.InsertClientPerson(ctx)
	_, _ = queries.InsertClientIdentification(ctx, sqlc.InsertClientIdentificationParams{
		PersonID:       person.ID,
		FullName:       "Plan Owner",
		ShortName:      "Owner",
		GenderIdentity: sqlc.GenderIdentityNonBinary,
		MaritalStatus:  sqlc.MaritalStatusSingle,
		BirthDate:      pgtype.Date{Time: time.Now(), Valid: true},
		CPF:            "55544433322",
	})
	client, _ := queries.InsertClientRecord(ctx, sqlc.InsertClientRecordParams{PersonID: person.ID})
	_, _ = queries.CreateCompanyClient(ctx, sqlc.CreateCompanyClientParams{CompanyID: company.ID, ClientID: client.ID})

	// Create Plan Type
	pt, err := queries.InsertPlanType(ctx, sqlc.InsertPlanTypeParams{
		Name:        "Service Plan Type",
		Description: pgtype.Text{String: "Type for testing", Valid: true},
	})
	require.NoError(t, err)

	// Create Plan
	plan, err := queries.InsertPlan(ctx, sqlc.InsertPlanParams{
		PlanTypeID:       pt.ID,
		Name:             "Monthly Premium",
		Description:      "Standard Monthly Plan",
		Package:          sqlc.ModulePackagePremium,
		Price:            mustNumeric(t, "49.90"),
		BillingCycleDays: 30,
		MaxUsers:         pgtype.Int4{Int32: 5, Valid: true},
		IsActive:         true,
	})
	require.NoError(t, err)

	t.Run("InsertAndGetClientPlan", func(t *testing.T) {
		cp, err := queries.InsertClientPlan(ctx, sqlc.InsertClientPlanParams{
			ClientID:  client.ID,
			PlanID:    plan.ID,
			StartedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
			PricePaid: mustNumeric(t, "49.90"),
			IsActive:  pgtype.Bool{Bool: true, Valid: true},
		})
		require.NoError(t, err)
		require.NotEqual(t, pgtype.UUID{}, cp.ID)

		res, err := queries.GetClientPlanByID(ctx, cp.ID)
		require.NoError(t, err)
		require.Equal(t, client.ID, res.ClientID)
		require.Equal(t, plan.ID, res.PlanID)
		require.Equal(t, "Monthly Premium", res.PlanName)
	})

	t.Run("ListClientPlansByCompany", func(t *testing.T) {
		list, err := queries.ListCompanyClientPlans(ctx, sqlc.ListCompanyClientPlansParams{
			CompanyID: company.ID,
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		require.NotEmpty(t, list)
		require.Equal(t, "Monthly Premium", list[0].PlanName)
	})
}
