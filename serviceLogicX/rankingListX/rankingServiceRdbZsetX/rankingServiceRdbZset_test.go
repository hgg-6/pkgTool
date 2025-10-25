package rankingServiceRdbZsetX

import (
	"context"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/cachex/cacheLocalx"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/cachex/cacheLocalx/cacheLocalRistrettox"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/logx/zerologx"
	"gitee.com/hgg_test/pkg_tool/v2/serviceLogicX/rankingListX/rankingServiceRdbZsetX/types"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"log"
	"os"
	"testing"
	"time"
)

func TestNewRankingService(t *testing.T) {
	// 1. 初始化全局服务
	localcache := newLocalCache()
	//defer localcache.Close()
	globalSvc := NewRankingService(10, newRedisCli(), localcache, newLogger())
	defer globalSvc.Stop()

	// 2. 获取 article 榜单
	articleSvc := globalSvc.WithBizType("article", types.HotScoreProvider{})

	// 3. 启动缓存刷新（可选）【本地缓存默认为15秒过期，自动刷新缓存开启的话，可小一些】
	articleSvc.StartRefresh(10 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 4. 用户点赞
	//for i := 0; i < 1000; i++ {
	//	_ = articleSvc.IncrScore(ctx, strconv.Itoa(i+1), float64(i+1.0), map[string]string{
	//		"title":  "rankTest_" + strconv.Itoa(i+1),
	//		"author": "李四",
	//	})
	//}

	// 5. 获取榜单（自动补全 Title）
	for i := 0; i < 3; i++ {
		now := time.Now()
		log.Println("开始获取榜单：", i+1)
		top100, _ := articleSvc.GetTopN(ctx, 5)
		for _, item := range top100 {
			fmt.Printf("ID: %s, Title: %s, Score: %.2f\n", item.BizID, item.Title, item.Score)
		}
		timeStop := time.Since(now).String()
		log.Println("获取榜单结束,耗时：", timeStop)

		time.Sleep(time.Second * 1)
	}
}

func newLogger() logx.Loggerx {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Level日志级别【可以考虑作为参数传】，测试传zerolog.InfoLevel/NoLevel不打印
	// 模块化: Str("module", "userService模块")
	logger := zerolog.New(os.Stderr).Level(zerolog.DebugLevel).With().CallerWithSkipFrameCount(4).Timestamp().Logger()

	return zerologx.NewZeroLogger(&logger)
}

func newRedisCli() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
}

func TestNewLocalCache(t *testing.T) {
	newLocalCache()
}
func newLocalCache() cacheLocalx.CacheLocalIn[string, []types.HotScore] {
	cache, err := ristretto.NewCache[string, []types.HotScore](&ristretto.Config[string, []types.HotScore]{
		NumCounters: 1e7,     // 按键跟踪次数为（10M）。
		MaxCost:     1 << 30, // 最大缓存成本（1GB）“位左移运算符”。
		BufferItems: 64,      // 每个Get缓冲区的键数。
	})
	if err != nil {
		return nil
	}
	localCache := cacheLocalRistrettox.NewCacheLocalRistrettoStr[string, []types.HotScore](cache)
	return localCache
}
