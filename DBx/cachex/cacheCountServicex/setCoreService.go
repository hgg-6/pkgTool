package cacheCountServicex

import (
	"context"
	"errors"
	"strings"
	"time"
)

/*
	set核心业务逻辑
*/

// SetCnt 操作计数方法，【biz不支持中间 : ，可以用 _ 替代，user_article而非user:article】
//   - 【1、SetCnt增加计数时，接收err，2、SetCnt增加计数时，直接接收err和res榜单，榜单每分钟更新】
//   - 如果开启了RankCount=true，实时计算热榜数据，调用者需自己维护redis、cache缓存过期后的回写操作 以及 热榜数据的数据库更新
//   - biz: 业务名【biz文章article服务 的 biz_id文章1】
//   - biz_Id: 业务id
//   - num 计数的多少，不存在num参数时，默认每次计数增加/减少1，【存在num, eg: num = 10，则每次增加计数10】
//   - 【增加或减少由CntOpt控制，true为增加，false为减少，默认为true】可调用SetCntOpt配置CntOpt增加或减少计数
func (i *Count[K, V]) SetCnt(ctx context.Context, biz string, bizId int64, num ...int64) *Count[K, V] {
	i.Error = nil
	switch i.RankCount {
	case true:
		if time.Now().After(time.UnixMilli(i.targetTime)) {
			// 过了一分钟，重置计数服务时间为1分钟后
			i.targetTime = time.Now().Add(time.Minute).UnixMilli()
			// 更新缓存数据
			er := i.setCnt(ctx, biz, bizId, num...).Error
			if er != nil {
				i.Error = er
				return i
			}
			// 过了一分钟，更新每分钟热榜重新计算
			rank, er := i.getCnt(ctx, biz, bizId, i.CntTypeConf)
			i.Rank = rank
			if er != nil {
				i.Error = er
				return i
			}
			return i
		} else {
			if len(i.Rank) == 0 {
				// 重置计数服务时间为1分钟后
				i.targetTime = time.Now().Add(time.Minute).UnixMilli()
				// 系统第一次计数，需要计算热榜
				rank, er := i.GetCnt(ctx, biz, bizId, i.CntTypeConf)
				i.Rank = rank
				if er != nil {
					i.Error = er
				}
			}
			// 没到一分钟，不计算热榜，正常更新缓存数据
			return i.setCnt(ctx, biz, bizId, num...)
		}
	case false:
		return i.setCnt(ctx, biz, bizId, num...)
	default:
		i.Error = errors.New("请先调用SetRankCount方法配置是否开启实时计算热榜数据")
		return i
	}
}

func (i *Count[K, V]) setCnt(ctx context.Context, biz string, bizId int64, num ...int64) *Count[K, V] {
	i.Error = nil

	// redis中缓存key数据
	key := i.key(biz, bizId)
	err := i.rdsCache(ctx, key, num...)
	if err != nil {
		i.Error = err
		return i
	}

	// 更新成功后，使排行榜本地缓存失效
	rankKey := i.rankKey(biz)
	_ = i.LocalCache.Del(rankKey)

	// 同步更新本地内存缓存中的计数值
	val, err := i.RedisCache.Get(ctx, key).Result()
	if err == nil {
		i.Error = i.LocalCache.Set(key, val, i.Expiration, i.Weight)
		return i
	}
	return i
}

// rdsCache 更新Redis缓存
func (i *Count[K, V]) rdsCache(ctx context.Context, key string, num ...int64) error {
	delta := int64(1)
	if len(num) > 0 {
		delta = num[0]
	}
	if !i.CntOpt {
		delta = -delta
	}

	rankKey := i.rankKeyFromCntKey(key)
	member := i.memberFromCntKey(key)

	err := i.tryEvalLua(ctx, key, rankKey, delta, member)
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "NOSCRIPT") {
		if loadErr := i.initLuaCntScripts(ctx); loadErr != nil {
			return loadErr
		}
		return i.tryEvalLua(ctx, key, rankKey, delta, member) // 重试
	}

	return err
}

func (i *Count[K, V]) memberFromCntKey(key string) string {
	parts := strings.Split(key, ":")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return key
}

func (i *Count[K, V]) tryEvalLua(ctx context.Context, key, rankKey string, delta int64, member string) error {
	return i.RedisCache.EvalSha(ctx, i.LuaCnt, []string{key, rankKey}, delta, member).Err()
}

// ===============================
// 每分钟调用一次，滑动窗口
var sliCtx = context.Background()

func (i *Count[K, V]) slideWindow(key string, newValue []RankItem) error {
	// 1. 左侧插入新值
	err := i.RedisCache.LPush(sliCtx, key, newValue).Err()
	if err != nil {
		return err
	}

	// 2. 保留最新的5个元素（0~4），其余截断
	err = i.RedisCache.LTrim(sliCtx, key, 0, 4).Err()
	if err != nil {
		return err
	}

	return nil
}
