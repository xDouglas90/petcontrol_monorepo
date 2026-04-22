package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

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

	modules, err := s.queries.ListTenantSettingsModulesByCompanyID(ctx, companyID)
	if err != nil {
		return CompanyUserPermissionsSnapshot{}, err
	}

	catalog, err := s.queries.ListTenantSettingsPermissionsByCompanyID(ctx, companyID)
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

	permissionGroups := buildCompanyUserPermissionGroupsFromRows(modules, catalog, permissions)

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
	_, err := s.queries.GetCompanyByID(ctx, companyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return CompanyUserPermissionsSnapshot{}, apperror.ErrNotFound
	}
	if err != nil {
		return CompanyUserPermissionsSnapshot{}, err
	}

	desiredCodes, err := s.normalizeManagedPermissionCodes(ctx, companyID, permissionCodes)
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

func (s *CompanyUserPermissionService) normalizeManagedPermissionCodes(ctx context.Context, companyID pgtype.UUID, values []string) (map[string]struct{}, error) {
	catalog, err := s.queries.ListTenantSettingsPermissionsByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	allowed := make(map[string]struct{}, len(catalog))
	for _, permission := range catalog {
		allowed[permission.Code] = struct{}{}
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

func buildCompanyUserPermissionGroupsFromRows(modules []sqlc.Module, rows []sqlc.ListTenantSettingsPermissionsByCompanyIDRow, permissions []CompanyUserPermissionItem) []CompanyUserPermissionGroup {
	byCode := make(map[string]CompanyUserPermissionItem, len(permissions))
	for _, permission := range permissions {
		byCode[permission.Code] = permission
	}

	rowsByModule := make(map[string][]sqlc.ListTenantSettingsPermissionsByCompanyIDRow, len(rows))
	for _, row := range rows {
		rowsByModule[row.ModuleCode] = append(rowsByModule[row.ModuleCode], row)
	}

	groups := make([]CompanyUserPermissionGroup, 0, len(modules))
	for _, module := range modules {
		moduleRows := rowsByModule[module.Code]
		if len(moduleRows) == 0 {
			continue
		}

		items := make([]CompanyUserPermissionItem, 0, len(moduleRows))
		for _, row := range moduleRows {
			permission, ok := byCode[row.Code]
			if !ok {
				continue
			}
			items = append(items, permission)
		}
		if len(items) == 0 {
			continue
		}

		slices.SortFunc(items, func(a, b CompanyUserPermissionItem) int {
			return strings.Compare(a.Code, b.Code)
		})

		groups = append(groups, CompanyUserPermissionGroup{
			ModuleCode:        module.Code,
			ModuleName:        module.Name,
			ModuleDescription: module.Description,
			MinPackage:        module.MinPackage,
			Permissions:       items,
		})
	}

	return groups
}
