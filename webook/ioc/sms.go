package ioc

import (
	"basic-project/webook/internal/service/sms"
	"basic-project/webook/internal/service/sms/memory"
	"basic-project/webook/internal/service/sms/metrics"
)

func InitSMSService() sms.Service {
	return metrics.NewPrometheusDecorator(memory.NewService())
}
