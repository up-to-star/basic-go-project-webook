package web

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/service"
	ijwt "basic-project/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"net/http"
)

type ArticleHandle struct {
	svc service.ArticleService
	ijwt.Handler
}

func NewArticleHandle(svc service.ArticleService, hdl ijwt.Handler) *ArticleHandle {
	return &ArticleHandle{
		svc:     svc,
		Handler: hdl,
	}
}

func (h *ArticleHandle) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)
	g.POST("/list", h.List)
}

func (h *ArticleHandle) List(ctx *gin.Context) {
	ctx.String(http.StatusOK, ctx.ClientIP())
}

func (h *ArticleHandle) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}

	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		zap.L().Error("article withdraw bind 失败", zap.Error(err))
		return
	}

	var claims ijwt.UserClaims
	tokenStr := h.ExtractToken(ctx)
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return ijwt.AtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统错误",
		})
		zap.L().Error("未发现用户信息，用户未登录", zap.Error(err))
		return
	}

	err = h.svc.Withdraw(ctx, domain.Article{
		Id: req.Id,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
	if err != nil {
		zap.L().Error("文章保存或更新出错", zap.Error(err))
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "OK",
		Data: req.Id,
	})

}

func (h *ArticleHandle) Publish(ctx *gin.Context) {
	var req ArticleReq
	err := ctx.Bind(&req)
	if err != nil {
		zap.L().Error("article publish 绑定失败", zap.Error(err))
		return
	}
	var claims ijwt.UserClaims
	tokenStr := h.ExtractToken(ctx)
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return ijwt.AtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统错误",
		})
		zap.L().Error("未发现用户信息，用户未登录", zap.Error(err))
		return
	}

	artId, err := h.svc.Publish(ctx, req.toDomain(claims.Uid))
	if err != nil {
		zap.L().Error("发表帖子出错", zap.Error(err))
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "OK",
		Data: artId,
	})
}

func (h *ArticleHandle) Edit(ctx *gin.Context) {
	var req ArticleReq
	err := ctx.Bind(&req)
	if err != nil {
		zap.L().Error("article edit 绑定失败", zap.Error(err))
		return
	}
	// 校验输入
	// 调用 service 代码
	var claims ijwt.UserClaims
	tokenStr := h.ExtractToken(ctx)
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return ijwt.AtKey, nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统错误",
		})
		zap.L().Error("未发现用户信息，用户未登录", zap.Error(err))
		return
	}
	artId, err := h.svc.Save(ctx, req.toDomain(claims.Uid))
	if err != nil {
		zap.L().Error("文章保存或更新出错", zap.Error(err))
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "OK",
		Data: artId,
	})
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req *ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}
