package cacheCountServiceX

import (
	"fmt"
	"time"
)

// Key 生成缓存键
func (i *Count[K, V]) Key(biz string, bizId int64) string {
	//return fmt.Sprintf("%s:%s:%d", i.ServiceTypeName, biz, bizId)
	return fmt.Sprintf("cnt:%s:%s:%d", i.ServiceTypeName, biz, bizId)
}

// RankKey 生成排行榜键
func (i *Count[K, V]) RankKey(biz string) string {
	//return fmt.Sprintf("%s:%s:rank", i.ServiceTypeName, biz)
	return fmt.Sprintf("rank_cnt:%s:%s:rank", i.ServiceTypeName, biz)
}

// SetCntOpt : true为增加，false为减少【默认为增加】
func (i *Count[K, V]) SetCntOpt(ctp bool) *Count[K, V] {
	i.CntOpt = ctp
	return i
}

// SetExpiration : 设置缓存过期时间【默认为5分钟】
func (i *Count[K, V]) SetExpiration(expiration time.Duration) *Count[K, V] {
	i.Expiration = expiration
	return i
}

// SetServiceTypeName : 设置服务名eg: like_cnt 【默认为count_service】
func (i *Count[K, V]) SetServiceTypeName(ServiceTypeName string) *Count[K, V] {
	i.ServiceTypeName = ServiceTypeName
	return i
}

// SetWeight : 设置本地缓存中数据权重【多用于缓存时间未过期，但是分配内存满了，需释放部分】
func (i *Count[K, V]) SetWeight(weight int64) *Count[K, V] {
	i.Weight = weight
	return i
}

// SetRankCacheExpiration 设置排行榜缓存过期时间
func (i *Count[K, V]) SetRankCacheExpiration(expiration time.Duration) *Count[K, V] {
	i.RankCacheExpiration = expiration
	return i
}

// SetLuaCnt : 设置Lua脚本，用于增减计数的 Lua 脚本
func (i *Count[K, V]) SetLuaCnt(LuaCnt string) *Count[K, V] {
	i.LuaCnt = LuaCnt
	// 如果需要，这里可以自动调用 SCRIPT LOAD 并设置 incrScriptSha
	return i
}

// SetGetLuaGetRank : 设置排行榜的 Lua 脚本, 用于获取排行榜的 Lua 脚本
func (i *Count[K, V]) SetGetLuaGetRank(LuaGetRank string) *Count[K, V] {
	i.LuaGetRank = LuaGetRank
	// 如果需要，这里可以自动调用 SCRIPT LOAD 并设置 getRankScriptSha
	return i
}

// SetRankCount : 设置是否统计排行榜数据
func (i *Count[K, V]) SetRankCount(rankCount bool) *Count[K, V] {
	i.RankCount = rankCount
	return i
}

// SetCntTypeConf : 设置获取排行榜数据时的参数
func (i *Count[K, V]) SetCntTypeConf(setCntTypeConf GetCntType) *Count[K, V] {
	i.CntTypeConf = setCntTypeConf
	return i
}
