package service

import (
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestIsUndefinedTableError(t *testing.T) {
	t.Run("returns true for undefined table with matching name", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:    "42P01",
			Message: `relation "pet_guardians" does not exist`,
		}

		if !isUndefinedTableError(err, "pet_guardians") {
			t.Fatalf("expected undefined table error to match")
		}
	})

	t.Run("returns false when table name does not match", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:    "42P01",
			Message: `relation "company_users" does not exist`,
		}

		if isUndefinedTableError(err, "pet_guardians") {
			t.Fatalf("expected undefined table error not to match different table")
		}
	})

	t.Run("returns false for non pg errors", func(t *testing.T) {
		if isUndefinedTableError(assertionError("boom"), "pet_guardians") {
			t.Fatalf("expected non pg error to not match")
		}
	})
}

type assertionError string

func (e assertionError) Error() string {
	return string(e)
}

func TestEnsureUserRoleCreationAllowed(t *testing.T) {
	require.NoError(t, ensureUserRoleCreationAllowed(sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeRoot))
	require.NoError(t, ensureUserRoleCreationAllowed(sqlc.UserRoleTypeInternal, sqlc.UserRoleTypeInternal))
	require.NoError(t, ensureUserRoleCreationAllowed(sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem))

	require.ErrorIs(t, ensureUserRoleCreationAllowed(sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeRoot), apperror.ErrForbidden)
	require.ErrorIs(t, ensureUserRoleCreationAllowed(sqlc.UserRoleTypeSystem, sqlc.UserRoleTypeInternal), apperror.ErrForbidden)
}

func TestFilterPeopleListItemsBySearch(t *testing.T) {
	items := []PeopleListItem{
		{
			ID:       testUUID(t),
			Kind:     sqlc.PersonKindClient,
			FullName: stringPointer("Maria Silva"),
			CPF:      stringPointer("12345678901"),
		},
		{
			ID:        testUUID(t),
			Kind:      sqlc.PersonKindSupplier,
			FullName:  stringPointer("Fornecedor XPTO"),
			ShortName: stringPointer("XPTO"),
		},
	}

	filtered := filterPeopleListItemsBySearch(items, "xpto")
	require.Len(t, filtered, 1)
	require.Equal(t, "Fornecedor XPTO", stringValue(filtered[0].FullName))
}

func TestSortablePeopleName(t *testing.T) {
	require.Equal(t, "ana lima", sortablePeopleName(PeopleListItem{
		ID:       testUUID(t),
		FullName: stringPointer(" Ana Lima "),
	}))
	require.Equal(t, "aninha", sortablePeopleName(PeopleListItem{
		ID:        testUUID(t),
		ShortName: stringPointer("Aninha"),
	}))
}

func TestPeopleListSortingPrefersNameThenCreatedAt(t *testing.T) {
	older := timestamptz(time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC))
	newer := timestamptz(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC))

	items := []PeopleListItem{
		{
			ID:        testUUIDFromString("00000000-0000-0000-0000-000000000003"),
			FullName:  stringPointer("Zulu"),
			CreatedAt: newer,
		},
		{
			ID:        testUUIDFromString("00000000-0000-0000-0000-000000000002"),
			FullName:  stringPointer("Ana"),
			CreatedAt: newer,
		},
		{
			ID:        testUUIDFromString("00000000-0000-0000-0000-000000000001"),
			FullName:  stringPointer("Ana"),
			CreatedAt: older,
		},
	}

	sort.SliceStable(items, func(i, j int) bool {
		leftName := sortablePeopleName(items[i])
		rightName := sortablePeopleName(items[j])
		if leftName != rightName {
			return leftName < rightName
		}

		leftCreatedAt := items[i].CreatedAt.Time
		rightCreatedAt := items[j].CreatedAt.Time
		if !leftCreatedAt.Equal(rightCreatedAt) {
			return leftCreatedAt.Before(rightCreatedAt)
		}

		return uuidKey(items[i].ID) < uuidKey(items[j].ID)
	})

	require.Equal(t, "00000000-0000-0000-0000-000000000001", uuidKey(items[0].ID))
	require.Equal(t, "00000000-0000-0000-0000-000000000002", uuidKey(items[1].ID))
	require.Equal(t, "00000000-0000-0000-0000-000000000003", uuidKey(items[2].ID))
}

func testUUID(t *testing.T) pgtype.UUID {
	t.Helper()
	return testUUIDFromString(uuid.NewString())
}

func testUUIDFromString(raw string) pgtype.UUID {
	parsed := uuid.MustParse(raw)
	var out pgtype.UUID
	copy(out.Bytes[:], parsed[:])
	out.Valid = true
	return out
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}
