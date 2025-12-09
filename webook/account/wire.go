//go:build wireinject

package main

import (
	"github.com/google/wire"
	"mbook/webook/account/grpc"
	"mbook/webook/account/ioc"
	"mbook/webook/account/repository"
	"mbook/webook/account/repository/dao"
	"mbook/webook/account/service"
	"mbook/webook/pkg/wego"
)

func Init() *wego.App {
	wire.Build(
		ioc.InitDB,
		ioc.InitLogger,
		ioc.InitEtcdClient,
		ioc.InitGRPCxServer,
		dao.NewCreditGORMDAO,
		repository.NewAccountRepository,
		service.NewAccountService,
		grpc.NewAccountServiceServer,
		wire.Struct(new(wego.App), "GRPCServer"))
	return new(wego.App)
}
