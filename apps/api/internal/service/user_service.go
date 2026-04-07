package service

import (
	"context"

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
