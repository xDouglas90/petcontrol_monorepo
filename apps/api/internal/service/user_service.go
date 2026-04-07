package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
)

type UserService struct {
	queries sqlc.Querier
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
