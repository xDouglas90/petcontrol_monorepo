package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type PlanService struct {
	queries sqlc.Querier
}

func NewPlanService(queries sqlc.Querier) *PlanService {
	return &PlanService{queries: queries}
}

func (s *PlanService) ListPlans(ctx context.Context) ([]sqlc.Plan, error) {
	return s.queries.ListPlans(ctx)
}

func (s *PlanService) ListPlanTypes(ctx context.Context) ([]sqlc.PlanType, error) {
	return s.queries.ListPlanTypes(ctx)
}

func (s *PlanService) ListPlansByPackage(ctx context.Context, modulePackage sqlc.ModulePackage) ([]sqlc.Plan, error) {
	return s.queries.ListPlansByPackage(ctx, modulePackage)
}

func (s *PlanService) GetPlan(ctx context.Context, planID pgtype.UUID) (sqlc.Plan, error) {
	plan, err := s.queries.GetPlanByID(ctx, planID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Plan{}, apperror.ErrNotFound
	}
	return plan, err
}

func (s *PlanService) GetCurrentPlan(ctx context.Context, companyID pgtype.UUID) (sqlc.Plan, error) {
	plan, err := s.queries.GetCurrentPlanByCompanyID(ctx, companyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Plan{}, apperror.ErrNotFound
	}
	return plan, err
}

func (s *PlanService) CreatePlanType(ctx context.Context, params sqlc.InsertPlanTypeParams) (sqlc.PlanType, error) {
	return s.queries.InsertPlanType(ctx, params)
}

func (s *PlanService) CreatePlan(ctx context.Context, params sqlc.InsertPlanParams) (sqlc.Plan, error) {
	return s.queries.InsertPlan(ctx, params)
}

func (s *PlanService) UpdatePlan(ctx context.Context, params sqlc.UpdatePlanParams) (sqlc.Plan, error) {
	rows, err := s.queries.UpdatePlan(ctx, params)
	if err != nil {
		return sqlc.Plan{}, err
	}
	if rows == 0 {
		return sqlc.Plan{}, apperror.ErrNotFound
	}
	return s.GetPlan(ctx, params.ID)
}
