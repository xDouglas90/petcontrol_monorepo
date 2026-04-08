package service

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
	"golang.org/x/crypto/bcrypt"
)

func newPGUUID(t *testing.T) pgtype.UUID {
	t.Helper()

	value := uuid.New()
	var out pgtype.UUID
	copy(out.Bytes[:], value[:])
	out.Valid = true
	return out
}

func newUserRows(id pgtype.UUID, email string, verified bool, isActive bool, role sqlc.UserRoleType, kind sqlc.UserKind) *pgxmock.Rows {
	var verifiedAt any = nil
	if verified {
		verifiedAt = time.Now().Add(-time.Minute)
	}

	return pgxmock.NewRows([]string{"id", "email", "email_verified", "email_verified_at", "role", "kind", "is_active", "created_at", "updated_at", "deleted_at"}).AddRow(id.String(), email, verified, verifiedAt, role, kind, isActive, time.Now().Add(-time.Hour), nil, nil)
}

func newUserAuthRows(userID pgtype.UUID, passwordHash string, attempts int16, lockedUntilValid bool) *pgxmock.Rows {
	var lockedUntil any = nil
	if lockedUntilValid {
		lockedUntil = time.Now().Add(15 * time.Minute)
	}

	return pgxmock.NewRows([]string{"user_id", "password_hash", "password_changed_at", "must_change_password", "login_attempts", "locked_until", "last_login_at", "created_at", "updated_at"}).AddRow(userID.String(), passwordHash, nil, false, attempts, lockedUntil, nil, time.Now().Add(-time.Hour), nil)
}

func newCompanyUserRows(companyID, userID pgtype.UUID) *pgxmock.Rows {
	return pgxmock.NewRows([]string{"id", "company_id", "user_id", "is_owner", "is_active", "joined_at", "left_at"}).AddRow(uuid.NewString(), companyID.String(), userID.String(), true, true, time.Now().Add(-time.Minute), nil)
}

func newAuthService(t *testing.T) (*AuthService, pgxmock.PgxPoolIface) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	return NewAuthService(sqlc.New(mock), "secret", time.Hour), mock
}

func validPasswordHash(t *testing.T, raw string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hash)
}

func TestAuthService_Login_Success(t *testing.T) {
	serviceUnderTest, mock := newAuthService(t)
	defer mock.Close()

	userID := newPGUUID(t)
	companyID := newPGUUID(t)
	password := "super-secret"
	hash := validPasswordHash(t, password)

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("owner@example.com").WillReturnRows(newUserRows(userID, "owner@example.com", true, true, sqlc.UserRoleTypeAdmin, sqlc.UserKindOwner))
	mock.ExpectQuery(`(?s)name: GetUserAuthByUserID`).WithArgs(userID).WillReturnRows(newUserAuthRows(userID, hash, 0, false))
	mock.ExpectExec(`(?s)name: ResetUserAuthLoginAttempts`).WithArgs(userID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery(`(?s)name: GetActiveCompanyUserByUserID`).WithArgs(userID).WillReturnRows(newCompanyUserRows(companyID, userID))
	mock.ExpectExec(`(?s)name: InsertLoginHistory`).WithArgs(userID, pgxmock.AnyArg(), "TestAgent/1.0", sqlc.LoginResultSuccess, pgtype.Text{}).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	result, err := serviceUnderTest.Login(context.Background(), " owner@example.com ", password, "127.0.0.1", "TestAgent/1.0")
	require.NoError(t, err)
	require.Equal(t, "Bearer", result.TokenType)
	require.Equal(t, userID.String(), result.UserID)
	require.Equal(t, companyID.String(), result.CompanyID)
	require.Equal(t, "admin", result.Role)
	require.Equal(t, "owner", result.Kind)
	require.NotEmpty(t, result.AccessToken)

	claims, err := appjwt.ParseToken("secret", result.AccessToken)
	require.NoError(t, err)
	require.Equal(t, result.UserID, claims.UserID)
	require.Equal(t, result.CompanyID, claims.CompanyID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_InvalidCredentials_LocksAccount(t *testing.T) {
	serviceUnderTest, mock := newAuthService(t)
	defer mock.Close()

	userID := newPGUUID(t)
	hash := validPasswordHash(t, "super-secret")

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("owner@example.com").WillReturnRows(newUserRows(userID, "owner@example.com", true, true, sqlc.UserRoleTypeAdmin, sqlc.UserKindOwner))
	mock.ExpectQuery(`(?s)name: GetUserAuthByUserID`).WithArgs(userID).WillReturnRows(newUserAuthRows(userID, hash, 4, false))
	mock.ExpectExec(`(?s)name: IncrementUserAuthLoginAttempts`).WithArgs(userID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(`(?s)name: SetUserAuthLockedUntil`).WithArgs(pgxmock.AnyArg(), userID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(`(?s)name: InsertLoginHistory`).WithArgs(userID, pgxmock.AnyArg(), "TestAgent/1.0", sqlc.LoginResultInvalidCredentials, pgtype.Text{String: "password mismatch", Valid: true}).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	_, err := serviceUnderTest.Login(context.Background(), "owner@example.com", "wrong-password", "127.0.0.1", "TestAgent/1.0")
	require.ErrorIs(t, err, apperror.ErrInvalidCredentials)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	serviceUnderTest, mock := newAuthService(t)
	defer mock.Close()

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("missing@example.com").WillReturnError(pgx.ErrNoRows)
	mock.ExpectExec(`(?s)name: InsertLoginHistory`).WithArgs(pgtype.UUID{}, pgxmock.AnyArg(), "TestAgent/1.0", sqlc.LoginResultInvalidCredentials, pgtype.Text{String: "user not found", Valid: true}).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	_, err := serviceUnderTest.Login(context.Background(), "missing@example.com", "secret", "127.0.0.1", "TestAgent/1.0")
	require.ErrorIs(t, err, apperror.ErrInvalidCredentials)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_MissingAuthProfile(t *testing.T) {
	serviceUnderTest, mock := newAuthService(t)
	defer mock.Close()

	userID := newPGUUID(t)

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("owner@example.com").WillReturnRows(newUserRows(userID, "owner@example.com", true, true, sqlc.UserRoleTypeAdmin, sqlc.UserKindOwner))
	mock.ExpectQuery(`(?s)name: GetUserAuthByUserID`).WithArgs(userID).WillReturnError(pgx.ErrNoRows)
	mock.ExpectExec(`(?s)name: InsertLoginHistory`).WithArgs(userID, pgxmock.AnyArg(), "TestAgent/1.0", sqlc.LoginResultInvalidCredentials, pgtype.Text{String: "auth profile not found", Valid: true}).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	_, err := serviceUnderTest.Login(context.Background(), "owner@example.com", "secret", "127.0.0.1", "TestAgent/1.0")
	require.ErrorIs(t, err, apperror.ErrInvalidCredentials)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_AccountInactive(t *testing.T) {
	serviceUnderTest, mock := newAuthService(t)
	defer mock.Close()

	userID := newPGUUID(t)

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("inactive@example.com").WillReturnRows(newUserRows(userID, "inactive@example.com", true, false, sqlc.UserRoleTypeAdmin, sqlc.UserKindOwner))
	mock.ExpectExec(`(?s)name: InsertLoginHistory`).WithArgs(userID, pgxmock.AnyArg(), "TestAgent/1.0", sqlc.LoginResultAccountInactive, pgtype.Text{String: "user inactive", Valid: true}).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	_, err := serviceUnderTest.Login(context.Background(), "inactive@example.com", "secret", "127.0.0.1", "TestAgent/1.0")
	require.ErrorIs(t, err, apperror.ErrAccountInactive)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_EmailUnverified(t *testing.T) {
	serviceUnderTest, mock := newAuthService(t)
	defer mock.Close()

	userID := newPGUUID(t)

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("unverified@example.com").WillReturnRows(newUserRows(userID, "unverified@example.com", false, true, sqlc.UserRoleTypeAdmin, sqlc.UserKindOwner))
	mock.ExpectExec(`(?s)name: InsertLoginHistory`).WithArgs(userID, pgxmock.AnyArg(), "TestAgent/1.0", sqlc.LoginResultEmailUnverified, pgtype.Text{String: "email not verified", Valid: true}).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	_, err := serviceUnderTest.Login(context.Background(), "unverified@example.com", "secret", "127.0.0.1", "TestAgent/1.0")
	require.ErrorIs(t, err, apperror.ErrEmailNotVerified)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_MissingCompanyMembership(t *testing.T) {
	serviceUnderTest, mock := newAuthService(t)
	defer mock.Close()

	userID := newPGUUID(t)
	hash := validPasswordHash(t, "super-secret")

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("owner@example.com").WillReturnRows(newUserRows(userID, "owner@example.com", true, true, sqlc.UserRoleTypeAdmin, sqlc.UserKindOwner))
	mock.ExpectQuery(`(?s)name: GetUserAuthByUserID`).WithArgs(userID).WillReturnRows(newUserAuthRows(userID, hash, 0, false))
	mock.ExpectExec(`(?s)name: ResetUserAuthLoginAttempts`).WithArgs(userID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery(`(?s)name: GetActiveCompanyUserByUserID`).WithArgs(userID).WillReturnError(pgx.ErrNoRows)

	_, err := serviceUnderTest.Login(context.Background(), "owner@example.com", "super-secret", "127.0.0.1", "TestAgent/1.0")
	require.ErrorIs(t, err, apperror.ErrForbidden)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_InvalidPayload(t *testing.T) {
	serviceUnderTest, mock := newAuthService(t)
	defer mock.Close()

	_, err := serviceUnderTest.Login(context.Background(), "  ", "", "127.0.0.1", "TestAgent/1.0")
	require.ErrorIs(t, err, apperror.ErrUnprocessableEntity)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_HasModuleAccess(t *testing.T) {
	serviceUnderTest, mock := newAuthService(t)
	defer mock.Close()

	companyID := newPGUUID(t)
	mock.ExpectQuery(`(?s)name: HasActiveCompanyModuleByCode`).WithArgs(companyID, "SCH").WillReturnRows(pgxmock.NewRows([]string{"has_access"}).AddRow(true))

	allowed, err := serviceUnderTest.HasModuleAccess(context.Background(), companyID, "SCH")
	require.NoError(t, err)
	require.True(t, allowed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthService_Login_AccountLocked(t *testing.T) {
	serviceUnderTest, mock := newAuthService(t)
	defer mock.Close()

	userID := newPGUUID(t)
	hash := validPasswordHash(t, "super-secret")

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("locked@example.com").WillReturnRows(newUserRows(userID, "locked@example.com", true, true, sqlc.UserRoleTypeAdmin, sqlc.UserKindOwner))
	mock.ExpectQuery(`(?s)name: GetUserAuthByUserID`).WithArgs(userID).WillReturnRows(newUserAuthRows(userID, hash, 0, true))
	mock.ExpectExec(`(?s)name: InsertLoginHistory`).WithArgs(userID, pgxmock.AnyArg(), "TestAgent/1.0", sqlc.LoginResultAccountLocked, pgtype.Text{String: "account locked", Valid: true}).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	_, err := serviceUnderTest.Login(context.Background(), "locked@example.com", "super-secret", "127.0.0.1", "TestAgent/1.0")
	require.ErrorIs(t, err, apperror.ErrAccountLocked)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestParseClientIP(t *testing.T) {
	t.Parallel()

	require.Equal(t, netip.MustParseAddr("127.0.0.1"), parseClientIP("not-an-ip"))
	require.Equal(t, netip.MustParseAddr("192.0.2.10"), parseClientIP("192.0.2.10"))
}
