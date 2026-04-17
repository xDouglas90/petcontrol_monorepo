package sqlc_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_Permissions_Integration(t *testing.T) {
	queries, ctx, _ := setupQueriesWithPool(t)

	t.Run("Lifecycle", func(t *testing.T) {
		code := "test:perm"
		arg := sqlc.InsertPermissionParams{
			Code:         code,
			DefaultRoles: []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin},
			Description:  pgtype.Text{String: "Testing integration", Valid: true},
		}

		// Insert
		_, err := queries.InsertPermission(ctx, arg)
		require.NoError(t, err)

		// Get
		perm, err := queries.GetPermissionByCode(ctx, code)
		require.NoError(t, err)
		require.Equal(t, code, perm.Code)

		// Update
		newDescription := "Updated description"
		_, err = queries.UpdatePermission(ctx, sqlc.UpdatePermissionParams{
			ID:          perm.ID,
			Description: pgtype.Text{String: newDescription, Valid: true},
		})
		require.NoError(t, err)

		updated, err := queries.GetPermissionByCode(ctx, code)
		require.NoError(t, err)
		require.Equal(t, newDescription, updated.Description.String)

		// List
		list, err := queries.ListPermissions(ctx, sqlc.ListPermissionsParams{Offset: 0, Limit: 10})
		require.NoError(t, err)
		found := false
		for _, p := range list {
			if p.Code == code {
				found = true
				break
			}
		}
		require.True(t, found)

		// Delete
		_, err = queries.DeletePermission(ctx, perm.ID)
		require.NoError(t, err)

		listAfter, _ := queries.ListPermissions(ctx, sqlc.ListPermissionsParams{Offset: 0, Limit: 10})
		foundAfter := false
		for _, p := range listAfter {
			if p.Code == code {
				foundAfter = true
				break
			}
		}
		require.False(t, foundAfter)
	})
}
