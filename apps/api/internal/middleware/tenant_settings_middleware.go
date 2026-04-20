package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

func RequireTenantSettingsPermission(queries sqlc.Querier, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok || claims.UserID == "" {
			JSONError(c, http.StatusForbidden, "user_context_required", "user context required")
			return
		}

		userID, err := parseUUID(claims.UserID)
		if err != nil {
			JSONError(c, http.StatusForbidden, "invalid_user_id", "invalid user_id in token")
			return
		}

		activeCodes, err := service.ListActiveTenantSettingsPermissionCodes(c.Request.Context(), queries, userID)
		if err != nil {
			JSONError(c, http.StatusInternalServerError, "tenant_settings_permission_verification_failed", "failed to verify tenant settings permissions")
			return
		}

		access := service.ComputeTenantSettingsAccess(claims.Role, activeCodes)
		if !access.CanView {
			JSONError(c, http.StatusForbidden, "tenant_settings_access_required", "tenant settings access required")
			return
		}

		if claims.Role == string(sqlc.UserRoleTypeAdmin) {
			c.Next()
			return
		}

		if permissionCode == "" || !slices.Contains(access.EditablePermissionCodes, permissionCode) {
			JSONError(c, http.StatusForbidden, "tenant_settings_permission_required", "tenant settings permission required")
			return
		}

		c.Next()
	}
}
