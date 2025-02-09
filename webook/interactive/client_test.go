package main

import (
	"context"
	intrv1 "github.com/basic-go-project-webook/webook/api/proto/gen/intr/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

func TestGRPCClient(t *testing.T) {
	cc, err := grpc.NewClient("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	client := intrv1.NewInteractiveServiceClient(cc)
	resp, err := client.Get(context.Background(), &intrv1.GetRequest{
		Biz:   "article",
		BizId: 1,
		Uid:   1,
	})
	assert.NoError(t, err)
	t.Log(resp.Intr)
}
