package ioc

import (
	"basic-project/webook/internal/service/sms"
	"basic-project/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
