package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func TestTenantSettingsPermissionGroupsByPackage(t *testing.T) {
	starterGroups := TenantSettingsPermissionGroupsByPackage(sqlc.ModulePackageStarter)
	starterCodes := TenantSettingsManageablePermissionCodesByPackage(sqlc.ModulePackageStarter)

	require.NotEmpty(t, starterGroups)
	require.Contains(t, starterCodes, "company_settings:edit")
	require.Contains(t, starterCodes, "users:create")
	require.Contains(t, starterCodes, "clients:view")
	require.Contains(t, starterCodes, "schedules:view")
	require.Contains(t, starterCodes, "services:view")
	require.NotContains(t, starterCodes, "pets:view")
	require.NotContains(t, starterCodes, "products:view")
	require.NotContains(t, starterCodes, "chat:view")
	require.NotContains(t, starterCodes, "logs:view")
	require.NotContains(t, starterCodes, "plans:view")
}

func TestTenantSettingsPermissionGroupsByPackage_TrialMatchesStarter(t *testing.T) {
	trialCodes := TenantSettingsManageablePermissionCodesByPackage(sqlc.ModulePackageTrial)
	starterCodes := TenantSettingsManageablePermissionCodesByPackage(sqlc.ModulePackageStarter)

	require.Equal(t, starterCodes, trialCodes)
}

func TestTenantSettingsPermissionGroupsByPackage_PremiumIncludesPremiumModules(t *testing.T) {
	premiumCodes := TenantSettingsManageablePermissionCodesByPackage(sqlc.ModulePackagePremium)

	require.Contains(t, premiumCodes, "company_settings:edit")
	require.Contains(t, premiumCodes, "pets:view")
	require.Contains(t, premiumCodes, "products:view")
	require.Contains(t, premiumCodes, "pickup_delivery:view")
	require.Contains(t, premiumCodes, "stock:view")
	require.Contains(t, premiumCodes, "daycare:view")
	require.Contains(t, premiumCodes, "hotel:view")
	require.Contains(t, premiumCodes, "chat:view")
	require.Contains(t, premiumCodes, "notifications:view")
	require.Contains(t, premiumCodes, "finances:view")
	require.Contains(t, premiumCodes, "suppliers:view")
	require.Contains(t, premiumCodes, "external_access:view")
}

func TestTenantSettingsPermissionGroupsExcludeInternalOnlyModules(t *testing.T) {
	groupCodes := make([]string, 0, len(TenantSettingsPermissionGroups()))
	for _, group := range TenantSettingsPermissionGroups() {
		groupCodes = append(groupCodes, group.ModuleCode)
	}

	require.NotContains(t, groupCodes, "TNT")
	require.NotContains(t, groupCodes, "AUD")
	require.NotContains(t, groupCodes, "ATL")
}
