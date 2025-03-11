package service

import (
	"context"
	"mbook/mbook/internal/domain"
	"mbook/mbook/internal/reposotory"
)

type UserService struct {
	repo *reposotory.UserRepository
}

func NewUserService(repo *reposotory.UserRepository) *UserService {
	return &UserService{repo: repo}
}
func (svc *UserService) Signup(ctx context.Context, user domain.User) error {
	return svc.repo.Create(ctx, user)
}
