package syncX

import "sync"

// Pool 是对 sync.Pool 的简单封装
// 会有一些性能损耗，但是基本可以忽略不计。担忧性能问题的可以参考
type Pool[T any] struct {
	p sync.Pool
}

// NewPool 创建一个 Pool 实例
// factory 必须返回 T 类型的值，并且不能返回 nil
func NewPool[T any](factory func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

// Get 取出一个元素
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

// Put 放回去一个元素
func (p *Pool[T]) Put(t T) {
	p.p.Put(t)
}
