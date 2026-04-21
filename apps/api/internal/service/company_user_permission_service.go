package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type CompanyUserPermissionItem struct {
	ID               pgtype.UUID
	Code             string
	Description      *string
	DefaultRoles     []sqlc.UserRoleType
	IsActive         bool
	IsDefaultForRole bool
	GrantedBy        pgtype.UUID
	GrantedAt        pgtype.Timestamptz
}

type CompanyUserPermissionGroup struct {
	ModuleCode        string
	ModuleName        string
	ModuleDescription string
	MinPackage        sqlc.ModulePackage
	Permissions       []CompanyUserPermissionItem
}

type CompanyUserPermissionsSnapshot struct {
	UserID           pgtype.UUID
	CompanyID        pgtype.UUID
	ActivePackage    sqlc.ModulePackage
	Role             sqlc.UserRoleType
	Kind             sqlc.UserKind
	IsOwner          bool
	IsActive         bool
	Permissions      []CompanyUserPermissionItem
	PermissionGroups []CompanyUserPermissionGroup
}

type CompanyUserPermissionService struct {
	queries sqlc.Querier
}

func NewCompanyUserPermissionService(queries sqlc.Querier) *CompanyUserPermissionService {
	return &CompanyUserPermissionService{queries: queries}
}

func (s *CompanyUserPermissionService) ListTenantSettingsPermissions(ctx context.Context, companyID pgtype.UUID, userID pgtype.UUID) (CompanyUserPermissionsSnapshot, error) {
	company, err := s.queries.GetCompanyByID(ctx, companyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return CompanyUserPermissionsSnapshot{}, apperror.ErrNotFound
	}
	if err != nil {
		return CompanyUserPermissionsSnapshot{}, err
	}

	membership, err := s.queries.GetCompanyUser(ctx, sqlc.GetCompanyUserParams{
		CompanyID: companyID,
		UserID:    userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return CompanyUserPermissionsSnapshot{}, apperror.ErrNotFound
	}
	if err != nil {
		return CompanyUserPermissionsSnapshot{}, err
	}

	user, err := s.queries.GetUserByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return CompanyUserPermissionsSnapshot{}, apperror.ErrNotFound
	}
	if err != nil {
		return CompanyUserPermissionsSnapshot{}, err
	}

	manageableCodes := TenantSettingsManageablePermissionCodesByPackage(company.ActivePackage)

	catalog, err := s.queries.ListPermissionsByCodes(ctx, manageableCodes)
	if err != nil {
		return CompanyUserPermissionsSnapshot{}, err
	}

	activePermissions, err := s.queries.ListPermissionsByUserID(ctx, sqlc.ListPermissionsByUserIDParams{
		UserID: userID,
		Offset: 0,
		Limit:  1000,
	})
	if err != nil {
		return CompanyUserPermissionsSnapshot{}, err
	}

	activeByCode := make(map[string]sqlc.ListPermissionsByUserIDRow, len(activePermissions))
	for _, item := range activePermissions {
		activeByCode[item.Code] = item
	}

	permissions := make([]CompanyUserPermissionItem, 0, len(catalog))
	for _, permission := range catalog {
		activeItem, isActive := activeByCode[permission.Code]
		permissions = append(permissions, CompanyUserPermissionItem{
			ID:               permission.ID,
			Code:             permission.Code,
			Description:      textValuePointer(permission.Description),
			DefaultRoles:     permission.DefaultRoles,
			IsActive:         isActive,
			IsDefaultForRole: roleInDefaultRoles(user.Role, permission.DefaultRoles),
			GrantedBy:        activeItem.GrantedBy,
			GrantedAt:        activeItem.GrantedAt,
		})
	}

	permissionGroups := buildCompanyUserPermissionGroups(company.ActivePackage, permissions)

	return CompanyUserPermissionsSnapshot{
		UserID:           userID,
		CompanyID:        companyID,
		ActivePackage:    company.ActivePackage,
		Role:             user.Role,
		Kind:             membership.Kind,
		IsOwner:          membership.IsOwner,
		IsActive:         membership.IsActive,
		Permissions:      permissions,
		PermissionGroups: permissionGroups,
	}, nil
}

func (s *CompanyUserPermissionService) UpdateTenantSettingsPermissions(ctx context.Context, companyID pgtype.UUID, actorUserID pgtype.UUID, targetUserID pgtype.UUID, permissionCodes []string) (CompanyUserPermissionsSnapshot, error) {
	company, err := s.queries.GetCompanyByID(ctx, companyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return CompanyUserPermissionsSnapshot{}, apperror.ErrNotFound
	}
	if err != nil {
		return CompanyUserPermissionsSnapshot{}, err
	}

	desiredCodes, err := normalizeManagedPermissionCodes(permissionCodes, company.ActivePackage)
	if err != nil {
		return CompanyUserPermissionsSnapshot{}, err
	}

	before, err := s.ListTenantSettingsPermissions(ctx, companyID, targetUserID)
	if err != nil {
		return CompanyUserPermissionsSnapshot{}, err
	}

	for _, permission := range before.Permissions {
		_, shouldBeActive := desiredCodes[permission.Code]

		switch {
		case shouldBeActive && !permission.IsActive:
			reactivated, err := s.queries.ReactivateUserPermission(ctx, sqlc.ReactivateUserPermissionParams{
				GrantedBy:    actorUserID,
				UserID:       targetUserID,
				PermissionID: permission.ID,
			})
			if err != nil {
				return CompanyUserPermissionsSnapshot{}, err
			}
			if reactivated == 0 {
				if _, err := s.queries.InsertUserPermission(ctx, sqlc.InsertUserPermissionParams{
					UserID:       targetUserID,
					PermissionID: permission.ID,
					GrantedBy:    actorUserID,
				}); err != nil {
					return CompanyUserPermissionsSnapshot{}, err
				}
			}
		case !shouldBeActive && permission.IsActive:
			if _, err := s.queries.DeleteUserPermission(ctx, sqlc.DeleteUserPermissionParams{
				RevokedBy:    actorUserID,
				UserID:       targetUserID,
				PermissionID: permission.ID,
			}); err != nil {
				return CompanyUserPermissionsSnapshot{}, err
			}
		}
	}

	return s.ListTenantSettingsPermissions(ctx, companyID, targetUserID)
}

func normalizeManagedPermissionCodes(values []string, pkg sqlc.ModulePackage) (map[string]struct{}, error) {
	allowedCodes := TenantSettingsManageablePermissionCodesByPackage(pkg)
	allowed := make(map[string]struct{}, len(allowedCodes))
	for _, code := range allowedCodes {
		allowed[code] = struct{}{}
	}

	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := allowed[value]; !ok {
			return nil, fmt.Errorf("%w: permission code %q is not manageable in this scope", apperror.ErrUnprocessableEntity, value)
		}
		result[value] = struct{}{}
	}

	return result, nil
}

func roleInDefaultRoles(role sqlc.UserRoleType, defaultRoles []sqlc.UserRoleType) bool {
	for _, candidate := range defaultRoles {
		if candidate == role {
			return true
		}
	}
	return false
}

func buildCompanyUserPermissionGroups(pkg sqlc.ModulePackage, permissions []CompanyUserPermissionItem) []CompanyUserPermissionGroup {
	byCode := make(map[string]CompanyUserPermissionItem, len(permissions))
	for _, permission := range permissions {
		byCode[permission.Code] = permission
	}

	groups := make([]CompanyUserPermissionGroup, 0)
	for _, group := range TenantSettingsPermissionGroupsByPackage(pkg) {
		items := make([]CompanyUserPermissionItem, 0)
		for _, code := range group.PermissionCodes {
			if permission, ok := byCode[code]; ok {
				items = append(items, permission)
			}
		}
		for _, prefix := range group.PermissionPrefixes {
			for _, permission := range permissions {
				if len(permission.Code) >= len(prefix) && permission.Code[:len(prefix)] == prefix {
					items = append(items, permission)
				}
			}
		}
		if len(items) == 0 {
			continue
		}
		groups = append(groups, CompanyUserPermissionGroup{
			ModuleCode:        group.ModuleCode,
			ModuleName:        group.ModuleName,
			ModuleDescription: group.ModuleDescription,
			MinPackage:        group.MinPackage,
			Permissions:       items,
		})
	}

	return groups
}
