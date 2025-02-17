package ioc

import (
	"github.com/basic-go-project-webook/webook/interactive/events"
	"github.com/basic-go-project-webook/webook/interactive/repository"
	"github.com/basic-go-project-webook/webook/interactive/repository/dao"
	"github.com/basic-go-project-webook/webook/pkg/kafkax"
	events2 "github.com/basic-go-project-webook/webook/pkg/migrator/events"
	"github.com/basic-go-project-webook/webook/pkg/migrator/events/fixer"
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

func InitFixerConsumer(src SrcDB, dst DstDB) *fixer.Consumer[dao.Interactive] {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	c, err := fixer.NewConsumer[dao.Interactive](cfg.Addr, "inconsistent_interactive", src, dst)
	if err != nil {
		panic(err)
	}
	return c
}

func InitInconsistentProducer() events2.Producer {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	return events2.NewKafkaProducer(cfg.Addr, "inconsistent_interactive")
}

func InitConsumers(c1 *events.InteractiveReadEventConsumer, fixConsumer *fixer.Consumer[dao.Interactive]) []kafkax.Consumer {
	return []kafkax.Consumer{c1, fixConsumer}
}
