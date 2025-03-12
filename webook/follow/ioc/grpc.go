package ioc

import (
	grpc2 "github.com/basic-go-project-webook/webook/follow/grpc"
	"github.com/basic-go-project-webook/webook/pkg/grpcx"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitGRPCXServer(followServer *grpc2.FollowServiceServer) *grpcx.Server {
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
	followServer.Register(server)
	return &grpcx.Server{
		Server:   server,
		Port:     cfg.Port,
		EtcdAddr: cfg.EtcdAddr,
		Name:     cfg.Name,
	}
}
