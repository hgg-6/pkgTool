package cacheLocalRistrettox

import (
	"errors"
	"github.com/hgg-6/pkgTool/v2/DBx/localCahceX"
	"github.com/dgraph-io/ristretto/v2"
	"time"
)

type CacheLocalRistrettoStr[K localCahceX.Key, V any] struct {
	cache *ristretto.Cache[K, V]
}

// NewCacheLocalRistrettoStr 是高性能、并发安全、带准入策略的内存缓存库【初始化参考测试用例 V1 版本】
func NewCacheLocalRistrettoStr[K localCahceX.Key, V any](cache *ristretto.Cache[K, V]) localCahceX.CacheLocalIn[K, V] {
	return &CacheLocalRistrettoStr[K, V]{
		cache: cache,
	}
}

// Set 设置本地缓存
//   - weight: 缓存权重
func (c *CacheLocalRistrettoStr[K, V]) Set(key K, value V, ttl time.Duration, weight int64) error {
	ok := c.cache.SetWithTTL(key, value, weight, ttl)
	if ok {
		return nil
	}
	return errors.New("set localCache fail error")
}

func (c *CacheLocalRistrettoStr[K, V]) Get(key K) (V, error) {
	// 先用 Get 判定 key 是否存在（GetTTL 对永久条目返回 (0, true)，
	// 仅用 TTL 判定会把所有无 TTL 的永久缓存误判为不存在）。
	value, ok := c.cache.Get(key)
	if !ok {
		var v V
		return v, errors.New("查询缓存失败")
	}
	return value, nil
}

func (c *CacheLocalRistrettoStr[K, V]) Del(key K) error {
	c.cache.Del(key)
	return nil
}

func (c *CacheLocalRistrettoStr[K, V]) Close() {
	c.cache.Close()
}

func (c *CacheLocalRistrettoStr[K, V]) WaitSet() {
	c.cache.Wait()
}

func (c *CacheLocalRistrettoStr[K, V]) initClose() {
	defer c.cache.Close()
}
