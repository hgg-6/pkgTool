package syncX

import (
	"fmt"
	"sync"
)

// RWMutexMap 带读写锁的并发安全泛型Map，且支持限制最大容量控制【仅用于写场景>=30%，读场景>=90%建议还是使用sync.map】
type RWMutexMap[K comparable, V any] struct {
	mu     sync.RWMutex
	items  map[K]V
	maxCap int // 最大容量，0表示无限制
}

// NewRWMutexMap 创建一个新的并发安全Map
//   - capacity: 初始容量，如果为0则使用默认容量
//   - maxCapacity: 最大容量限制，如果为0则表示无限制
func NewRWMutexMap[K comparable, V any](capacity, maxCapacity int) *RWMutexMap[K, V] {
	if capacity < 0 {
		capacity = 0
	}
	if maxCapacity < 0 {
		maxCapacity = 0
	}

	return &RWMutexMap[K, V]{
		items:  make(map[K]V, capacity),
		maxCap: maxCapacity,
	}
}

// Set 设置键值对，如果容量已满则返回false，key存在则更新值，key不存在则添加
func (m *RWMutexMap[K, V]) Set(key K, value V) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查容量限制
	if m.maxCap > 0 {
		// 如果key不存在且已达到最大容量，则拒绝添加
		if _, exists := m.items[key]; !exists && len(m.items) >= m.maxCap {
			return false
		}
	}

	m.items[key] = value
	return true
}

// Get 获取值
func (m *RWMutexMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, exists := m.items[key]
	return value, exists
}

// Delete 删除键值对
func (m *RWMutexMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.items, key)
}

// Len 返回当前元素数量
func (m *RWMutexMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.items)
}

// CapMax 返回容量限制
func (m *RWMutexMap[K, V]) CapMax() int {
	return m.maxCap
}

// Has 检查key是否存在
func (m *RWMutexMap[K, V]) Has(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.items[key]
	return exists
}

// Keys 返回所有key的切片
func (m *RWMutexMap[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]K, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回所有value的切片
func (m *RWMutexMap[K, V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	values := make([]V, 0, len(m.items))
	for _, v := range m.items {
		values = append(values, v)
	}
	return values
}

// Clear 清空map
func (m *RWMutexMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清空map但保持容量
	for k := range m.items {
		delete(m.items, k)
	}
}

// Range 遍历map，如果函数返回false则停止遍历
func (m *RWMutexMap[K, V]) Range(f func(key K, value V) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for k, v := range m.items {
		if !f(k, v) {
			break
		}
	}
}

// SetMaxCapacity 设置最大容量限制
func (m *RWMutexMap[K, V]) SetMaxCapacity(maxCap int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.maxCap = maxCap
	if maxCap > 0 {
		// 如果当前元素数量超过新设置的最大容量，需要删除多余的元素
		if len(m.items) > maxCap {
			// 简单的实现：删除任意元素直到满足容量限制
			// 在实际应用中可能需要更复杂的淘汰策略
			toDelete := len(m.items) - maxCap
			for k := range m.items {
				delete(m.items, k)
				toDelete--
				if toDelete <= 0 {
					break
				}
			}
		}
	}
}

// IsFull 检查是否已满（仅在设置了最大容量时有效）
func (m *RWMutexMap[K, V]) IsFull() bool {
	if m.maxCap <= 0 {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items) >= m.maxCap
}

// String 实现Stringer接口
func (m *RWMutexMap[K, V]) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf("RWMutexMap{len: %d, cap: %d, items: %v}", len(m.items), m.maxCap, m.items)
}
