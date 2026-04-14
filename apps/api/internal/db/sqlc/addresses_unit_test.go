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

func TestQueries_Addresses_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()
	errExpected := errors.New("db error")

	t.Run("InsertAddress", func(t *testing.T) {
		arg := sqlc.InsertAddressParams{
			ZipCode:  "12345-678",
			Street:   "Main St",
			Number:   "100",
			District: "Central",
			City:     "Tech City",
			State:    "TC",
			Country:  "Mockland",
		}

		// Success
		mock.ExpectExec(`(?s)name: InsertAddress`).
			WithArgs(arg.ZipCode, arg.Street, arg.Number, arg.Complement, arg.District, arg.City, arg.State, arg.Country).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		rows, err := queries.InsertAddress(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		// Database error
		mock.ExpectExec(`(?s)name: InsertAddress`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.InsertAddress(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("GetAddress", func(t *testing.T) {
		id := uuidValue()

		// Success
		rows := pgxmock.NewRows([]string{"zip_code", "street", "number", "complement", "district", "city", "state", "country", "created_at", "updated_at"}).
			AddRow("12345-678", "Main St", "100", pgtype.Text{String: "Apt 1", Valid: true}, "Central", "Tech City", "TC", "Mockland", pgtype.Timestamptz{Valid: true}, pgtype.Timestamptz{Valid: true})

		mock.ExpectQuery(`(?s)name: GetAddress`).
			WithArgs(id).
			WillReturnRows(rows)

		addr, err := queries.GetAddress(ctx, id)
		require.NoError(t, err)
		require.Equal(t, "12345-678", addr.ZipCode)

		// Database error
		mock.ExpectQuery(`(?s)name: GetAddress`).
			WithArgs(id).
			WillReturnError(errExpected)

		_, err = queries.GetAddress(ctx, id)
		require.ErrorIs(t, err, errExpected)
	})

	t.Run("UpdateAddress", func(t *testing.T) {
		arg := sqlc.UpdateAddressParams{
			ID:      uuidValue(),
			ZipCode: pgtype.Text{String: "87654-321", Valid: true},
			Street:  pgtype.Text{String: "New St", Valid: true},
		}

		// Success
		mock.ExpectExec(`(?s)name: UpdateAddress`).
			WithArgs(arg.ZipCode, arg.Street, arg.Number, arg.Complement, arg.District, arg.City, arg.State, arg.Country, arg.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		rows, err := queries.UpdateAddress(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)

		// Database error
		mock.ExpectExec(`(?s)name: UpdateAddress`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errExpected)

		_, err = queries.UpdateAddress(ctx, arg)
		require.ErrorIs(t, err, errExpected)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
