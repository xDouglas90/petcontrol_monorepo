package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func RequireCompanyOwner(queries sqlc.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		companyID, ok := GetCompanyID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "company context required"})
			return
		}

		claims, ok := GetClaims(c)
		if !ok || claims.UserID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user context required"})
			return
		}

		if claims.Role == "root" {
			c.Next()
			return
		}

		userID, err := parseUUID(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid user_id in token"})
			return
		}

		membership, err := queries.GetCompanyUser(c.Request.Context(), sqlc.GetCompanyUserParams{
			CompanyID: companyID,
			UserID:    userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "company ownership required"})
				return
			}

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to verify company ownership"})
			return
		}

		if !membership.IsActive || !membership.IsOwner {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "company ownership required"})
			return
		}

		c.Next()
	}
}
