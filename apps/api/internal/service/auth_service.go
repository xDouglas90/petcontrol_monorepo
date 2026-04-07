package service

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
)

type LoginResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	UserID      string `json:"user_id"`
	CompanyID   string `json:"company_id"`
	Role        string `json:"role"`
	Kind        string `json:"kind"`
}

type AuthService struct {
	queries      sqlc.Querier
	jwtSecret    string
	jwtTTL       time.Duration
	maxAttempts  int16
	lockDuration time.Duration
}

func NewAuthService(queries sqlc.Querier, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{
		queries:      queries,
		jwtSecret:    jwtSecret,
		jwtTTL:       jwtTTL,
		maxAttempts:  5,
		lockDuration: 15 * time.Minute,
	}
}

func (s *AuthService) Login(ctx context.Context, email string, password string, clientIP string, userAgent string) (LoginResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return LoginResult{}, apperror.ErrUnprocessableEntity
	}

	ipAddr := parseClientIP(clientIP)
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = s.logAttempt(ctx, pgtype.UUID{}, ipAddr, userAgent, sqlc.LoginResultInvalidCredentials, "user not found")
			return LoginResult{}, apperror.ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if !user.IsActive {
		_ = s.logAttempt(ctx, user.ID, ipAddr, userAgent, sqlc.LoginResultAccountInactive, "user inactive")
		return LoginResult{}, apperror.ErrAccountInactive
	}

	if !user.EmailVerified {
		_ = s.logAttempt(ctx, user.ID, ipAddr, userAgent, sqlc.LoginResultEmailUnverified, "email not verified")
		return LoginResult{}, apperror.ErrEmailNotVerified
	}

	authData, err := s.queries.GetUserAuthByUserID(ctx, user.ID)
	if err != nil {
		return LoginResult{}, err
	}

	now := time.Now()
	if authData.LockedUntil.Valid && authData.LockedUntil.Time.After(now) {
		_ = s.logAttempt(ctx, user.ID, ipAddr, userAgent, sqlc.LoginResultAccountLocked, "account locked")
		return LoginResult{}, apperror.ErrAccountLocked
	}

	if bcrypt.CompareHashAndPassword([]byte(authData.PasswordHash), []byte(password)) != nil {
		_ = s.queries.IncrementUserAuthLoginAttempts(ctx, user.ID)
		if authData.LoginAttempts+1 >= s.maxAttempts {
			_ = s.queries.SetUserAuthLockedUntil(ctx, sqlc.SetUserAuthLockedUntilParams{
				UserID: user.ID,
				LockedUntil: pgtype.Timestamptz{
					Time:  now.Add(s.lockDuration),
					Valid: true,
				},
			})
		}
		_ = s.logAttempt(ctx, user.ID, ipAddr, userAgent, sqlc.LoginResultInvalidCredentials, "password mismatch")
		return LoginResult{}, apperror.ErrInvalidCredentials
	}

	if err := s.queries.ResetUserAuthLoginAttempts(ctx, user.ID); err != nil {
		return LoginResult{}, err
	}

	membership, err := s.queries.GetActiveCompanyUserByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginResult{}, apperror.ErrForbidden
		}
		return LoginResult{}, err
	}

	claims := appjwt.Claims{
		UserID:    pgUUIDToString(user.ID),
		CompanyID: pgUUIDToString(membership.CompanyID),
		Role:      string(user.Role),
		Kind:      string(user.Kind),
	}
	accessToken, err := appjwt.GenerateToken(s.jwtSecret, s.jwtTTL, claims)
	if err != nil {
		return LoginResult{}, err
	}

	_ = s.logAttempt(ctx, user.ID, ipAddr, userAgent, sqlc.LoginResultSuccess, "")

	return LoginResult{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		UserID:      claims.UserID,
		CompanyID:   claims.CompanyID,
		Role:        claims.Role,
		Kind:        claims.Kind,
	}, nil
}

func (s *AuthService) HasModuleAccess(ctx context.Context, companyID pgtype.UUID, moduleCode string) (bool, error) {
	return s.queries.HasActiveCompanyModuleByCode(ctx, sqlc.HasActiveCompanyModuleByCodeParams{
		CompanyID: companyID,
		Code:      moduleCode,
	})
}

func (s *AuthService) logAttempt(ctx context.Context, userID pgtype.UUID, ip netip.Addr, userAgent string, result sqlc.LoginResult, detail string) error {
	params := sqlc.InsertLoginHistoryParams{
		UserID:    userID,
		IPAddress: ip,
		UserAgent: userAgent,
		Result:    result,
	}
	if detail != "" {
		params.FailureDetail = pgtype.Text{String: detail, Valid: true}
	}
	return s.queries.InsertLoginHistory(ctx, params)
}

func parseClientIP(raw string) netip.Addr {
	ip, err := netip.ParseAddr(raw)
	if err == nil {
		return ip
	}
	return netip.MustParseAddr("127.0.0.1")
}

func pgUUIDToString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	u, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return ""
	}
	return u.String()
}
