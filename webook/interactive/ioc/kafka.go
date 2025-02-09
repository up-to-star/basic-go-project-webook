package ioc

import (
	"github.com/basic-go-project-webook/webook/interactive/events"
	"github.com/basic-go-project-webook/webook/interactive/repository"
	"github.com/basic-go-project-webook/webook/pkg/kafkax"
	"github.com/spf13/viper"
)

func InitInteractiveReadEventConsumer(repo repository.InteractiveRepository) *events.InteractiveReadEventConsumer {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	return events.NewInteractiveReadEventConsumer(cfg.Addr, repo)
}

func InitConsumers(c1 *events.InteractiveReadEventConsumer) []kafkax.Consumer {
	return []kafkax.Consumer{c1}
}
