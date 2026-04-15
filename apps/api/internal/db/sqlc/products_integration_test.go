package sqlc_test

import (
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_Products_Integration(t *testing.T) {
	queries, ctx, _ := setupQueriesWithPool(t)

	t.Run("InsertProduct - Success", func(t *testing.T) {
		arg := sqlc.InsertProductParams{
			Name:        "Dog Food",
			Description: pgtype.Text{String: "Premium dog food 10kg", Valid: true},
			Quantity:    pgtype.Int4{Int32: 50, Valid: true},
		}

		res, err := queries.InsertProduct(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Name, res.Name)
		require.Equal(t, int32(50), res.Quantity)
	})

	t.Run("UpdateProduct - Success", func(t *testing.T) {
		p := mustCreateProduct(t, queries)

		arg := sqlc.UpdateProductParams{
			ID:   p.ID,
			Name: pgtype.Text{String: "Cat Food", Valid: true},
		}

		res, err := queries.UpdateProduct(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Name.String, res.Name)
	})

	t.Run("UpdateProductQuantity - Success", func(t *testing.T) {
		p := mustCreateProduct(t, queries)

		arg := sqlc.UpdateProductQuantityParams{
			ID:       p.ID,
			Quantity: 25,
		}

		res, err := queries.UpdateProductQuantity(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, int32(25), res.Quantity)
	})

	t.Run("DeleteProduct - Success", func(t *testing.T) {
		p := mustCreateProduct(t, queries)

		res, err := queries.DeleteProduct(ctx, p.ID)
		require.NoError(t, err)
		require.True(t, res.DeletedAt.Valid)

		// Verify it's not found
		_, err = queries.GetProductByID(ctx, p.ID)
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("GetProductByID - Success", func(t *testing.T) {
		p := mustCreateProduct(t, queries)

		res, err := queries.GetProductByID(ctx, p.ID)
		require.NoError(t, err)
		require.Equal(t, p.Name, res.Name)
	})
}
