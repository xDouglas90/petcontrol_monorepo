package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

func RequireModule(queries sqlc.Querier, moduleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := moduleCode
		if code == "" {
			code = c.Param("code")
		}
		if code == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "module code is required"})
			return
		}

		companyID, ok := GetCompanyID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "company context required"})
			return
		}

		hasAccess, err := queries.HasActiveCompanyModuleByCode(c.Request.Context(), sqlc.HasActiveCompanyModuleByCodeParams{
			CompanyID: companyID,
			Code:      code,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to verify module access"})
			return
		}
		if !hasAccess {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "module not available for company"})
			return
		}

		c.Next()
	}
}
