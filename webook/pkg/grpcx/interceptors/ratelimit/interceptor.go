package ratelimit

import (
	"context"
	"github.com/basic-go-project-webook/webook/pkg/ratelimit"
	"github.com/google/martian/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InterceptorBuilder 整个应用限流，也是集群级限流
type InterceptorBuilder struct {
	limiter ratelimit.Limiter
	key     string
}

func NewInterceptorBuilder(limiter ratelimit.Limiter, key string) *InterceptorBuilder {
	return &InterceptorBuilder{
		limiter: limiter,
		key:     key,
	}
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			// 判定限流出现问题，激进的做法，直接过， 保守做法返回
			// 保守的做法，返回限流错误
			log.Errorf("限流器判断限流失败: %v", err)
			return nil, status.Errorf(codes.ResourceExhausted, "限流器判断限流失败")
		}
		if limited {
			return nil, status.Errorf(codes.ResourceExhausted, "限流器判断限流")
		}
		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			// 判定限流出现问题，激进的做法，直接过， 保守做法返回
			// 保守的做法，返回限流错误
			log.Errorf("限流器判断限流失败: %v", err)
			return status.Errorf(codes.ResourceExhausted, "限流器判断限流失败")
		}
		if limited {
			return status.Errorf(codes.ResourceExhausted, "限流器判断限流")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
