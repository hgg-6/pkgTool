package cacheLocalRistrettox

import (
	"github.com/dgraph-io/ristretto/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

/*
	✅ 最佳实践建议：【Set时的cost配置建议】
	【场景:】							【推荐 cost 设置:】
	缓存 []byte、string					cost = len(value)
	缓存结构体							估算大小，或统一设为 1（按条数）
	缓存不同重要性数据						按业务权重设置 cost（VIP=10, 普通=1）
	不确定大小							固定 cost=1，靠 MaxCost 控制总条数
*/

const (
	VipUserCost = 10
	UserCost    = 1
)

/*
	 “位左移运算符”
	1. 表示 2 的幂，语义清晰
	1 << 10 → 1024 			→ 1KB
	1 << 20 → 1,048,576 	→ 1MB
	1 << 30 → 1,073,741,824 → 1GB
	1 << 40 → 1TB
*/

func TestRistretto(t *testing.T) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: 1e7,     // 按键跟踪次数为（10M）。
		MaxCost:     1 << 30, // 最大缓存成本（1GB）“位左移运算符”。
		BufferItems: 64,      // 每个Get缓冲区的键数。
	})
	assert.NoError(t, err)
	defer cache.Close()

	// 设置一个成本为1的值
	//cache.Set("key", "value", 1)
	cache.SetWithTTL("key", "value", 1, time.Second*5)

	// 等待值通过缓冲区【除非有重要的缓存，实时性要求较高，要堵塞直至等待缓冲通过】
	cache.Wait()

	// 从缓存中获取值
	value, ok := cache.Get("key")
	if !ok {
		t.Log("missing value")
	}
	t.Log("local cache: ", value)

	time.Sleep(time.Second * 1)
	v, ok := cache.GetTTL("key1")
	if !ok {
		t.Log("missing value")
	}
	t.Log("local cache v: ", v)
	if v <= 0 {
		t.Log("value TTL is no duration")
	} else {
		value, _ = cache.Get("key")
		t.Log("local cache: ", value)
	}

	// 缓存中的del值
	cache.Del("key")
}

func TestRistrettoV1(t *testing.T) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,     // 按键跟踪次数为（10M）。
		MaxCost:     1 << 30, // 最大缓存成本（1GB）“位左移运算符”。
		BufferItems: 64,      // 每个Get缓冲区的键数。
	})
	//defer cache.Close()

	assert.NoError(t, err)
	ca := NewCacheLocalRistrettoStr[string, any](cache)
	defer ca.Close()

	// 打印时间单位为微秒
	t.Log("set cache: ", time.Now().UnixMicro())
	err = ca.Set("key", "value", time.Second*5, VipUserCost)
	// 等待值通过缓冲区
	ca.WaitSet()
	//time.Sleep(time.Second * 1)

	assert.NoError(t, err)
	t.Log("set cache ok: ", time.Now().UnixMicro())
	t.Log("get cache: ", time.Now().UnixMicro())
	val, err := ca.Get("key")
	assert.NoError(t, err)
	t.Log("val: ", val)
	t.Log("get cache ok: ", time.Now().UnixMicro())
	ca.Del("key")
}
