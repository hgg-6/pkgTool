package queueX

import (
	"container/heap"
	"sync"
)

// PriorityQueue 泛型优先级队列（最小堆/最大堆由less函数决定）
type PriorityQueue[T comparable] struct {
	items []T               // 堆元素存储
	less  func(a, b T) bool // 堆排序规则（a < b 则返回true，对应最小堆）
	lock  sync.Mutex        // 线程安全锁
}

// NewPriorityQueue 创建优先级队列（传入比较函数）
// 示例：time.Time的最小堆 → func(a, b time.Time) bool { return a.Before(b) }
func NewPriorityQueue[T comparable](less func(a, b T) bool) *PriorityQueue[T] {
	pq := &PriorityQueue[T]{items: make([]T, 0), less: less}
	heap.Init(pq) // 初始化堆结构
	return pq
}

// Len
//   - 返回队列长度
func (pq *PriorityQueue[T]) Len() int { return len(pq.items) }

// Less 堆排序规则
func (pq *PriorityQueue[T]) Less(i, j int) bool { return pq.less(pq.items[i], pq.items[j]) }

// Swap 交换元素
func (pq *PriorityQueue[T]) Swap(i, j int) { pq.items[i], pq.items[j] = pq.items[j], pq.items[i] }

// Push 入队（线程安全）
func (pq *PriorityQueue[T]) Push(x any) {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	item := x.(T)
	pq.items = append(pq.items, item)
	heap.Fix(pq, len(pq.items)-1) // 插入后调整堆
}

// Pop 出队（线程安全，返回堆顶元素）
func (pq *PriorityQueue[T]) Pop() any {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	old := pq.items
	n := len(old)
	item := old[n-1]
	pq.items = old[0 : n-1]
	heap.Init(pq) // 弹出后调整堆
	return item
}

// Peek 查看堆顶元素（不删除，线程安全）
func (pq *PriorityQueue[T]) Peek() (T, bool) {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	if pq.Len() == 0 {
		var zero T
		return zero, false
	}
	return pq.items[0], true
}

// Remove 删除指定元素（线程安全，依赖T的comparable约束）
func (pq *PriorityQueue[T]) Remove(x T) bool {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	for i, item := range pq.items {
		if item == x {
			heap.Remove(pq, i) // 调用heap包删除并调整堆
			return true
		}
	}
	return false
}
