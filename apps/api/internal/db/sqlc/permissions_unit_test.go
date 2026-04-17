package sqlc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_Permissions_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()
	errExpected := errors.New("db error")

	t.Run("GetPermissionByCode", func(t *testing.T) {
		code := "user:read"
		rows := pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "created_at", "updated_at"}).
			AddRow(uuidValue(), code, pgtype.Text{String: "Allows reading user info", Valid: true}, []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin}, pgtype.Timestamptz{Valid: true}, pgtype.Timestamptz{Valid: true})

		mock.ExpectQuery(`(?s)name: GetPermissionByCode`).
			WithArgs(code).
			WillReturnRows(rows)

		res, err := queries.GetPermissionByCode(ctx, code)
		require.NoError(t, err)
		require.Equal(t, code, res.Code)

		// Failure
		mock.ExpectQuery(`(?s)name: GetPermissionByCode`).
			WithArgs(code).
			WillReturnError(errExpected)

		_, err = queries.GetPermissionByCode(ctx, code)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("InsertPermission", func(t *testing.T) {
		arg := sqlc.InsertPermissionParams{
			Code:         "pet:write",
			DefaultRoles: []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin},
			Description:  pgtype.Text{String: "Allows creating pets", Valid: true},
		}

		mock.ExpectExec(`(?s)name: InsertPermission`).
			WithArgs(arg.Code, arg.Description, arg.DefaultRoles).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		rows, err := queries.InsertPermission(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		// Failure
		mock.ExpectExec(`(?s)name: InsertPermission`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.InsertPermission(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("ListPermissions", func(t *testing.T) {
		mock.ExpectQuery(`(?s)name: ListPermissions`).
			WithArgs(int32(0), int32(10)).
			WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "created_at", "updated_at"}).
				AddRow(uuidValue(), "p1", pgtype.Text{}, []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin}, pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		res, err := queries.ListPermissions(ctx, sqlc.ListPermissionsParams{Offset: 0, Limit: 10})
		require.NoError(t, err)
		require.Len(t, res, 1)

		// Failure
		mock.ExpectQuery(`(?s)name: ListPermissions`).
			WithArgs(int32(0), int32(10)).
			WillReturnError(errExpected)

		_, err = queries.ListPermissions(ctx, sqlc.ListPermissionsParams{Offset: 0, Limit: 10})
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("UpdatePermission", func(t *testing.T) {
		arg := sqlc.UpdatePermissionParams{
			ID:           uuidValue(),
			Code:         pgtype.Text{String: "new-code", Valid: true},
			DefaultRoles: []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin},
			Description:  pgtype.Text{String: "New Desc", Valid: true},
		}

		mock.ExpectExec(`(?s)name: UpdatePermission`).
			WithArgs(arg.Code, arg.Description, arg.DefaultRoles, arg.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		rows, err := queries.UpdatePermission(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		// Failure
		mock.ExpectExec(`(?s)name: UpdatePermission`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.UpdatePermission(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("DeletePermission", func(t *testing.T) {
		id := uuidValue()
		mock.ExpectExec(`(?s)name: DeletePermission`).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		rows, err := queries.DeletePermission(ctx, id)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		// Failure
		mock.ExpectExec(`(?s)name: DeletePermission`).
			WithArgs(id).
			WillReturnError(errExpected)

		_, err = queries.DeletePermission(ctx, id)
		require.ErrorIs(t, err, errExpected)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
