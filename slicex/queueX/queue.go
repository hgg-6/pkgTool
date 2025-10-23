package queueX

import (
	"container/heap"
	"sync"
)

type PriorityQueue[T any] struct {
	items []T
	less  func(a, b T) bool
	lock  sync.Mutex
}

// NewPriorityQueue 创建优先级队列
func NewPriorityQueue[T any](less func(a, b T) bool) *PriorityQueue[T] {
	pq := &PriorityQueue[T]{items: make([]T, 0), less: less}
	heap.Init(pq)
	return pq
}

func (pq *PriorityQueue[T]) Len() int { return len(pq.items) }

func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return pq.less(pq.items[i], pq.items[j])
}

func (pq *PriorityQueue[T]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Push 供 heap.Push 调用，只负责追加元素（不加锁！）
func (pq *PriorityQueue[T]) Push(x any) {
	pq.items = append(pq.items, x.(T))
}

// Pop 供 heap.Pop 调用，只负责弹出最后一个元素（不加锁！）
func (pq *PriorityQueue[T]) Pop() any {
	old := pq.items
	n := len(old)
	item := old[n-1]
	pq.items = old[0 : n-1]
	return item
}

// --- 公共线程安全 API 尽量保证调用下API---

// Enqueue 入队（线程安全）
func (pq *PriorityQueue[T]) Enqueue(item T) {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	heap.Push(pq, item) // 会调用上面的 Push + 调整堆
}

// Dequeue 出队（线程安全），返回堆顶元素
func (pq *PriorityQueue[T]) Dequeue() (T, bool) {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	if len(pq.items) == 0 {
		var zero T
		return zero, false
	}
	item := heap.Pop(pq) // 会调用上面的 Pop + 调整堆
	return item.(T), true
}

// Peek 查看堆顶（线程安全）
func (pq *PriorityQueue[T]) Peek() (T, bool) {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	if len(pq.items) == 0 {
		var zero T
		return zero, false
	}
	return pq.items[0], true
}

// RemoveIf 删除指定元素（线程安全）传入一个“等于”函数
func (pq *PriorityQueue[T]) RemoveIf(predicate func(T) bool) bool {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	for i, item := range pq.items {
		if predicate(item) {
			heap.Remove(pq, i)
			return true
		}
	}
	return false
}
