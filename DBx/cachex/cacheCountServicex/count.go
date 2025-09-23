// Package cacheCountServicex   基于redis缓存计数服务【一套基于redis和本地缓存的计数服务，并且维护了一个榜单的数据，优先命中本地缓存】
package cacheCountServicex

import (
	"context"
	_ "embed"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/cachex/cacheLocalx"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var (
	//go:embed lua/cnt.lua
	LuaCnt string // 缓存计数lua脚本

	//go:embed lua/get_rank.lua
	LuaGetRank string // 获取排名的lua脚本
)

// RankItem 排行榜项
//   - BizID 业务id, 业务的唯一标识。eg: 帖子ID、用户ID或任何你需要排名对象的唯一标识。
//   - Score 业务计数、得分值、点赞数等等, 该业务项的得分。这个值决定了在排行榜中的位置，值越高排名通常越靠前。
//   - Rank 排名, 该业务项在排行榜中的具体名次。eg: 分数最高的项目 Rank 为 1。
type RankItem struct {
	// BizID 业务id, 业务的唯一标识。eg: 帖子ID、用户ID或任何你需要排名对象的唯一标识。
	BizID int64 `json:"biz_id"`

	// Score 业务计数、得分值、点赞数等等, 该业务项的得分、点赞数等等。这个值决定了在排行榜中的位置，值越高排名通常越靠前。
	Score int64 `json:"score"`

	// Rank 排名, 该业务项在排行榜中的具体名次。eg: 分数最高的项目 Rank 为 1。
	Rank int64 `json:"rank"`
}

// Count 计数服务
type Count[K cacheLocalx.Key, V any] struct {
	RedisCache redis.Cmdable
	initOnce   sync.Once // 初始化锁
	LocalCache cacheLocalx.CacheLocalIn[string, string]
	// 缓存计数操作，true为增加，false为减少
	CntOpt bool

	// ===========本地缓存参数===========
	// 本地缓存过期时间Expiration、RankCacheExpiration【redis缓存过期时间在lua脚本中调整】
	//	- 缓存过期时间【默认为5分钟】
	Expiration time.Duration
	// 排行榜本地&redis缓存过期时间，默认1分钟，可set自由调整
	RankCacheExpiration time.Duration

	// 缓存计数字段服务名【eg: like_cnt】
	ServiceTypeName string
	// 本地缓存中数据权重【多用于缓存时间未过期，但是分配内存满了，需释放部分】
	Weight int64

	// ===========redis缓存参数===========
	LuaCnt     string // redis缓存计数的lua脚本
	LuaGetRank string // redis获取排名的lua脚本

	// ===========热榜计算开关，默认true===========
	RankCount   bool
	CntTypeConf GetCntType // 获取计数参数, 默认offset=0, Limit=100为获取的条数, 前100条数据为排行榜数据
	targetTime  int64      // 计数服务中，获取排行榜数据时，指定时间戳，默认为当前时间戳一分钟后

	Error error
	Rank  []RankItem
}

// NewCount 创建一个计数服务
//   - 计数服务名,ServiceTypeName建议显示set设置，每个业务用不同的
func NewCount[K cacheLocalx.Key, V any](redisCache redis.Cmdable, localCache cacheLocalx.CacheLocalIn[string, string]) *Count[K, V] {
	return &Count[K, V]{
		RedisCache: redisCache,
		LocalCache: localCache,
		CntOpt:     true,
		// 本地缓存
		Expiration:          time.Minute * 5, // 缓存过期时间【默认为5分钟】
		RankCacheExpiration: time.Minute,     // 排行榜本地缓存默认1分钟，确保数据新鲜度，可set自由调整

		// 计数服务名
		ServiceTypeName: "count_service",
		Weight:          10, // 默认权重为10，计数服务一般是高频访问数据，所以权重给大一些

		LuaCnt:     LuaCnt,     // 缓存计数lua脚本，默认使用缓存计数lua脚本【可替换自定义lua】
		LuaGetRank: LuaGetRank, // 获取排名的lua脚本，默认使用获取排名的lua脚本【可替换自定义lua】

		RankCount: true,
		CntTypeConf: GetCntType{
			Offset: 1,
			Limit:  10,
		},
		targetTime: time.Now().UnixMilli(),
	}
}

// initLuaCntScripts 在初始化时预加载脚本并获取 SHA
func (i *Count[K, V]) initLuaCntScripts(ctx context.Context) error {
	var err error
	i.initOnce.Do(func() {
		i.LuaCnt, err = i.RedisCache.ScriptLoad(ctx, i.LuaCnt).Result()
	})
	return err
}

// initLuaGetRankScripts 在初始化时预加载脚本并获取 SHA
func (i *Count[K, V]) initLuaGetRankScripts(ctx context.Context) error {
	sha, err := i.RedisCache.ScriptLoad(ctx, i.LuaGetRank).Result()
	if err != nil {
		return err
	}
	i.LuaGetRank = sha
	return nil
}

// ResErr 获取计数服务错误【普通写入直接返回err】，写入时RankCount=true实时计算排行榜数据时使用ResRank
func (i *Count[K, V]) ResErr() error {
	return i.Error
}

// ResRank 获取计数服务错误【普通写入直接返回err】，写入时RankCount=true实时计算排行榜数据时使用ResRank
func (i *Count[K, V]) ResRank() ([]RankItem, error) {
	return i.Rank, i.Error
}

// =====================================================================================
// =====================================================================================
// =====================================================================================
// =====================================================================================

// DelCnt 删除计数, 默认为全部删除【localCache、RedisCache】
func (i *Count[K, V]) DelCnt(ctx context.Context, biz string, bizId int64) error {
	// 删除本地缓存计数
	key := i.Key(biz, bizId)
	_ = i.LocalCache.Del(key)
	// 删除本地缓存排行榜
	rankKey := i.RankKey(biz)
	_ = i.LocalCache.Del(rankKey)
	// 删除redis缓存计数
	err := i.RedisCache.Del(ctx, key).Err()
	if err == nil {
		return i.RedisCache.Del(ctx, rankKey).Err()
	}
	return err
}
