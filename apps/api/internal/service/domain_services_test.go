package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func newDomainUUID(t *testing.T) pgtype.UUID {
	t.Helper()

	value := uuid.New()
	var out pgtype.UUID
	copy(out.Bytes[:], value[:])
	out.Valid = true
	return out
}

func TestCompanyService_GetCurrentCompany(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewCompanyService(queries)

	companyID := newDomainUUID(t)
	mock.ExpectQuery(`(?s)name: GetCompanyByID`).WithArgs(companyID).WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).AddRow(companyID.String(), "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, newDomainUUID(t).String(), sqlc.ModulePackageStarter, true, time.Now(), nil, nil))

	company, err := serviceUnderTest.GetCurrentCompany(context.Background(), companyID)
	require.NoError(t, err)
	require.Equal(t, companyID, company.ID)
	require.Equal(t, sqlc.ModulePackageStarter, company.ActivePackage)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPlanService_GetCurrentPlan(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewPlanService(queries)

	companyID := newDomainUUID(t)
	planTypeID := newDomainUUID(t)
	mock.ExpectQuery(`(?s)name: GetCurrentPlanByCompanyID`).WithArgs(companyID).WillReturnRows(pgxmock.NewRows([]string{"id", "plan_type_id", "name", "description", "package", "price", "billing_cycle_days", "max_users", "is_active", "image_url", "created_at", "updated_at", "deleted_at"}).AddRow(newDomainUUID(t).String(), planTypeID.String(), "Starter", "starter plan", sqlc.ModulePackageStarter, "99.90", int32(30), pgtype.Int4{Int32: 5, Valid: true}, true, nil, time.Now(), nil, nil))

	plan, err := serviceUnderTest.GetCurrentPlan(context.Background(), companyID)
	require.NoError(t, err)
	require.Equal(t, "Starter", plan.Name)
	require.Equal(t, sqlc.ModulePackageStarter, plan.Package)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModuleService_ListActiveModulesByCompanyID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewModuleService(queries)

	companyID := newDomainUUID(t)
	mock.ExpectQuery(`(?s)name: ListActiveModulesByCompanyID`).WithArgs(companyID).WillReturnRows(pgxmock.NewRows([]string{"id", "code", "name", "description", "min_package", "is_active", "created_at", "updated_at", "deleted_at"}).AddRow(newDomainUUID(t).String(), "SCH", "Scheduling", "Scheduling", sqlc.ModulePackageStarter, true, time.Now(), nil, nil))

	modules, err := serviceUnderTest.ListActiveModulesByCompanyID(context.Background(), companyID)
	require.NoError(t, err)
	require.Len(t, modules, 1)
	require.Equal(t, "SCH", modules[0].Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyUserService_CreateAndDeactivate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewCompanyUserService(queries)

	companyID := newDomainUUID(t)
	userID := newDomainUUID(t)
	createdID := newDomainUUID(t)

	mock.ExpectQuery(`(?s)name: CreateCompanyUser`).WithArgs(companyID, userID, true, true).WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "is_owner", "is_active", "joined_at", "left_at"}).AddRow(createdID.String(), companyID.String(), userID.String(), true, true, time.Now(), nil))
	created, err := serviceUnderTest.CreateCompanyUser(context.Background(), sqlc.CreateCompanyUserParams{CompanyID: companyID, UserID: userID, IsOwner: true, IsActive: true})
	require.NoError(t, err)
	require.Equal(t, createdID, created.ID)

	mock.ExpectExec(`(?s)name: DeactivateCompanyUser`).WithArgs(companyID, userID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	require.NoError(t, serviceUnderTest.DeactivateCompanyUser(context.Background(), companyID, userID))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyService_DeleteCompany_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewCompanyService(queries)

	companyID := newDomainUUID(t)
	mock.ExpectExec(`(?s)name: DeleteCompany`).WithArgs(companyID).WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	err = serviceUnderTest.DeleteCompany(context.Background(), companyID)
	require.ErrorIs(t, err, apperror.ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}
