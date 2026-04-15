package sqlc

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestQueries_Services_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := New(mock)
	ctx := context.Background()

	t.Run("CreateService - Success", func(t *testing.T) {
		arg := CreateServiceParams{
			TypeID:      pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Title:       "Test Service",
			Description: "Test Description",
			Price:       pgtype.Numeric{Valid: true},
			IsActive:    true,
		}

		mock.ExpectQuery("INSERT INTO services").
			WithArgs(arg.TypeID, arg.Title, arg.Description, pgxmock.AnyArg(), arg.Price, pgxmock.AnyArg(), pgxmock.AnyArg(), arg.IsActive).
			WillReturnRows(pgxmock.NewRows([]string{"id", "type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{2}, Valid: true}, arg.TypeID, arg.Title, arg.Description, nil, nil, nil, nil, true, nil, nil, nil))

		res, err := queries.CreateService(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title, res.Title)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateServiceByIDAndCompanyID - Success", func(t *testing.T) {
		arg := UpdateServiceByIDAndCompanyIDParams{
			ID:        pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			CompanyID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Title:     pgtype.Text{String: "Updated Service", Valid: true},
		}

		mock.ExpectQuery("UPDATE services s").
			WithArgs(pgxmock.AnyArg(), arg.Title, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), arg.ID, arg.CompanyID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "type_id", "title", "description", "notes", "price", "discount_rate", "image_url", "is_active", "created_at", "updated_at", "deleted_at"}).
				AddRow(arg.ID, pgtype.UUID{}, arg.Title.String, "", nil, nil, nil, nil, true, nil, nil, nil))

		res, err := queries.UpdateServiceByIDAndCompanyID(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.Title.String, res.Title)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeactivateCompanyService - Success", func(t *testing.T) {
		arg := DeactivateCompanyServiceParams{
			CompanyID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			ServiceID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		}

		mock.ExpectQuery("UPDATE company_services").
			WithArgs(arg.CompanyID, arg.ServiceID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "service_id", "is_active", "created_at", "updated_at"}).
				AddRow(pgtype.UUID{Bytes: [16]byte{3}, Valid: true}, arg.CompanyID, arg.ServiceID, false, nil, nil))

		res, err := queries.DeactivateCompanyService(ctx, arg)
		require.NoError(t, err)
		require.False(t, res.IsActive)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
