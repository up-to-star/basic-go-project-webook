package ioc

import (
	grpc2 "github.com/basic-go-project-webook/webook/interactive/grpc"
	"github.com/basic-go-project-webook/webook/pkg/grpcx"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitGRPCXServer(intrSvc *grpc2.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		EtcdAddr string `yaml:"etcdAddr"`
		Port     int    `json:"port"`
		Name     string `yaml:"name"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	intrSvc.Register(server)
	return &grpcx.Server{
		Server:   server,
		Port:     cfg.Port,
		Name:     cfg.Name,
		EtcdAddr: cfg.EtcdAddr,
	}
}
