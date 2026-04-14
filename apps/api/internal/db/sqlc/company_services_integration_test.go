package sqlc_test

import (
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_CompanyServices_Integration(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)

	t.Run("InsertCompanyService - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		st := mustCreateServiceType(t, queries)
		service := mustCreateService(t, queries, st.ID)

		arg := sqlc.InsertCompanyServiceParams{
			CompanyID: company.ID,
			ServiceID: service.ID,
			IsActive:  pgtype.Bool{Bool: true, Valid: true},
		}

		res, err := queries.InsertCompanyService(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, company.ID, res.CompanyID)
		require.Equal(t, service.ID, res.ServiceID)
	})

	t.Run("UpdateCompanyService - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		st := mustCreateServiceType(t, queries)
		service := mustCreateService(t, queries, st.ID)
		mustCreateCompanyService(t, queries, company.ID, service.ID)

		arg := sqlc.UpdateCompanyServiceParams{
			CompanyID: company.ID,
			ServiceID: service.ID,
			IsActive:  pgtype.Bool{Bool: false, Valid: true},
		}

		res, err := queries.UpdateCompanyService(ctx, arg)
		require.NoError(t, err)
		require.False(t, res.IsActive)
	})

	t.Run("DeleteCompanyService - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		st := mustCreateServiceType(t, queries)
		service := mustCreateService(t, queries, st.ID)
		mustCreateCompanyService(t, queries, company.ID, service.ID)

		res, err := queries.DeleteCompanyService(ctx, sqlc.DeleteCompanyServiceParams{
			CompanyID: company.ID,
			ServiceID: service.ID,
		})
		require.NoError(t, err)
		require.Equal(t, service.ID, res.ServiceID)

		// Verify it's not found
		_, err = queries.GetCompanyService(ctx, sqlc.GetCompanyServiceParams{
			CompanyID: company.ID,
			ServiceID: service.ID,
		})
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})

	t.Run("ListActiveCompanyServices - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		st := mustCreateServiceType(t, queries)
		s1 := mustCreateService(t, queries, st.ID)
		s2 := mustCreateService(t, queries, st.ID)

		mustCreateCompanyService(t, queries, company.ID, s1.ID)
		mustCreateCompanyService(t, queries, company.ID, s2.ID)

		res, err := queries.ListActiveCompanyServices(ctx, sqlc.ListActiveCompanyServicesParams{
			CompanyID: company.ID,
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		require.Len(t, res, 2)
		_ = pool
	})
}
