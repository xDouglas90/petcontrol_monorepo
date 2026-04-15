package sqlc

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestQueries_ServicePlans_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := New(mock)
	ctx := context.Background()

	t.Run("InsertServicePlan - Success", func(t *testing.T) {
		arg := InsertServicePlanParams{
			PlanTypeID:  pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Title:       "Basic Plan",
			Description: "Basic service plan",
			Price:       pgtype.Numeric{Valid: true},
			IsActive:    pgtype.Bool{Bool: true, Valid: true},
		}

		mock.ExpectQuery("INSERT INTO service_plans").
			WithArgs(arg.PlanTypeID, arg.Title, arg.Description, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), arg.IsActive).
			WillReturnRows(pgxmock.NewRows([]string{"id", "plan_type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{2}, Valid: true}, arg.PlanTypeID, arg.Title, arg.Description, nil, nil, nil, nil, true, nil, nil, nil))

		res, err := queries.InsertServicePlan(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title, res.Title)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateServicePlan - Success", func(t *testing.T) {
		arg := UpdateServicePlanParams{
			ID:    pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Title: pgtype.Text{String: "Standard Plan", Valid: true},
		}

		mock.ExpectQuery("UPDATE service_plans").
			WithArgs(pgxmock.AnyArg(), arg.Title, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), arg.ID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "plan_type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(arg.ID, pgtype.UUID{}, arg.Title.String, "", nil, nil, nil, nil, true, nil, nil, nil))

		res, err := queries.UpdateServicePlan(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title.String, res.Title)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteServicePlan - Success", func(t *testing.T) {
		id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}

		mock.ExpectQuery("UPDATE service_plans SET deleted_at").
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "plan_type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(id, pgtype.UUID{}, "Deleted", "", nil, nil, nil, nil, false, nil, nil, pgtype.Timestamptz{Valid: true}))

		res, err := queries.DeleteServicePlan(ctx, id)
		require.NoError(t, err)
		require.True(t, res.DeletedAt.Valid)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
