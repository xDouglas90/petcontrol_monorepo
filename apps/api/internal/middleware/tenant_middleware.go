package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const companyIDContextKey = "company_id"

func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			JSONError(c, 403, "tenant_context_not_available", "tenant context not available")
			return
		}
		if claims.CompanyID == "" {
			JSONError(c, 403, "company_id_missing", "company_id missing in token")
			return
		}

		companyID, err := parseUUID(claims.CompanyID)
		if err != nil {
			JSONError(c, 403, "invalid_company_id", "invalid company_id in token")
			return
		}

		c.Set(companyIDContextKey, companyID)
		c.Next()
	}
}

func GetCompanyID(c *gin.Context) (pgtype.UUID, bool) {
	value, ok := c.Get(companyIDContextKey)
	if !ok {
		return pgtype.UUID{}, false
	}
	companyID, ok := value.(pgtype.UUID)
	return companyID, ok
}

func parseUUID(raw string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return pgtype.UUID{}, err
	}
	var out pgtype.UUID
	copy(out.Bytes[:], parsed[:])
	out.Valid = true
	return out, nil
}
