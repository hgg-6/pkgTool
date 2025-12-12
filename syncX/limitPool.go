package syncX

import (
	"sync/atomic"
)

// LimitPool 是对 Pool 的简单封装允许用户通过控制一段时间内对Pool的令牌申请次数来间接控制Pool中对象的内存总占用量
type LimitPool[T any] struct {
	pool   *Pool[T]
	tokens *atomic.Int32
}

// NewLimitPool 创建一个 LimitPool 实例
// maxTokens 表示一段时间内的允许发放的最大令牌数
// factory 必须返回 T 类型的值，并且不能返回 nil
func NewLimitPool[T any](maxTokens int, factory func() T) *LimitPool[T] {
	var tokens atomic.Int32
	tokens.Add(int32(maxTokens))
	return &LimitPool[T]{
		pool:   NewPool[T](factory),
		tokens: &tokens,
	}
}

// Get 取出一个元素
// 如果返回值是 true，则代表确实从 Pool 里面取出来了一个
// 否则是新建了一个
func (l *LimitPool[T]) Get() (T, bool) {
	// 使用循环来避免竞态条件
	for {
		current := l.tokens.Load()
		if current <= 0 {
			var zero T
			return zero, false
		}
		// 尝试原子地减少令牌计数
		if l.tokens.CompareAndSwap(current, current-1) {
			// 成功获取令牌，从池中获取对象
			return l.pool.Get(), true
		}
		// 如果CAS失败，说明有其他goroutine修改了令牌计数，重试
	}
}

// Put 放回去一个元素
func (l *LimitPool[T]) Put(t T) {
	l.pool.Put(t)
	l.tokens.Add(1)
}
