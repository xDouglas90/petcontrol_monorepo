package sqlc_test

import (
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_ServiceTypes_Integration(t *testing.T) {
	queries, ctx, _ := setupQueriesWithPool(t)

	t.Run("InsertServiceType - Success", func(t *testing.T) {
		arg := sqlc.InsertServiceTypeParams{
			Name:        "Bath & Brush",
			Description: pgtype.Text{String: "Standard bath and brushing service", Valid: true},
		}

		res, err := queries.InsertServiceType(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Name, res.Name)
		require.True(t, res.ID.Valid)
	})

	t.Run("UpdateServiceType - Success", func(t *testing.T) {
		st := mustCreateServiceType(t, queries)

		arg := sqlc.UpdateServiceTypeParams{
			ID:   st.ID,
			Name: pgtype.Text{String: "Premium Bath", Valid: true},
		}

		res, err := queries.UpdateServiceType(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Name.String, res.Name)
	})

	t.Run("DeleteServiceType - Success", func(t *testing.T) {
		st := mustCreateServiceType(t, queries)

		res, err := queries.DeleteServiceType(ctx, st.ID)
		require.NoError(t, err)
		require.True(t, res.DeletedAt.Valid)

		// Verify it's not found in GetServiceTypeByID because of deleted_at IS NULL filter
		_, err = queries.GetServiceTypeByID(ctx, st.ID)
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("GetServiceTypeByID - Success", func(t *testing.T) {
		st := mustCreateServiceType(t, queries)

		res, err := queries.GetServiceTypeByID(ctx, st.ID)
		require.NoError(t, err)
		require.Equal(t, st.Name, res.Name)
	})
}
