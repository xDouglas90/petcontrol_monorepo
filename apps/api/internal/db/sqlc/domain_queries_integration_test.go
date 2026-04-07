package sqlc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/test/integration"
)

func setupQueriesWithPool(t *testing.T) (*sqlc.Queries, context.Context, *pgxpool.Pool) {
	t.Helper()

	setup := integration.SetupPostgresWithMigrations(t)
	return sqlc.New(setup.Pool), setup.Ctx, setup.Pool
}

func mustNumeric(t *testing.T, value string) pgtype.Numeric {
	t.Helper()
	var n pgtype.Numeric
	require.NoError(t, n.Scan(value))
	return n
}

func mustCreateResponsiblePerson(t *testing.T, pool interface {
	QueryRow(context.Context, string, ...interface{}) pgx.Row
},
) pgtype.UUID {
	t.Helper()

	var id pgtype.UUID
	err := pool.QueryRow(context.Background(),
		"INSERT INTO people(kind, is_active, has_system_user) VALUES ('responsible', TRUE, FALSE) RETURNING id",
	).Scan(&id)
	require.NoError(t, err)
	require.True(t, id.Valid)
	return id
}

func mustCreateCompany(t *testing.T, queries *sqlc.Queries, pool interface {
	QueryRow(context.Context, string, ...interface{}) pgx.Row
},
) sqlc.Company {
	t.Helper()

	company, err := queries.InsertCompany(context.Background(), sqlc.InsertCompanyParams{
		Slug:           fmt.Sprintf("company-%d", time.Now().UnixNano()),
		Name:           "PetControl Co",
		FantasyName:    "PetControl",
		Cnpj:           fmt.Sprintf("%014d", time.Now().UnixNano()%100000000000000),
		FoundationDate: pgtype.Date{},
		LogoUrl:        pgtype.Text{},
		ResponsibleID:  mustCreateResponsiblePerson(t, pool),
	})
	require.NoError(t, err)
	return company
}

func TestQueries_Modules_Integration(t *testing.T) {
	queries, _, _ := setupQueriesWithPool(t)

	created, err := queries.CreateModule(context.Background(), sqlc.CreateModuleParams{
		Code:        fmt.Sprintf("M%07d", time.Now().UnixNano()%10000000),
		Name:        "Module Test",
		Description: "module integration test",
		MinPackage:  sqlc.ModulePackageStarter,
		IsActive:    true,
	})
	require.NoError(t, err)

	gotByCode, err := queries.GetModuleByCode(context.Background(), created.Code)
	require.NoError(t, err)
	require.Equal(t, created.ID, gotByCode.ID)

	listed, err := queries.ListModules(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, listed)

	updated, err := queries.UpdateModule(context.Background(), sqlc.UpdateModuleParams{
		Code:        pgtype.Text{String: created.Code, Valid: true},
		Name:        pgtype.Text{String: "Module Test Updated", Valid: true},
		Description: pgtype.Text{String: "module integration updated", Valid: true},
		MinPackage:  sqlc.NullModulePackage{ModulePackage: sqlc.ModulePackageBasic, Valid: true},
		IsActive:    pgtype.Bool{Bool: true, Valid: true},
		ID:          created.ID,
	})
	require.NoError(t, err)
	require.Equal(t, "Module Test Updated", updated.Name)
	require.Equal(t, sqlc.ModulePackageBasic, updated.MinPackage)

	affected, err := queries.DeleteModule(context.Background(), created.ID)
	require.NoError(t, err)
	require.EqualValues(t, 1, affected)

	_, err = queries.GetModuleByCode(context.Background(), created.Code)
	require.Error(t, err)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}

func TestQueries_Plans_Integration(t *testing.T) {
	queries, _, _ := setupQueriesWithPool(t)

	planType, err := queries.InsertPlanType(context.Background(), sqlc.InsertPlanTypeParams{
		Name:        fmt.Sprintf("Type-%d", time.Now().UnixNano()),
		Description: pgtype.Text{String: "plan type integration", Valid: true},
	})
	require.NoError(t, err)

	plan, err := queries.InsertPlan(context.Background(), sqlc.InsertPlanParams{
		PlanTypeID:       planType.ID,
		Name:             fmt.Sprintf("Plan-%d", time.Now().UnixNano()),
		Description:      "plan integration",
		Package:          sqlc.ModulePackageStarter,
		Price:            mustNumeric(t, "129.90"),
		BillingCycleDays: 30,
		MaxUsers:         pgtype.Int4{Int32: 10, Valid: true},
		IsActive:         true,
		ImageUrl:         pgtype.Text{},
	})
	require.NoError(t, err)

	gotPlan, err := queries.GetPlanByID(context.Background(), plan.ID)
	require.NoError(t, err)
	require.Equal(t, plan.ID, gotPlan.ID)

	planTypes, err := queries.ListPlanTypes(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, planTypes)

	plans, err := queries.ListPlans(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, plans)

	rows, err := queries.UpdatePlan(context.Background(), sqlc.UpdatePlanParams{
		PlanTypeID:       planType.ID,
		Name:             pgtype.Text{String: "Plan Updated", Valid: true},
		Description:      pgtype.Text{String: "plan integration updated", Valid: true},
		Package:          sqlc.NullModulePackage{ModulePackage: sqlc.ModulePackageBasic, Valid: true},
		Price:            mustNumeric(t, "199.90"),
		BillingCycleDays: pgtype.Int4{Int32: 60, Valid: true},
		MaxUsers:         pgtype.Int4{Int32: 20, Valid: true},
		IsActive:         pgtype.Bool{Bool: true, Valid: true},
		ImageUrl:         pgtype.Text{String: "https://example.com/plan.png", Valid: true},
		ID:               plan.ID,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	updatedPlan, err := queries.GetPlanByID(context.Background(), plan.ID)
	require.NoError(t, err)
	require.Equal(t, "Plan Updated", updatedPlan.Name)
	require.Equal(t, sqlc.ModulePackageBasic, updatedPlan.Package)
}

func TestQueries_Companies_Integration(t *testing.T) {
	queries, _, pool := setupQueriesWithPool(t)

	created := mustCreateCompany(t, queries, pool)

	gotByID, err := queries.GetCompanyByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, gotByID.ID)

	gotBySlug, err := queries.GetCompanyBySlug(context.Background(), created.Slug)
	require.NoError(t, err)
	require.Equal(t, created.ID, gotBySlug.ID)

	list, err := queries.ListCompanies(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, list)

	byPackage, err := queries.ListCompaniesByPackage(context.Background(), created.ActivePackage)
	require.NoError(t, err)
	require.NotEmpty(t, byPackage)

	newResponsible := mustCreateResponsiblePerson(t, pool)
	rows, err := queries.UpdateCompany(context.Background(), sqlc.UpdateCompanyParams{
		Slug:           pgtype.Text{String: created.Slug, Valid: true},
		Name:           pgtype.Text{String: "PetControl Updated", Valid: true},
		FantasyName:    pgtype.Text{String: "PetControl Upd", Valid: true},
		CNPJ:           pgtype.Text{String: created.Cnpj, Valid: true},
		FoundationDate: pgtype.Date{},
		LogoURL:        pgtype.Text{},
		ResponsibleID:  newResponsible,
		ActivePackage:  sqlc.NullModulePackage{ModulePackage: sqlc.ModulePackageBasic, Valid: true},
		IsActive:       pgtype.Bool{Bool: true, Valid: true},
		ID:             created.ID,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	affected, err := queries.DeleteCompany(context.Background(), created.ID)
	require.NoError(t, err)
	require.EqualValues(t, 1, affected)

	_, err = queries.GetCompanyByID(context.Background(), created.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}

func TestQueries_CompanyUsers_Integration(t *testing.T) {
	queries, _, pool := setupQueriesWithPool(t)

	company := mustCreateCompany(t, queries, pool)
	user := insertDefaultUser(t, queries, uniqueEmail("company-user"))

	created, err := queries.CreateCompanyUser(context.Background(), sqlc.CreateCompanyUserParams{
		CompanyID: company.ID,
		UserID:    user.ID,
		IsOwner:   true,
		IsActive:  true,
	})
	require.NoError(t, err)

	gotByID, err := queries.GetCompanyUserByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, gotByID.ID)

	gotByPair, err := queries.GetCompanyUser(context.Background(), sqlc.GetCompanyUserParams{
		CompanyID: company.ID,
		UserID:    user.ID,
	})
	require.NoError(t, err)
	require.Equal(t, created.ID, gotByPair.ID)

	list, err := queries.ListCompanyUsersByCompanyID(context.Background(), company.ID)
	require.NoError(t, err)
	require.NotEmpty(t, list)

	err = queries.DeactivateCompanyUser(context.Background(), sqlc.DeactivateCompanyUserParams{
		CompanyID: company.ID,
		UserID:    user.ID,
	})
	require.NoError(t, err)

	activeList, err := queries.ListCompanyUsersByCompanyID(context.Background(), company.ID)
	require.NoError(t, err)
	require.Len(t, activeList, 0)
}
