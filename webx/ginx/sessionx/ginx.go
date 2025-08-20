package sessionx

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type MiddlewareGinx interface {
	SessionCookie(name string) gin.HandlerFunc
	SessionRedis(name string) gin.HandlerFunc
	SessionMemstore(name string) gin.HandlerFunc // memstore基于内存实现
}

type middlewareGinx struct {
	Store sessions.Store
}

// NewMiddlewareGinx 创建一个中间件,【store】需要【NewStore】一个【cookie实现就cookie.NewStore，redis实现就Redis.NewStore。。。】
func NewMiddlewareGinx(store sessions.Store) MiddlewareGinx {
	return &middlewareGinx{Store: store}
}

// SessionMemstore 基于内存实现【注册到gin中间件, server.use()】
func (m *middlewareGinx) SessionMemstore(name string) gin.HandlerFunc {
	return sessions.Sessions(name, m.Store)
}

// SessionRedis 基于redis实现【注册到gin中间件, server.use()】
func (m *middlewareGinx) SessionRedis(name string) gin.HandlerFunc {
	return sessions.Sessions(name, m.Store)
}

// SessionCookie 基于cookie实现【注册到gin中间件, server.use()】
func (m *middlewareGinx) SessionCookie(name string) gin.HandlerFunc {
	return sessions.Sessions(name, m.Store)
}
