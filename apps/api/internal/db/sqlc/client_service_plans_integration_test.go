package sqlc_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_ClientServicePlans_Integration(t *testing.T) {
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
		FullName:       "Service Plan Client",
		ShortName:      "SP Client",
		GenderIdentity: sqlc.GenderIdentityWomanCisgender,
		MaritalStatus:  sqlc.MaritalStatusSingle,
		BirthDate:      pgtype.Date{Time: time.Now(), Valid: true},
		CPF:            "66655544433",
	})
	client, _ := queries.InsertClientRecord(ctx, sqlc.InsertClientRecordParams{PersonID: person.ID})
	_, _ = queries.CreateCompanyClient(ctx, sqlc.CreateCompanyClientParams{CompanyID: company.ID, ClientID: client.ID})

	// Create Plan Type
	pt, _ := queries.InsertPlanType(ctx, sqlc.InsertPlanTypeParams{
		Name: "Service Type",
	})

	// Create Service Plan
	sp, err := queries.InsertServicePlan(ctx, sqlc.InsertServicePlanParams{
		PlanTypeID:   pt.ID,
		Title:        "Bath Pack 5",
		Description:  "5 Baths pack",
		Price:        mustNumeric(t, "150.00"),
		DiscountRate: mustNumeric(t, "10.00"),
		IsActive:     pgtype.Bool{Bool: true, Valid: true},
	})
	require.NoError(t, err)

	t.Run("InsertAndGetClientServicePlan", func(t *testing.T) {
		csp, err := queries.InsertClientServicePlan(ctx, sqlc.InsertClientServicePlanParams{
			ClientID:      client.ID,
			ServicePlanID: sp.ID,
			StartedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 3, 0), Valid: true},
			PricePaid:     mustNumeric(t, "135.00"),
			IsActive:      pgtype.Bool{Bool: true, Valid: true},
		})
		require.NoError(t, err)
		require.NotEqual(t, pgtype.UUID{}, csp.ID)

		res, err := queries.GetClientServicePlan(ctx, sqlc.GetClientServicePlanParams{
			ClientID:      client.ID,
			ServicePlanID: sp.ID,
		})
		require.NoError(t, err)
		require.Equal(t, csp.ID, res.ID)
	})

	t.Run("ListClientServicePlans", func(t *testing.T) {
		list, err := queries.ListClientServicePlans(ctx, sqlc.ListClientServicePlansParams{
			ClientID: client.ID,
			Limit:    10,
			Offset:   0,
		})
		require.NoError(t, err)
		require.NotEmpty(t, list)
		require.Equal(t, "Bath Pack 5", list[0].ServicePlanTitle)
	})
}
