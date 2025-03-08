package repository

import (
	"context"
	"database/sql"
	"github.com/basic-go-project-webook/webook/comment/domain"
	"github.com/basic-go-project-webook/webook/comment/repository/dao"
	"golang.org/x/sync/errgroup"
	"time"
)

type CommentRepository interface {
	CreateComment(ctx context.Context, comment domain.Comment) error
	DeleteComment(ctx context.Context, comment domain.Comment) error
	FindByBiz(ctx context.Context, biz string, bizId int64, limit int64, minId int64) ([]domain.Comment, error)
	GetMoreReplies(ctx context.Context, rid int64, limit int64, maxId int64) ([]domain.Comment, error)
	GetCommentByIds(ctx context.Context, ids []int64) ([]domain.Comment, error)
}

type CachedCommentRepository struct {
	dao dao.CommentDAO
}

func (c *CachedCommentRepository) GetCommentByIds(ctx context.Context, ids []int64) ([]domain.Comment, error) {
	comments, err := c.dao.GetCommentByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(comments))
	for _, comment := range comments {
		res = append(res, c.toDomain(comment))
	}
	return res, nil
}

func (c *CachedCommentRepository) GetMoreReplies(ctx context.Context, rid int64, limit int64, maxId int64) ([]domain.Comment, error) {
	comments, err := c.dao.FindRepliesByRid(ctx, rid, limit, maxId)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(comments))
	for _, comment := range comments {
		res = append(res, c.toDomain(comment))
	}
	return res, nil
}

func (c *CachedCommentRepository) FindByBiz(ctx context.Context, biz string, bizId int64, limit int64, minId int64) ([]domain.Comment, error) {
	comments, err := c.dao.FindByBiz(ctx, biz, bizId, limit, minId)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, len(comments))
	downgrade := ctx.Value("downgrade") == "true"
	var eg errgroup.Group

	for i, dc := range comments {
		dc := dc
		idx := i
		// 降级，只查询一级评论
		if downgrade {
			res[idx] = c.toDomain(dc)
			continue
		}
		eg.Go(func() error {
			res[idx] = c.toDomain(dc)
			subComments, err := c.dao.FindRepliesByPid(ctx, dc.Id, 0, 3)
			if err != nil {
				return err
			}
			res[idx].Children = make([]domain.Comment, 0, len(subComments))
			for _, sc := range subComments {
				res[idx].Children = append(res[idx].Children, c.toDomain(sc))
			}
			return nil
		})
	}
	err = eg.Wait()
	return res, err
}

func (c *CachedCommentRepository) DeleteComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Delete(ctx, dao.Comment{
		Id: comment.Id,
	})
}

func (c *CachedCommentRepository) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Insert(ctx, c.toEntity(comment))
}

func NewCommentRepository(dao dao.CommentDAO) CommentRepository {
	return &CachedCommentRepository{
		dao: dao,
	}
}

func (c *CachedCommentRepository) toEntity(comment domain.Comment) dao.Comment {
	now := time.Now()
	res := dao.Comment{
		Id:      comment.Id,
		Uid:     comment.Commentator.Id,
		Biz:     comment.Biz,
		BizId:   comment.BizId,
		Content: comment.Content,
		Ctime:   now.UnixMilli(),
		Utime:   now.UnixMilli(),
	}
	if comment.ParentComment != nil {
		res.PID = sql.NullInt64{
			Int64: comment.ParentComment.Id,
			Valid: true,
		}
	}
	if comment.RootComment != nil {
		res.RootId = sql.NullInt64{
			Int64: comment.RootComment.Id,
			Valid: true,
		}
	}
	return res
}

func (c *CachedCommentRepository) toDomain(comment dao.Comment) domain.Comment {
	res := domain.Comment{
		Id: comment.Id,
		Commentator: domain.User{
			Id: comment.Uid,
		},
		Content: comment.Content,
		Biz:     comment.Biz,
		BizId:   comment.BizId,
		Ctime:   time.UnixMilli(comment.Ctime),
		Utime:   time.UnixMilli(comment.Utime),
	}
	if comment.PID.Valid {
		res.ParentComment = &domain.Comment{
			Id: comment.PID.Int64,
		}
	}
	if comment.RootId.Valid {
		res.RootComment = &domain.Comment{
			Id: comment.RootId.Int64,
		}
	}
	return res
}
