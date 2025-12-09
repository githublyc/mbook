//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"mbook/webook/account/grpc"
	"mbook/webook/account/repository"
	"mbook/webook/account/repository/dao"
	"mbook/webook/account/service"
)

func InitAccountService() *grpc.AccountServiceServer {
	wire.Build(InitTestDB,
		dao.NewCreditGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer)
	return new(grpc.AccountServiceServer)
}
