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

//func NewCacheLocalRistrettoStrV1[K cachex.Key, V any](cache *ristretto.Cache[K, V]) *CacheLocalRistrettoStr[K, V] {
//	return &CacheLocalRistrettoStr[K, V]{cache: cache}
//}

func (c *CacheLocalRistrettoStr[K, V]) Set(key K, value V, ttl time.Duration, weight ...int64) error {
	if len(weight) <= 0 {
		return errors.New("no weight, set localCache need -weight-, 没有权重信息，Ristretto设置本地缓存需要权重信息")
	}
	ok := c.cache.SetWithTTL(key, value, weight[0], ttl)
	if ok {
		return nil
	}
	return errors.New("set localCache fail error")
}

func (c *CacheLocalRistrettoStr[K, V]) Get(key K) (V, error) {
	var v V
	val, ok := c.cache.GetTTL(key)
	if !ok {
		return v, errors.New("get localCache error, no key --> value, 查询缓存失败, Key不存在")
	} else if val <= time.Duration(0) {
		c.cache.Del(key)
		return v, errors.New("get localCache error, key --> value is expired, 查询缓存失败, 缓存已过期")
	}
	if value, isok := c.cache.Get(key); isok {
		return value, nil
	}
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
