package service

import (
	"slices"

	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type TenantSettingsPermissionGroup struct {
	ModuleCode         string
	ModuleName         string
	ModuleDescription  string
	MinPackage         sqlc.ModulePackage
	PermissionPrefixes []string
	PermissionCodes    []string
}

var tenantSettingsPermissionGroups = []TenantSettingsPermissionGroup{
	{
		ModuleCode:        "CFG",
		ModuleName:        "Configurações",
		ModuleDescription: "Configurações institucionais, plano, pagamentos, notificações, integrações e segurança do tenant.",
		MinPackage:        sqlc.ModulePackageStarter,
		PermissionCodes: []string{
			TenantSettingsPermissionCompanyEdit,
			TenantSettingsPermissionPlanEdit,
			TenantSettingsPermissionPaymentEdit,
			TenantSettingsPermissionNotificationEdit,
			TenantSettingsPermissionIntegrationEdit,
			TenantSettingsPermissionSecurityEdit,
		},
	},
	{
		ModuleCode:         "UCR",
		ModuleName:         "Usuários",
		ModuleDescription:  "Gestão de usuários vinculados ao tenant.",
		MinPackage:         sqlc.ModulePackageStarter,
		PermissionPrefixes: []string{"users:"},
	},
	{
		ModuleCode:         "CLI",
		ModuleName:         "Clientes",
		ModuleDescription:  "Cadastro e manutenção de clientes.",
		MinPackage:         sqlc.ModulePackageStarter,
		PermissionPrefixes: []string{"clients:"},
	},
	{
		ModuleCode:         "SCH",
		ModuleName:         "Agendamentos",
		ModuleDescription:  "Fluxo operacional de agenda.",
		MinPackage:         sqlc.ModulePackageStarter,
		PermissionPrefixes: []string{"schedules:"},
	},
	{
		ModuleCode:         "SVC",
		ModuleName:         "Serviços",
		ModuleDescription:  "Catálogo e manutenção de serviços.",
		MinPackage:         sqlc.ModulePackageStarter,
		PermissionPrefixes: []string{"services:"},
	},
	{
		ModuleCode:         "PET",
		ModuleName:         "Pets",
		ModuleDescription:  "Cadastro e gestão de pets.",
		MinPackage:         sqlc.ModulePackageBasic,
		PermissionPrefixes: []string{"pets:"},
	},
	{
		ModuleCode:         "RPT",
		ModuleName:         "Relatórios",
		ModuleDescription:  "Relatórios operacionais disponíveis ao tenant.",
		MinPackage:         sqlc.ModulePackageBasic,
		PermissionPrefixes: []string{"reports:"},
	},
	{
		ModuleCode:         "PRD",
		ModuleName:         "Produtos",
		ModuleDescription:  "Cadastro e gestão de produtos.",
		MinPackage:         sqlc.ModulePackageEssential,
		PermissionPrefixes: []string{"products:"},
	},
	{
		ModuleCode:         "DLV",
		ModuleName:         "Tele-busca/Entrega de Pets",
		ModuleDescription:  "Fluxos de pickup e delivery de pets.",
		MinPackage:         sqlc.ModulePackageEssential,
		PermissionPrefixes: []string{"pickup_delivery:"},
	},
	{
		ModuleCode:         "INV",
		ModuleName:         "Estoque",
		ModuleDescription:  "Controle de estoque do tenant.",
		MinPackage:         sqlc.ModulePackageEssential,
		PermissionPrefixes: []string{"stock:"},
	},
	{
		ModuleCode:         "PDC",
		ModuleName:         "Creche de Pets",
		ModuleDescription:  "Operação de creche para pets.",
		MinPackage:         sqlc.ModulePackagePremium,
		PermissionPrefixes: []string{"daycare:"},
	},
	{
		ModuleCode:         "PHO",
		ModuleName:         "Hotel de Pets",
		ModuleDescription:  "Operação de hospedagem para pets.",
		MinPackage:         sqlc.ModulePackagePremium,
		PermissionPrefixes: []string{"hotel:"},
	},
	{
		ModuleCode:         "CHT",
		ModuleName:         "Chat",
		ModuleDescription:  "Recursos de chat do tenant.",
		MinPackage:         sqlc.ModulePackagePremium,
		PermissionPrefixes: []string{"chat:"},
	},
	{
		ModuleCode:         "NTF",
		ModuleName:         "Notificações",
		ModuleDescription:  "Envio e gestão de notificações.",
		MinPackage:         sqlc.ModulePackagePremium,
		PermissionPrefixes: []string{"notifications:"},
	},
	{
		ModuleCode:         "FIN",
		ModuleName:         "Financeiro",
		ModuleDescription:  "Operações financeiras do tenant.",
		MinPackage:         sqlc.ModulePackagePremium,
		PermissionPrefixes: []string{"finances:"},
	},
	{
		ModuleCode:         "SUP",
		ModuleName:         "Fornecedores",
		ModuleDescription:  "Cadastro e relacionamento com fornecedores.",
		MinPackage:         sqlc.ModulePackagePremium,
		PermissionPrefixes: []string{"suppliers:"},
	},
	{
		ModuleCode:         "EUA",
		ModuleName:         "Acesso de Usuários Externos",
		ModuleDescription:  "Gestão de acessos externos vinculados ao tenant.",
		MinPackage:         sqlc.ModulePackagePremium,
		PermissionPrefixes: []string{"external_access:"},
	},
}

func TenantSettingsPermissionGroups() []TenantSettingsPermissionGroup {
	return slices.Clone(tenantSettingsPermissionGroups)
}

func TenantSettingsPermissionGroupsByPackage(pkg sqlc.ModulePackage) []TenantSettingsPermissionGroup {
	groups := make([]TenantSettingsPermissionGroup, 0, len(tenantSettingsPermissionGroups))
	for _, group := range tenantSettingsPermissionGroups {
		if tenantSettingsPackageIncludes(pkg, group.MinPackage) {
			groups = append(groups, group)
		}
	}

	return groups
}

func TenantSettingsManageablePermissionCodesByPackage(pkg sqlc.ModulePackage) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, 64)

	for _, group := range TenantSettingsPermissionGroupsByPackage(pkg) {
		for _, code := range group.PermissionCodes {
			if _, ok := seen[code]; ok {
				continue
			}
			seen[code] = struct{}{}
			result = append(result, code)
		}

		for _, prefix := range group.PermissionPrefixes {
			for _, code := range tenantSettingsPermissionCodesByPrefix(prefix) {
				if _, ok := seen[code]; ok {
					continue
				}
				seen[code] = struct{}{}
				result = append(result, code)
			}
		}
	}

	slices.Sort(result)
	return result
}

func tenantSettingsPermissionCodesByPrefix(prefix string) []string {
	result := make([]string, 0)
	for _, code := range tenantSettingsAllManageablePermissionCodes {
		if len(code) >= len(prefix) && code[:len(prefix)] == prefix {
			result = append(result, code)
		}
	}
	slices.Sort(result)
	return result
}

func tenantSettingsPackageIncludes(current sqlc.ModulePackage, min sqlc.ModulePackage) bool {
	return tenantSettingsPackageRank(current) >= tenantSettingsPackageRank(min)
}

func tenantSettingsPackageRank(pkg sqlc.ModulePackage) int {
	switch pkg {
	case sqlc.ModulePackageTrial:
		return 1
	case sqlc.ModulePackageStarter:
		return 1
	case sqlc.ModulePackageBasic:
		return 2
	case sqlc.ModulePackageEssential:
		return 3
	case sqlc.ModulePackagePremium:
		return 4
	case sqlc.ModulePackageInternal:
		return 5
	default:
		return 0
	}
}

var tenantSettingsAllManageablePermissionCodes = []string{
	TenantSettingsPermissionCompanyEdit,
	TenantSettingsPermissionPlanEdit,
	TenantSettingsPermissionPaymentEdit,
	TenantSettingsPermissionNotificationEdit,
	TenantSettingsPermissionIntegrationEdit,
	TenantSettingsPermissionSecurityEdit,
	"users:create",
	"users:view",
	"users:update",
	"users:delete",
	"users:restore",
	"users:block",
	"users:unblock",
	"clients:create",
	"clients:view",
	"clients:update",
	"clients:delete",
	"clients:restore",
	"clients:deactivate",
	"clients:reactivate",
	"schedules:create",
	"schedules:view",
	"schedules:update",
	"schedules:delete",
	"schedules:deactivate",
	"schedules:reactivate",
	"services:create",
	"services:view",
	"services:update",
	"services:delete",
	"services:deactivate",
	"services:reactivate",
	"pets:create",
	"pets:view",
	"pets:update",
	"pets:delete",
	"pets:deactivate",
	"pets:reactivate",
	"reports:create",
	"reports:view",
	"reports:update",
	"reports:delete",
	"reports:restore",
	"reports:deactivate",
	"reports:reactivate",
	"products:create",
	"products:view",
	"products:update",
	"products:delete",
	"products:deactivate",
	"products:reactivate",
	"pickup_delivery:create",
	"pickup_delivery:view",
	"pickup_delivery:update",
	"pickup_delivery:delete",
	"pickup_delivery:restore",
	"pickup_delivery:deactivate",
	"pickup_delivery:reactivate",
	"stock:create",
	"stock:view",
	"stock:update",
	"stock:delete",
	"stock:deactivate",
	"stock:reactivate",
	"daycare:create",
	"daycare:view",
	"daycare:update",
	"daycare:delete",
	"daycare:restore",
	"daycare:deactivate",
	"daycare:reactivate",
	"hotel:create",
	"hotel:view",
	"hotel:update",
	"hotel:delete",
	"hotel:restore",
	"hotel:deactivate",
	"hotel:reactivate",
	"chat:create",
	"chat:view",
	"chat:update",
	"chat:delete",
	"chat:restore",
	"chat:deactivate",
	"chat:reactivate",
	"notifications:create",
	"notifications:view",
	"notifications:update",
	"notifications:delete",
	"notifications:restore",
	"notifications:deactivate",
	"notifications:reactivate",
	"finances:create",
	"finances:view",
	"finances:update",
	"finances:delete",
	"finances:restore",
	"finances:deactivate",
	"finances:reactivate",
	"suppliers:create",
	"suppliers:view",
	"suppliers:update",
	"suppliers:delete",
	"suppliers:restore",
	"suppliers:deactivate",
	"suppliers:reactivate",
	"external_access:create",
	"external_access:view",
	"external_access:update",
	"external_access:delete",
	"external_access:deactivate",
	"external_access:reactivate",
}
