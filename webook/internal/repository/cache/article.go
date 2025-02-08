package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/basic-go-project-webook/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error
	DeleteFirstPage(ctx context.Context, uid int64) error
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, art domain.Article) error
	SetPub(ctx context.Context, art domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	Del(ctx context.Context, id int64) error
}

type RedisArticleCache struct {
	client redis.Cmdable
}

func (r *RedisArticleCache) Del(ctx context.Context, id int64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	data, err := r.client.Get(ctx, r.pubKey(id)).Bytes()
	if err != nil {
		zap.L().Warn("缓存中获取发布文章失败", zap.Int64("id", id), zap.Error(err))
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(data, &art)
	return art, err
}

func (r *RedisArticleCache) SetPub(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.pubKey(art.Id), data, time.Minute*10).Err()
}

func (r *RedisArticleCache) Set(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(art.Id), data, time.Minute).Err()
}

func (r *RedisArticleCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err != nil {
		zap.L().Error("缓存根据文章id获取文章失败", zap.Int64("id", id), zap.Error(err))
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal(data, &art)
	return art, err
}

func (r *RedisArticleCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	data, err := r.client.Get(ctx, r.firstKey(uid)).Bytes()
	if err != nil {
		zap.L().Error("缓存获取文章失败", zap.Int64("uid", uid), zap.Error(err))
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(data, &arts)
	return arts, err
}

func (r *RedisArticleCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}

	data, err := json.Marshal(arts)
	if err != nil {
		zap.L().Error("arts json marshal failed", zap.Error(err))
		return err
	}

	return r.client.Set(ctx, r.firstKey(uid), data, time.Minute*10).Err()
}

func (r *RedisArticleCache) DeleteFirstPage(ctx context.Context, uid int64) error {
	return r.client.Del(ctx, r.firstKey(uid)).Err()
}

func (r *RedisArticleCache) firstKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}

func (r *RedisArticleCache) key(id int64) string {
	return fmt.Sprintf("article:detail:%d", id)
}

func (r *RedisArticleCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:detail:%d", id)
}

func NewRedisArticleCache(client redis.Cmdable) ArticleCache {
	return &RedisArticleCache{
		client: client,
	}
}
