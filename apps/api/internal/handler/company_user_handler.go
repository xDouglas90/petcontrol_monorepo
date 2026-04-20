package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type CompanyUserHandler struct {
	service           *service.CompanyUserService
	permissionService *service.CompanyUserPermissionService
}

type createCompanyUserRequest struct {
	UserID  string `json:"user_id"`
	IsOwner bool   `json:"is_owner"`
}

type updateCompanyUserPermissionsRequest struct {
	PermissionCodes []string `json:"permission_codes"`
}

func NewCompanyUserHandler(service *service.CompanyUserService, permissionService ...*service.CompanyUserPermissionService) *CompanyUserHandler {
	handler := &CompanyUserHandler{service: service}
	if len(permissionService) > 0 {
		handler.permissionService = permissionService[0]
	}
	return handler
}

// List godoc
// @Summary List company users
// @Description Lists users linked to the authenticated tenant company.
// @Tags company_users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} CompanyUserListResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /company-users [get]
func (h *CompanyUserHandler) List(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	users, err := h.service.ListCompanyUsersWithProfile(c.Request.Context(), companyID)
	if err != nil {
		middleware.JSONError(c, 500, "list_company_users_failed", "failed to list company users")
		return
	}

	middleware.JSONData(c, 200, mapCompanyUsers(users))
}

// Create godoc
// @Summary Create company user link
// @Description Links an existing user to the authenticated tenant company. Requires company owner access.
// @Tags company_users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CompanyUserCreateRequestDoc true "Company user payload"
// @Success 201 {object} CompanyUserItemResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /company-users [post]
func (h *CompanyUserHandler) Create(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	var req createCompanyUserRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == "" {
		middleware.JSONError(c, 422, "invalid_request_body", "invalid request body")
		return
	}

	userID, err := parseUUID(req.UserID)
	if err != nil {
		middleware.JSONError(c, 422, "invalid_user_id", "invalid user_id")
		return
	}

	kind := sqlc.UserKindEmployee
	if req.IsOwner {
		kind = sqlc.UserKindOwner
	}

	created, err := h.service.CreateCompanyUser(c.Request.Context(), sqlc.CreateCompanyUserParams{
		CompanyID: companyID,
		UserID:    userID,
		Kind:      kind,
		IsOwner:   req.IsOwner,
		IsActive:  pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "create_company_user_failed", "failed to create company user")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionCreate,
		EntityTable: "company_users",
		EntityID:    created.ID,
		CompanyID:   companyID,
		OldData:     nil,
		NewData:     created,
	})

	middleware.JSONData(c, 201, created)
}

// Deactivate godoc
// @Summary Deactivate company user link
// @Description Deactivates a user link for the authenticated tenant company. Requires company owner access.
// @Tags company_users
// @Security BearerAuth
// @Produce json
// @Param user_id path string true "User ID"
// @Success 204
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /company-users/{user_id} [delete]
func (h *CompanyUserHandler) Deactivate(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	userID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		middleware.JSONError(c, 422, "invalid_user_id", "invalid user_id")
		return
	}

	before, err := h.service.GetCompanyUser(c.Request.Context(), companyID, userID)
	if err != nil {
		status := apperror.HTTPStatus(err)
		code := "get_company_user_failed"
		if errors.Is(err, apperror.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
			code = "company_user_not_found"
		}
		middleware.JSONError(c, status, code, "failed to load company user")
		return
	}

	if err := h.service.DeactivateCompanyUser(c.Request.Context(), companyID, userID); err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "deactivate_company_user_failed", "failed to deactivate company user")
		return
	}

	after, err := h.service.GetCompanyUser(c.Request.Context(), companyID, userID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_company_user_failed", "failed to load company user")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionDeactivate,
		EntityTable: "company_users",
		EntityID:    before.ID,
		CompanyID:   companyID,
		OldData:     before,
		NewData:     after,
	})

	c.Status(204)
}

// ListPermissions godoc
// @Summary List company user manageable permissions
// @Description Lists the configurable tenant settings permissions for a user linked to the authenticated company. Admin only.
// @Tags company_users
// @Security BearerAuth
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} CompanyUserPermissionsResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /company-users/{user_id}/permissions [get]
func (h *CompanyUserHandler) ListPermissions(c *gin.Context) {
	companyID, claims, ok := h.requireAdminCompanyContext(c)
	if !ok {
		return
	}
	if h.permissionService == nil {
		middleware.JSONError(c, http.StatusInternalServerError, "permissions_service_unavailable", "permissions service unavailable")
		return
	}

	targetUserID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_user_id", "invalid user_id")
		return
	}

	snapshot, err := h.permissionService.ListTenantSettingsPermissions(c.Request.Context(), companyID, targetUserID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "list_company_user_permissions_failed", "failed to list company user permissions")
		return
	}

	middleware.JSONData(c, http.StatusOK, mapCompanyUserPermissionsSnapshot(snapshot, claims))
}

// UpdatePermissions godoc
// @Summary Update company user manageable permissions
// @Description Updates the configurable tenant settings permissions for a user linked to the authenticated company. Admin only.
// @Tags company_users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param request body CompanyUserPermissionsUpdateRequestDoc true "Permissions payload"
// @Success 200 {object} CompanyUserPermissionsResponseDoc
// @Failure 403 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 422 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /company-users/{user_id}/permissions [patch]
func (h *CompanyUserHandler) UpdatePermissions(c *gin.Context) {
	companyID, claims, ok := h.requireAdminCompanyContext(c)
	if !ok {
		return
	}
	if h.permissionService == nil {
		middleware.JSONError(c, http.StatusInternalServerError, "permissions_service_unavailable", "permissions service unavailable")
		return
	}

	targetUserID, err := parseUUID(c.Param("user_id"))
	if err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_user_id", "invalid user_id")
		return
	}

	var req updateCompanyUserPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	actorUserID, err := parseUUID(claims.UserID)
	if err != nil {
		middleware.JSONError(c, http.StatusForbidden, "invalid_user_id", "invalid user_id in token")
		return
	}

	before, err := h.permissionService.ListTenantSettingsPermissions(c.Request.Context(), companyID, targetUserID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_company_user_permissions_failed", "failed to load company user permissions")
		return
	}

	after, err := h.permissionService.UpdateTenantSettingsPermissions(c.Request.Context(), companyID, actorUserID, targetUserID, req.PermissionCodes)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "update_company_user_permissions_failed", "failed to update company user permissions")
		return
	}

	middleware.AddAuditEntry(c, middleware.AuditEntry{
		Action:      sqlc.LogActionUpdate,
		EntityTable: "user_permissions",
		EntityID:    targetUserID,
		CompanyID:   companyID,
		OldData:     mapCompanyUserPermissionsSnapshot(before, claims),
		NewData:     mapCompanyUserPermissionsSnapshot(after, claims),
	})

	middleware.JSONData(c, http.StatusOK, mapCompanyUserPermissionsSnapshot(after, claims))
}

func (h *CompanyUserHandler) requireAdminCompanyContext(c *gin.Context) (pgtype.UUID, appjwt.Claims, bool) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		middleware.JSONError(c, http.StatusForbidden, "company_context_required", "company context required")
		return pgtype.UUID{}, appjwt.Claims{}, false
	}

	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID == "" {
		middleware.JSONError(c, http.StatusForbidden, "user_context_required", "user context required")
		return pgtype.UUID{}, appjwt.Claims{}, false
	}

	if claims.Role != string(sqlc.UserRoleTypeAdmin) {
		middleware.JSONError(c, http.StatusForbidden, "admin_required", "admin required")
		return pgtype.UUID{}, appjwt.Claims{}, false
	}

	return companyID, claims, true
}

func mapCompanyUserPermissionsSnapshot(snapshot service.CompanyUserPermissionsSnapshot, claims appjwt.Claims) map[string]any {
	permissions := make([]map[string]any, 0, len(snapshot.Permissions))
	for _, item := range snapshot.Permissions {
		permissions = append(permissions, map[string]any{
			"id":                  uuidToString(item.ID),
			"code":                item.Code,
			"description":         item.Description,
			"default_roles":       userRolesToStrings(item.DefaultRoles),
			"is_active":           item.IsActive,
			"is_default_for_role": item.IsDefaultForRole,
			"granted_by":          nullableUUID(item.GrantedBy),
			"granted_at":          nullableTimestamptz(item.GrantedAt),
		})
	}

	return map[string]any{
		"user_id":      uuidToString(snapshot.UserID),
		"company_id":   uuidToString(snapshot.CompanyID),
		"role":         string(snapshot.Role),
		"kind":         string(snapshot.Kind),
		"is_owner":     snapshot.IsOwner,
		"is_active":    snapshot.IsActive,
		"managed_by":   claims.UserID,
		"permissions":  permissions,
		"scope":        "tenant_settings",
	}
}

func nullableUUID(value pgtype.UUID) *string {
	if !value.Valid {
		return nil
	}
	id := uuidToString(value)
	return &id
}

func userRolesToStrings(values []sqlc.UserRoleType) []string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, string(value))
	}
	return items
}
