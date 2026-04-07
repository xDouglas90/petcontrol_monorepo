package jwt

import (
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    string `json:"user_id"`
	CompanyID string `json:"company_id"`
	Role      string `json:"role"`
	Kind      string `json:"kind"`
	jwtv5.RegisteredClaims
}

func GenerateToken(secret string, ttl time.Duration, claims Claims) (string, error) {
	now := time.Now()
	claims.RegisteredClaims = jwtv5.RegisteredClaims{
		IssuedAt:  jwtv5.NewNumericDate(now),
		ExpiresAt: jwtv5.NewNumericDate(now.Add(ttl)),
		Subject:   claims.UserID,
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(secret string, token string) (Claims, error) {
	var claims Claims
	parsed, err := jwtv5.ParseWithClaims(token, &claims, func(t *jwtv5.Token) (any, error) {
		if t.Method != jwtv5.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Method.Alg())
		}
		return []byte(secret), nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !parsed.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}

	return claims, nil
}
