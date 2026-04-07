package jwt

import (
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndParseToken(t *testing.T) {
	t.Parallel()

	claims := Claims{
		UserID:    "user-1",
		CompanyID: "company-1",
		Role:      "admin",
		Kind:      "owner",
	}

	token, err := GenerateToken("secret", time.Hour, claims)
	require.NoError(t, err)

	parsed, err := ParseToken("secret", token)
	require.NoError(t, err)
	require.Equal(t, claims.UserID, parsed.UserID)
	require.Equal(t, claims.CompanyID, parsed.CompanyID)
	require.Equal(t, claims.Role, parsed.Role)
	require.Equal(t, claims.Kind, parsed.Kind)
	require.Equal(t, claims.UserID, parsed.Subject)
	require.WithinDuration(t, time.Now().Add(time.Hour), parsed.ExpiresAt.Time, 2*time.Second)
}

func TestParseTokenWithWrongSecret(t *testing.T) {
	t.Parallel()

	token, err := GenerateToken("secret", time.Hour, Claims{UserID: "user-1"})
	require.NoError(t, err)

	_, err = ParseToken("different-secret", token)
	require.Error(t, err)
}

func TestParseTokenRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	_, err := ParseToken("secret", "not-a-token")
	require.Error(t, err)
}

func TestGenerateTokenUsesHS256(t *testing.T) {
	t.Parallel()

	token, err := GenerateToken("secret", time.Minute, Claims{UserID: "user-1"})
	require.NoError(t, err)

	parsed, _, err := jwtv5.NewParser().ParseUnverified(token, &Claims{})
	require.NoError(t, err)
	require.Equal(t, jwtv5.SigningMethodHS256.Alg(), parsed.Method.Alg())
}
