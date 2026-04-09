package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func RequireCompanyOwner(queries sqlc.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		companyID, ok := GetCompanyID(c)
		if !ok {
			JSONError(c, 403, "company_context_required", "company context required")
			return
		}

		claims, ok := GetClaims(c)
		if !ok || claims.UserID == "" {
			JSONError(c, 403, "user_context_required", "user context required")
			return
		}

		if claims.Role == "root" {
			c.Next()
			return
		}

		userID, err := parseUUID(claims.UserID)
		if err != nil {
			JSONError(c, 403, "invalid_user_id", "invalid user_id in token")
			return
		}

		membership, err := queries.GetCompanyUser(c.Request.Context(), sqlc.GetCompanyUserParams{
			CompanyID: companyID,
			UserID:    userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				JSONError(c, 403, "company_ownership_required", "company ownership required")
				return
			}

			JSONError(c, 500, "company_ownership_verification_failed", "failed to verify company ownership")
			return
		}

		if !membership.IsActive || !membership.IsOwner {
			JSONError(c, 403, "company_ownership_required", "company ownership required")
			return
		}

		c.Next()
	}
}
