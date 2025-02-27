package circuitbreaker

import (
	"context"
	"github.com/go-kratos/aegis/circuitbreaker"
	"google.golang.org/grpc"
)

type InterceptorBuilder struct {
	breaker circuitbreaker.CircuitBreaker
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if b.breaker.Allow() == nil {
			resp, err = handler(ctx, req)
			if err != nil {
				// 系统错误或者是业务错误
				b.breaker.MarkFailed()
			} else {
				b.breaker.MarkSuccess()
			}
		}
		// 触发熔断器
		b.breaker.MarkFailed()
		return nil, err
	}
}
