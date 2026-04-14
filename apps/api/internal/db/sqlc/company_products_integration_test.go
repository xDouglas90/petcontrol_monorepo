package sqlc_test

import (

	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)



func TestQueries_CompanyProducts_Integration(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)

	t.Run("InsertCompanyProduct - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		product := mustCreateProduct(t, queries)

		arg := sqlc.InsertCompanyProductParams{
			CompanyID:    company.ID,
			ProductID:    product.ID,
			Kind:         sqlc.ProductKindCustomer,
			CostPerUnit:  mustNumeric(t, "10.00"),
			SalePrice:    mustNumeric(t, "20.00"),
			HasStock:     pgtype.Bool{Bool: true, Valid: true},
			ForSale:      pgtype.Bool{Bool: true, Valid: true},
			ProfitMargin: mustNumeric(t, "50.00"),
		}

		res, err := queries.InsertCompanyProduct(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, company.ID, res.CompanyID)
		require.Equal(t, product.ID, res.ProductID)
	})

	t.Run("UpdateCompanyProduct - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		product := mustCreateProduct(t, queries)
		cp := mustCreateCompanyProduct(t, queries, company.ID, product.ID)

		arg := sqlc.UpdateCompanyProductParams{
			ID:        cp.ID,
			CompanyID: company.ID,
			SalePrice: mustNumeric(t, "30.00"),
		}

		res, err := queries.UpdateCompanyProduct(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.ID, res.ID)
	})

	t.Run("DeleteCompanyProduct - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		product := mustCreateProduct(t, queries)
		cp := mustCreateCompanyProduct(t, queries, company.ID, product.ID)

		res, err := queries.DeleteCompanyProduct(ctx, sqlc.DeleteCompanyProductParams{
			ID:        cp.ID,
			CompanyID: company.ID,
		})
		require.NoError(t, err)
		require.True(t, res.DeletedAt.Valid)

		// Verify it's not found
		_, err = queries.GetCompanyProductByID(ctx, sqlc.GetCompanyProductByIDParams{
			ID:        cp.ID,
			CompanyID: company.ID,
		})
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("ListCompanyProducts - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		p1 := mustCreateProduct(t, queries)
		p2 := mustCreateProduct(t, queries)
		mustCreateCompanyProduct(t, queries, company.ID, p1.ID)
		mustCreateCompanyProduct(t, queries, company.ID, p2.ID)

		res, err := queries.ListCompanyProducts(ctx, sqlc.ListCompanyProductsParams{
			CompanyID: company.ID,
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		require.Len(t, res, 2)
		_ = pool
	})
}
