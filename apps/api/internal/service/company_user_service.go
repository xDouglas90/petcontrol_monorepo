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

func NewCompanyUserService(queries sqlc.Querier) *CompanyUserService {
	return &CompanyUserService{queries: queries}
}

func (s *CompanyUserService) ListCompanyUsers(ctx context.Context, companyID pgtype.UUID) ([]sqlc.CompanyUser, error) {
	return s.queries.ListCompanyUsersByCompanyID(ctx, companyID)
}

func (s *CompanyUserService) GetCompanyUser(ctx context.Context, companyID pgtype.UUID, userID pgtype.UUID) (sqlc.CompanyUser, error) {
	companyUser, err := s.queries.GetCompanyUser(ctx, sqlc.GetCompanyUserParams{CompanyID: companyID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.CompanyUser{}, apperror.ErrNotFound
	}
	return companyUser, err
}

func (s *CompanyUserService) CreateCompanyUser(ctx context.Context, params sqlc.CreateCompanyUserParams) (sqlc.CompanyUser, error) {
	return s.queries.CreateCompanyUser(ctx, params)
}

func (s *CompanyUserService) DeactivateCompanyUser(ctx context.Context, companyID pgtype.UUID, userID pgtype.UUID) error {
	return s.queries.DeactivateCompanyUser(ctx, sqlc.DeactivateCompanyUserParams{CompanyID: companyID, UserID: userID})
}
