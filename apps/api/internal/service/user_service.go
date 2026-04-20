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
	UserID    pgtype.UUID
	PersonID  pgtype.UUID
	FullName  *string
	ShortName *string
	ImageURL  *string
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
	profile, err := s.queries.GetUserProfile(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return CurrentUserProfile{UserID: userID}, nil
	}
	if err != nil {
		return CurrentUserProfile{}, err
	}

	person, err := s.queries.GetPerson(ctx, profile.PersonID)
	if errors.Is(err, pgx.ErrNoRows) {
		return CurrentUserProfile{
			UserID:   profile.UserID,
			PersonID: profile.PersonID,
		}, nil
	}
	if err != nil {
		return CurrentUserProfile{}, err
	}

	return CurrentUserProfile{
		UserID:    profile.UserID,
		PersonID:  profile.PersonID,
		FullName:  textValuePointer(person.FullName),
		ShortName: textValuePointer(person.ShortName),
		ImageURL:  textValuePointer(person.ImageUrl),
	}, nil
}

func textValuePointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}
