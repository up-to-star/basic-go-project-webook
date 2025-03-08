//go:build wireinject

package main

import (
	grpc2 "github.com/basic-go-project-webook/webook/comment/grpc"
	"github.com/basic-go-project-webook/webook/comment/ioc"
	"github.com/basic-go-project-webook/webook/comment/repository"
	"github.com/basic-go-project-webook/webook/comment/repository/dao"
	"github.com/basic-go-project-webook/webook/comment/service"
	"github.com/google/wire"
)

var thirdPorivder = wire.NewSet(
	ioc.InitDB,
)

var serviceProvider = wire.NewSet(
	dao.NewCommentDAO,
	repository.NewCommentRepository,
	service.NewCommentService,
	grpc2.NewCommentServiceServer,
)

func InitApp() *App {
	wire.Build(
		thirdPorivder,
		serviceProvider,
		ioc.InitGRPCXServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
