//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "k8s config",
	},
	Redis: RedisConfig{
		Addr: "k8s config",
	},
}
