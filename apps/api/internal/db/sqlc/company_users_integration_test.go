package sqlc_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_CompanyUsers_Integration(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)

	company := mustCreateCompany(t, queries, pool)
	user := insertDefaultUser(t, queries, uniqueEmail("company-user-life"))

	t.Run("Lifecycle", func(t *testing.T) {
		// Create
		kind := sqlc.UserKindEmployee
		created, err := queries.CreateCompanyUser(ctx, sqlc.CreateCompanyUserParams{
			CompanyID: company.ID,
			UserID:    user.ID,
			Kind:      kind,
			IsActive:  pgtype.Bool{Bool: true, Valid: true},
		})
		require.NoError(t, err)
		require.Equal(t, kind, created.Kind)

		// Get by ID
		gotByID, err := queries.GetCompanyUserByID(ctx, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, gotByID.ID)

		// Get by pair
		gotByPair, err := queries.GetCompanyUser(ctx, sqlc.GetCompanyUserParams{
			CompanyID: company.ID,
			UserID:    user.ID,
		})
		require.NoError(t, err)
		require.Equal(t, created.ID, gotByPair.ID)

		// Get active by UserID
		gotActive, err := queries.GetActiveCompanyUserByUserID(ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, gotActive.ID)

		// List by Company
		list, err := queries.ListCompanyUsersByCompanyID(ctx, company.ID)
		require.NoError(t, err)
		require.NotEmpty(t, list)

		// List by Kind
		byKind, err := queries.ListCompanyUsersByKind(ctx, sqlc.ListCompanyUsersByKindParams{
			CompanyID: company.ID,
			Kind:      sqlc.UserKindEmployee,
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		require.Len(t, byKind, 1)

		// Deactivate
		err = queries.DeactivateCompanyUser(ctx, sqlc.DeactivateCompanyUserParams{
			CompanyID: company.ID,
			UserID:    user.ID,
		})
		require.NoError(t, err)

		// Should not be active anymore
		activeList, _ := queries.ListCompanyUsersByCompanyID(ctx, company.ID)
		require.Empty(t, activeList)
	})
}
