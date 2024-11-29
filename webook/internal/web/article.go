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
}

func (h *ArticleHandle) Edit(ctx *gin.Context) {
	type Req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
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
			Msg:  "系统异常",
		})
		zap.L().Error("未发现用户信息，用户未登录", zap.Error(err))
		return
	}
	artId, err := h.svc.Save(ctx, domain.Article{
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: claims.Uid,
		},
	})
	if err != nil {
		zap.L().Error("文章保存出错", zap.Error(err))
		ctx.JSON(http.StatusOK, Result{
			Code: 0,
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