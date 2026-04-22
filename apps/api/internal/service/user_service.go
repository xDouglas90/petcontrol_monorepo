package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type UserService struct {
	queries sqlc.Querier
}

type CurrentUserProfile struct {
	UserID         pgtype.UUID
	PersonID       pgtype.UUID
	FullName       *string
	ShortName      *string
	ImageURL       *string
	SettingsAccess TenantSettingsAccess
}

func NewUserService(queries sqlc.Querier) *UserService {
	return &UserService{queries: queries}
}

func (s *UserService) ListUsers(ctx context.Context, limit int32, offset int32) ([]sqlc.User, error) {
	return s.queries.ListUsersBasic(ctx, sqlc.ListUsersBasicParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (s *UserService) ListCompanyUsers(ctx context.Context, companyID pgtype.UUID) ([]sqlc.CompanyUser, error) {
	return s.queries.ListCompanyUsersByCompanyID(ctx, companyID)
}

func (s *UserService) GetCurrentUserProfile(ctx context.Context, userID pgtype.UUID) (CurrentUserProfile, error) {
	activeSettingsPermissionCodes, err := ListActiveTenantSettingsPermissionCodes(ctx, s.queries, userID)
	if err != nil {
		return CurrentUserProfile{}, err
	}

	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return CurrentUserProfile{}, err
	}

	settingsAccess := ComputeTenantSettingsAccess(string(user.Role), activeSettingsPermissionCodes)

	profile, err := s.queries.GetUserProfile(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return CurrentUserProfile{
			UserID:         userID,
			SettingsAccess: settingsAccess,
		}, nil
	}
	if err != nil {
		return CurrentUserProfile{}, err
	}

	person, err := s.queries.GetPerson(ctx, profile.PersonID)
	if errors.Is(err, pgx.ErrNoRows) {
		return CurrentUserProfile{
			UserID:         profile.UserID,
			PersonID:       profile.PersonID,
			SettingsAccess: settingsAccess,
		}, nil
	}
	if err != nil {
		return CurrentUserProfile{}, err
	}

	return CurrentUserProfile{
		UserID:         profile.UserID,
		PersonID:       profile.PersonID,
		FullName:       textValuePointer(person.FullName),
		ShortName:      textValuePointer(person.ShortName),
		ImageURL:       textValuePointer(person.ImageUrl),
		SettingsAccess: settingsAccess,
	}, nil
}

func textValuePointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}
