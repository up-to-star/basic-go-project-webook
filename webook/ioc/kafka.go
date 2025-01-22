package ioc

import (
	"basic-project/webook/internal/events"
	"basic-project/webook/internal/events/article"
	"basic-project/webook/internal/repository"
	"github.com/spf13/viper"
)

func InitProducer() article.Producer {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	return article.NewKafkaProducer(cfg.Addr)
}

func InitInteractiveReadEventConsumer(repo repository.InteractiveRepository) *article.InteractiveReadEventConsumer {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	return article.NewInteractiveReadEventConsumer(cfg.Addr, repo)
}

func InitConsumers(c1 *article.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
