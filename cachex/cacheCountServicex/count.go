// Package CountCachex  缓存计数服务【一套基于redis和本地缓存的计数服务，并且维护了一个榜单的数据，优先命中本地缓存】
package cacheCountServicex

import (
	"context"
	_ "embed"
	"gitee.com/hgg_test/pkg_tool/v2/cachex"
	"github.com/redis/go-redis/v9"
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
	// Score 业务计数、得分值、点赞数等等, 该业务项的得分。这个值决定了在排行榜中的位置，值越高排名通常越靠前。
	Score int64 `json:"score"`
	// Rank 排名, 该业务项在排行榜中的具体名次。eg: 分数最高的项目 Rank 为 1。
	Rank int64 `json:"rank"`
}

// Count 计数服务
type Count[K cachex.Key, V any] struct {
	RedisCache redis.Cmdable
	LocalCache cachex.CacheLocalIn[string, string]
	// 缓存计数操作，true为增加，false为减少
	CntOpt bool

	// ===========本地缓存参数===========
	// 本地缓存过期时间Expiration、RankCacheExpiration【redis缓存过期时间在lua脚本中调整】
	//	- 缓存过期时间【默认为5分钟】
	Expiration time.Duration
	// 排行榜本地缓存过期时间
	RankCacheExpiration time.Duration

	// 缓存计数字段服务名【eg: like_cnt】
	ServiceTypeName string
	// 本地缓存中数据权重【多用于缓存时间未过期，但是分配内存满了，需释放部分】
	Weight int64

	// ===========redis缓存参数===========
	LuaCnt     string // redis缓存计数的lua脚本
	LuaGetRank string // redis获取排名的lua脚本
}

// NewCount 创建一个计数服务
func NewCount[K cachex.Key, V any](redisCache redis.Cmdable, localCache cachex.CacheLocalIn[string, string]) *Count[K, V] {
	return &Count[K, V]{
		RedisCache: redisCache,
		LocalCache: localCache,
		CntOpt:     true,
		// 本地缓存
		Expiration:          time.Minute * 5, // 缓存过期时间【默认为5分钟】
		RankCacheExpiration: time.Second * 5, // 排行榜本地缓存默认5秒，确保数据新鲜度

		// 计数服务名
		ServiceTypeName: "count_service",
		Weight:          10, // 默认权重为10，计数服务一般是高频访问数据，所以权重给大一些

		LuaCnt:     LuaCnt,     // 缓存计数lua脚本，默认使用缓存计数lua脚本【可替换自定义lua】
		LuaGetRank: LuaGetRank, // 获取排名的lua脚本，默认使用获取排名的lua脚本【可替换自定义lua】
	}
}

// SetCnt 操作计数方法，调用者需自己维护redis、cache缓存过期后的回写操作
//   - biz: 业务名
//   - biz_Id: 业务id
//   - num 计数的多少，不存在num参数时，默认每次计数增加/减少1，【存在num, eg: num = 10，则每次增加计数10】
//   - 【增加或减少由CntOpt控制，true为增加，false为减少，默认为true】可调用SetCntOpt配置CntOpt增加或减少计数
func (i *Count[K, V]) SetCnt(ctx context.Context, biz string, bizId int64, num ...int64) error {
	// redis中缓存key数据
	key := i.key(biz, bizId)
	err := i.rdsCache(ctx, key, num...)
	if err != nil {
		return err
	}

	// 更新成功后，使排行榜缓存失效
	rankKey := i.rankKey(biz)
	_ = i.LocalCache.Del(rankKey)

	// 同步更新本地内存缓存中的计数值
	val, err := i.RedisCache.Get(ctx, key).Result()
	if err == nil {
		return i.LocalCache.Set(key, val, i.Expiration, i.Weight)
	}

	return nil
}

// GetCntType GetCnt 获取缓存计数数据
//   - 实现类似与可以获取前100的数据效果，那就是offset=0，limit=100为获取的条数
type GetCntType struct {
	offset int64
	limit  int64
}

// GetCnt 获取计数或排行榜
//   - biz: 业务名、biz_Id: 业务id
//   - opt: 查询选项，是查询单个服务计数还是排行榜取值范围【单个服务计数可不传opt】
//   - 【返回值[]RankItem注意】当opt为空时，返回查询的单个业务计数【返回值排名无效，为0值，忽略即可，[]RankItem[0].Rank】
func (i *Count[K, V]) GetCnt(ctx context.Context, biz string, bizId int64, opt ...GetCntType) ([]RankItem, error) {
	// 如果没有指定查询选项，返回单个业务的计数
	if len(opt) == 0 {
		if cnt, err := i.getSingleCnt(ctx, biz, bizId); err == nil {
			return []RankItem{{
				BizID: bizId,
				Rank:  0,
				Score: cnt,
			}}, nil
		}
	}

	// 获取排行榜
	return i.getRankList(ctx, biz, opt[0])
}

//// GetCnt 获取计数或排行榜
////   - biz: 业务名、biz_Id: 业务id
////   - opt: 查询选项，是查询单个服务计数还是排行榜取值范围【单个服务计数可不传opt】
//func (i *Count[K, V]) GetCnt(ctx context.Context, biz string, bizId int64, opt ...GetCntType) (interface{}, error) {
//	// 如果没有指定查询选项，返回单个业务的计数
//	if len(opt) == 0 {
//		return i.getSingleCnt(ctx, biz, bizId)
//	}
//
//	// 获取排行榜
//	return i.getRankList(ctx, biz, opt[0])
//}
