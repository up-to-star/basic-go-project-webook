package grpcx

import (
	"context"
	"fmt"
	"github.com/basic-go-project-webook/webook/pkg/netx"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
)

type Server struct {
	*grpc.Server
	Port     int
	EtcdAddr string
	Name     string
	client   *etcdv3.Client
	kaCancel func()
}

func (s *Server) Serve() error {
	addr := ":" + strconv.Itoa(s.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 注册服务
	err = s.register()
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	client, err := etcdv3.NewFromURL(s.EtcdAddr)
	if err != nil {
		return err
	}
	s.client = client
	em, err := endpoints.NewManager(client, "service/"+s.Name)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d", netx.GetOutBoundIP(), s.Port)
	key := fmt.Sprintf("service/%s/%s", s.Name, addr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 租期
	var ttl int64 = 5
	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr:     addr,
		Metadata: leaseResp.ID,
	}, etcdv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	// 续租
	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	ch, err := client.KeepAlive(kaCtx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for kaResp := range ch {
			// 记录日志
			zap.L().Debug(kaResp.String())
		}
	}()
	return nil
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.client != nil {
		err := s.client.Close()
		if err != nil {
			return err
		}
	}
	s.GracefulStop()
	return nil
}
