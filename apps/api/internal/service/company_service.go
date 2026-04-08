package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type CompanyService struct {
	queries sqlc.Querier
}

func NewCompanyService(queries sqlc.Querier) *CompanyService {
	return &CompanyService{queries: queries}
}

func (s *CompanyService) ListCompanies(ctx context.Context) ([]sqlc.Company, error) {
	return s.queries.ListCompanies(ctx)
}

func (s *CompanyService) GetCompany(ctx context.Context, companyID pgtype.UUID) (sqlc.Company, error) {
	company, err := s.queries.GetCompanyByID(ctx, companyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Company{}, apperror.ErrNotFound
	}
	return company, err
}

func (s *CompanyService) GetCurrentCompany(ctx context.Context, companyID pgtype.UUID) (sqlc.Company, error) {
	return s.GetCompany(ctx, companyID)
}

func (s *CompanyService) CreateCompany(ctx context.Context, params sqlc.InsertCompanyParams) (sqlc.Company, error) {
	return s.queries.InsertCompany(ctx, params)
}

func (s *CompanyService) UpdateCompany(ctx context.Context, params sqlc.UpdateCompanyParams) (sqlc.Company, error) {
	rows, err := s.queries.UpdateCompany(ctx, params)
	if err != nil {
		return sqlc.Company{}, err
	}
	if rows == 0 {
		return sqlc.Company{}, apperror.ErrNotFound
	}
	return s.GetCompany(ctx, params.ID)
}

func (s *CompanyService) DeleteCompany(ctx context.Context, companyID pgtype.UUID) error {
	rows, err := s.queries.DeleteCompany(ctx, companyID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperror.ErrNotFound
	}
	return nil
}
