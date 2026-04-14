package sqlc_test

import (
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_ServicePlans_Integration(t *testing.T) {
	queries, ctx, _ := setupQueriesWithPool(t)

	t.Run("InsertServicePlan - Success", func(t *testing.T) {
		pt := mustCreatePlanType(t, queries)
		arg := sqlc.InsertServicePlanParams{
			PlanTypeID:   pt.ID,
			Title:        "Monthly Basic",
			Description:  "Basic monthly plan",
			Price:        mustNumeric(t, "49.90"),
			DiscountRate: mustNumeric(t, "0.00"),
			IsActive:     pgtype.Bool{Bool: true, Valid: true},
		}

		res, err := queries.InsertServicePlan(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title, res.Title)
		require.True(t, res.ID.Valid)
	})

	t.Run("UpdateServicePlan - Success", func(t *testing.T) {
		pt := mustCreatePlanType(t, queries)
		sp := mustInsertServicePlan(t, queries, pt.ID)

		arg := sqlc.UpdateServicePlanParams{
			ID:    sp.ID,
			Title: pgtype.Text{String: "Annual Basic", Valid: true},
			Price: mustNumeric(t, "499.00"),
		}

		res, err := queries.UpdateServicePlan(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title.String, res.Title)
	})

	t.Run("DeleteServicePlan - Success", func(t *testing.T) {
		pt := mustCreatePlanType(t, queries)
		sp := mustInsertServicePlan(t, queries, pt.ID)

		res, err := queries.DeleteServicePlan(ctx, sp.ID)
		require.NoError(t, err)
		require.True(t, res.DeletedAt.Valid)

		// Verify it's not found
		_, err = queries.GetServicePlanByID(ctx, sp.ID)
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("GetServicePlanByID - Success", func(t *testing.T) {
		pt := mustCreatePlanType(t, queries)
		sp := mustInsertServicePlan(t, queries, pt.ID)

		res, err := queries.GetServicePlanByID(ctx, sp.ID)
		require.NoError(t, err)
		require.Equal(t, sp.Title, res.Title)
	})
}
