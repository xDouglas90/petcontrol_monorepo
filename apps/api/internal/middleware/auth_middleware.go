package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
)

const claimsContextKey = "auth_claims"

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			JSONError(c, 401, "missing_bearer_token", "missing bearer token")
			return
		}

		claims, err := appjwt.ParseToken(secret, token)
		if err != nil {
			JSONError(c, 401, "invalid_token", "invalid token")
			return
		}

		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

func GetClaims(c *gin.Context) (appjwt.Claims, bool) {
	value, ok := c.Get(claimsContextKey)
	if !ok {
		return appjwt.Claims{}, false
	}
	claims, ok := value.(appjwt.Claims)
	return claims, ok
}

func extractBearerToken(authorization string) string {
	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
