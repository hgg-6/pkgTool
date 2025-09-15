package cacheCountServicex

import (
	"context"
	"encoding/json"
	"errors"
	"gitee.com/hgg_test/pkg_tool/v2/convertx/toanyx"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
)

/*
	此包为核心业务逻辑
*/

// getRankList 获取排行榜
func (i *Count[K, V]) getRankList(ctx context.Context, biz string, opt GetCntType) ([]RankItem, error) {
	rankKey := i.rankKey(biz)

	// 先尝试从本地缓存获取排行榜
	if cacheData, err := i.LocalCache.Get(rankKey); err == nil {
		var rankList []RankItem
		if err := json.Unmarshal([]byte(cacheData), &rankList); err == nil {
			// 返回请求的范围
			start, end := i.calculateRange(opt, len(rankList))
			return rankList[start:end], nil
		}
	}

	// 本地缓存没有或解析失败，从Redis获取
	rankList, err := i.getRankFromRedis(ctx, biz, opt)
	if err != nil {
		return nil, err
	}

	// 缓存到本地
	if data, err := json.Marshal(rankList); err == nil {
		err = i.LocalCache.Set(rankKey, string(data), i.RankCacheExpiration, i.Weight)
		if err != nil {
			return nil, err
		}
	}

	// 返回请求的范围
	start, end := i.calculateRange(opt, len(rankList))
	return rankList[start:end], nil
}

// calculateRange 计算返回的范围
//   - opt.offset: 起始索引，负数表示从末尾开始计算
//   - opt.limit: 获取的记录数
//   - total: 排行榜的总记录数
func (i *Count[K, V]) calculateRange(opt GetCntType, total int) (int, int) {
	start := int(opt.offset)
	if start < 0 {
		start = 0
	}
	if start >= total {
		start = total
	}

	end := start + int(opt.limit)
	if end > total {
		end = total
	}

	return start, end
}

// getRankFromRedis 从Redis获取排行榜
func (i *Count[K, V]) getRankFromRedis(ctx context.Context, biz string, opt GetCntType) ([]RankItem, error) {
	// 使用Lua脚本原子性地获取排名和分数
	keys := []string{i.rankKey(biz)}
	args := []interface{}{opt.offset, opt.offset + opt.limit - 1}

	result, err := i.RedisCache.Eval(ctx, i.LuaGetRank, keys, args...).Result()
	if err != nil {
		return nil, err
	}

	// 解析Lua脚本返回的结果
	items, ok := result.([]interface{})
	if !ok {
		return nil, errors.New("invalid result type from Lua script，Lua脚本的结果类型无效")
	}

	rankList := make([]RankItem, 0, len(items)/2)
	for j := 0; j < len(items); j += 2 {
		bizIdStr, ok1 := items[j].(string)
		scoreStr, ok2 := items[j+1].(string)
		if !ok1 || !ok2 {
			continue
		}

		bizId, err1 := strconv.ParseInt(bizIdStr, 10, 64)
		score, err2 := strconv.ParseInt(scoreStr, 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}

		rankList = append(rankList, RankItem{
			BizID: bizId,
			Score: score,
			Rank:  int64(len(rankList) + 1),
		})
	}

	return rankList, nil
}

// getSingleCnt 获取单个业务的计数
func (i *Count[K, V]) getSingleCnt(ctx context.Context, biz string, bizId int64) (int64, error) {
	key := i.key(biz, bizId)

	// 先尝试从本地缓存获取
	if val, err := i.LocalCache.Get(key); err == nil {
		//if cnt, err := strconv.ParseInt(val, 10, 64); err == nil {
		//	return cnt, nil
		//}
		if cnt, ok := toanyx.ToAny[int64](val); ok {
			return cnt, nil
		}
	}

	// 本地缓存没有，从Redis获取
	val, err := i.RedisCache.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil // 键不存在，返回0
		}
		return 0, err
	}

	// 解析并缓存到本地
	//cnt, err := strconv.ParseInt(val, 10, 64)
	cnt, ok := toanyx.ToAny[int64](val)
	if !ok {
		return 0, errors.New("val[string] --> cnt[int64]解析转换错误")
	}

	err = i.LocalCache.Set(key, val, i.Expiration, i.Weight)
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

// rdsCache 更新Redis缓存
func (i *Count[K, V]) rdsCache(ctx context.Context, key string, num ...int64) error {
	// 存在num参数时，使用num参数计数增加/减少指定值
	delta := int64(1)
	if len(num) > 0 {
		delta = num[0]
	}

	if !i.CntOpt {
		delta = -delta
	}

	// 使用Lua脚本原子性地更新计数和排行榜
	rankKey := i.rankKeyFromCntKey(key)
	//return i.RedisCache.Eval(ctx, i.LuaCnt, []string{key, rankKey}, delta).Err()
	// 【固定lua脚本】先尝试使用 EVALSHA 提升性能
	err := i.RedisCache.EvalSha(ctx, i.LuaCnt, []string{key, rankKey}, delta).Err()
	if err != nil && strings.Contains(err.Error(), "NOSCRIPT") {
		// 如果脚本未加载，使用 EVAL 并重新加载脚本
		err = i.RedisCache.Eval(ctx, i.LuaCnt, []string{key, rankKey}, delta).Err()
		if err == nil {
			// 重新加载脚本
			i.initLuaCntScripts(ctx)
		}
		return err
	}
	return err
}

// rankKeyFromCntKey 从计数键提取排行榜键
func (i *Count[K, V]) rankKeyFromCntKey(cntKey string) string {
	// 解析计数键，提取业务名
	// 格式: ServiceTypeName:biz:bizId
	// 我们要提取ServiceTypeName和biz部分
	lastColon := -1
	prevColon := -1
	for idx, char := range cntKey {
		if char == ':' {
			prevColon = lastColon
			lastColon = idx
		}
	}

	if prevColon == -1 {
		return cntKey + ":rank"
	}

	return cntKey[:lastColon] + ":rank"
}

// initLuaCntScripts 在初始化时预加载脚本并获取 SHA
func (i *Count[K, V]) initLuaCntScripts(ctx context.Context) error {
	sha, err := i.RedisCache.ScriptLoad(ctx, i.LuaCnt).Result()
	if err != nil {
		return err
	}
	i.LuaCnt = sha
	return nil
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
