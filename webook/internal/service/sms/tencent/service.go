package tencent

import (
	"context"
	"errors"
	"fmt"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	appId    *string
	signName *string
	client   *sms.Client
}

func NewService(appId, signName string, client *sms.Client) *Service {
	return &Service{
		appId:    &appId,
		signName: &signName,
		client:   client,
	}
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = &biz
	req.PhoneNumberSet = str2strPtr(numbers...)
	req.TemplateParamSet = str2strPtr(args...)
	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}

	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return errors.New(fmt.Sprintf("send sms failed, status code: %s, status message: %s",
				*status.Code, *status.Message))
		}
	}
	return nil
}

func str2strPtr(src ...string) []*string {
	res := make([]*string, len(src))
	for i, s := range src {
		res[i] = &s
	}
	return res
}
