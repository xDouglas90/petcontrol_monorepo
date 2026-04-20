package service

import (
	"context"
	"slices"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

const (
	TenantSettingsPermissionCompanyEdit      = "company_settings:edit"
	TenantSettingsPermissionPlanEdit         = "plan_settings:edit"
	TenantSettingsPermissionPaymentEdit      = "payment_settings:edit"
	TenantSettingsPermissionNotificationEdit = "notification_settings:edit"
	TenantSettingsPermissionIntegrationEdit  = "integration_settings:edit"
	TenantSettingsPermissionSecurityEdit     = "security_settings:edit"
)

var tenantSettingsPermissionCodes = []string{
	TenantSettingsPermissionCompanyEdit,
	TenantSettingsPermissionPlanEdit,
	TenantSettingsPermissionPaymentEdit,
	TenantSettingsPermissionNotificationEdit,
	TenantSettingsPermissionIntegrationEdit,
	TenantSettingsPermissionSecurityEdit,
}

type TenantSettingsAccess struct {
	CanView                 bool
	CanManagePermissions    bool
	ActivePermissionCodes   []string
	EditablePermissionCodes []string
}

func TenantSettingsPermissionCodes() []string {
	return slices.Clone(tenantSettingsPermissionCodes)
}

func ComputeTenantSettingsAccess(role string, activePermissionCodes []string) TenantSettingsAccess {
	active := intersectTenantSettingsPermissionCodes(activePermissionCodes)

	switch role {
	case string(sqlc.UserRoleTypeAdmin):
		return TenantSettingsAccess{
			CanView:                 true,
			CanManagePermissions:    true,
			ActivePermissionCodes:   TenantSettingsPermissionCodes(),
			EditablePermissionCodes: TenantSettingsPermissionCodes(),
		}
	case string(sqlc.UserRoleTypeSystem):
		return TenantSettingsAccess{
			CanView:                 len(active) > 0,
			CanManagePermissions:    false,
			ActivePermissionCodes:   active,
			EditablePermissionCodes: active,
		}
	default:
		return TenantSettingsAccess{
			CanView:                 false,
			CanManagePermissions:    false,
			ActivePermissionCodes:   []string{},
			EditablePermissionCodes: []string{},
		}
	}
}

func ListActiveTenantSettingsPermissionCodes(ctx context.Context, queries sqlc.Querier, userID pgtype.UUID) ([]string, error) {
	permissions, err := queries.ListPermissionsByUserID(ctx, sqlc.ListPermissionsByUserIDParams{
		UserID: userID,
		Offset: 0,
		Limit:  1000,
	})
	if err != nil {
		return nil, err
	}

	codes := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		if slices.Contains(tenantSettingsPermissionCodes, permission.Code) {
			codes = append(codes, permission.Code)
		}
	}

	slices.Sort(codes)
	return codes, nil
}

func intersectTenantSettingsPermissionCodes(codes []string) []string {
	filtered := make([]string, 0, len(codes))
	for _, code := range codes {
		if slices.Contains(tenantSettingsPermissionCodes, code) {
			filtered = append(filtered, code)
		}
	}
	slices.Sort(filtered)
	return filtered
}
