package sqlc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func createTestAddressParams() sqlc.InsertAddressParams {
	return sqlc.InsertAddressParams{
		ZipCode:    "01001-000",
		Street:     "Praça da Sé",
		Number:     "s/n",
		Complement: pgtype.Text{String: "Lado Par", Valid: true},
		District:   "Sé",
		City:       "São Paulo",
		State:      "SP",
		Country:    "Brasil",
	}
}

func mustInsertAddressDirectly(t *testing.T, pool interface {
	QueryRow(context.Context, string, ...interface{}) pgx.Row
},
) pgtype.UUID {
	t.Helper()

	var id pgtype.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO addresses(zip_code, street, number, complement, district, city, state, country)
         VALUES ('01001-000', 'Praça da Sé', 's/n', 'Lado Par', 'Sé', 'São Paulo', 'SP', 'Brasil')
         RETURNING id`,
	).Scan(&id)
	require.NoError(t, err)
	require.True(t, id.Valid)
	return id
}

func TestQueries_Addresses_Insert(t *testing.T) {
	queries, ctx, _ := setupQueriesWithPool(t)

	t.Run("success", func(t *testing.T) {
		arg := createTestAddressParams()
		rowsAffected, err := queries.InsertAddress(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rowsAffected)
	})

	t.Run("failure_invalid_state_length", func(t *testing.T) {
		arg := createTestAddressParams()
		// O campo state é char(2) no banco. Passar uma string maior causará erro 22001 (truncation).
		arg.State = "SÃO PAULO"
		_, err := queries.InsertAddress(ctx, arg)
		require.Error(t, err)

		var pgErr *pgconn.PgError
		require.ErrorAs(t, err, &pgErr)
		require.Equal(t, "22001", pgErr.Code)
	})
}

func TestQueries_Addresses_Get(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)

	t.Run("success", func(t *testing.T) {
		id := mustInsertAddressDirectly(t, pool)

		address, err := queries.GetAddress(ctx, id)
		require.NoError(t, err)
		require.Equal(t, "01001-000", address.ZipCode)
		require.True(t, address.CreatedAt.Valid)
	})

	t.Run("failure_not_found", func(t *testing.T) {
		missingID := pgtype.UUID{Valid: true}
		_, err := queries.GetAddress(ctx, missingID)
		require.Error(t, err)
		require.True(t, errors.Is(err, pgx.ErrNoRows))
	})
}

func TestQueries_Addresses_Update(t *testing.T) {
	queries, ctx, pool := setupQueriesWithPool(t)

	t.Run("success", func(t *testing.T) {
		id := mustInsertAddressDirectly(t, pool)

		arg := sqlc.UpdateAddressParams{
			ID:      id,
			ZipCode: pgtype.Text{String: "01002-000", Valid: true},
			Street:  pgtype.Text{String: "Rua Direita", Valid: true},
			Number:  pgtype.Text{String: "123", Valid: true},
			City:    pgtype.Text{String: "São Paulo", Valid: true},
			State:   pgtype.Text{String: "SP", Valid: true},
			Country: pgtype.Text{String: "Brasil", Valid: true},
		}

		rowsAffected, err := queries.UpdateAddress(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rowsAffected)

		updated, err := queries.GetAddress(ctx, id)
		require.NoError(t, err)
		require.Equal(t, "01002-000", updated.ZipCode)
		require.Equal(t, "Rua Direita", updated.Street)
	})

	t.Run("failure_not_found", func(t *testing.T) {
		missingID := pgtype.UUID{Valid: true}
		arg := sqlc.UpdateAddressParams{
			ID:      missingID,
			Street:  pgtype.Text{String: "Nowhere", Valid: true},
			City:    pgtype.Text{String: "Lost", Valid: true},
			State:   pgtype.Text{String: "??", Valid: true},
			Country: pgtype.Text{String: "Empty", Valid: true},
		}
		rowsAffected, err := queries.UpdateAddress(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 0, rowsAffected)
	})

	t.Run("partial_update_success", func(t *testing.T) {
		id := mustInsertAddressDirectly(t, pool)

		// Vamos atualizar apenas a rua, mantendo o zip_code e numero originais via COALESCE
		arg := sqlc.UpdateAddressParams{
			ID:     id,
			Street: pgtype.Text{String: "Rua Nova Apenas", Valid: true},
			// Outros campos ficam como Valid: false (NULL no SQL)
		}

		rowsAffected, err := queries.UpdateAddress(ctx, arg)
		require.NoError(t, err)
		require.EqualValues(t, 1, rowsAffected)

		updated, err := queries.GetAddress(ctx, id)
		require.NoError(t, err)
		require.Equal(t, "Rua Nova Apenas", updated.Street)
		require.Equal(t, "01001-000", updated.ZipCode) // Valor original deve ter sido preservado pelo COALESCE
		require.Equal(t, "s/n", updated.Number)        // Valor original preservado
	})
}
