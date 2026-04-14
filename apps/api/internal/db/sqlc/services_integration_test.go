package sqlc_test

import (
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)


func TestQueries_Services_Integration(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)

	t.Run("CreateService - Success", func(t *testing.T) {
		st := mustCreateServiceType(t, queries)
		arg := sqlc.CreateServiceParams{
			TypeID:       st.ID,
			Title:        "Grooming",
			Description:  "Full grooming service",
			Price:        pgtype.Numeric{Valid: true},
			DiscountRate: pgtype.Numeric{Valid: true},
			IsActive:     true,
		}

		res, err := queries.CreateService(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title, res.Title)
		require.True(t, res.ID.Valid)
	})

	t.Run("UpdateServiceByIDAndCompanyID - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		st := mustCreateServiceType(t, queries)
		service := mustCreateService(t, queries, st.ID)
		mustCreateCompanyService(t, queries, company.ID, service.ID)

		arg := sqlc.UpdateServiceByIDAndCompanyIDParams{
			ID:        service.ID,
			CompanyID: company.ID,
			Title:     pgtype.Text{String: "Extreme Grooming", Valid: true},
			Price:     pgtype.Numeric{Valid: true},
		}

		res, err := queries.UpdateServiceByIDAndCompanyID(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title.String, res.Title)
	})

	t.Run("DeactivateCompanyService - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		st := mustCreateServiceType(t, queries)
		service := mustCreateService(t, queries, st.ID)
		mustCreateCompanyService(t, queries, company.ID, service.ID)

		arg := sqlc.DeactivateCompanyServiceParams{
			CompanyID: company.ID,
			ServiceID: service.ID,
		}

		res, err := queries.DeactivateCompanyService(ctx, arg)
		require.NoError(t, err)
		require.False(t, res.IsActive)

		// Verify it's actually inactive in the DB
		got, err := queries.GetCompanyService(ctx, sqlc.GetCompanyServiceParams{
			CompanyID: company.ID,
			ServiceID: service.ID,
		})
		require.NoError(t, err)
		require.False(t, got.IsActive)
	})

	t.Run("GetServiceByIDAndCompanyID - Success", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		st := mustCreateServiceType(t, queries)
		service := mustCreateService(t, queries, st.ID)
		mustCreateCompanyService(t, queries, company.ID, service.ID)

		res, err := queries.GetServiceByIDAndCompanyID(ctx, sqlc.GetServiceByIDAndCompanyIDParams{
			CompanyID: company.ID,
			ID:        service.ID,
		})
		require.NoError(t, err)
		require.Equal(t, service.Title, res.Title)
		require.True(t, res.IsActive)
	})

	t.Run("GetServiceByIDAndCompanyID - Failure-NotFound", func(t *testing.T) {
		company := mustCreateCompany(t, queries, pool)
		_, err := queries.GetServiceByIDAndCompanyID(ctx, sqlc.GetServiceByIDAndCompanyIDParams{
			CompanyID: company.ID,
			ID:        pgtype.UUID{Valid: true},
		})
		require.Error(t, err)
		require.ErrorIs(t, err, pgx.ErrNoRows)
	})
}
