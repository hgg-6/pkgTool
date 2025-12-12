package limitX

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"gitee.com/hgg_test/pkg_tool/limiter"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"github.com/gin-gonic/gin"
)

type Builder struct {
	prefix string
	//cmd      redis.Cmdable
	//interval time.Duration
	//// 阈值
	//rate int
	limiter limiter.Limiter
	log     logx.Loggerx
}

// NewRedisBuilder 【注册到gin中间件, server.use()】
// ratelimitx.NewBuilder(limiter.NewRedisSlideWindowKLimiter(redisClient, time.Second, 1000)).Build(),
// 限流中间件，注册到 gin框架,使用 redis，100次请求/秒。传三个参数，第一个为redis客户端，第二个为限流时间，第三个为限流次数。
func NewRedisBuilder(l limiter.Limiter, logh logx.Loggerx) *Builder {
	return &Builder{
		prefix:  "ip-limiter",
		limiter: l,
		log:     logh,
	}
}

func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limiter.Limit(ctx, fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP()))
		if err != nil {
			log.Println(err)
			b.log.Error("限流出错，redis有误", logx.Error(err))
			// 这一步很有意思，就是如果这边出错了
			// 要怎么办？
			// 保守做法：因为借助于 Redis 来做限流，那么 Redis 崩溃了，为了防止系统崩溃，直接限流
			ctx.AbortWithStatus(http.StatusInternalServerError)
			// 激进做法：虽然 Redis 崩溃了，但是这个时候还是要尽量服务正常的用户，所以不限流
			// ctx.Next()
			return
		}
		if limited {
			if b.log != nil {
				b.log.Info("请求被限流",
					logx.String("client_ip", ctx.ClientIP()),
					logx.String("path", ctx.Request.URL.Path))
			}
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}
