package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type CompanySystemConfigService struct {
	queries sqlc.Querier
}

func NewCompanySystemConfigService(queries sqlc.Querier) *CompanySystemConfigService {
	return &CompanySystemConfigService{queries: queries}
}

func (s *CompanySystemConfigService) GetCurrent(ctx context.Context, companyID pgtype.UUID) (sqlc.GetCompanySystemConfigRow, error) {
	item, err := s.queries.GetCompanySystemConfig(ctx, companyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.GetCompanySystemConfigRow{}, apperror.ErrNotFound
	}
	return item, err
}

func (s *CompanySystemConfigService) UpdateCurrent(ctx context.Context, params sqlc.UpdateCompanySystemConfigParams) (sqlc.GetCompanySystemConfigRow, error) {
	rows, err := s.queries.UpdateCompanySystemConfig(ctx, params)
	if err != nil {
		return sqlc.GetCompanySystemConfigRow{}, err
	}
	if rows == 0 {
		return sqlc.GetCompanySystemConfigRow{}, apperror.ErrNotFound
	}
	return s.GetCurrent(ctx, params.CompanyID)
}
