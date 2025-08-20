package sessionx

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type SessionMiddlewareGinx interface {
	Sessionx(name string) gin.HandlerFunc
}

type middlewareGinx struct {
	Store sessions.Store
}

// NewMiddlewareGinx 创建一个中间件,【store】需要【NewStore】一个【cookie实现就cookie.NewStore，redis实现就Redis.NewStore。。。】
//
//	【业务逻辑一般设置】
//	sess := sessions.Default(ctx)
//	sess.Set("userId", u.Id) // 设置用户id，把用户Id放入session中
//	sess.Options(sessions.Options{
//		MaxAge: 600, // 设置session过期时间，单位为秒
//	})
//	err := sess.Save() // 保存session，save才会以上sess设置才生效
func NewSessionMiddlewareGinx(store sessions.Store) SessionMiddlewareGinx {
	return &middlewareGinx{Store: store}
}

// SessionRedis 基于redis实现【注册到gin中间件, server.use()】
func (m *middlewareGinx) Sessionx(name string) gin.HandlerFunc {
	return sessions.Sessions(name, m.Store)
}
