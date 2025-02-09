package client

import (
	"context"
	intrv1 "github.com/basic-go-project-webook/webook/api/proto/gen/intr/v1"
	"google.golang.org/grpc"
	"math/rand"
	"sync/atomic"
)

type GrayScaleInteractiveServiceClient struct {
	remote intrv1.InteractiveServiceClient
	local  intrv1.InteractiveServiceClient
	// 控制流量, 随机数 + 阈值
	threshold atomic.Int32
}

func NewGrayScaleInteractiveServiceClient(remote intrv1.InteractiveServiceClient, local intrv1.InteractiveServiceClient) *GrayScaleInteractiveServiceClient {
	res := &GrayScaleInteractiveServiceClient{
		remote:    remote,
		local:     local,
		threshold: atomic.Int32{},
	}
	res.threshold.Store(50)
	return res
}

func (g *GrayScaleInteractiveServiceClient) Like(ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption) (*intrv1.LikeResponse, error) {
	return g.client().Like(ctx, in, opts...)
}

func (g *GrayScaleInteractiveServiceClient) CancelLike(ctx context.Context, in *intrv1.CancelLikeRequest, opts ...grpc.CallOption) (*intrv1.CancelLikeResponse, error) {
	return g.client().CancelLike(ctx, in, opts...)
}

func (g *GrayScaleInteractiveServiceClient) IncrReadCnt(ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv1.IncrReadCntResponse, error) {
	return g.client().IncrReadCnt(ctx, in, opts...)
}

func (g *GrayScaleInteractiveServiceClient) Collect(ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption) (*intrv1.CollectResponse, error) {
	return g.client().Collect(ctx, in, opts...)
}

func (g *GrayScaleInteractiveServiceClient) Get(ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption) (*intrv1.GetResponse, error) {
	return g.client().Get(ctx, in, opts...)
}

func (g *GrayScaleInteractiveServiceClient) GetByIds(ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
	return g.client().GetByIds(ctx, in, opts...)
}

func (g *GrayScaleInteractiveServiceClient) UpdateThreshold(threshold int32) {
	g.threshold.Store(threshold)
}

func (g *GrayScaleInteractiveServiceClient) client() intrv1.InteractiveServiceClient {
	threshold := g.threshold.Load()
	// [0, 100)的随机数
	num := rand.Int31n(100)
	if num <= threshold {
		return g.remote
	}
	return g.local
}
