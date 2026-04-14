package sqlc_test

import (

	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)



func TestQueries_CompanyServicePlans_Integration(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)

	t.Run("InsertCompanyServicePlan - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		pt := mustCreatePlanType(t, queries)
		sp := mustInsertServicePlan(t, queries, pt.ID)

		arg := sqlc.InsertCompanyServicePlanParams{
			CompanyID:     company.ID,
			ServicePlanID: sp.ID,
			IsActive:      pgtype.Bool{Bool: true, Valid: true},
		}

		res, err := queries.InsertCompanyServicePlan(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, company.ID, res.CompanyID)
		require.Equal(t, sp.ID, res.ServicePlanID)
	})

	t.Run("UpdateCompanyServicePlan - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		pt := mustCreatePlanType(t, queries)
		sp := mustInsertServicePlan(t, queries, pt.ID)
		mustCreateCompanyServicePlan(t, queries, company.ID, sp.ID)

		arg := sqlc.UpdateCompanyServicePlanParams{
			CompanyID:     company.ID,
			ServicePlanID: sp.ID,
			IsActive:      pgtype.Bool{Bool: false, Valid: true},
		}

		res, err := queries.UpdateCompanyServicePlan(ctx, arg)
		require.NoError(t, err)
		require.False(t, res.IsActive)
	})

	t.Run("DeleteCompanyServicePlan - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		pt := mustCreatePlanType(t, queries)
		sp := mustInsertServicePlan(t, queries, pt.ID)
		mustCreateCompanyServicePlan(t, queries, company.ID, sp.ID)

		res, err := queries.DeleteCompanyServicePlan(ctx, sqlc.DeleteCompanyServicePlanParams{
			CompanyID:     company.ID,
			ServicePlanID: sp.ID,
		})
		require.NoError(t, err)
		require.Equal(t, sp.ID, res.ServicePlanID)

		// Verify it's not found
		_, err = queries.GetCompanyServicePlan(ctx, sqlc.GetCompanyServicePlanParams{
			CompanyID:     company.ID,
			ServicePlanID: sp.ID,
		})
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("ListActiveCompanyServicePlans - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		pt := mustCreatePlanType(t, queries)
		sp1 := mustInsertServicePlan(t, queries, pt.ID)
		sp2 := mustInsertServicePlan(t, queries, pt.ID)

		mustCreateCompanyServicePlan(t, queries, company.ID, sp1.ID)
		mustCreateCompanyServicePlan(t, queries, company.ID, sp2.ID)

		res, err := queries.ListActiveCompanyServicePlans(ctx, sqlc.ListActiveCompanyServicePlansParams{
			CompanyID: company.ID,
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		require.Len(t, res, 2)
		_ = pool
	})
}
