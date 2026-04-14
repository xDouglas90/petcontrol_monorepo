package sqlc

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestQueries_Products_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := New(mock)
	ctx := context.Background()

	t.Run("InsertProduct - Success", func(t *testing.T) {
		arg := InsertProductParams{
			Name:        "Shampoo",
			Description: pgtype.Text{String: "Pet shampoo", Valid: true},
			Quantity:    pgtype.Int4{Int32: 100, Valid: true},
		}

		mock.ExpectQuery("INSERT INTO products").
			WithArgs(arg.Name, pgxmock.AnyArg(), arg.Description, pgxmock.AnyArg(), pgxmock.AnyArg(), arg.Quantity).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "batch_number", "description", "image_url", "expiration_date", "quantity", "created_at", "updated_at", "deleted_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, arg.Name, nil, arg.Description, nil, nil, int32(100), nil, nil, nil))

		res, err := queries.InsertProduct(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Name, res.Name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateProduct - Success", func(t *testing.T) {
		arg := UpdateProductParams{
			ID:   pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Name: pgtype.Text{String: "Conditioner", Valid: true},
		}

		mock.ExpectQuery("UPDATE products").
			WithArgs(arg.Name, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), arg.ID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "batch_number", "description", "image_url", "expiration_date", "quantity", "created_at", "updated_at", "deleted_at"}).
				AddRow(arg.ID, arg.Name.String, nil, nil, nil, nil, int32(0), nil, nil, nil))

		res, err := queries.UpdateProduct(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Name.String, res.Name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateProductQuantity - Success", func(t *testing.T) {
		arg := UpdateProductQuantityParams{
			ID:       pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Quantity: 50,
		}

		mock.ExpectQuery("UPDATE products SET quantity").
			WithArgs(arg.Quantity, arg.ID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "batch_number", "description", "image_url", "expiration_date", "quantity", "created_at", "updated_at", "deleted_at"}).
				AddRow(arg.ID, "Shampoo", nil, nil, nil, nil, int32(50), nil, nil, nil))

		res, err := queries.UpdateProductQuantity(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, int32(50), res.Quantity)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteProduct - Success", func(t *testing.T) {
		id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}

		mock.ExpectQuery("UPDATE products SET deleted_at").
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "batch_number", "description", "image_url", "expiration_date", "quantity", "created_at", "updated_at", "deleted_at"}).
				AddRow(id, "Deleted", nil, nil, nil, nil, int32(0), nil, nil, pgtype.Timestamptz{Valid: true}))

		res, err := queries.DeleteProduct(ctx, id)
		require.NoError(t, err)
		require.True(t, res.DeletedAt.Valid)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
