package syncX

import "sync"

// Map 是对 sync.Map 的一个泛型封装
// 要注意，K 必须是 comparable 的，并且谨慎使用指针作为 K。
// 使用指针的情况下，两个 key 是否相等，仅仅取决于它们的地址
// 而不是地址指向的值。可以参考 Load 测试。
// 注意，key 不存在和 key 存在但是值恰好为零值（如 nil），是两码事
type Map[K comparable, V any] struct {
	m sync.Map
}

// NewMap 为了防止有时使用时忘记&取地址，所以又加了New构造
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{}
}

// Load 加载键值对
func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	var anyVal any
	anyVal, ok = m.m.Load(key)
	if anyVal != nil {
		value = anyVal.(V)
	}
	return
}

// Store 存储键值对
func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

// LoadOrStore 加载或者存储一个键值对
// true 代表是加载的，false 代表执行了 store
func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	var anyVal any
	anyVal, loaded = m.m.LoadOrStore(key, value)
	if anyVal != nil {
		actual = anyVal.(V)
	}
	return
}

// LoadOrStoreFunc 是一个优化，也就是使用该方法能够避免无意义的创建实例。
// 如果你的初始化过程非常消耗资源，那么使用这个方法是有价值的。
// 它的代价就是 Key 不存在的时候会多一次 Load 调用。
// 当 fn 返回 error 的时候，LoadOrStoreFunc 也会返回 error。
func (m *Map[K, V]) LoadOrStoreFunc(key K, fn func() (V, error)) (actual V, loaded bool, err error) {
	val, ok := m.Load(key)
	if ok {
		return val, true, nil
	}
	val, err = fn()
	if err != nil {
		return
	}
	actual, loaded = m.LoadOrStore(key, val)
	return
}

// LoadAndDelete 加载并且删除一个键值对
func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	var anyVal any
	anyVal, loaded = m.m.LoadAndDelete(key)
	if anyVal != nil {
		value = anyVal.(V)
	}
	return
}

// Delete 删除键值对
func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// Range 遍历, f 不能为 nil
// 传入 f 的时候，K 和 V 直接使用对应的类型，如果 f 返回 false，那么就会中断遍历
func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool {
		var (
			k K
			v V
		)
		if value != nil {
			v = value.(V)
		}
		if key != nil {
			k = key.(K)
		}
		return f(k, v)
	})
}

// IsEmpty 判断 Map 是否为空
//   - true : Map 为空
//   - false : Map 不为空
func (m *Map[K, V]) IsEmpty() bool {
	empty := true
	m.Range(func(key K, value V) bool {
		empty = false
		return false // 立即停止遍历
	})
	return empty
}
