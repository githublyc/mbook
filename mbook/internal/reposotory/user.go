package reposotory

import (
	"context"
	"mbook/mbook/internal/domain"
	"mbook/mbook/internal/reposotory/dao"
)

type UserRepository struct {
	dao *dao.UserDao
}

func NewUserRepository(dao *dao.UserDao) *UserRepository {
	return &UserRepository{dao: dao}

}
func (repo *UserRepository) Create(ctx context.Context, user domain.User) error {
	return repo.dao.Insert(ctx, dao.User{Email: user.Email, Password: user.Password})
}
