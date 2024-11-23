package jwt

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var (
	AtKey = []byte("BTv_D7]5q+f)9MTLwAA'5N!PJ6d6PNQQ")
	RtKey = []byte("BTv_D7]5q+f)9MTLwAA'5N!PJ6d6xyad")
)

type RedisJwtHandler struct {
	cmd redis.Cmdable
}

func NewRedisJwtHandler(cmd redis.Cmdable) Handler {
	return &RedisJwtHandler{
		cmd: cmd,
	}
}

func (r *RedisJwtHandler) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

func (r *RedisJwtHandler) CheckSession(ctx *gin.Context, ssid string) error {
	_, err := r.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	return err
}

func (r *RedisJwtHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New()
	err := r.SetJWTToken(ctx, uid, ssid.String())
	if err != nil {
		return err
	}
	err = r.SetRefreshToken(ctx, uid, ssid.String())
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisJwtHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(RtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (r *RedisJwtHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(AtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (r *RedisJwtHandler) ClearToken(ctx *gin.Context) error {
	// 清除token
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	tokenStr := r.ExtractToken(ctx)
	var claims UserClaims
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return AtKey, nil
	})
	if err != nil {
		return err
	}

	err = r.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid), "", time.Hour*24*7).Err()
	return err
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}
