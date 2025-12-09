package repository

import (
	"context"
	"mbook/webook/account/domain"
)

type AccountRepository interface {
	AddCredit(ctx context.Context, c domain.Credit) error
}
