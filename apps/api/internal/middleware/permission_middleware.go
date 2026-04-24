package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func RequirePermission(queries sqlc.Querier, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if permissionCode == "" {
			JSONError(c, http.StatusBadRequest, "permission_code_required", "permission code is required")
			return
		}

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

		permissions, err := queries.ListPermissionsByUserID(c.Request.Context(), sqlc.ListPermissionsByUserIDParams{
			UserID: userID,
			Offset: 0,
			Limit:  1000,
		})
		if err != nil {
			JSONError(c, http.StatusInternalServerError, "permission_verification_failed", "failed to verify permissions")
			return
		}

		codes := make([]string, 0, len(permissions))
		for _, permission := range permissions {
			codes = append(codes, permission.Code)
		}
		if !slices.Contains(codes, permissionCode) {
			JSONError(c, http.StatusForbidden, "permission_required", "permission required")
			return
		}

		c.Next()
	}
}
