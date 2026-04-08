package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type ModuleService struct {
	queries sqlc.Querier
}

func NewModuleService(queries sqlc.Querier) *ModuleService {
	return &ModuleService{queries: queries}
}

func (s *ModuleService) ListModules(ctx context.Context) ([]sqlc.Module, error) {
	return s.queries.ListModules(ctx)
}

func (s *ModuleService) ListActiveModulesByCompanyID(ctx context.Context, companyID pgtype.UUID) ([]sqlc.Module, error) {
	return s.queries.ListActiveModulesByCompanyID(ctx, companyID)
}

func (s *ModuleService) GetModuleByCode(ctx context.Context, code string) (sqlc.Module, error) {
	module, err := s.queries.GetModuleByCode(ctx, code)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Module{}, apperror.ErrNotFound
	}
	return module, err
}

func (s *ModuleService) CreateModule(ctx context.Context, params sqlc.CreateModuleParams) (sqlc.Module, error) {
	return s.queries.CreateModule(ctx, params)
}

func (s *ModuleService) UpdateModule(ctx context.Context, params sqlc.UpdateModuleParams) (sqlc.Module, error) {
	module, err := s.queries.UpdateModule(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Module{}, apperror.ErrNotFound
	}
	return module, err
}

func (s *ModuleService) DeleteModule(ctx context.Context, moduleID pgtype.UUID) error {
	rows, err := s.queries.DeleteModule(ctx, moduleID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return apperror.ErrNotFound
	}
	return nil
}
