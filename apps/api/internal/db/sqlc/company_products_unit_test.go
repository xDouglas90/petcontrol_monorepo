package sqlc

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestQueries_CompanyProducts_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := New(mock)
	ctx := context.Background()

	t.Run("InsertCompanyProduct - Success", func(t *testing.T) {
		arg := InsertCompanyProductParams{
			CompanyID:   pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			ProductID:   pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
			Kind:        ProductKindService,
			CostPerUnit: pgtype.Numeric{Valid: true},
			SalePrice:   pgtype.Numeric{Valid: true},
		}

		mock.ExpectQuery("INSERT INTO company_products").
			WithArgs(arg.CompanyID, arg.ProductID, arg.Kind, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "product_id", "kind", "has_stock", "for_sale", "cost_per_unit", "profit_margin", "sale_price", "created_at", "updated_at", "deleted_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{3}, Valid: true}, arg.CompanyID, arg.ProductID, arg.Kind, true, true, pgtype.Numeric{Valid: true}, pgtype.Numeric{Valid: true}, pgtype.Numeric{Valid: true}, nil, nil, nil))

		res, err := queries.InsertCompanyProduct(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.CompanyID, res.CompanyID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateCompanyProduct - Success", func(t *testing.T) {
		arg := UpdateCompanyProductParams{
			ID:        pgtype.UUID{Bytes: [16]byte{3}, Valid: true},
			CompanyID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			SalePrice: pgtype.Numeric{Valid: true},
		}

		mock.ExpectQuery("UPDATE company_products").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), arg.SalePrice, arg.ID, arg.CompanyID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "product_id", "kind", "has_stock", "for_sale", "cost_per_unit", "profit_margin", "sale_price", "created_at", "updated_at", "deleted_at"}).
				AddRow(arg.ID, arg.CompanyID, pgtype.UUID{}, ProductKindService, true, true, pgtype.Numeric{Valid: true}, pgtype.Numeric{Valid: true}, pgtype.Numeric{Valid: true}, nil, nil, nil))

		res, err := queries.UpdateCompanyProduct(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.ID, res.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteCompanyProduct - Success", func(t *testing.T) {
		arg := DeleteCompanyProductParams{
			ID:        pgtype.UUID{Bytes: [16]byte{3}, Valid: true},
			CompanyID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		}

		mock.ExpectQuery("UPDATE company_products SET deleted_at").
			WithArgs(arg.ID, arg.CompanyID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "product_id", "kind", "has_stock", "for_sale", "cost_per_unit", "profit_margin", "sale_price", "created_at", "updated_at", "deleted_at"}).
				AddRow(arg.ID, arg.CompanyID, pgtype.UUID{}, ProductKindService, true, true, pgtype.Numeric{Valid: true}, pgtype.Numeric{Valid: true}, pgtype.Numeric{Valid: true}, nil, nil, pgtype.Timestamptz{Valid: true}))

		res, err := queries.DeleteCompanyProduct(ctx, arg)
		require.NoError(t, err)
		require.True(t, res.DeletedAt.Valid)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
