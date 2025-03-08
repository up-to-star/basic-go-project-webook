package service

import (
	"context"
	"github.com/basic-go-project-webook/webook/comment/domain"
	"github.com/basic-go-project-webook/webook/comment/repository"
)

type CommentService interface {
	GetCommentList(ctx context.Context, biz string, bizId int64, limit int64, minId int64) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, id int64) error
	GetMoreReplies(ctx context.Context, rid int64, limit int64, maxId int64) ([]domain.Comment, error)
	CreateComment(ctx context.Context, comment domain.Comment) error
}

type commentService struct {
	repo repository.CommentRepository
}

func NewCommentService(repo repository.CommentRepository) CommentService {
	return &commentService{
		repo: repo,
	}
}

func (c *commentService) GetCommentList(ctx context.Context, biz string, bizId int64, limit int64, minId int64) ([]domain.Comment, error) {
	list, err := c.repo.FindByBiz(ctx, biz, bizId, limit, minId)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (c *commentService) DeleteComment(ctx context.Context, id int64) error {
	return c.repo.DeleteComment(ctx, domain.Comment{
		Id: id,
	})
}

func (c *commentService) GetMoreReplies(ctx context.Context, rid int64, limit int64, maxId int64) ([]domain.Comment, error) {
	return c.repo.GetMoreReplies(ctx, rid, limit, maxId)
}

func (c *commentService) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.repo.CreateComment(ctx, comment)
}
