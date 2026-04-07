package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const companyIDContextKey = "company_id"

func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "tenant context not available"})
			return
		}
		if claims.CompanyID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "company_id missing in token"})
			return
		}

		companyID, err := parseUUID(claims.CompanyID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid company_id in token"})
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
