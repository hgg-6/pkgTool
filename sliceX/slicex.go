// Package sliceX 提供切片操作的工具函数，包括过滤、映射、去重、集合运算等。
package sliceX

type NumberOrString interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 | uintptr |
		float32 | float64 |
		string
}

// Max 切片里面最大值
// 如果切片为空，返回零值
func Max[T NumberOrString](a []T) T {
	val, _ := MaxSafe(a)
	return val
}

// MaxSafe 切片里面最大值，返回值和是否成功
func MaxSafe[T NumberOrString](a []T) (T, bool) {
	if len(a) < 1 {
		var zero T
		return zero, false
	}
	m := a[0]
	for _, v := range a {
		m = max(m, v)
	}
	return m, true
}

// Min 切片里面最小值
// 如果切片为空，返回零值
func Min[T NumberOrString](a []T) T {
	val, _ := MinSafe(a)
	return val
}

// MinSafe 切片里面最小值，返回值和是否成功
func MinSafe[T NumberOrString](a []T) (T, bool) {
	if len(a) < 1 {
		var zero T
		return zero, false
	}
	m := a[0]
	for _, v := range a {
		m = min(m, v)
	}
	return m, true
}

// Sum 切片求和
func Sum[T NumberOrString](s []T) T {
	var a T
	for _, v := range s {
		a += v
	}
	return a
}
