package ioc

import (
	"github.com/basic-go-project-webook/webook/internal/service/oauth2/wechat"
	"os"
)

func InitOAuth2WechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有环境变量 WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewWechatService(appId, appSecret)
}
