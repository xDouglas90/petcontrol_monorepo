package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
	"golang.org/x/crypto/bcrypt"
)

func authServiceWithMock(t *testing.T) (*service.AuthService, pgxmock.PgxPoolIface) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	return service.NewAuthService(sqlc.New(mock), "secret", time.Hour), mock
}

func newHandlerUUID(t *testing.T) pgtype.UUID {
	t.Helper()

	value := uuid.New()
	var out pgtype.UUID
	copy(out.Bytes[:], value[:])
	out.Valid = true
	return out
}

func handlerUserRows(id pgtype.UUID, email string, verified bool, isActive bool) *pgxmock.Rows {
	var verifiedAt any = nil
	if verified {
		verifiedAt = time.Now().Add(-time.Minute)
	}

	return pgxmock.NewRows([]string{"id", "email", "email_verified", "email_verified_at", "role", "kind", "is_active", "created_at", "updated_at", "deleted_at"}).AddRow(id.String(), email, verified, verifiedAt, sqlc.UserRoleTypeAdmin, sqlc.UserKindOwner, isActive, time.Now().Add(-time.Hour), nil, nil)
}

func handlerUserAuthRows(userID pgtype.UUID, passwordHash string) *pgxmock.Rows {
	return pgxmock.NewRows([]string{"user_id", "password_hash", "password_changed_at", "must_change_password", "login_attempts", "locked_until", "last_login_at", "created_at", "updated_at"}).AddRow(userID.String(), passwordHash, nil, false, int16(0), nil, nil, time.Now().Add(-time.Hour), nil)
}

func handlerValidHash(t *testing.T, password string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hash)
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceUnderTest, mock := authServiceWithMock(t)
	defer mock.Close()

	h := NewAuthHandler(serviceUnderTest)
	router := gin.New()
	router.POST("/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnprocessableEntity, res.Code)
	require.Contains(t, res.Body.String(), "invalid request body")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceUnderTest, mock := authServiceWithMock(t)
	defer mock.Close()

	userID := newHandlerUUID(t)
	companyID := newHandlerUUID(t)
	password := "integration-secret"
	hash := handlerValidHash(t, password)

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("owner@example.com").WillReturnRows(handlerUserRows(userID, "owner@example.com", true, true))
	mock.ExpectQuery(`(?s)name: GetUserAuthByUserID`).WithArgs(userID).WillReturnRows(handlerUserAuthRows(userID, hash))
	mock.ExpectExec(`(?s)name: ResetUserAuthLoginAttempts`).WithArgs(userID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery(`(?s)name: GetActiveCompanyUserByUserID`).WithArgs(userID).WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "is_owner", "is_active", "joined_at", "left_at"}).AddRow(uuid.NewString(), companyID.String(), userID.String(), true, true, time.Now(), nil))
	mock.ExpectExec(`(?s)name: InsertLoginHistory`).WithArgs(userID, pgxmock.AnyArg(), "HandlerTest/1.0", sqlc.LoginResultSuccess, pgtype.Text{}).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	h := NewAuthHandler(serviceUnderTest)
	router := gin.New()
	router.POST("/login", h.Login)

	body := map[string]string{"email": " owner@example.com ", "password": password}
	payload, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HandlerTest/1.0")
	req.RemoteAddr = "127.0.0.1:1234"
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "access_token")
	require.Contains(t, res.Body.String(), companyID.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceUnderTest, mock := authServiceWithMock(t)
	defer mock.Close()

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("missing@example.com").WillReturnError(pgx.ErrNoRows)
	mock.ExpectExec(`(?s)name: InsertLoginHistory`).WithArgs(pgtype.UUID{}, pgxmock.AnyArg(), "HandlerTest/1.0", sqlc.LoginResultInvalidCredentials, pgtype.Text{String: "user not found", Valid: true}).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	h := NewAuthHandler(serviceUnderTest)
	router := gin.New()
	router.POST("/login", h.Login)

	body := map[string]string{"email": "missing@example.com", "password": "secret"}
	payload, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HandlerTest/1.0")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	require.Contains(t, res.Body.String(), "invalid credentials")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthHandler_Login_MissingCompanyMembership(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceUnderTest, mock := authServiceWithMock(t)
	defer mock.Close()

	userID := newHandlerUUID(t)
	password := "integration-secret"
	hash := handlerValidHash(t, password)

	mock.ExpectQuery(`(?s)name: GetUserByEmail`).WithArgs("owner@example.com").WillReturnRows(handlerUserRows(userID, "owner@example.com", true, true))
	mock.ExpectQuery(`(?s)name: GetUserAuthByUserID`).WithArgs(userID).WillReturnRows(handlerUserAuthRows(userID, hash))
	mock.ExpectExec(`(?s)name: ResetUserAuthLoginAttempts`).WithArgs(userID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectQuery(`(?s)name: GetActiveCompanyUserByUserID`).WithArgs(userID).WillReturnError(pgx.ErrNoRows)

	h := NewAuthHandler(serviceUnderTest)
	router := gin.New()
	router.POST("/login", h.Login)

	body := map[string]string{"email": "owner@example.com", "password": password}
	payload, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HandlerTest/1.0")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.Contains(t, res.Body.String(), "no active company membership")
	require.NoError(t, mock.ExpectationsWereMet())
}
