package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type CompanyUserService struct {
	queries sqlc.Querier
}

type CompanyUserWithProfile struct {
	ID        pgtype.UUID
	CompanyID pgtype.UUID
	UserID    pgtype.UUID
	Kind      sqlc.UserKind
	Role      sqlc.UserRoleType
	IsOwner   bool
	IsActive  bool
	JoinedAt  pgtype.Timestamptz
	LeftAt    pgtype.Timestamptz
	FullName  *string
	ShortName *string
	ImageURL  *string
}

func NewCompanyUserService(queries sqlc.Querier) *CompanyUserService {
	return &CompanyUserService{queries: queries}
}

func (s *CompanyUserService) ListCompanyUsers(ctx context.Context, companyID pgtype.UUID) ([]sqlc.CompanyUser, error) {
	return s.queries.ListCompanyUsersByCompanyID(ctx, companyID)
}

func (s *CompanyUserService) ListCompanyUsersWithProfile(ctx context.Context, companyID pgtype.UUID) ([]CompanyUserWithProfile, error) {
	items, err := s.queries.ListCompanyUsersByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	result := make([]CompanyUserWithProfile, 0, len(items))
	for _, item := range items {
		user, err := s.queries.GetUserByID(ctx, item.UserID)
		if err != nil {
			return nil, err
		}

		profile := CompanyUserWithProfile{
			ID:        item.ID,
			CompanyID: item.CompanyID,
			UserID:    item.UserID,
			Kind:      item.Kind,
			Role:      user.Role,
			IsOwner:   item.IsOwner,
			IsActive:  item.IsActive,
			JoinedAt:  item.CreatedAt,
			LeftAt:    item.DeletedAt,
		}

		userProfile, err := s.queries.GetUserProfile(ctx, item.UserID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		if err == nil {
			person, personErr := s.queries.GetPerson(ctx, userProfile.PersonID)
			if personErr != nil && !errors.Is(personErr, pgx.ErrNoRows) {
				return nil, personErr
			}
			if personErr == nil {
				profile.FullName = textValuePointer(person.FullName)
				profile.ShortName = textValuePointer(person.ShortName)
				profile.ImageURL = textValuePointer(person.ImageUrl)
			}
		}

		result = append(result, profile)
	}

	return result, nil
}

func (s *CompanyUserService) GetCompanyUser(ctx context.Context, companyID pgtype.UUID, userID pgtype.UUID) (sqlc.CompanyUser, error) {
	companyUser, err := s.queries.GetCompanyUser(ctx, sqlc.GetCompanyUserParams{CompanyID: companyID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.CompanyUser{}, apperror.ErrNotFound
	}
	return companyUser, err
}

func (s *CompanyUserService) CreateCompanyUser(ctx context.Context, params sqlc.CreateCompanyUserParams) (sqlc.CompanyUser, error) {
	user, err := s.queries.GetUserByID(ctx, params.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.CompanyUser{}, apperror.ErrNotFound
	}
	if err != nil {
		return sqlc.CompanyUser{}, err
	}

	if user.Role == sqlc.UserRoleTypeRoot || user.Role == sqlc.UserRoleTypeInternal {
		return sqlc.CompanyUser{}, apperror.ErrUnprocessableEntity
	}

	companyUser, err := s.queries.CreateCompanyUser(ctx, params)
	if err != nil {
		return sqlc.CompanyUser{}, mapClientDBError(err)
	}

	return companyUser, nil
}

func (s *CompanyUserService) DeactivateCompanyUser(ctx context.Context, companyID pgtype.UUID, userID pgtype.UUID) error {
	return s.queries.DeactivateCompanyUser(ctx, sqlc.DeactivateCompanyUserParams{CompanyID: companyID, UserID: userID})
}
