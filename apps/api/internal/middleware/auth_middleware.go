package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
)

const claimsContextKey = "auth_claims"

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
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

func extractBearerToken(c *gin.Context) string {
	// 1. Try Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	// 2. Try query parameter (common for WebSockets)
	return c.Query("token")
}
