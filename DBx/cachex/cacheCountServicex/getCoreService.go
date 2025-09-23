package cacheCountServicex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/convertx/toanyx"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
)

/*
	get核心业务逻辑
*/

// GetCntType GetCnt 获取缓存计数数据
//   - 实现类似与可以获取前100的数据效果，那就是offset=1，Limit=100为获取的条数，开区间[]
type GetCntType struct {
	Offset int64
	Limit  int64
}

// GetCnt 获取计数或排行榜【获取排行榜，调用者如果SetCnt只获取了err，建议一分钟调用一次getCnt获取榜单自行计算存储数据库，因为缓存计数数据只缓存5分钟】
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
		return nil, fmt.Errorf("获取计数数据失败, biz: %s, biz_id: %d", biz, bizId)
	}

	// 获取排行榜
	rankItem, err := i.getRankList(ctx, biz, opt[0])
	if err != nil || len(rankItem) <= 0 {
		return nil, fmt.Errorf("获取biz: %s, biz_id: %d 排行榜数据失败 / 未获取到排行榜【%d--%d】的数据, err: %v", biz, bizId, opt[0].Offset, opt[0].Limit, err)
	}
	return rankItem, nil
}

// getCnt 获取计数或排行榜
//   - biz: 业务名、biz_Id: 业务id
//   - opt: 查询选项，是查询单个服务计数还是排行榜取值范围【单个服务计数可不传opt】
//   - 【返回值[]RankItem注意】当opt为空时，返回查询的单个业务计数【返回值排名无效，为0值，忽略即可，[]RankItem[0].Rank】
func (i *Count[K, V]) getCnt(ctx context.Context, biz string, bizId int64, opt ...GetCntType) ([]RankItem, error) {
	// 如果没有指定查询选项，返回单个业务的计数
	if len(opt) == 0 {
		if cnt, err := i.getSingleCnt(ctx, biz, bizId); err == nil {
			return []RankItem{{
				BizID: bizId,
				Rank:  0,
				Score: cnt,
			}}, nil
		}
		return nil, fmt.Errorf("获取计数数据失败, biz: %s, biz_id: %d", biz, bizId)
	}

	// 获取排行榜
	rankItem, err := i.getRankList(ctx, biz, opt[0])
	if err != nil || len(rankItem) <= 0 {
		return nil, fmt.Errorf("获取biz: %s, biz_id: %d 排行榜数据失败 / 排行榜【%d--%d】的数据为空, err: %v", biz, bizId, opt[0].Offset, opt[0].Limit, err)
	}
	return rankItem, nil
}

// getRankList 获取排行榜
func (i *Count[K, V]) getRankList(ctx context.Context, biz string, opt GetCntType) ([]RankItem, error) {
	rankKey := i.RankKey(biz)

	// 先尝试从本地缓存获取排行榜
	if cacheData, err := i.LocalCache.Get(rankKey); err == nil {
		var rankList []RankItem
		if err := json.Unmarshal([]byte(cacheData), &rankList); err == nil {
			// 返回请求的范围
			start, end := i.calculateRange(opt, len(rankList))
			return rankList[start:end], nil
		}
	}
	// 如果本地缓存没有，尝试从Redis获取
	//if cacheData, err := i.RedisCache.Get(ctx, RankKey).Result(); err == nil {
	//	var rankList []RankItem
	//	if err := json.Unmarshal([]byte(cacheData), &rankList); err == nil {
	//		// 返回请求的范围
	//		start, end := i.calculateRange(opt, len(rankList))
	//		return rankList[start:end], nil
	//	}
	//}

	// 本地缓存没有或解析失败，从Redis重新计算获取榜单
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
		//err = i.RedisCache.Set(ctx, RankKey, data, i.RankCacheExpiration).Err()
		//if err != nil {
		//	return nil, err
		//}
	}

	// 返回请求的范围
	start, end := i.calculateRange(opt, len(rankList))
	return rankList[start:end], nil
}

// calculateRange 计算返回的范围
//   - opt.Offset: 起始索引，负数表示从末尾开始计算
//   - opt.Limit: 获取的记录数
//   - total: 排行榜的总记录数
func (i *Count[K, V]) calculateRange(opt GetCntType, total int) (int, int) {
	start := int(opt.Offset)
	//if start < 0 {
	//	start = 0
	//}
	if start > total {
		start = total
	}

	end := opt.Offset + opt.Limit
	if int(end) > total {
		end = int64(total)
	}

	return start, int(end)
}

// getRankFromRedis 从Redis获取排行榜
func (i *Count[K, V]) getRankFromRedis(ctx context.Context, biz string, opt GetCntType) ([]RankItem, error) {
	// 使用Lua脚本原子性地获取排名和分数
	keys := []string{i.RankKey(biz)}
	args := []interface{}{opt.Offset, opt.Offset + opt.Limit - 1}

	result, err := i.RedisCache.Eval(ctx, i.LuaGetRank, keys, args...).Result()
	if err != nil {
		return nil, err
	}

	// 解析Lua脚本返回的结果
	items, ok := result.([]interface{})
	if !ok {
		return nil, errors.New("invalid result type from Lua script，Lua脚本的结果类型无效")
	}
	if len(items)%2 != 0 {
		return nil, errors.New("invalid items length from ZREVRANGE WITHSCORES")
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

		realRank := opt.Offset + int64(j/2) + 1
		rankList = append(rankList, RankItem{
			BizID: bizId,
			Score: score,
			Rank:  realRank,
		})
	}
	return rankList, nil
}

// getSingleCnt 获取单个业务的计数
func (i *Count[K, V]) getSingleCnt(ctx context.Context, biz string, bizId int64) (int64, error) {
	key := i.Key(biz, bizId)

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
			return 0, errors.New("查询redis数据，键不存在") // 键不存在
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

// rankKeyFromCntKey 从计数键提取排行榜键
func (i *Count[K, V]) rankKeyFromCntKeyV1(cntKey string) string {
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
		//return cntKey + ":rank"
		return "rank_cnt:" + cntKey + ":rank"
	}

	//return cntKey[:lastColon] + ":rank"
	return "rank_cnt:" + cntKey[:lastColon] + ":rank"
}
func (i *Count[K, V]) rankKeyFromCntKey(cntKey string) string {
	// 目标：从cntKey中提取 ServiceTypeName 和 biz（cntKey格式：cnt:ServiceTypeName:biz:bizId）
	// 步骤1：按冒号分割字符串，得到各字段切片
	parts := strings.Split(cntKey, ":")

	// 步骤2：校验格式有效性（至少需要4个字段才能提取ServiceTypeName和biz）
	if len(parts) < 4 {
		// 格式不符合预期时，回退原有兼容逻辑（避免panic）
		return "rank_cnt:" + cntKey + ":rank"
	}

	// 步骤3：提取目标字段（parts[1]=ServiceTypeName，parts[2]=biz）
	serviceTypeName := parts[1]
	biz := parts[2]

	// 步骤4：构造最终的rankKey（格式：rank_cnt:ServiceTypeName:biz:rank）
	return fmt.Sprintf("rank_cnt:%s:%s:rank", serviceTypeName, biz)
}
