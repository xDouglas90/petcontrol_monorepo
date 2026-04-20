package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// Current godoc
// @Summary Get current user
// @Description Returns the authenticated user profile required by the Web shell.
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} CurrentUserResponseDoc
// @Failure 401 {object} APIErrorResponseDoc
// @Failure 404 {object} APIErrorResponseDoc
// @Failure 500 {object} APIErrorResponseDoc
// @Router /users/me [get]
func (h *UserHandler) Current(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID == "" || claims.CompanyID == "" {
		middleware.JSONError(c, http.StatusUnauthorized, "auth_claims_required", "auth claims required")
		return
	}

	userID, err := parseUUID(claims.UserID)
	if err != nil {
		middleware.JSONError(c, http.StatusUnauthorized, "invalid_user_context", "invalid user context")
		return
	}

	profile, err := h.service.GetCurrentUserProfile(c.Request.Context(), userID)
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "get_current_user_failed", "failed to get current user")
		return
	}

	middleware.JSONData(c, http.StatusOK, gin.H{
		"user_id":    claims.UserID,
		"company_id": claims.CompanyID,
		"person_id":  uuidToString(profile.PersonID),
		"role":       claims.Role,
		"kind":       claims.Kind,
		"full_name":  profile.FullName,
		"short_name": profile.ShortName,
		"image_url":  profile.ImageURL,
		"settings_access": gin.H{
			"can_view":                  profile.SettingsAccess.CanView,
			"can_manage_permissions":    profile.SettingsAccess.CanManagePermissions,
			"active_permission_codes":   profile.SettingsAccess.ActivePermissionCodes,
			"editable_permission_codes": profile.SettingsAccess.EditablePermissionCodes,
		},
	})
}

func (h *UserHandler) List(c *gin.Context) {
	limit := int32(20)
	offset := int32(0)

	if rawLimit := c.Query("limit"); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	if rawOffset := c.Query("offset"); rawOffset != "" {
		if parsed, err := strconv.Atoi(rawOffset); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	users, err := h.service.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *UserHandler) ListCompanyUsers(c *gin.Context) {
	companyID, ok := middleware.GetCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	users, err := h.service.ListCompanyUsers(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list company users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}
