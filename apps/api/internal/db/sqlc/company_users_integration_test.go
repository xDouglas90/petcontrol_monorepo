package sqlc_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
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

	t.Run("RejectsRootAndInternalRoles", func(t *testing.T) {
		testCases := []struct {
			name string
			role sqlc.UserRoleType
		}{
			{name: "root", role: sqlc.UserRoleTypeRoot},
			{name: "internal", role: sqlc.UserRoleTypeInternal},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				user, err := queries.InsertUser(ctx, sqlc.InsertUserParams{
					Email:           uniqueEmail("company-user-" + tc.name),
					EmailVerified:   true,
					EmailVerifiedAt: pgtype.Timestamptz{},
					Role:            tc.role,
					IsActive:        true,
				})
				require.NoError(t, err)

				_, err = queries.CreateCompanyUser(ctx, sqlc.CreateCompanyUserParams{
					CompanyID: company.ID,
					UserID:    user.ID,
					Kind:      sqlc.UserKindEmployee,
					IsActive:  pgtype.Bool{Bool: true, Valid: true},
				})
				require.Error(t, err)

				var pgErr *pgconn.PgError
				require.True(t, errors.As(err, &pgErr))
				require.Equal(t, "23514", pgErr.Code)
				require.Equal(t, "chk_company_users_no_root_internal", pgErr.ConstraintName)
			})
		}
	})
}
