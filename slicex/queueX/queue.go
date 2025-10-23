package queueX

import (
	"container/heap"
	"sync"
)

// PriorityQueue 泛型优先队列
type PriorityQueue[T any] struct {
	items    []T
	less     func(a, b T) bool
	capacity int // <=0 表示无界
	lock     sync.Mutex
}

// NewPriorityQueue 创建优先队列【支持批量入队】
// capacity <= 0 表示无界队列；>0 表示有界队列（最大容量为 capacity）
func NewPriorityQueue[T any](less func(a, b T) bool, capacity int) *PriorityQueue[T] {
	pq := &PriorityQueue[T]{
		items:    make([]T, 0, maxx(0, capacity)),
		less:     less,
		capacity: capacity,
	}
	heap.Init(pq)
	return pq
}

// ---------- heap.Interface（不加锁）----------

// Len 返回当前元素数量
//   - 【未加锁，尽量使用下面加锁API保证线程安全】
func (pq *PriorityQueue[T]) Len() int { return len(pq.items) }

// Less 比较两个元素
//   - 【未加锁，尽量使用下面加锁API保证线程安全】
func (pq *PriorityQueue[T]) Less(i, j int) bool {
	return pq.less(pq.items[i], pq.items[j])
}

// Swap 交换两个元素
//   - 【未加锁，尽量使用下面加锁API保证线程安全】
func (pq *PriorityQueue[T]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Push 供 heap.Push 调用，只负责追加元素（不加锁！）
//   - 【未加锁，尽量使用下面加锁API保证线程安全】
func (pq *PriorityQueue[T]) Push(x any) {
	pq.items = append(pq.items, x.(T))
}

// Pop 供 heap.Pop 调用，只负责弹出最后一个元素（不加锁！）
//   - 【未加锁，尽量使用下面加锁API保证线程安全】
func (pq *PriorityQueue[T]) Pop() any {
	old := pq.items
	n := len(old)
	item := old[n-1]
	pq.items = old[:n-1]
	return item
}

// ---------- 公共线程安全 API ----------

// Enqueue 入队单个元素，有界队列满时返回 false
//   - 【已加锁】
//   - 返回 true ：入队成功
func (pq *PriorityQueue[T]) Enqueue(item T) bool {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.capacity > 0 && len(pq.items) >= pq.capacity {
		return false
	}
	heap.Push(pq, item)
	return true
}

// EnqueueBatch 批量入队，返回每个元素是否成功入队
//   - 【已加锁】
//   - 有界队列：按顺序尝试入队，满则后续全部失败
func (pq *PriorityQueue[T]) EnqueueBatch(items []T) []bool {
	if len(items) == 0 {
		return nil
	}

	pq.lock.Lock()
	defer pq.lock.Unlock()

	results := make([]bool, len(items))

	for i, item := range items {
		if pq.capacity > 0 && len(pq.items) >= pq.capacity {
			// 队列已满，后续全部失败
			for j := i; j < len(items); j++ {
				results[j] = false
			}
			break
		}
		heap.Push(pq, item)
		results[i] = true
	}

	return results
}

// Dequeue 出队堆顶元素
//   - 【已加锁】
func (pq *PriorityQueue[T]) Dequeue() (T, bool) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if len(pq.items) == 0 {
		var zero T
		return zero, false
	}
	item := heap.Pop(pq)
	return item.(T), true
}

// Peek 查看堆顶（不删除）
//   - 【已加锁】
func (pq *PriorityQueue[T]) Peek() (T, bool) {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	if len(pq.items) == 0 {
		var zero T
		return zero, false
	}
	return pq.items[0], true
}

// RemoveIf 删除第一个满足条件的元素
//   - 【已加锁】
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

// Size 返回当前元素数量
//   - 【已加锁】
func (pq *PriorityQueue[T]) Size() int {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	return len(pq.items)
}

// IsFull 仅对有界队列有意义
//   - 【已加锁】
//   - 返回 true 表示已满
func (pq *PriorityQueue[T]) IsFull() bool {
	if pq.capacity <= 0 {
		return false
	}
	pq.lock.Lock()
	defer pq.lock.Unlock()
	return len(pq.items) >= pq.capacity
}

// IsEmpty 判断是否为空
//   - 【已加锁】
func (pq *PriorityQueue[T]) IsEmpty() bool {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	return len(pq.items) == 0
}

// ---------- 辅助函数 ----------

func maxx(a, b int) int {
	if a > b {
		return a
	}
	return b
}
