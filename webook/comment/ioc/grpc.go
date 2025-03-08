package ioc

import (
	grpc2 "github.com/basic-go-project-webook/webook/comment/grpc"
	"github.com/basic-go-project-webook/webook/pkg/grpcx"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitGRPCXServer(commentSvc *grpc2.CommentServiceServer) *grpcx.Server {
	type Config struct {
		EtcdAddr string `yaml:"etcdAddr"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc", &cfg)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	commentSvc.Register(server)

	return &grpcx.Server{
		Server:   server,
		Port:     cfg.Port,
		Name:     cfg.Name,
		EtcdAddr: cfg.EtcdAddr,
	}
}
