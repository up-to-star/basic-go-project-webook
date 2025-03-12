package grpc

import (
	"context"
	"github.com/basic-go-project-webook/webook/api/proto/gen/follow/v1"
	"github.com/basic-go-project-webook/webook/follow/domain"
	"github.com/basic-go-project-webook/webook/follow/service"
	"google.golang.org/grpc"
)

type FollowServiceServer struct {
	followv1.UnimplementedFollowServiceServer
	svc service.FollowService
}

func NewFollowServiceServer(svc service.FollowService) *FollowServiceServer {
	return &FollowServiceServer{
		svc: svc,
	}
}

func (f *FollowServiceServer) Follow(ctx context.Context, request *followv1.FollowRequest) (*followv1.FollowResponse, error) {
	err := f.svc.Follow(ctx, request.GetFollowee(), request.GetFollower())
	return &followv1.FollowResponse{}, err
}

func (f *FollowServiceServer) CancelFollow(ctx context.Context, request *followv1.CancelFollowRequest) (*followv1.CancelFollowResponse, error) {
	err := f.svc.CancelFollow(ctx, request.GetFollowee(), request.GetFollower())
	return &followv1.CancelFollowResponse{}, err
}

func (f *FollowServiceServer) GetFollowee(ctx context.Context, request *followv1.GetFolloweeRequest) (*followv1.GetFolloweeResponse, error) {
	relationList, err := f.svc.GetFollowee(ctx, request.GetFollower(), request.GetOffset(), request.GetLimit())
	if err != nil {
		return nil, err
	}
	res := make([]*followv1.FollowRelation, 0, len(relationList))
	for _, relation := range relationList {
		res = append(res, f.toDTO(relation))
	}
	return &followv1.GetFolloweeResponse{
		FollowRelation: res,
	}, nil
}

func (f *FollowServiceServer) FollowInfo(ctx context.Context, request *followv1.FollowInfoRequest) (*followv1.FollowInfoResponse, error) {
	relation, err := f.svc.FollowInfo(ctx, request.GetFollower(), request.GetFollowee())
	if err != nil {
		return nil, err
	}
	return &followv1.FollowInfoResponse{
		FollowRelation: f.toDTO(relation),
	}, nil
}

func (f *FollowServiceServer) GetFollowStatics(ctx context.Context, request *followv1.GetFollowStaticsRequest) (*followv1.GetFollowStaticsResponse, error) {
	statics, err := f.svc.GetFollowStatics(ctx, request.GetUid())
	if err != nil {
		return nil, err
	}
	return &followv1.GetFollowStaticsResponse{
		FollowerCnt:  statics.Followees,
		FollowingCnt: statics.Followers,
	}, nil
}

func (f *FollowServiceServer) toDTO(domainRelation domain.FollowRelation) *followv1.FollowRelation {
	return &followv1.FollowRelation{
		Followee: domainRelation.Followee,
		Follower: domainRelation.Follower,
	}
}

func (f *FollowServiceServer) Register(server *grpc.Server) {
	followv1.RegisterFollowServiceServer(server, f)
}
