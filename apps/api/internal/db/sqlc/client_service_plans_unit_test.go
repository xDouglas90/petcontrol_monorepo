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

func TestQueries_ClientServicePlans_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()

	t.Run("InsertClientServicePlan", func(t *testing.T) {
		arg := sqlc.InsertClientServicePlanParams{
			ClientID:      uuidValue(),
			ServicePlanID: uuidValue(),
			StartedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpiresAt:     pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Valid: true},
			PricePaid:     mustNumeric(t, "50.00"),
			IsActive:      pgtype.Bool{Bool: true, Valid: true},
		}

		mock.ExpectQuery(`(?s)INSERT INTO client_service_plans`).
			WithArgs(arg.ClientID, arg.ServicePlanID, arg.StartedAt, arg.ExpiresAt, arg.PricePaid, arg.IsActive).
			WillReturnRows(pgxmock.NewRows([]string{"id", "client_id", "service_plan_id", "started_at", "expires_at", "price_paid", "is_active", "created_at", "updated_at"}).
				AddRow(uuidValue(), arg.ClientID, arg.ServicePlanID, arg.StartedAt, arg.ExpiresAt, arg.PricePaid, true, time.Now(), pgtype.Timestamptz{}))

		res, err := queries.InsertClientServicePlan(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.ClientID, res.ClientID)
	})

	t.Run("ListClientServicePlans", func(t *testing.T) {
		clientID := uuidValue()
		mock.ExpectQuery(`(?s)SELECT.*?FROM.*?client_service_plans csp`).
			WithArgs(clientID, int32(0), int32(10)).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "client_id", "service_plan_id", "started_at", "expires_at", "price_paid", "is_active", "created_at", "updated_at",
				"service_plan_title", "service_plan_description", "service_plan_price", "service_plan_discount_rate", "service_plan_image_url",
			}).AddRow(
				uuidValue(), clientID, uuidValue(), time.Now(), time.Now().AddDate(0, 1, 0), mustNumeric(t, "50.00"), true, time.Now(), pgtype.Timestamptz{},
				"Service Pack", "Basic pack", mustNumeric(t, "50.00"), mustNumeric(t, "0.00"), pgtype.Text{},
			))

		res, err := queries.ListClientServicePlans(ctx, sqlc.ListClientServicePlansParams{
			ClientID: clientID,
			Limit:    10,
			Offset:   0,
		})
		require.NoError(t, err)
		require.NotEmpty(t, res)
		require.Equal(t, "Service Pack", res[0].ServicePlanTitle)
	})
}
