package sqlc

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestQueries_CompanyServicePlans_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := New(mock)
	ctx := context.Background()

	t.Run("InsertCompanyServicePlan - Success", func(t *testing.T) {
		arg := InsertCompanyServicePlanParams{
			CompanyID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			ServicePlanID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
			IsActive:      pgtype.Bool{Bool: true, Valid: true},
		}

		mock.ExpectQuery("INSERT INTO company_service_plans").
			WithArgs(arg.CompanyID, arg.ServicePlanID, arg.IsActive).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "service_plan_id", "is_active", "created_at", "updated_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{3}, Valid: true}, arg.CompanyID, arg.ServicePlanID, true, nil, nil))

		res, err := queries.InsertCompanyServicePlan(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.CompanyID, res.CompanyID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateCompanyServicePlan - Success", func(t *testing.T) {
		arg := UpdateCompanyServicePlanParams{
			CompanyID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			ServicePlanID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
			IsActive:      pgtype.Bool{Bool: false, Valid: true},
		}

		mock.ExpectQuery("UPDATE company_service_plans").
			WithArgs(arg.IsActive, arg.CompanyID, arg.ServicePlanID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "service_plan_id", "is_active", "created_at", "updated_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{3}, Valid: true}, arg.CompanyID, arg.ServicePlanID, false, nil, nil))

		res, err := queries.UpdateCompanyServicePlan(ctx, arg)
		require.NoError(t, err)
		require.False(t, res.IsActive)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteCompanyServicePlan - Success", func(t *testing.T) {
		arg := DeleteCompanyServicePlanParams{
			CompanyID:     pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			ServicePlanID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		}

		mock.ExpectQuery("DELETE FROM company_service_plans").
			WithArgs(arg.CompanyID, arg.ServicePlanID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "service_plan_id", "is_active", "created_at", "updated_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{3}, Valid: true}, arg.CompanyID, arg.ServicePlanID, true, nil, nil))

		res, err := queries.DeleteCompanyServicePlan(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.ServicePlanID, res.ServicePlanID)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
