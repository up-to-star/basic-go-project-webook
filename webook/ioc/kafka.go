package ioc

import (
	events2 "basic-project/webook/interactive/events"
	"basic-project/webook/interactive/repository"
	"basic-project/webook/internal/events"
	"basic-project/webook/internal/events/article"
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

func InitInteractiveReadEventConsumer(repo repository.InteractiveRepository) *events2.InteractiveReadEventConsumer {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	return events2.NewInteractiveReadEventConsumer(cfg.Addr, repo)
}

func InitConsumers(c1 *events2.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
