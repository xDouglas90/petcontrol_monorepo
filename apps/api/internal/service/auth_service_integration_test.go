package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
	"github.com/xdouglas90/petcontrol_monorepo/test/integration"
	"golang.org/x/crypto/bcrypt"
)

func insertResponsiblePerson(t *testing.T, pool *pgxpool.Pool) pgtype.UUID {
	t.Helper()

	var id pgtype.UUID
	err := pool.QueryRow(context.Background(), "INSERT INTO people(kind, is_active, has_system_user) VALUES ('responsible', TRUE, FALSE) RETURNING id").Scan(&id)
	require.NoError(t, err)
	require.True(t, id.Valid)
	return id
}

func insertUserAuth(t *testing.T, pool *pgxpool.Pool, userID pgtype.UUID, rawPassword string) {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), `INSERT INTO user_auth(user_id, password_hash, must_change_password, login_attempts) VALUES ($1, $2, FALSE, 0)`, userID, string(hash))
	require.NoError(t, err)
}

func insertCompanyModule(t *testing.T, pool *pgxpool.Pool, companyID, moduleID pgtype.UUID) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `INSERT INTO company_modules(company_id, module_id, is_active) VALUES ($1, $2, TRUE)`, companyID, moduleID)
	require.NoError(t, err)
}

func createIntegrationUser(t *testing.T, queries *sqlc.Queries, email string) sqlc.User {
	t.Helper()

	user, err := queries.InsertUser(context.Background(), sqlc.InsertUserParams{
		Email:           email,
		EmailVerified:   true,
		EmailVerifiedAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
		Role:            sqlc.UserRoleTypeAdmin,
		IsActive:        true,
	})
	require.NoError(t, err)
	return user
}

func createIntegrationCompany(t *testing.T, queries *sqlc.Queries, pool *pgxpool.Pool) sqlc.Company {
	t.Helper()

	company, err := queries.InsertCompany(context.Background(), sqlc.InsertCompanyParams{
		Slug:           fmt.Sprintf("company-%d", time.Now().UnixNano()),
		Name:           "PetControl",
		FantasyName:    "PetControl",
		Cnpj:           fmt.Sprintf("%014d", time.Now().UnixNano()%100000000000000),
		FoundationDate: pgtype.Date{},
		LogoUrl:        pgtype.Text{},
		ResponsibleID:  insertResponsiblePerson(t, pool),
	})
	require.NoError(t, err)
	return company
}

func createIntegrationModule(t *testing.T, queries *sqlc.Queries, code string) sqlc.Module {
	t.Helper()

	module, err := queries.CreateModule(context.Background(), sqlc.CreateModuleParams{
		Code:        code,
		Name:        "Scheduling",
		Description: "scheduling module",
		MinPackage:  sqlc.ModulePackageStarter,
		IsActive:    true,
	})
	require.NoError(t, err)
	return module
}

func TestAuthService_Login_Integration(t *testing.T) {
	setup := integration.SetupPostgresWithMigrations(t)
	queries := sqlc.New(setup.Pool)

	user := createIntegrationUser(t, queries, fmt.Sprintf("auth-%d@example.com", time.Now().UnixNano()))
	insertUserAuth(t, setup.Pool, user.ID, "integration-secret")
	company := createIntegrationCompany(t, queries, setup.Pool)
	_, err := queries.CreateCompanyUser(context.Background(), sqlc.CreateCompanyUserParams{
		CompanyID: company.ID,
		UserID:    user.ID,
		Kind:      sqlc.UserKindOwner,
		IsOwner:   true,
		IsActive:  pgtype.Bool{Bool: true, Valid: true},
	})
	require.NoError(t, err)
	module := createIntegrationModule(t, queries, fmt.Sprintf("M%06d", time.Now().UnixNano()%1000000))
	insertCompanyModule(t, setup.Pool, company.ID, module.ID)

	authService := NewAuthService(queries, "integration-secret", time.Hour)
	result, err := authService.Login(context.Background(), user.Email, "integration-secret", "127.0.0.1", "IntegrationTest/1.0")
	require.NoError(t, err)
	require.Equal(t, user.ID.String(), result.UserID)
	require.Equal(t, company.ID.String(), result.CompanyID)
	require.NotEmpty(t, result.AccessToken)

	claims, err := appjwt.ParseToken("integration-secret", result.AccessToken)
	require.NoError(t, err)
	require.Equal(t, result.UserID, claims.UserID)
	require.Equal(t, result.CompanyID, claims.CompanyID)

	allowed, err := authService.HasModuleAccess(context.Background(), company.ID, module.Code)
	require.NoError(t, err)
	require.True(t, allowed)

	var loginCount int
	err = setup.Pool.QueryRow(context.Background(), `SELECT count(*) FROM login_history WHERE user_id = $1 AND result = 'success'`, user.ID).Scan(&loginCount)
	require.NoError(t, err)
	require.Equal(t, 1, loginCount)
}

func TestAuthService_Login_IntegrationMissingCompanyMembership(t *testing.T) {
	setup := integration.SetupPostgresWithMigrations(t)
	queries := sqlc.New(setup.Pool)

	user := createIntegrationUser(t, queries, fmt.Sprintf("auth-missing-%d@example.com", time.Now().UnixNano()))
	insertUserAuth(t, setup.Pool, user.ID, "integration-secret")

	authService := NewAuthService(queries, "integration-secret", time.Hour)
	_, err := authService.Login(context.Background(), user.Email, "integration-secret", "127.0.0.1", "IntegrationTest/1.0")
	require.ErrorIs(t, err, apperror.ErrForbidden)

	var loginCount int
	err = setup.Pool.QueryRow(context.Background(), `SELECT count(*) FROM login_history WHERE user_id = $1`, user.ID).Scan(&loginCount)
	require.NoError(t, err)
	require.Equal(t, 0, loginCount)
}
