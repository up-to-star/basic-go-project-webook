package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/basic-go-project-webook/webook/follow/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
)

var (
	//go:embed lua/update_cnt.lua
	updateScript string
)

var ErrKeyNotExist = redis.Nil

const (
	fieldFollowerCnt = "follower_cnt"
	fieldFolloweeCnt = "followee_cnt"
)

type FollowCache interface {
	StaticsInfo(ctx context.Context, uid int64) (domain.FollowStatics, error)
	SetStaticsInfo(ctx context.Context, uid int64, statics domain.FollowStatics) error
	Follow(ctx context.Context, follower, followee int64) error
	CancelFollow(ctx context.Context, follower, followee int64) error
}

type RedisFollowCache struct {
	client redis.Cmdable
}

func (r *RedisFollowCache) StaticsInfo(ctx context.Context, uid int64) (domain.FollowStatics, error) {
	data, err := r.client.HGetAll(ctx, r.staticsKey(uid)).Result()
	if err != nil {
		return domain.FollowStatics{}, err
	}
	if len(data) == 0 {
		return domain.FollowStatics{}, ErrKeyNotExist
	}
	var res domain.FollowStatics
	res.Followers, _ = strconv.ParseInt(data[fieldFollowerCnt], 10, 64)
	res.Followees, _ = strconv.ParseInt(data[fieldFolloweeCnt], 10, 64)
	return res, nil
}

func (r *RedisFollowCache) SetStaticsInfo(ctx context.Context, uid int64, statics domain.FollowStatics) error {
	return r.client.HSet(ctx, r.staticsKey(uid),
		fieldFollowerCnt, statics.Followers, fieldFolloweeCnt, statics.Followees).Err()
}

func (r *RedisFollowCache) Follow(ctx context.Context, follower, followee int64) error {
	return r.updateStaticsInfo(ctx, follower, followee, 1)
}

func (r *RedisFollowCache) CancelFollow(ctx context.Context, follower, followee int64) error {
	return r.updateStaticsInfo(ctx, follower, followee, -1)
}

func (r *RedisFollowCache) updateStaticsInfo(ctx context.Context, follower int64, followee int64, delta int) error {
	return r.client.Eval(ctx, updateScript,
		[]string{r.staticsKey(follower), r.staticsKey(followee)}, fieldFolloweeCnt, fieldFollowerCnt, delta).Err()
}

func (r *RedisFollowCache) staticsKey(uid int64) string {
	return fmt.Sprintf("follow:statics:%d", uid)
}

func NewRedisFollowCache(client redis.Cmdable) FollowCache {
	return &RedisFollowCache{
		client: client,
	}
}
