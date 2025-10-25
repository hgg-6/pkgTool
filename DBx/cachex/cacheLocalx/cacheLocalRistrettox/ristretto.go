package cacheLocalRistrettox

import (
	"errors"
	"gitee.com/hgg_test/pkg_tool/v2/DBx/cachex/cacheLocalx"
	"github.com/dgraph-io/ristretto/v2"
	"time"
)

type CacheLocalRistrettoStr[K cacheLocalx.Key, V any] struct {
	cache *ristretto.Cache[K, V]
}

// NewCacheLocalRistrettoStr 是高性能、并发安全、带准入策略的内存缓存库【初始化参考测试用例 V1 版本】
func NewCacheLocalRistrettoStr[K cacheLocalx.Key, V any](cache *ristretto.Cache[K, V]) cacheLocalx.CacheLocalIn[K, V] {
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
	val, ok := c.cache.GetTTL(key)
	if !ok || val <= time.Duration(0) {
		var v V
		return v, errors.New("查询缓存失败")
	}
	if value, isok := c.cache.Get(key); isok {
		return value, nil
	}
	var v V
	return v, errors.New("get localCache error, no key --> value, 查询缓存失败, Key不存在")
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

// init方法会被自动调用
func (c *CacheLocalRistrettoStr[K, V]) initClose() {
	defer c.cache.Close()
}
