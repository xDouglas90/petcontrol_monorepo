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

func TestQueries_CompanyUsers_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()
	errExpected := errors.New("db error")

	t.Run("CreateCompanyUser", func(t *testing.T) {
		arg := sqlc.CreateCompanyUserParams{
			CompanyID: uuidValue(),
			UserID:    uuidValue(),
			Kind:      sqlc.UserKindEmployee,
			IsActive:  pgtype.Bool{Bool: true, Valid: true},
		}

		rows := pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(uuidValue(), arg.CompanyID, arg.UserID, arg.Kind, arg.IsActive.Bool, pgtype.Timestamptz{}, pgtype.Timestamptz{}, pgtype.Timestamptz{})

		mock.ExpectQuery(`(?s)name: CreateCompanyUser`).
			WithArgs(arg.CompanyID, arg.UserID, arg.Kind, arg.IsActive).
			WillReturnRows(rows)

		res, err := queries.CreateCompanyUser(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Kind, res.Kind)

		// Failure
		mock.ExpectQuery(`(?s)name: CreateCompanyUser`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.CreateCompanyUser(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("DeactivateCompanyUser", func(t *testing.T) {
		arg := sqlc.DeactivateCompanyUserParams{
			CompanyID: uuidValue(),
			UserID:    uuidValue(),
		}

		mock.ExpectExec(`(?s)name: DeactivateCompanyUser`).
			WithArgs(arg.CompanyID, arg.UserID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := queries.DeactivateCompanyUser(ctx, arg)
		require.NoError(t, err)

		// Failure
		mock.ExpectExec(`(?s)name: DeactivateCompanyUser`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		err = queries.DeactivateCompanyUser(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("GetActiveCompanyUserByUserID", func(t *testing.T) {
		userID := uuidValue()
		mock.ExpectQuery(`(?s)name: GetActiveCompanyUserByUserID`).
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(uuidValue(), uuidValue(), userID, sqlc.UserKindEmployee, true, pgtype.Timestamptz{}, pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		res, err := queries.GetActiveCompanyUserByUserID(ctx, userID)
		require.NoError(t, err)
		require.True(t, res.IsActive)

		// Failure
		mock.ExpectQuery(`(?s)name: GetActiveCompanyUserByUserID`).
			WithArgs(userID).
			WillReturnError(errExpected)

		_, err = queries.GetActiveCompanyUserByUserID(ctx, userID)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("ListCompanyUsersByKind", func(t *testing.T) {
		arg := sqlc.ListCompanyUsersByKindParams{
			CompanyID: uuidValue(),
			Kind:      sqlc.UserKindEmployee,
			Limit:     10,
			Offset:    0,
		}

		mock.ExpectQuery(`(?s)name: ListCompanyUsersByKind`).
			WithArgs(arg.CompanyID, arg.Kind, arg.Offset, arg.Limit).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(uuidValue(), arg.CompanyID, uuidValue(), arg.Kind, true, pgtype.Timestamptz{}, pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		res, err := queries.ListCompanyUsersByKind(ctx, arg)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
