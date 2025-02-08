package ioc

import (
	events2 "github.com/basic-go-project-webook/webook/interactive/events"
	"github.com/basic-go-project-webook/webook/interactive/repository"
	"github.com/basic-go-project-webook/webook/internal/events"
	"github.com/basic-go-project-webook/webook/internal/events/article"
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
