package sqlc

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestQueries_ServiceTypes_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := New(mock)
	ctx := context.Background()

	t.Run("InsertServiceType - Success", func(t *testing.T) {
		arg := InsertServiceTypeParams{
			Name:        "Test Type",
			Description: pgtype.Text{String: "Test Desc", Valid: true},
		}

		mock.ExpectQuery("INSERT INTO service_types").
			WithArgs(arg.Name, arg.Description).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at", "deleted_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, arg.Name, arg.Description, nil, nil, nil))

		res, err := queries.InsertServiceType(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Name, res.Name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateServiceType - Success", func(t *testing.T) {
		arg := UpdateServiceTypeParams{
			ID:   pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Name: pgtype.Text{String: "Updated Type", Valid: true},
		}

		mock.ExpectQuery("UPDATE service_types").
			WithArgs(arg.Name, pgxmock.AnyArg(), arg.ID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at", "deleted_at"}).
				AddRow(arg.ID, arg.Name.String, pgtype.Text{}, nil, nil, nil))

		res, err := queries.UpdateServiceType(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Name.String, res.Name)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteServiceType - Success", func(t *testing.T) {
		id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}

			mock.ExpectQuery("UPDATE service_types SET deleted_at").
				WithArgs(id).
				WillReturnRows(pgxmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at", "deleted_at"}).
					AddRow(id, "Deleted", pgtype.Text{}, nil, nil, pgtype.Timestamptz{Valid: true}))

		res, err := queries.DeleteServiceType(ctx, id)
		require.NoError(t, err)
		require.True(t, res.DeletedAt.Valid)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
