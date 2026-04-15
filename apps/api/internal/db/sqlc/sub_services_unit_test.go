package sqlc

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestQueries_SubServices_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := New(mock)
	ctx := context.Background()

	t.Run("InsertSubService - Success", func(t *testing.T) {
		arg := InsertSubServiceParams{
			ServiceID:   pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			TypeID:      pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
			Title:       "Sub-Service X",
			Description: "Description X",
			Price:       pgtype.Numeric{Valid: true},
		}

		mock.ExpectQuery("INSERT INTO sub_services").
			WithArgs(arg.ServiceID, arg.TypeID, arg.Title, arg.Description, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"id", "service_id", "type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{3}, Valid: true}, arg.ServiceID, arg.TypeID, arg.Title, arg.Description, nil, nil, nil, nil, true, nil, nil, nil))

		res, err := queries.InsertSubService(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title, res.Title)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateSubService - Success", func(t *testing.T) {
		arg := UpdateSubServiceParams{
			ID:    pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Title: pgtype.Text{String: "Updated Sub-Service", Valid: true},
		}

		mock.ExpectQuery("UPDATE sub_services").
			WithArgs(pgxmock.AnyArg(), arg.Title, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), arg.ID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "service_id", "type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(arg.ID, pgtype.UUID{}, pgtype.UUID{}, arg.Title.String, "", nil, nil, nil, nil, true, nil, nil, nil))

		res, err := queries.UpdateSubService(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title.String, res.Title)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteSubService - Success", func(t *testing.T) {
		id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}

		mock.ExpectQuery("UPDATE sub_services SET deleted_at").
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "service_id", "type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(id, pgtype.UUID{}, pgtype.UUID{}, "Deleted", "", nil, nil, nil, nil, false, nil, nil, pgtype.Timestamptz{Valid: true}))

		res, err := queries.DeleteSubService(ctx, id)
		require.NoError(t, err)
		require.True(t, res.DeletedAt.Valid)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
