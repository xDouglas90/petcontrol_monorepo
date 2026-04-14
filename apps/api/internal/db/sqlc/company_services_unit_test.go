package sqlc

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestQueries_CompanyServices_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := New(mock)
	ctx := context.Background()

	t.Run("InsertCompanyService - Success", func(t *testing.T) {
		arg := InsertCompanyServiceParams{
			CompanyID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			ServiceID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
			IsActive:  pgtype.Bool{Bool: true, Valid: true},
		}

		mock.ExpectQuery("INSERT INTO company_services").
			WithArgs(arg.CompanyID, arg.ServiceID, arg.IsActive).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "service_id", "is_active", "created_at", "updated_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{3}, Valid: true}, arg.CompanyID, arg.ServiceID, true, nil, nil))

		res, err := queries.InsertCompanyService(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.CompanyID, res.CompanyID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateCompanyService - Success", func(t *testing.T) {
		arg := UpdateCompanyServiceParams{
			CompanyID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			ServiceID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
			IsActive:  pgtype.Bool{Bool: false, Valid: true},
		}

		mock.ExpectQuery("UPDATE company_services").
			WithArgs(arg.IsActive, arg.CompanyID, arg.ServiceID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "service_id", "is_active", "created_at", "updated_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{3}, Valid: true}, arg.CompanyID, arg.ServiceID, false, nil, nil))

		res, err := queries.UpdateCompanyService(ctx, arg)
		require.NoError(t, err)
		require.False(t, res.IsActive)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteCompanyService - Success", func(t *testing.T) {
		arg := DeleteCompanyServiceParams{
			CompanyID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			ServiceID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		}

		mock.ExpectQuery("DELETE FROM company_services").
			WithArgs(arg.CompanyID, arg.ServiceID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "service_id", "is_active", "created_at", "updated_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{3}, Valid: true}, arg.CompanyID, arg.ServiceID, true, nil, nil))

		res, err := queries.DeleteCompanyService(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.ServiceID, res.ServiceID)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
