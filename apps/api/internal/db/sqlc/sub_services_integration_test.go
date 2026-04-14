package sqlc_test

import (
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_SubServices_Integration(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)

	t.Run("InsertSubService - Success", func(t *testing.T) {
		st := mustCreateServiceType(t, queries)
		service := mustCreateService(t, queries, st.ID)

		arg := sqlc.InsertSubServiceParams{
			ServiceID:    service.ID,
			TypeID:       st.ID,
			Title:        "Teeth Cleaning",
			Description:  "Deep teeth cleaning sub-service",
			Price:        mustNumeric(t, "25.00"),
			DiscountRate: mustNumeric(t, "0.00"),
			IsActive:     pgtype.Bool{Bool: true, Valid: true},
		}

		res, err := queries.InsertSubService(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title, res.Title)
		require.True(t, res.ID.Valid)
	})

	t.Run("UpdateSubService - Success", func(t *testing.T) {
		st := mustCreateServiceType(t, queries)
		service := mustCreateService(t, queries, st.ID)
		ss := mustCreateSubService(t, queries, service.ID, st.ID)

		arg := sqlc.UpdateSubServiceParams{
			ID:    ss.ID,
			Title: pgtype.Text{String: "Extreme Teeth Cleaning", Valid: true},
			Price: mustNumeric(t, "35.00"),
		}

		res, err := queries.UpdateSubService(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title.String, res.Title)
	})

	t.Run("DeleteSubService - Success", func(t *testing.T) {
		st := mustCreateServiceType(t, queries)
		service := mustCreateService(t, queries, st.ID)
		ss := mustCreateSubService(t, queries, service.ID, st.ID)

		res, err := queries.DeleteSubService(ctx, ss.ID)
		require.NoError(t, err)
		require.True(t, res.DeletedAt.Valid)

		// Verify it's not found
		_, err = queries.GetSubServiceByID(ctx, ss.ID)
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("ListSubServicesByServiceID - Success", func(t *testing.T) {
		st := mustCreateServiceType(t, queries)
		service := mustCreateService(t, queries, st.ID)
		mustCreateSubService(t, queries, service.ID, st.ID)
		mustCreateSubService(t, queries, service.ID, st.ID)

		res, err := queries.ListSubServicesByServiceID(ctx, sqlc.ListSubServicesByServiceIDParams{
			ServiceID: service.ID,
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		require.Len(t, res, 2)
		_ = pool
	})
}
