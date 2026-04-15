package sqlc_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestQueries_ClientPlans_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()

	t.Run("InsertClientPlan", func(t *testing.T) {
		arg := sqlc.InsertClientPlanParams{
			ClientID:  uuidValue(),
			PlanID:    uuidValue(),
			StartedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().AddDate(1, 0, 0), Valid: true},
			PricePaid: mustNumeric(t, "199.90"),
			IsActive:  pgtype.Bool{Bool: true, Valid: true},
		}

		mock.ExpectQuery(`(?s)INSERT INTO client_plans`).
			WithArgs(arg.ClientID, arg.PlanID, arg.StartedAt, arg.ExpiresAt, arg.PricePaid, arg.IsActive).
			WillReturnRows(pgxmock.NewRows([]string{"id", "client_id", "plan_id", "started_at", "expires_at", "price_paid", "is_active", "created_at", "updated_at"}).
				AddRow(uuidValue(), arg.ClientID, arg.PlanID, arg.StartedAt, arg.ExpiresAt, arg.PricePaid, true, time.Now(), pgtype.Timestamptz{}))

		res, err := queries.InsertClientPlan(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.ClientID, res.ClientID)
	})

	t.Run("GetClientPlanByID", func(t *testing.T) {
		id := uuidValue()
		mock.ExpectQuery(`(?s)SELECT.*?FROM.*?client_plans cp`).
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "client_id", "plan_id", "started_at", "expires_at", "price_paid", "is_active", "created_at", "updated_at",
				"plan_name", "plan_description", "plan_price", "plan_billing_cycle_days", "plan_max_users", "plan_is_active", "plan_image_url", "plan_created_at", "plan_updated_at",
				"identifications_full_name", "identifications_short_name", "identifications_gender_identity", "identifications_marital_status", "identifications_image_url", "identifications_birth_date", "identifications_cpf", "identifications_created_at", "identifications_updated_at",
				"client_since", "recommended_by", "client_notes",
			}).AddRow(
				id, uuidValue(), uuidValue(), time.Now(), time.Now().AddDate(1, 0, 0), mustNumeric(t, "199.90"), true, time.Now(), pgtype.Timestamptz{},
				"Premium Plan", "Best plan", mustNumeric(t, "199.90"), 365, pgtype.Int4{Int32: 5, Valid: true}, true, pgtype.Text{}, time.Now(), pgtype.Timestamptz{},
				"John Doe", "John", sqlc.GenderIdentityManCisgender, sqlc.MaritalStatusSingle, pgtype.Text{}, pgtype.Date{}, "12345678901", time.Now(), pgtype.Timestamptz{},
				pgtype.Date{}, pgtype.UUID{}, pgtype.Text{},
			))

		res, err := queries.GetClientPlanByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, id, res.ID)
		require.Equal(t, "Premium Plan", res.PlanName)
	})
}
