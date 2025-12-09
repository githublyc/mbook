package service

import (
	"context"
	"mbook/webook/account/domain"
)

type AccountService interface {
	Credit(ctx context.Context, cr domain.Credit) error
}
