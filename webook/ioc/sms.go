package ioc

import (
	"github.com/basic-go-project-webook/webook/internal/service/sms"
	"github.com/basic-go-project-webook/webook/internal/service/sms/memory"
	"github.com/basic-go-project-webook/webook/internal/service/sms/metrics"
)

func InitSMSService() sms.Service {
	return metrics.NewPrometheusDecorator(memory.NewService())
}
