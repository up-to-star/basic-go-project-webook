package ioc

import (
	intrv1 "github.com/basic-go-project-webook/webook/api/proto/gen/intr/v1"
	"github.com/basic-go-project-webook/webook/interactive/service"
	"github.com/basic-go-project-webook/webook/internal/web/client"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitIntrGRPCClient(svc service.InteractiveService) intrv1.InteractiveServiceClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool   `yaml:"secure"`
		Threshold int32  `yaml:"threshold"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	var ops []grpc.DialOption
	if cfg.Secure {

	} else {
		ops = append(ops, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if err != nil {
		panic(err)
	}
	cc, err := grpc.NewClient(cfg.Addr, ops...)
	if err != nil {
		panic(err)
	}
	remote := intrv1.NewInteractiveServiceClient(cc)
	local := client.NewInteractiveServiceAdapter(svc)
	res := client.NewGrayScaleInteractiveServiceClient(remote, local)
	viper.OnConfigChange(func(e fsnotify.Event) {
		var cfg Config
		err := viper.UnmarshalKey("grpc.client.intr", &cfg)
		if err != nil {
			zap.L().Error("unmarshal config failed", zap.Error(err))
		}
		res.UpdateThreshold(cfg.Threshold)
	})
	return res
}
