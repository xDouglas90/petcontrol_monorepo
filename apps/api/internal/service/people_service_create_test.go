package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/queue"
	"github.com/xdouglas90/petcontrol_monorepo/test/integration"
)

type peoplePublisherStub struct {
	peopleCalled  bool
	peopleCalls   int
	peoplePayload queue.PersonAccessCredentialsPayload
	failTimes     int
}

func (p *peoplePublisherStub) EnqueueDummyNotification(context.Context, queue.DummyNotificationPayload) error {
	return nil
}

func (p *peoplePublisherStub) EnqueueScheduleConfirmation(context.Context, queue.ScheduleConfirmationPayload) error {
	return nil
}

func (p *peoplePublisherStub) EnqueuePersonAccessCredentials(_ context.Context, payload queue.PersonAccessCredentialsPayload) error {
	p.peopleCalled = true
	p.peopleCalls++
	p.peoplePayload = payload
	if p.failTimes > 0 {
		p.failTimes--
		return context.DeadlineExceeded
	}
	return nil
}

func (p *peoplePublisherStub) Close() error { return nil }

func createIntegrationUserWithRole(t *testing.T, queries *sqlc.Queries, email string, role sqlc.UserRoleType) sqlc.User {
	t.Helper()

	user, err := queries.InsertUser(context.Background(), sqlc.InsertUserParams{
		Email:           email,
		EmailVerified:   true,
		EmailVerifiedAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
		Role:            role,
		IsActive:        true,
	})
	require.NoError(t, err)
	return user
}

func attachCompanyUser(t *testing.T, queries *sqlc.Queries, companyID, userID pgtype.UUID, kind sqlc.UserKind, isOwner bool) {
	t.Helper()

	_, err := queries.CreateCompanyUser(context.Background(), sqlc.CreateCompanyUserParams{
		CompanyID: companyID,
		UserID:    userID,
		Kind:      kind,
		IsOwner:   isOwner,
		IsActive:  pgtype.Bool{Bool: true, Valid: true},
	})
	require.NoError(t, err)
}

func mustCreatePermissionForRole(t *testing.T, queries *sqlc.Queries, code string, role sqlc.UserRoleType) sqlc.Permission {
	t.Helper()

	rows, err := queries.InsertPermission(context.Background(), sqlc.InsertPermissionParams{
		Code:         code,
		Description:  pgtype.Text{String: "test permission", Valid: true},
		DefaultRoles: []sqlc.UserRoleType{role},
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	permission, err := queries.GetPermissionByCode(context.Background(), code)
	require.NoError(t, err)
	return permission
}

func attachPermissionToModule(t *testing.T, pool interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
}, moduleID, permissionID pgtype.UUID) {
	t.Helper()

	_, err := pool.Exec(context.Background(), "INSERT INTO module_permissions(module_id, permission_id) VALUES ($1, $2)", moduleID, permissionID)
	require.NoError(t, err)
}

func parsePeopleNumeric(t *testing.T, raw string) pgtype.Numeric {
	t.Helper()

	var value pgtype.Numeric
	err := value.Scan(raw)
	require.NoError(t, err)
	return value
}

func parsePeopleDate(t *testing.T, raw string) pgtype.Date {
	t.Helper()

	var value pgtype.Date
	err := value.Scan(raw)
	require.NoError(t, err)
	return value
}

func mustParsePeopleDate(raw string) pgtype.Date {
	var value pgtype.Date
	if err := value.Scan(raw); err != nil {
		panic(err)
	}
	return value
}

func baseCreatePersonInput(companyID, actorUserID pgtype.UUID, actorRole sqlc.UserRoleType, kind sqlc.PersonKind) CreatePersonInput {
	return CreatePersonInput{
		CompanyID:      companyID,
		ActorUserID:    actorUserID,
		ActorRole:      actorRole,
		Kind:           kind,
		FullName:       "Pessoa Teste",
		ShortName:      "Pessoa",
		GenderIdentity: sqlc.GenderIdentityWomanCisgender,
		MaritalStatus:  sqlc.MaritalStatusSingle,
		BirthDate:      mustParsePeopleDate("1992-06-15"),
		CPF:            "12345678901",
		Email:          "pessoa.teste@petcontrol.local",
		Cellphone:      "+5511999990001",
		HasWhatsapp:    true,
		IsActive:       true,
	}
}

func countRowsByQuery(t *testing.T, pool *pgxpool.Pool, query string, args ...any) int {
	t.Helper()

	var count int
	err := pool.QueryRow(context.Background(), query, args...).Scan(&count)
	require.NoError(t, err)
	return count
}

func linkedUserIDByPerson(t *testing.T, pool *pgxpool.Pool, personID pgtype.UUID) pgtype.UUID {
	t.Helper()

	var userID pgtype.UUID
	err := pool.QueryRow(context.Background(), "SELECT user_id FROM user_profiles WHERE person_id = $1", personID).Scan(&userID)
	require.NoError(t, err)
	return userID
}

func TestCreateClientWithoutUser(t *testing.T) {
	setup := integration.SetupPostgresWithMigrations(t)
	queries := sqlc.New(setup.Pool)

	company := createIntegrationCompany(t, queries, setup.Pool)
	admin := createIntegrationUserWithRole(t, queries, fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano()), sqlc.UserRoleTypeAdmin)
	attachCompanyUser(t, queries, company.ID, admin.ID, sqlc.UserKindOwner, true)

	serviceUnderTest := NewPeopleService(setup.Pool, queries, nil)
	input := baseCreatePersonInput(company.ID, admin.ID, sqlc.UserRoleTypeAdmin, sqlc.PersonKindClient)
	input.FullName = "Maria Silva"
	input.ShortName = "Maria"
	input.Email = "maria.silva@petcontrol.local"
	input.ClientSince = datePtr(parsePeopleDate(t, "2026-04-01"))
	input.Notes = stringPtr("Cliente recorrente")

	created, err := serviceUnderTest.CreatePerson(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, sqlc.PersonKindClient, created.Kind)
	require.False(t, created.HasSystemUser)
	require.NotNil(t, created.ClientDetails)
	require.Nil(t, created.LinkedUser)

	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM people WHERE id = $1", created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM people_identifications WHERE person_id = $1", created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM people_contacts WHERE person_id = $1", created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_people WHERE company_id = $1 AND person_id = $2", company.ID, created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, `
		SELECT count(*)
		FROM company_clients cc
		INNER JOIN clients c ON c.id = cc.client_id
		WHERE cc.company_id = $1 AND c.person_id = $2
	`, company.ID, created.ID))
	require.Equal(t, 0, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM user_profiles WHERE person_id = $1", created.ID))
}

func TestCreateClientWithUserBySystem(t *testing.T) {
	setup := integration.SetupPostgresWithMigrations(t)
	queries := sqlc.New(setup.Pool)

	company := createIntegrationCompany(t, queries, setup.Pool)
	module := createIntegrationModule(t, queries, fmt.Sprintf("PPL%06d", time.Now().UnixNano()%1000000))
	insertCompanyModule(t, setup.Pool, company.ID, module.ID)
	permission := mustCreatePermissionForRole(t, queries, fmt.Sprintf("people:client:%d", time.Now().UnixNano()), sqlc.UserRoleTypeCommon)
	attachPermissionToModule(t, setup.Pool, module.ID, permission.ID)

	systemUser := createIntegrationUserWithRole(t, queries, fmt.Sprintf("system-%d@example.com", time.Now().UnixNano()), sqlc.UserRoleTypeSystem)
	attachCompanyUser(t, queries, company.ID, systemUser.ID, sqlc.UserKindEmployee, false)

	publisher := &peoplePublisherStub{failTimes: 2}
	serviceUnderTest := NewPeopleService(setup.Pool, queries, publisher)
	input := baseCreatePersonInput(company.ID, systemUser.ID, sqlc.UserRoleTypeSystem, sqlc.PersonKindClient)
	input.FullName = "Cliente com acesso"
	input.ShortName = "Cliente"
	input.Email = "cliente.acesso@petcontrol.local"
	input.CPF = "12345678902"
	input.HasSystemUser = true

	created, err := serviceUnderTest.CreatePerson(context.Background(), input)
	require.NoError(t, err)
	require.True(t, created.HasSystemUser)
	require.NotNil(t, created.LinkedUser)
	require.Equal(t, sqlc.UserRoleTypeCommon, created.LinkedUser.Role)
	require.Equal(t, sqlc.UserKindClient, created.LinkedUser.Kind)

	userID := linkedUserIDByPerson(t, setup.Pool, created.ID)
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_people WHERE company_id = $1 AND person_id = $2", company.ID, created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, `
		SELECT count(*)
		FROM company_clients cc
		INNER JOIN clients c ON c.id = cc.client_id
		WHERE cc.company_id = $1 AND c.person_id = $2
	`, company.ID, created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM user_profiles WHERE person_id = $1", created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_users WHERE company_id = $1 AND user_id = $2 AND kind = 'client'", company.ID, userID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM user_auth WHERE user_id = $1 AND must_change_password = TRUE", userID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM user_permissions WHERE user_id = $1", userID))

	require.True(t, publisher.peopleCalled)
	require.Equal(t, 3, publisher.peopleCalls)
	require.Equal(t, created.ID.String(), publisher.peoplePayload.PersonID)
	require.Equal(t, created.LinkedUser.UserID.String(), publisher.peoplePayload.UserID)
	require.Equal(t, "cliente.acesso@petcontrol.local", publisher.peoplePayload.RecipientEmail)
	require.Equal(t, string(sqlc.UserRoleTypeCommon), publisher.peoplePayload.Role)
}

func TestCreateEmployeeWithUser(t *testing.T) {
	setup := integration.SetupPostgresWithMigrations(t)
	queries := sqlc.New(setup.Pool)

	company := createIntegrationCompany(t, queries, setup.Pool)
	module := createIntegrationModule(t, queries, fmt.Sprintf("HR%06d", time.Now().UnixNano()%1000000))
	insertCompanyModule(t, setup.Pool, company.ID, module.ID)
	permission := mustCreatePermissionForRole(t, queries, fmt.Sprintf("people:employee:%d", time.Now().UnixNano()), sqlc.UserRoleTypeSystem)
	attachPermissionToModule(t, setup.Pool, module.ID, permission.ID)

	admin := createIntegrationUserWithRole(t, queries, fmt.Sprintf("admin-employee-%d@example.com", time.Now().UnixNano()), sqlc.UserRoleTypeAdmin)
	attachCompanyUser(t, queries, company.ID, admin.ID, sqlc.UserKindOwner, true)

	publisher := &peoplePublisherStub{}
	serviceUnderTest := NewPeopleService(setup.Pool, queries, publisher)
	input := baseCreatePersonInput(company.ID, admin.ID, sqlc.UserRoleTypeAdmin, sqlc.PersonKindEmployee)
	input.FullName = "Ana Funcionária"
	input.ShortName = "Ana"
	input.Email = "ana.funcionaria@petcontrol.local"
	input.CPF = "12345678903"
	input.HasSystemUser = true
	input.Employment = &PersonEmploymentInput{
		Role:          "Banhista",
		AdmissionDate: parsePeopleDate(t, "2026-01-10"),
		Salary:        parsePeopleNumeric(t, "2500.75"),
	}
	input.Finance = &PersonFinanceInput{
		BankName:         "Banco Pet",
		BankCode:         stringPtr("001"),
		BankBranch:       "1234",
		BankAccount:      "56789",
		BankAccountDigit: "0",
		BankAccountType:  sqlc.BankAccountKindChecking,
		HasPix:           true,
		PixKey:           stringPtr("ana.funcionaria@petcontrol.local"),
		PixKeyType:       pixKeyTypePtr(sqlc.PixKeyKindEmail),
	}
	input.EmployeeDocs = &PersonEmployeeDocumentsInput{
		RG:          "123456789",
		IssuingBody: "SSP",
		IssuingDate: parsePeopleDate(t, "2010-01-10"),
		CTPS:        "123456",
		CTPSSeries:  "001",
		CTPSState:   "SP",
		PIS:         "12345678901",
		Graduation:  sqlc.GraduationLevelCollegeComplete,
	}
	input.EmployeeBenefits = &PersonEmployeeBenefitsInput{
		MealTicket:            true,
		MealTicketValue:       parsePeopleNumeric(t, "350.00"),
		TransportVoucher:      true,
		TransportVoucherQty:   2,
		TransportVoucherValue: parsePeopleNumeric(t, "220.50"),
		ValidFrom:             parsePeopleDate(t, "2026-01-10"),
	}

	created, err := serviceUnderTest.CreatePerson(context.Background(), input)
	require.NoError(t, err)
	require.True(t, created.HasSystemUser)
	require.NotNil(t, created.LinkedUser)
	require.Equal(t, sqlc.UserRoleTypeSystem, created.LinkedUser.Role)
	require.Equal(t, sqlc.UserKindEmployee, created.LinkedUser.Kind)
	require.NotNil(t, created.EmployeeDetails)
	require.NotNil(t, created.EmployeeDocuments)
	require.NotNil(t, created.EmployeeBenefits)
	require.NotNil(t, created.Finance)

	userID := linkedUserIDByPerson(t, setup.Pool, created.ID)
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_people WHERE company_id = $1 AND person_id = $2", company.ID, created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_employees WHERE company_id = $1 AND person_id = $2", company.ID, created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM user_profiles WHERE person_id = $1", created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM people_finances WHERE person_id = $1", created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM employee_documents WHERE person_id = $1", created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, `
		SELECT count(*)
		FROM employee_benefits eb
		INNER JOIN company_employees ce ON ce.id = eb.company_employee_id
		WHERE ce.company_id = $1 AND ce.person_id = $2
	`, company.ID, created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_users WHERE company_id = $1 AND user_id = $2 AND kind = 'employee'", company.ID, userID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM user_permissions WHERE user_id = $1", userID))
	require.True(t, publisher.peopleCalled)
	require.Equal(t, string(sqlc.UserRoleTypeSystem), publisher.peoplePayload.Role)
}

func TestCreateOutsourcedEmployeeWithUser(t *testing.T) {
	setup := integration.SetupPostgresWithMigrations(t)
	queries := sqlc.New(setup.Pool)

	company := createIntegrationCompany(t, queries, setup.Pool)
	module := createIntegrationModule(t, queries, fmt.Sprintf("OPS%06d", time.Now().UnixNano()%1000000))
	insertCompanyModule(t, setup.Pool, company.ID, module.ID)
	permission := mustCreatePermissionForRole(t, queries, fmt.Sprintf("people:outsourced:%d", time.Now().UnixNano()), sqlc.UserRoleTypeSystem)
	attachPermissionToModule(t, setup.Pool, module.ID, permission.ID)

	admin := createIntegrationUserWithRole(t, queries, fmt.Sprintf("admin-outsourced-%d@example.com", time.Now().UnixNano()), sqlc.UserRoleTypeAdmin)
	attachCompanyUser(t, queries, company.ID, admin.ID, sqlc.UserKindOwner, true)

	publisher := &peoplePublisherStub{}
	serviceUnderTest := NewPeopleService(setup.Pool, queries, publisher)
	input := baseCreatePersonInput(company.ID, admin.ID, sqlc.UserRoleTypeAdmin, sqlc.PersonKindOutsourcedEmployee)
	input.FullName = "Carlos Terceiro"
	input.ShortName = "Carlos"
	input.Email = "carlos.terceiro@petcontrol.local"
	input.CPF = "12345678904"
	input.HasSystemUser = true
	input.Employment = &PersonEmploymentInput{
		Role:          "Motorista",
		AdmissionDate: parsePeopleDate(t, "2026-02-01"),
		Salary:        parsePeopleNumeric(t, "3100.00"),
	}
	input.Finance = &PersonFinanceInput{
		BankName:         "Banco Terceiro",
		BankBranch:       "8888",
		BankAccount:      "12345",
		BankAccountDigit: "6",
		BankAccountType:  sqlc.BankAccountKindSavings,
	}
	input.EmployeeDocs = &PersonEmployeeDocumentsInput{
		RG:          "987654321",
		IssuingBody: "SSP",
		IssuingDate: parsePeopleDate(t, "2012-04-20"),
		CTPS:        "654321",
		CTPSSeries:  "002",
		CTPSState:   "SP",
		PIS:         "10987654321",
		Graduation:  sqlc.GraduationLevelHighComplete,
	}

	created, err := serviceUnderTest.CreatePerson(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, created.LinkedUser)
	require.Equal(t, sqlc.UserRoleTypeSystem, created.LinkedUser.Role)
	require.Equal(t, sqlc.UserKindOutsourcedEmployee, created.LinkedUser.Kind)
	require.NotNil(t, created.EmployeeDetails)
	require.NotNil(t, created.EmployeeDocuments)
	require.NotNil(t, created.Finance)

	userID := linkedUserIDByPerson(t, setup.Pool, created.ID)
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_employees WHERE company_id = $1 AND person_id = $2", company.ID, created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM user_profiles WHERE person_id = $1", created.ID))
	require.Equal(t, 1, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_users WHERE company_id = $1 AND user_id = $2 AND kind = 'outsourced_employee'", company.ID, userID))
	require.True(t, publisher.peopleCalled)
}

func TestBlockSystemCreatingUserForSupplier(t *testing.T) {
	setup := integration.SetupPostgresWithMigrations(t)
	queries := sqlc.New(setup.Pool)

	company := createIntegrationCompany(t, queries, setup.Pool)
	systemUser := createIntegrationUserWithRole(t, queries, fmt.Sprintf("system-supplier-%d@example.com", time.Now().UnixNano()), sqlc.UserRoleTypeSystem)
	attachCompanyUser(t, queries, company.ID, systemUser.ID, sqlc.UserKindEmployee, false)

	serviceUnderTest := NewPeopleService(setup.Pool, queries, nil)
	input := baseCreatePersonInput(company.ID, systemUser.ID, sqlc.UserRoleTypeSystem, sqlc.PersonKindSupplier)
	input.FullName = "Fornecedor Bloqueado"
	input.ShortName = "Fornecedor"
	input.Email = "fornecedor.bloqueado@petcontrol.local"
	input.CPF = "12345678905"
	input.HasSystemUser = true

	_, err := serviceUnderTest.CreatePerson(context.Background(), input)
	require.ErrorIs(t, err, apperror.ErrUnprocessableEntity)
	require.Equal(t, 0, countRowsByQuery(t, setup.Pool, `
		SELECT count(*)
		FROM people_identifications
		WHERE cpf = $1
	`, input.CPF))
	require.Equal(t, 0, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_people"))
	require.Equal(t, 0, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM user_profiles"))
}

func TestCreatePerson_RollsBackTransactionOnEmployeeFailure(t *testing.T) {
	setup := integration.SetupPostgresWithMigrations(t)
	queries := sqlc.New(setup.Pool)

	company := createIntegrationCompany(t, queries, setup.Pool)
	admin := createIntegrationUserWithRole(t, queries, fmt.Sprintf("admin-rollback-%d@example.com", time.Now().UnixNano()), sqlc.UserRoleTypeAdmin)
	attachCompanyUser(t, queries, company.ID, admin.ID, sqlc.UserKindOwner, true)

	serviceUnderTest := NewPeopleService(setup.Pool, queries, nil)
	input := baseCreatePersonInput(company.ID, admin.ID, sqlc.UserRoleTypeAdmin, sqlc.PersonKindEmployee)
	input.FullName = "Falha Empregado"
	input.ShortName = "Falha"
	input.Email = "falha.empregado@petcontrol.local"
	input.CPF = "12345678906"

	_, err := serviceUnderTest.CreatePerson(context.Background(), input)
	require.ErrorIs(t, err, apperror.ErrUnprocessableEntity)
	require.Equal(t, 0, countRowsByQuery(t, setup.Pool, `
		SELECT count(*)
		FROM people_identifications
		WHERE cpf = $1
	`, input.CPF))
	require.Equal(t, 0, countRowsByQuery(t, setup.Pool, `
		SELECT count(*)
		FROM people_contacts
		WHERE email = $1
	`, input.Email))
	require.Equal(t, 0, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_people"))
	require.Equal(t, 0, countRowsByQuery(t, setup.Pool, "SELECT count(*) FROM company_employees"))
}

func datePtr(value pgtype.Date) *pgtype.Date {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func pixKeyTypePtr(value sqlc.PixKeyKind) *sqlc.PixKeyKind {
	return &value
}
