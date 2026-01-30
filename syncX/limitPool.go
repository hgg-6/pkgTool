// Package syncX 提供并发安全的数据结构，是对标准库 sync 的扩展。
package syncX

import (
	"sync/atomic"
)

// LimitPool 是对 Pool 的简单封装允许用户通过控制一段时间内对Pool的令牌申请次数来间接控制Pool中对象的内存总占用量
type LimitPool[T any] struct {
	pool      *Pool[T]
	used      *atomic.Int32 // 当前已借出的对象数量
	maxTokens int32         // 最大允许借出的对象数量
}

// NewLimitPool 创建一个 LimitPool 实例
// maxTokens 表示一段时间内的允许发放的最大令牌数
// factory 必须返回 T 类型的值，并且不能返回 nil
func NewLimitPool[T any](maxTokens int, factory func() T) *LimitPool[T] {
	if maxTokens < 0 {
		maxTokens = 0
	}
	var used atomic.Int32
	return &LimitPool[T]{
		pool:      NewPool[T](factory),
		used:      &used,
		maxTokens: int32(maxTokens),
	}
}

// Get 取出一个元素
// 如果返回值是 true，则代表成功借出对象（可能来自池或新建）
// 如果返回 false，表示已达到最大限制，无法借出更多对象
func (l *LimitPool[T]) Get() (T, bool) {
	// 使用循环来避免竞态条件
	for {
		used := l.used.Load()
		if used >= l.maxTokens {
			var zero T
			return zero, false
		}
		// 尝试原子地增加已使用计数
		if l.used.CompareAndSwap(used, used+1) {
			// 成功获取令牌，从池中获取对象
			return l.pool.Get(), true
		}
		// 如果CAS失败，说明有其他goroutine修改了计数，重试
	}
}

// Put 放回去一个元素
func (l *LimitPool[T]) Put(t T) {
	l.pool.Put(t)
	// 减少已使用计数，确保计数不小于0
	for {
		used := l.used.Load()
		if used <= 0 {
			// 计数已经为0或负数，不再减少
			return
		}
		if l.used.CompareAndSwap(used, used-1) {
			return
		}
		// CAS失败，重试
	}
}
