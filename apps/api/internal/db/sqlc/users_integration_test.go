package sqlc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/test/integration"
)

func setupTestDB(t *testing.T) (*sqlc.Queries, func()) {
	t.Helper()

	setup := integration.SetupPostgresWithMigrations(t)
	queries := sqlc.New(setup.Pool)

	cleanup := func() {}
	return queries, cleanup
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@example.com", prefix, time.Now().UnixNano())
}

func insertDefaultUser(t *testing.T, queries *sqlc.Queries, email string) sqlc.User {
	t.Helper()

	user, err := queries.InsertUser(context.Background(), sqlc.InsertUserParams{
		Email:           email,
		EmailVerified:   false,
		EmailVerifiedAt: pgtype.Timestamptz{},
		Role:            sqlc.UserRoleTypeAdmin,
		Kind:            sqlc.UserKindOwner,
		IsActive:        true,
	})
	require.NoError(t, err)
	return user
}

func TestQueries_InsertUser(t *testing.T) {
	queries, cleanup := setupTestDB(t)
	defer cleanup()

	email := uniqueEmail("insert")
	got := insertDefaultUser(t, queries, email)

	require.True(t, got.ID.Valid)
	require.Equal(t, email, got.Email)
	require.Equal(t, sqlc.UserRoleTypeAdmin, got.Role)
	require.Equal(t, sqlc.UserKindOwner, got.Kind)
	require.True(t, got.CreatedAt.Valid)
	require.False(t, got.DeletedAt.Valid)
}

func TestQueries_GetUserByEmail(t *testing.T) {
	queries, cleanup := setupTestDB(t)
	defer cleanup()

	created := insertDefaultUser(t, queries, uniqueEmail("by-email"))
	got, err := queries.GetUserByEmail(context.Background(), created.Email)
	require.NoError(t, err)

	require.Equal(t, created.ID, got.ID)
	require.Equal(t, created.Email, got.Email)
}

func TestQueries_GetUserByID(t *testing.T) {
	queries, cleanup := setupTestDB(t)
	defer cleanup()

	created := insertDefaultUser(t, queries, uniqueEmail("by-id"))
	got, err := queries.GetUserByID(context.Background(), created.ID)
	require.NoError(t, err)

	require.Equal(t, created.ID, got.ID)
	require.Equal(t, created.Email, got.Email)
}

func TestQueries_ListUsersBasic(t *testing.T) {
	queries, cleanup := setupTestDB(t)
	defer cleanup()

	_ = insertDefaultUser(t, queries, uniqueEmail("list-1"))
	_ = insertDefaultUser(t, queries, uniqueEmail("list-2"))

	items, err := queries.ListUsersBasic(context.Background(), sqlc.ListUsersBasicParams{
		Offset: 0,
		Limit:  2,
	})
	require.NoError(t, err)
	require.Len(t, items, 2)
	for _, item := range items {
		require.True(t, item.IsActive)
		require.False(t, item.DeletedAt.Valid)
	}
}

func TestQueries_UpdateUser(t *testing.T) {
	queries, cleanup := setupTestDB(t)
	defer cleanup()

	created := insertDefaultUser(t, queries, uniqueEmail("update"))

	updated, err := queries.UpdateUser(context.Background(), sqlc.UpdateUserParams{
		Email:         pgtype.Text{String: uniqueEmail("updated"), Valid: true},
		EmailVerified: pgtype.Bool{Bool: true, Valid: true},
		Role: sqlc.NullUserRoleType{
			UserRoleType: sqlc.UserRoleTypeManager,
			Valid:        true,
		},
		Kind: sqlc.NullUserKind{
			UserKind: sqlc.UserKindStaff,
			Valid:    true,
		},
		IsActive: pgtype.Bool{Bool: true, Valid: true},
		ID:       created.ID,
	})
	require.NoError(t, err)

	require.Equal(t, sqlc.UserRoleTypeManager, updated.Role)
	require.Equal(t, sqlc.UserKindStaff, updated.Kind)
	require.True(t, updated.EmailVerified)
	require.True(t, updated.UpdatedAt.Valid)
}

func TestQueries_DeleteUser(t *testing.T) {
	queries, cleanup := setupTestDB(t)
	defer cleanup()

	created := insertDefaultUser(t, queries, uniqueEmail("delete"))

	err := queries.DeleteUser(context.Background(), created.ID)
	require.NoError(t, err)

	got, err := queries.GetUserByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.False(t, got.IsActive)
	require.True(t, got.DeletedAt.Valid)
}

func TestQueries_InsertUser_DuplicateEmail(t *testing.T) {
	queries, cleanup := setupTestDB(t)
	defer cleanup()

	email := uniqueEmail("duplicate")
	_ = insertDefaultUser(t, queries, email)

	_, err := queries.InsertUser(context.Background(), sqlc.InsertUserParams{
		Email:           email,
		EmailVerified:   false,
		EmailVerifiedAt: pgtype.Timestamptz{},
		Role:            sqlc.UserRoleTypeAdmin,
		Kind:            sqlc.UserKindOwner,
		IsActive:        true,
	})
	require.Error(t, err)

	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "23505", pgErr.Code)
}

func TestQueries_GetUserByEmail_NotFound(t *testing.T) {
	queries, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := queries.GetUserByEmail(context.Background(), uniqueEmail("missing-email"))
	require.Error(t, err)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}

func TestQueries_GetUserByID_NotFound(t *testing.T) {
	queries, cleanup := setupTestDB(t)
	defer cleanup()

	missingID := pgtype.UUID{Valid: true}
	_, err := queries.GetUserByID(context.Background(), missingID)
	require.Error(t, err)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}

func TestQueries_UpdateUser_NotFound(t *testing.T) {
	queries, cleanup := setupTestDB(t)
	defer cleanup()

	missingID := pgtype.UUID{Valid: true}
	_, err := queries.UpdateUser(context.Background(), sqlc.UpdateUserParams{
		Email:         pgtype.Text{String: uniqueEmail("updated-missing"), Valid: true},
		EmailVerified: pgtype.Bool{Bool: true, Valid: true},
		Role: sqlc.NullUserRoleType{
			UserRoleType: sqlc.UserRoleTypeManager,
			Valid:        true,
		},
		Kind: sqlc.NullUserKind{
			UserKind: sqlc.UserKindStaff,
			Valid:    true,
		},
		IsActive: pgtype.Bool{Bool: true, Valid: true},
		ID:       missingID,
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}
