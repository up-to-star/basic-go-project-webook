package auth

import (
	"context"
	"errors"
	"github.com/basic-go-project-webook/webook/internal/service/sms"
	"github.com/golang-jwt/jwt/v5"
)

type SMSService struct {
	svc sms.Service
	key string
}

// Send biz 代表线下申请业务方的token
func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	var tc TokenClaims
	token, err := jwt.ParseWithClaims(biz, &tc, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("invalid token")
	}
	return s.svc.Send(ctx, tc.Tpl, args, numbers...)
}

type TokenClaims struct {
	jwt.RegisteredClaims
	Tpl string
}
