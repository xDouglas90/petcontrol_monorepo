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

func TestQueries_Clients_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	ctx := context.Background()

	t.Run("InsertClientRecord", func(t *testing.T) {
		arg := sqlc.InsertClientRecordParams{
			PersonID:    uuidValue(),
			ClientSince: pgtype.Date{Time: time.Now(), Valid: true},
			Notes:       pgtype.Text{String: "First client", Valid: true},
		}

		mock.ExpectQuery(`(?s)INSERT INTO clients`).
			WithArgs(arg.PersonID, arg.ClientSince, arg.Notes).
			WillReturnRows(pgxmock.NewRows([]string{"id", "person_id", "client_since", "recommended_by", "notes", "created_at", "updated_at", "deleted_at"}).
				AddRow(uuidValue(), arg.PersonID, arg.ClientSince, pgtype.UUID{}, arg.Notes, time.Now(), pgtype.Timestamptz{}, pgtype.Timestamptz{}))

		res, err := queries.InsertClientRecord(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.PersonID, res.PersonID)
	})

	t.Run("GetClientByIDAndCompanyID", func(t *testing.T) {
		arg := sqlc.GetClientByIDAndCompanyIDParams{
			CompanyID: uuidValue(),
			ID:        uuidValue(),
		}

		mock.ExpectQuery(`(?s)SELECT.*?FROM.*?company_clients`).
			WithArgs(arg.CompanyID, arg.ID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "person_id", "company_id", "full_name", "short_name", "gender_identity", "marital_status", "birth_date", "cpf", "email", "phone", "cellphone", "has_whatsapp", "client_since", "notes", "is_active", "created_at", "updated_at", "joined_at", "left_at",
			}).AddRow(
				arg.ID, uuidValue(), arg.CompanyID, "John Client", "John", sqlc.GenderIdentityManCisgender, sqlc.MaritalStatusSingle, pgtype.Date{Time: time.Now(), Valid: true}, "12345678901", "client@test.com", pgtype.Text{}, "11999998888", true, pgtype.Date{Time: time.Now(), Valid: true}, pgtype.Text{}, true, time.Now(), pgtype.Timestamptz{}, time.Now(), pgtype.Timestamptz{},
			))

		res, err := queries.GetClientByIDAndCompanyID(ctx, arg)
		require.NoError(t, err)
		require.Equal(t, arg.ID, res.ID)
		require.Equal(t, "John Client", res.FullName)
	})

	t.Run("DeactivateClient", func(t *testing.T) {
		arg := sqlc.DeactivateClientParams{
			CompanyID: uuidValue(),
			ClientID:  uuidValue(),
		}

		mock.ExpectExec(`(?s)UPDATE.*?company_clients`).
			WithArgs(arg.CompanyID, arg.ClientID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		rows, err := queries.DeactivateClient(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rows)
	})
}
