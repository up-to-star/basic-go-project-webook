package grpc

import (
	"context"
	"github.com/basic-go-project-webook/webook/api/proto/gen/comment/v1"
	"github.com/basic-go-project-webook/webook/comment/domain"
	"github.com/basic-go-project-webook/webook/comment/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math"
)

type CommentServiceServer struct {
	commentv1.UnimplementedCommentServiceServer
	svc service.CommentService
}

func NewCommentServiceServer(svc service.CommentService) *CommentServiceServer {
	return &CommentServiceServer{
		svc: svc,
	}
}

func (c *CommentServiceServer) Register(server *grpc.Server) {
	commentv1.RegisterCommentServiceServer(server, c)
}

func (c *CommentServiceServer) GetCommentList(ctx context.Context, request *commentv1.GetCommentListRequest) (*commentv1.GetCommentListResponse, error) {
	minId := request.GetMinId()
	if minId <= 0 {
		minId = math.MaxInt64
	}
	domainComments, err := c.svc.GetCommentList(ctx, request.GetBiz(), request.GetBizId(), request.GetLimit(), minId)
	if err != nil {
		return nil, err
	}
	return &commentv1.GetCommentListResponse{
		Comments: c.toDTO(domainComments),
	}, nil
}

func (c *CommentServiceServer) DeleteComment(ctx context.Context, request *commentv1.DeleteCommentRequest) (*commentv1.DeleteCommentResponse, error) {
	err := c.svc.DeleteComment(ctx, request.GetId())
	return &commentv1.DeleteCommentResponse{}, err
}

func (c *CommentServiceServer) CreateComment(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	request.GetComment()
	err := c.svc.CreateComment(ctx, c.toDomain(request.GetComment()))
	return &commentv1.CreateCommentResponse{}, err
}

func (c *CommentServiceServer) GetMoreReplies(ctx context.Context, request *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	comments, err := c.svc.GetMoreReplies(ctx, request.GetRid(), request.GetLimit(), request.GetMaxId())
	if err != nil {
		return nil, err
	}
	return &commentv1.GetMoreRepliesResponse{
		Comments: c.toDTO(comments),
	}, nil
}

func (c *CommentServiceServer) toDomain(comment *commentv1.Comment) domain.Comment {
	domainComment := domain.Comment{
		Id:      comment.GetId(),
		Content: comment.GetContent(),
		Biz:     comment.GetBiz(),
		BizId:   comment.GetBizId(),
		Commentator: domain.User{
			Id: comment.GetUid(),
		},
	}
	if comment.GetParentComment() != nil {
		domainComment.ParentComment = &domain.Comment{
			Id: comment.GetParentComment().GetId(),
		}
	}

	if comment.GetRootComment() != nil {
		domainComment.RootComment = &domain.Comment{
			Id: comment.GetRootComment().GetId(),
		}
	}
	return domainComment
}

func (c *CommentServiceServer) toDTO(comments []domain.Comment) []*commentv1.Comment {
	res := make([]*commentv1.Comment, 0, len(comments))
	for _, comment := range comments {
		rpcComment := &commentv1.Comment{
			Id:      comment.Id,
			Content: comment.Content,
			Uid:     comment.Commentator.Id,
			Biz:     comment.Biz,
			BizId:   comment.BizId,
			Ctime:   timestamppb.New(comment.Ctime),
			Utime:   timestamppb.New(comment.Utime),
		}
		if comment.RootComment != nil {
			rpcComment.ParentComment = &commentv1.Comment{
				Id: comment.RootComment.Id,
			}
		}
		if comment.RootComment != nil {
			rpcComment.RootComment = &commentv1.Comment{
				Id: comment.RootComment.Id,
			}
		}
		res = append(res, rpcComment)

	}
	rpcCommentMap := make(map[int64]*commentv1.Comment, len(res))
	for _, comment := range res {
		rpcCommentMap[comment.GetId()] = comment
	}

	for _, comment := range comments {
		if comment.ParentComment != nil {
			rpcComment := rpcCommentMap[comment.Id]
			if comment.RootComment != nil {
				val, ok := rpcCommentMap[comment.ParentComment.Id]
				if ok {
					rpcComment.RootComment = val
				}
			}
		}
		if comment.RootComment != nil {
			rpcComment := rpcCommentMap[comment.Id]
			if comment.RootComment != nil {
				val, ok := rpcCommentMap[comment.Id]
				if ok {
					rpcComment.RootComment = val
				}
			}
		}
	}

	return res
}
