package middleware

import (
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
			JSONError(c, 400, "module_code_required", "module code is required")
			return
		}

		companyID, ok := GetCompanyID(c)
		if !ok {
			JSONError(c, 403, "company_context_required", "company context required")
			return
		}

		hasAccess, err := queries.HasActiveCompanyModuleByCode(c.Request.Context(), sqlc.HasActiveCompanyModuleByCodeParams{
			CompanyID: companyID,
			Code:      code,
		})
		if err != nil {
			JSONError(c, 500, "module_access_verification_failed", "failed to verify module access")
			return
		}
		if !hasAccess {
			JSONError(c, 403, "module_not_available", "module not available for company")
			return
		}

		c.Next()
	}
}
