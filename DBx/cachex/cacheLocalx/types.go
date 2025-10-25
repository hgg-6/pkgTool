package cacheLocalx

import (
	"time"
)

type Key interface {
	uint64 | string | []byte | byte | int | uint | int32 | uint32 | int64
}

/*
	本地缓存：
	cost权重建议：【Set时的cost权重配置建议】【一般用于缓存过期时间没到，但是内存阈值到了，根据weight权重释放一些本地缓存】
	【场景:】							【推荐 cost 设置:】
	缓存 []byte、string					cost = len(value)
	缓存结构体							估算大小，或统一设为 1（按条数）
	缓存不同重要性数据						按业务权重设置 cost（VIP=10, 普通=1）
	不确定大小							固定 cost=1，靠 MaxCost 控制总条数
*/

// CacheLocalIn 抽象缓存接口
type CacheLocalIn[K Key, V any] interface {
	Set(key K, value V, ttl time.Duration, weight int64) error
	Get(key K) (V, error)
	Del(key K) error

	// WaitSet 等待值通过缓冲区【除非有重要的缓存，实时性要求特别较高，要堵塞直至等待缓冲写入通过，否则不用管 也就毫秒纳秒级甚至还不到】
	WaitSet()
	// Close 关闭会停止所有goroutines并关闭所有频道。【ristretto实现一定记着】 defer cache.Close()
	Close()
}
