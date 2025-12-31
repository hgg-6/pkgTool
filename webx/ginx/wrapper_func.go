package ginx

import (
	"net/http"
	"strconv"

	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	vector *prometheus.CounterVec
	L      logx.Loggerx
)

func NewLogMdlHandlerFunc(l logx.Loggerx) {
	L = l
	if L != nil {
		L.Info("init log prometheus middleware success")
	}
}

func InitCounter(opt prometheus.CounterOpts) {
	vector = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(vector)
}

// WrapBodyAndClaims bizFn 就是你的业务逻辑
func WrapBodyAndClaims[Req any, Claims jwt.Claims](bizFn func(ctx *gin.Context, req Req, uc Claims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			if L != nil {
				L.Error("输入错误", logx.Error(err))
			}
			ctx.JSON(http.StatusBadRequest, Result{Code: http.StatusBadRequest, Msg: "请求参数错误"})
			return
		}
		if L != nil {
			L.Debug("输入参数", logx.Field{Key: "req", Value: req})
		}
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, req, uc)
		if vector != nil {
			vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		}
		if err != nil && L != nil {
			L.Error("执行业务逻辑失败", logx.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[Req any](bizFn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			if L != nil {
				L.Error("输入错误", logx.Error(err))
			}
			ctx.JSON(http.StatusBadRequest, Result{Code: http.StatusBadRequest, Msg: "请求参数错误"})
			return
		}
		if L != nil {
			L.Debug("输入参数", logx.Field{Key: "req", Value: req})
		}
		res, err := bizFn(ctx, req)
		if vector != nil {
			vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		}
		if err != nil && L != nil {
			L.Error("执行业务逻辑失败", logx.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapClaims[Claims any](
	bizFn func(ctx *gin.Context, uc Claims) (Result, error),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, uc)
		if vector != nil {
			vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		}
		if err != nil && L != nil {
			L.Error("执行业务逻辑失败", logx.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func Wrap(fn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil && L != nil {
			// 开始处理 error，其实就是记录一下日志
			L.Error("处理业务逻辑出错",
				logx.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logx.String("route", ctx.FullPath()),
				logx.Error(err))
		}
		if vector != nil {
			vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		}
		ctx.JSON(http.StatusOK, res)
	}
}
