package cacheCountServicex

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/cachex/cacheLocalx/cacheLocalRistrettox"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 创建redis缓存
	redisCache := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	// 创建ristretto缓存
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: 1e7,     // 按键跟踪次数为（10M）。
		MaxCost:     1 << 30, // 最大缓存成本（1GB）“位左移运算符”。
		BufferItems: 64,      // 每个Get缓冲区的键数。
	})
	assert.NoError(t, err)

	// 创建基于ristretto本地缓存
	localCache := cacheLocalRistrettox.NewCacheLocalRistrettoStr(ristrettoCache)
	// 缓存关闭，勿忘
	defer localCache.Close()

	// 创建计数服务
	countCache := NewCount[string, string](redisCache, localCache)
	// 模拟文章Biz点赞增加计数，文章id【bizId】：1、2、3、4、5、6、7、8、9。。。。不同文章的不同点赞数量
	for i := 1; i <= 20; i++ {
		// 模拟用户点赞，文章1计数1，文章2计数2，文章3计数3，，，，，，文章20计数20
		for w := 1; w < i; w++ {
			err = countCache.SetCnt(ctx, "userArticle", int64(i))
		}
	}

	// 获取单个文章服务Biz=userArticle，BizId=11的点赞计数
	serviceCntOne, err := countCache.GetCnt(ctx, "userArticle", 12)
	assert.NoError(t, err)
	t.Log("文章服务中BizId=11计数: ", serviceCntOne[0].Score)
	t.Log("文章服务中BizId=11计数: ", serviceCntOne[0])

	// 获取文章服务中点赞数前10的数据，排行榜点赞数前10热度的数据
	serviceCntSlice, err := countCache.GetCnt(ctx, "userArticle", 1, GetCntType{
		//offset: 0,
		limit: 10,
	})
	assert.NoError(t, err)
	t.Log("文章服务中前10的数据：", serviceCntSlice)
	t.Log("文章服务中第一点赞的数据，业务id-BizId为：", serviceCntSlice[0].BizID)
}
