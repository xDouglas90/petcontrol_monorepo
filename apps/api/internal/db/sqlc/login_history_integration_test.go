package sqlc_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_LoginHistory_Integration(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)

	// Setup: Create a user
	email := uniqueEmail("login-hist")
	user, err := queries.InsertUser(ctx, sqlc.InsertUserParams{
		Email: email,
		Role:  sqlc.UserRoleTypeAdmin,
	})
	require.NoError(t, err)

	// Helper to insert directly and return ID if needed,
	// but login_history has uuid default so we scan it.
	insertAndGetID := func(result sqlc.LoginResult) pgtype.UUID {
		var id pgtype.UUID
		err := pool.QueryRow(ctx,
			"INSERT INTO login_history(user_id, ip_address, user_agent, result) VALUES ($1, $2, $3, $4) RETURNING id",
			user.ID, "127.0.0.1", "TestAgent", result,
		).Scan(&id)
		require.NoError(t, err)
		return id
	}

	t.Run("GetLoginHistoryByID", func(t *testing.T) {
		id := insertAndGetID(sqlc.LoginResultSuccess)

		res, err := queries.GetLoginHistoryByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, id, res.ID)
		require.Equal(t, user.ID, res.UserID)
		require.Equal(t, sqlc.LoginResultSuccess, res.Result)
	})

	t.Run("ListFilters", func(t *testing.T) {
		// Clean start not needed as setupQueriesWithPool creates fresh db,
		// but let's insert a few.
		insertAndGetID(sqlc.LoginResultSuccess)
		insertAndGetID(sqlc.LoginResultInvalidCredentials)
		insertAndGetID(sqlc.LoginResultSuccess)

		// List by User
		list, err := queries.ListLoginHistoryByUserID(ctx, sqlc.ListLoginHistoryByUserIDParams{
			UserID: user.ID,
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(list), 3)

		// List by Result (Success)
		list, err = queries.ListLoginHistoryByResult(ctx, sqlc.ListLoginHistoryByResultParams{
			Result: sqlc.LoginResultSuccess,
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		for _, item := range list {
			require.Equal(t, sqlc.LoginResultSuccess, item.Result)
		}
	})
}
