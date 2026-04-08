package sqlc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func uuidValue() pgtype.UUID {
	return pgtype.UUID{Valid: true}
}

func TestQueries_Modules_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	errExpected := errors.New("db error")

	mock.ExpectQuery(`(?s)name: CreateModule`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.CreateModule(context.Background(), sqlc.CreateModuleParams{})
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: GetModuleByCode`).WithArgs(pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.GetModuleByCode(context.Background(), "SCH")
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: ListModules`).WillReturnError(errExpected)
	_, err = queries.ListModules(context.Background())
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: ListActiveModulesByCompanyID`).WithArgs(pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.ListActiveModulesByCompanyID(context.Background(), uuidValue())
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: UpdateModule`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.UpdateModule(context.Background(), sqlc.UpdateModuleParams{})
	require.ErrorIs(t, err, errExpected)

	mock.ExpectExec(`(?s)name: DeleteModule`).WithArgs(pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	rows, err := queries.DeleteModule(context.Background(), uuidValue())
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueries_Plans_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	errExpected := errors.New("db error")

	mock.ExpectQuery(`(?s)name: InsertPlanType`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.InsertPlanType(context.Background(), sqlc.InsertPlanTypeParams{})
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: ListPlanTypes`).WillReturnError(errExpected)
	_, err = queries.ListPlanTypes(context.Background())
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: InsertPlan`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.InsertPlan(context.Background(), sqlc.InsertPlanParams{})
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: GetPlanByID`).WithArgs(pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.GetPlanByID(context.Background(), uuidValue())
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: GetCurrentPlanByCompanyID`).WithArgs(pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.GetCurrentPlanByCompanyID(context.Background(), uuidValue())
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: ListPlans`).WillReturnError(errExpected)
	_, err = queries.ListPlans(context.Background())
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: ListPlansByPackage`).WithArgs(pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.ListPlansByPackage(context.Background(), sqlc.ModulePackageStarter)
	require.ErrorIs(t, err, errExpected)

	mock.ExpectExec(`(?s)name: UpdatePlan`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	rows, err := queries.UpdatePlan(context.Background(), sqlc.UpdatePlanParams{})
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueries_Companies_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	errExpected := errors.New("db error")

	mock.ExpectQuery(`(?s)name: InsertCompany`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.InsertCompany(context.Background(), sqlc.InsertCompanyParams{})
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: GetCompanyByID`).WithArgs(pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.GetCompanyByID(context.Background(), uuidValue())
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: GetCompanyBySlug`).WithArgs(pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.GetCompanyBySlug(context.Background(), "slug")
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: ListCompanies`).WillReturnError(errExpected)
	_, err = queries.ListCompanies(context.Background())
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: ListCompaniesByPackage`).WithArgs(pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.ListCompaniesByPackage(context.Background(), sqlc.ModulePackageStarter)
	require.ErrorIs(t, err, errExpected)

	mock.ExpectExec(`(?s)name: UpdateCompany`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	rows, err := queries.UpdateCompany(context.Background(), sqlc.UpdateCompanyParams{})
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	mock.ExpectExec(`(?s)name: DeleteCompany`).WithArgs(pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	rows, err = queries.DeleteCompany(context.Background(), uuidValue())
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueries_CompanyUsers_Unit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	errExpected := errors.New("db error")

	mock.ExpectQuery(`(?s)name: CreateCompanyUser`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.CreateCompanyUser(context.Background(), sqlc.CreateCompanyUserParams{})
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: GetCompanyUserByID`).WithArgs(pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.GetCompanyUserByID(context.Background(), uuidValue())
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: GetCompanyUser`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.GetCompanyUser(context.Background(), sqlc.GetCompanyUserParams{})
	require.ErrorIs(t, err, errExpected)

	mock.ExpectQuery(`(?s)name: ListCompanyUsersByCompanyID`).WithArgs(pgxmock.AnyArg()).WillReturnError(errExpected)
	_, err = queries.ListCompanyUsersByCompanyID(context.Background(), uuidValue())
	require.ErrorIs(t, err, errExpected)

	mock.ExpectExec(`(?s)name: DeactivateCompanyUser`).WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	err = queries.DeactivateCompanyUser(context.Background(), sqlc.DeactivateCompanyUserParams{})
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
