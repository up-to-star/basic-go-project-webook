package web

import (
	"basic-project/webook/internal/domain"
	"basic-project/webook/internal/service"
	ijwt "basic-project/webook/internal/web/jwt"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandle struct {
	svc service.ArticleService
	ijwt.Handler
	intrSvc service.InteractiveService
	biz     string
}

func NewArticleHandle(svc service.ArticleService, hdl ijwt.Handler, intr service.InteractiveService) *ArticleHandle {
	return &ArticleHandle{
		svc:     svc,
		Handler: hdl,
		intrSvc: intr,
		biz:     "article",
	}
}

func (h *ArticleHandle) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)
	g.POST("/list", h.List)
	g.GET("/detail/:id", h.Detail)

	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	pub.POST("/like", h.Like)
}

func (h *ArticleHandle) List(ctx *gin.Context) {
	var req Page
	err := ctx.Bind(&req)
	if err != nil {
		zap.L().Error("绑定出错", zap.Error(err))
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

	arts, err := h.svc.List(ctx, claims.Uid, req.Limit, req.Offset)
	if err != nil {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统错误",
		})
		zap.L().Error("查找文章列表失败", zap.Error(err))
	}
	ctx.JSON(http.StatusOK, &Result{
		Code: 0,
		Msg:  "OK",
		Data: toArticleVOs(arts),
	})

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
		Data: strconv.FormatInt(artId, 10),
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

func (h *ArticleHandle) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "参数错误",
		})
		zap.L().Warn("参数错误", zap.Error(err))
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		zap.L().Error("系统错误, 查询文章失败", zap.Error(err))
		return
	}
	var claims ijwt.UserClaims
	tokenStr := h.ExtractToken(ctx)
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return ijwt.AtKey, nil
	})
	if err != nil || !token.Valid || claims.Uid != art.Author.Id {
		ctx.JSON(http.StatusOK, &Result{
			Code: 5,
			Msg:  "系统错误",
		})
		zap.L().Error("非法用户信息", zap.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "OK",
		Data: ArticleVO{
			Id:         strconv.FormatInt(id, 10),
			Title:      art.Title,
			Abstract:   art.Abstract(),
			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			Status:     art.Status.ToUint8(),
			Ctime:      art.Ctime.Format("2006-01-02 15:04:05"),
			Utime:      art.Utime.Format("2006-01-02 15:04:05"),
		},
	})
}

func (h *ArticleHandle) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "id 参数错误",
		})
		zap.L().Warn("文章查询失败，id参数不对", zap.Error(err))
		return
	}

	art, err := h.svc.GetPubById(ctx, id)
	// 更新阅读数量
	go func() {
		newCtx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()
		er := h.intrSvc.IncrReadCnt(newCtx, h.biz, art.Id)
		if er != nil {
			zap.L().Error("更新阅读次数失败", zap.Error(er), zap.Int64("art_id", art.Id))
		}
	}()

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		zap.L().Error("文章查询失败，系统错误", zap.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "OK",
		Data: ArticleVO{
			Id:         strconv.FormatInt(id, 10),
			Title:      art.Title,
			Abstract:   art.Abstract(),
			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			Status:     art.Status.ToUint8(),
			Ctime:      art.Ctime.Format("2006-01-02 15:04:05"),
			Utime:      art.Utime.Format("2006-01-02 15:04:05"),
		},
	})
}

func (h *ArticleHandle) Like(ctx *gin.Context) {
	type Req struct {
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		zap.L().Error("article Like Bind 出错", zap.Error(err))
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
	var err1 error
	if req.Like {
		err1 = h.intrSvc.Like(ctx, h.biz, req.Id, claims.Uid)
	} else {
		err1 = h.intrSvc.CancelLike(ctx, h.biz, req.Id)
	}
	if err1 != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		zap.L().Error("系统错误", zap.Error(err1), zap.Int64("uid", claims.Uid), zap.Int64("id", req.Id))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "OK",
	})
}

type ArticleReq struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req *ArticleReq) toDomain(uid int64) domain.Article {
	id, _ := strconv.ParseInt(req.Id, 10, 64)
	return domain.Article{
		Id:      id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	}
}
