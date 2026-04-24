package service

import (
	"context"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestCompanyUserPermissionService_ListTenantSettingsPermissionsBuildsGroupsFromDatabase(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewCompanyUserPermissionService(queries)

	companyID := newDomainUUID(t)
	userID := newDomainUUID(t)
	cfgModuleID := newDomainUUID(t)
	ucrModuleID := newDomainUUID(t)
	companyPermissionID := newDomainUUID(t)
	planPermissionID := newDomainUUID(t)
	usersViewPermissionID := newDomainUUID(t)
	adminUserID := newDomainUUID(t)
	now := time.Now().UTC()

	mock.ExpectQuery(`(?s)name: GetCompanyByID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(companyID, "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, newDomainUUID(t), sqlc.ModulePackageStarter, true, now, nil, nil))
	mock.ExpectQuery(`(?s)name: GetCompanyUser`).
		WithArgs(companyID, userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company_id", "user_id", "kind", "is_owner", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(newDomainUUID(t), companyID, userID, sqlc.UserKindEmployee, false, true, now, nil, nil))
	mock.ExpectQuery(`(?s)name: GetUserByID`).
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "email", "email_verified", "email_verified_at", "role", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(userID, "system@petcontrol.local", true, now, sqlc.UserRoleTypeSystem, true, now, nil, nil))
	mock.ExpectQuery(`(?s)name: ListTenantSettingsModulesByCompanyID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "name", "description", "min_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(cfgModuleID, "CFG", "Configurações", "Configurações do tenant", sqlc.ModulePackageStarter, true, now, nil, nil).
			AddRow(ucrModuleID, "UCR", "Usuários", "Gestão de usuários", sqlc.ModulePackageStarter, true, now, nil, nil))
	mock.ExpectQuery(`(?s)name: ListTenantSettingsPermissionsByCompanyID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"module_id", "module_code", "module_name", "module_description", "module_min_package", "id", "code", "description", "default_roles", "created_at", "updated_at"}).
			AddRow(cfgModuleID, "CFG", "Configurações", "Configurações do tenant", sqlc.ModulePackageStarter, companyPermissionID, "company_settings:edit", "Editar configurações de negócios", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin}, now, nil).
			AddRow(cfgModuleID, "CFG", "Configurações", "Configurações do tenant", sqlc.ModulePackageStarter, planPermissionID, "plan_settings:edit", "Editar configurações de plano", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem}, now, nil).
			AddRow(ucrModuleID, "UCR", "Usuários", "Gestão de usuários", sqlc.ModulePackageStarter, usersViewPermissionID, "users:view", "Visualizar usuário", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeInternal, sqlc.UserRoleTypeAdmin}, now, nil))
	mock.ExpectQuery(`(?s)name: ListPermissionsByUserID`).
		WithArgs(userID, int32(0), int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "description", "default_roles", "granted_by", "granted_at", "revoked_by", "revoked_at"}).
			AddRow(planPermissionID, "plan_settings:edit", "Editar configurações de plano", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin, sqlc.UserRoleTypeSystem}, adminUserID, now, nil, nil))

	snapshot, err := serviceUnderTest.ListTenantSettingsPermissions(context.Background(), companyID, userID)
	require.NoError(t, err)
	require.Len(t, snapshot.Permissions, 3)
	require.Len(t, snapshot.PermissionGroups, 2)
	require.Equal(t, "CFG", snapshot.PermissionGroups[0].ModuleCode)
	require.Equal(t, "UCR", snapshot.PermissionGroups[1].ModuleCode)
	require.Equal(t, "company_settings:edit", snapshot.PermissionGroups[0].Permissions[0].Code)
	require.Equal(t, "plan_settings:edit", snapshot.PermissionGroups[0].Permissions[1].Code)
	require.True(t, snapshot.PermissionGroups[0].Permissions[1].IsActive)
	require.True(t, snapshot.PermissionGroups[0].Permissions[1].IsDefaultForRole)
	require.Equal(t, "users:view", snapshot.PermissionGroups[1].Permissions[0].Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCompanyUserPermissionService_UpdateTenantSettingsPermissionsRejectsPermissionOutsideTenantCatalog(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	queries := sqlc.New(mock)
	serviceUnderTest := NewCompanyUserPermissionService(queries)

	companyID := newDomainUUID(t)
	actorUserID := newDomainUUID(t)
	targetUserID := newDomainUUID(t)
	cfgModuleID := newDomainUUID(t)
	companyPermissionID := newDomainUUID(t)
	now := time.Now().UTC()

	mock.ExpectQuery(`(?s)name: GetCompanyByID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"id", "slug", "name", "fantasy_name", "cnpj", "foundation_date", "logo_url", "responsible_id", "active_package", "is_active", "created_at", "updated_at", "deleted_at"}).
			AddRow(companyID, "petcontrol", "PetControl", "PetControl", "12345678000195", nil, nil, newDomainUUID(t), sqlc.ModulePackageStarter, true, now, nil, nil))
	mock.ExpectQuery(`(?s)name: ListTenantSettingsPermissionsByCompanyID`).
		WithArgs(companyID).
		WillReturnRows(pgxmock.NewRows([]string{"module_id", "module_code", "module_name", "module_description", "module_min_package", "id", "code", "description", "default_roles", "created_at", "updated_at"}).
			AddRow(cfgModuleID, "CFG", "Configurações", "Configurações do tenant", sqlc.ModulePackageStarter, companyPermissionID, "company_settings:edit", "Editar configurações de negócios", []sqlc.UserRoleType{sqlc.UserRoleTypeRoot, sqlc.UserRoleTypeAdmin}, now, nil))

	_, err = serviceUnderTest.UpdateTenantSettingsPermissions(context.Background(), companyID, actorUserID, targetUserID, []string{"chat:view"})
	require.ErrorIs(t, err, apperror.ErrUnprocessableEntity)
	require.NoError(t, mock.ExpectationsWereMet())
}
