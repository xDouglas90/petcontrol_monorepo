package sqlc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@test.com", prefix, time.Now().UnixNano())
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

func mustAttachModuleToCompany(t *testing.T, pool interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}, companyID, moduleID pgtype.UUID,
) {
	t.Helper()

	_, err := pool.Exec(context.Background(),
		"INSERT INTO company_modules(company_id, module_id, is_active) VALUES ($1, $2, TRUE)",
		companyID,
		moduleID,
	)
	require.NoError(t, err)
}

func mustAttachPermissionToModule(t *testing.T, pool interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}, moduleID, permissionID pgtype.UUID,
) {
	t.Helper()

	_, err := pool.Exec(context.Background(),
		"INSERT INTO module_permissions(module_id, permission_id) VALUES ($1, $2)",
		moduleID,
		permissionID,
	)
	require.NoError(t, err)
}

func mustCreatePermission(t *testing.T, queries *sqlc.Queries, code, description string, defaultRoles []sqlc.UserRoleType) sqlc.Permission {
	t.Helper()

	rows, err := queries.InsertPermission(context.Background(), sqlc.InsertPermissionParams{
		Code:         code,
		Description:  pgtype.Text{String: description, Valid: true},
		DefaultRoles: defaultRoles,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	permission, err := queries.GetPermissionByCode(context.Background(), code)
	require.NoError(t, err)
	return permission
}

func mustCreateActiveSubscription(t *testing.T, pool interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}, companyID, planID pgtype.UUID, price string,
) {
	t.Helper()

	_, err := pool.Exec(context.Background(),
		"INSERT INTO company_subscriptions(company_id, plan_id, started_at, expires_at, is_active, price_paid) VALUES ($1, $2, now() - interval '1 day', now() + interval '30 days', TRUE, $3)",
		companyID,
		planID,
		price,
	)
	require.NoError(t, err)
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
	queries, _, pool := setupQueriesWithPool(t)

	planType, err := queries.InsertPlanType(context.Background(), sqlc.InsertPlanTypeParams{
		Name:        fmt.Sprintf("Type-%d", time.Now().UnixNano()),
		Description: pgtype.Text{String: "plan type integration", Valid: true},
	})
	require.NoError(t, err)

	secondType, err := queries.InsertPlanType(context.Background(), sqlc.InsertPlanTypeParams{
		Name:        fmt.Sprintf("TypeB-%d", time.Now().UnixNano()),
		Description: pgtype.Text{String: "plan type integration b", Valid: true},
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

	plansByPackage, err := queries.ListPlansByPackage(context.Background(), sqlc.ModulePackageStarter)
	require.NoError(t, err)
	require.NotEmpty(t, plansByPackage)

	company := mustCreateCompany(t, queries, pool)
	mustCreateActiveSubscription(t, pool, company.ID, plan.ID, "129.90")

	currentPlan, err := queries.GetCurrentPlanByCompanyID(context.Background(), company.ID)
	require.NoError(t, err)
	require.Equal(t, plan.ID, currentPlan.ID)

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
	_ = secondType
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

func TestQueries_Modules_Integration_ActiveByCompany(t *testing.T) {
	queries, _, pool := setupQueriesWithPool(t)

	company := mustCreateCompany(t, queries, pool)
	module, err := queries.CreateModule(context.Background(), sqlc.CreateModuleParams{
		Code:        "SCH001",
		Name:        "Scheduling",
		Description: "active module integration",
		MinPackage:  sqlc.ModulePackageStarter,
		IsActive:    true,
	})
	require.NoError(t, err)

	mustAttachModuleToCompany(t, pool, company.ID, module.ID)

	activeModules, err := queries.ListActiveModulesByCompanyID(context.Background(), company.ID)
	require.NoError(t, err)
	require.Len(t, activeModules, 1)
	require.Equal(t, module.ID, activeModules[0].ID)
}

func TestQueries_TenantSettingsCatalog_IntegrationFiltersInternalMissingAndPremiumModules(t *testing.T) {
	queries, _, pool := setupQueriesWithPool(t)

	company := mustCreateCompany(t, queries, pool)

	rows, err := queries.UpdateCompany(context.Background(), sqlc.UpdateCompanyParams{
		Slug:           pgtype.Text{String: company.Slug, Valid: true},
		Name:           pgtype.Text{String: company.Name, Valid: true},
		FantasyName:    pgtype.Text{String: company.FantasyName, Valid: true},
		CNPJ:           pgtype.Text{String: company.Cnpj, Valid: true},
		FoundationDate: pgtype.Date{},
		LogoURL:        pgtype.Text{},
		ResponsibleID:  company.ResponsibleID,
		ActivePackage:  sqlc.NullModulePackage{ModulePackage: sqlc.ModulePackageStarter, Valid: true},
		IsActive:       pgtype.Bool{Bool: true, Valid: true},
		ID:             company.ID,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	cfgModule, err := queries.CreateModule(context.Background(), sqlc.CreateModuleParams{
		Code:        fmt.Sprintf("CFG%06d", time.Now().UnixNano()%1000000),
		Name:        "Configurações",
		Description: "Configurações do tenant",
		MinPackage:  sqlc.ModulePackageStarter,
		IsActive:    true,
	})
	require.NoError(t, err)
	premiumModule, err := queries.CreateModule(context.Background(), sqlc.CreateModuleParams{
		Code:        fmt.Sprintf("FIN%06d", time.Now().UnixNano()%1000000),
		Name:        "Financeiro",
		Description: "Financeiro premium",
		MinPackage:  sqlc.ModulePackagePremium,
		IsActive:    true,
	})
	require.NoError(t, err)
	internalModule, err := queries.CreateModule(context.Background(), sqlc.CreateModuleParams{
		Code:        fmt.Sprintf("AUD%06d", time.Now().UnixNano()%1000000),
		Name:        "Logs",
		Description: "Logs internos",
		MinPackage:  sqlc.ModulePackageInternal,
		IsActive:    true,
	})
	require.NoError(t, err)
	orphanModule, err := queries.CreateModule(context.Background(), sqlc.CreateModuleParams{
		Code:        fmt.Sprintf("EMP%06d", time.Now().UnixNano()%1000000),
		Name:        "Sem Permissoes",
		Description: "Modulo sem permissoes vinculadas",
		MinPackage:  sqlc.ModulePackageStarter,
		IsActive:    true,
	})
	require.NoError(t, err)

	cfgPermission := mustCreatePermission(t, queries, fmt.Sprintf("company_settings:test:%d", time.Now().UnixNano()), "Editar configurações de negócios", []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin})
	premiumPermission := mustCreatePermission(t, queries, fmt.Sprintf("finances:test:%d", time.Now().UnixNano()), "Visualizar financeiro", []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin})
	internalPermission := mustCreatePermission(t, queries, fmt.Sprintf("logs:test:%d", time.Now().UnixNano()), "Visualizar logs", []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin})
	unlinkedPermission := mustCreatePermission(t, queries, fmt.Sprintf("orphan:test:%d", time.Now().UnixNano()), "Permissão sem módulo", []sqlc.UserRoleType{sqlc.UserRoleTypeAdmin})

	mustAttachPermissionToModule(t, pool, cfgModule.ID, cfgPermission.ID)
	mustAttachPermissionToModule(t, pool, premiumModule.ID, premiumPermission.ID)
	mustAttachPermissionToModule(t, pool, internalModule.ID, internalPermission.ID)
	_ = unlinkedPermission

	mustAttachModuleToCompany(t, pool, company.ID, cfgModule.ID)
	mustAttachModuleToCompany(t, pool, company.ID, premiumModule.ID)
	mustAttachModuleToCompany(t, pool, company.ID, internalModule.ID)
	mustAttachModuleToCompany(t, pool, company.ID, orphanModule.ID)

	modules, err := queries.ListTenantSettingsModulesByCompanyID(context.Background(), company.ID)
	require.NoError(t, err)
	require.Len(t, modules, 1)
	require.Equal(t, cfgModule.ID, modules[0].ID)

	permissions, err := queries.ListTenantSettingsPermissionsByCompanyID(context.Background(), company.ID)
	require.NoError(t, err)
	require.Len(t, permissions, 1)
	require.Equal(t, cfgModule.ID, permissions[0].ModuleID)
	require.Equal(t, cfgPermission.ID, permissions[0].ID)
	require.NotEqual(t, premiumPermission.ID, permissions[0].ID)
	require.NotEqual(t, internalPermission.ID, permissions[0].ID)
}

func mustCreateServiceType(t *testing.T, queries *sqlc.Queries) sqlc.ServiceType {
	t.Helper()
	st, err := queries.CreateServiceType(context.Background(), sqlc.CreateServiceTypeParams{
		Name:        fmt.Sprintf("Type-%d", time.Now().UnixNano()),
		Description: pgtype.Text{String: "service type test", Valid: true},
	})
	require.NoError(t, err)
	return st
}

func mustCreateService(t *testing.T, queries *sqlc.Queries, typeID pgtype.UUID) sqlc.Service {
	t.Helper()
	s, err := queries.CreateService(context.Background(), sqlc.CreateServiceParams{
		TypeID:       typeID,
		Title:        fmt.Sprintf("Service-%d", time.Now().UnixNano()),
		Description:  "service test",
		Price:        mustNumeric(t, "50.00"),
		DiscountRate: mustNumeric(t, "0.00"),
		IsActive:     true,
	})
	require.NoError(t, err)
	return s
}

func mustCreateSubService(t *testing.T, queries *sqlc.Queries, serviceID, typeID pgtype.UUID) sqlc.SubService {
	t.Helper()
	ss, err := queries.InsertSubService(context.Background(), sqlc.InsertSubServiceParams{
		ServiceID:    serviceID,
		TypeID:       typeID,
		Title:        fmt.Sprintf("SubService-%d", time.Now().UnixNano()),
		Description:  "sub service test",
		Price:        mustNumeric(t, "25.00"),
		DiscountRate: mustNumeric(t, "0.00"),
		IsActive:     pgtype.Bool{Bool: true, Valid: true},
	})
	require.NoError(t, err)
	return ss
}

func mustCreatePlanType(t *testing.T, queries *sqlc.Queries) sqlc.PlanType {
	t.Helper()
	pt, err := queries.InsertPlanType(context.Background(), sqlc.InsertPlanTypeParams{
		Name:        fmt.Sprintf("PlanType-%d", time.Now().UnixNano()),
		Description: pgtype.Text{String: "plan type test", Valid: true},
	})
	require.NoError(t, err)
	return pt
}

func mustInsertServicePlan(t *testing.T, queries *sqlc.Queries, planTypeID pgtype.UUID) sqlc.ServicePlan {
	t.Helper()
	sp, err := queries.InsertServicePlan(context.Background(), sqlc.InsertServicePlanParams{
		PlanTypeID:   planTypeID,
		Title:        fmt.Sprintf("Plan-%d", time.Now().UnixNano()),
		Description:  "service plan test",
		Price:        mustNumeric(t, "150.00"),
		DiscountRate: mustNumeric(t, "0.00"),
		IsActive:     pgtype.Bool{Bool: true, Valid: true},
	})
	require.NoError(t, err)
	return sp
}

func mustCreateProduct(t *testing.T, queries *sqlc.Queries) sqlc.Product {
	t.Helper()
	p, err := queries.InsertProduct(context.Background(), sqlc.InsertProductParams{
		Name:        fmt.Sprintf("Product-%d", time.Now().UnixNano()),
		Description: pgtype.Text{String: "product test", Valid: true},
		Quantity:    pgtype.Int4{Int32: 100, Valid: true},
	})
	require.NoError(t, err)
	return p
}

func mustCreateCompanyService(t *testing.T, queries *sqlc.Queries, companyID, serviceID pgtype.UUID) sqlc.CompanyService {
	t.Helper()
	cs, err := queries.InsertCompanyService(context.Background(), sqlc.InsertCompanyServiceParams{
		CompanyID: companyID,
		ServiceID: serviceID,
		IsActive:  pgtype.Bool{Bool: true, Valid: true},
	})
	require.NoError(t, err)
	return cs
}

func mustCreateCompanyProduct(t *testing.T, queries *sqlc.Queries, companyID, productID pgtype.UUID) sqlc.CompanyProduct {
	t.Helper()
	cp, err := queries.InsertCompanyProduct(context.Background(), sqlc.InsertCompanyProductParams{
		CompanyID:    companyID,
		ProductID:    productID,
		Kind:         sqlc.ProductKindCustomer,
		CostPerUnit:  mustNumeric(t, "5.00"),
		SalePrice:    mustNumeric(t, "15.00"),
		HasStock:     pgtype.Bool{Bool: true, Valid: true},
		ForSale:      pgtype.Bool{Bool: true, Valid: true},
		ProfitMargin: mustNumeric(t, "150.00"),
	})
	require.NoError(t, err)
	return cp
}

func mustCreateCompanyServicePlan(t *testing.T, queries *sqlc.Queries, companyID, servicePlanID pgtype.UUID) sqlc.CompanyServicePlan {
	t.Helper()
	csp, err := queries.InsertCompanyServicePlan(context.Background(), sqlc.InsertCompanyServicePlanParams{
		CompanyID:     companyID,
		ServicePlanID: servicePlanID,
		IsActive:      pgtype.Bool{Bool: true, Valid: true},
	})
	require.NoError(t, err)
	return csp
}
