package sqlc_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_Auth_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()
	errExpected := errors.New("db error")

	t.Run("GetUserAuthByUserID", func(t *testing.T) {
		userID := uuidValue()

		// Success
		rows := pgxmock.NewRows([]string{"user_id", "password_hash", "password_changed_at", "must_change_password", "login_attempts", "locked_until", "last_login_at", "created_at", "updated_at"}).
			AddRow(userID, "hash", pgtype.Timestamptz{Valid: true}, false, int32(0), pgtype.Timestamptz{}, pgtype.Timestamptz{}, pgtype.Timestamptz{Valid: true}, pgtype.Timestamptz{Valid: true})

		mock.ExpectQuery(`(?s)name: GetUserAuthByUserID`).
			WithArgs(userID).
			WillReturnRows(rows)

		res, err := queries.GetUserAuthByUserID(ctx, userID)
		require.NoError(t, err)
		require.Equal(t, userID, res.UserID)

		// Failure
		mock.ExpectQuery(`(?s)name: GetUserAuthByUserID`).
			WithArgs(userID).
			WillReturnError(errExpected)

		_, err = queries.GetUserAuthByUserID(ctx, userID)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("HasActiveCompanyModuleByCode", func(t *testing.T) {
		arg := sqlc.HasActiveCompanyModuleByCodeParams{
			CompanyID: uuidValue(),
			Code:      "SCH",
		}

		mock.ExpectQuery(`(?s)name: HasActiveCompanyModuleByCode`).
			WithArgs(arg.CompanyID, arg.Code).
			WillReturnRows(pgxmock.NewRows([]string{"has_access"}).AddRow(true))

		has, err := queries.HasActiveCompanyModuleByCode(ctx, arg)
		require.NoError(t, err)
		require.True(t, has)

		mock.ExpectQuery(`(?s)name: HasActiveCompanyModuleByCode`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.HasActiveCompanyModuleByCode(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("IncrementUserAuthLoginAttempts", func(t *testing.T) {
		userID := uuidValue()

		mock.ExpectExec(`(?s)name: IncrementUserAuthLoginAttempts`).
			WithArgs(userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := queries.IncrementUserAuthLoginAttempts(ctx, userID)
		require.NoError(t, err)

		mock.ExpectExec(`(?s)name: IncrementUserAuthLoginAttempts`).
			WithArgs(pgxmock.AnyArg()).
			WillReturnError(errExpected)

		err = queries.IncrementUserAuthLoginAttempts(ctx, userID)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("InsertLoginHistory", func(t *testing.T) {
		arg := sqlc.InsertLoginHistoryParams{
			UserID:    uuidValue(),
			IPAddress: netip.MustParseAddr("127.0.0.1"),
			UserAgent: "Mozilla",
			Result:    sqlc.LoginResultSuccess,
		}

		mock.ExpectExec(`(?s)name: InsertLoginHistory`).
			WithArgs(arg.UserID, arg.IPAddress, arg.UserAgent, arg.Result, arg.FailureDetail).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := queries.InsertLoginHistory(ctx, arg)
		require.NoError(t, err)

		mock.ExpectExec(`(?s)name: InsertLoginHistory`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		err = queries.InsertLoginHistory(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("ResetUserAuthLoginAttempts", func(t *testing.T) {
		userID := uuidValue()

		mock.ExpectExec(`(?s)name: ResetUserAuthLoginAttempts`).
			WithArgs(userID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := queries.ResetUserAuthLoginAttempts(ctx, userID)
		require.NoError(t, err)
	})

	t.Run("SetUserAuthLockedUntil", func(t *testing.T) {
		arg := sqlc.SetUserAuthLockedUntilParams{
			UserID:      uuidValue(),
			LockedUntil: pgtype.Timestamptz{Valid: true},
		}

		mock.ExpectExec(`(?s)name: SetUserAuthLockedUntil`).
			WithArgs(arg.LockedUntil, arg.UserID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := queries.SetUserAuthLockedUntil(ctx, arg)
		require.NoError(t, err)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
