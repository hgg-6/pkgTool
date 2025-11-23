package sliceX

type NumberOrString interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 | uintptr |
		float32 | float64 |
		string
}

// Max 切片里面最大值
func Max[T NumberOrString](a []T) T {
	if len(a) < 1 {
		var s T
		return s
	}
	m := a[0]
	for _, v := range a {
		m = max(m, v)
	}
	return m
}

// Min 切片里面最小值
func Min[T NumberOrString](a []T) T {
	if len(a) < 1 {
		var s T
		return s
	}
	m := a[0]
	for _, v := range a {
		m = min(m, v)
	}
	return m
}

// Sum 切片求和
func Sum[T NumberOrString](s []T) T {
	var a T
	for _, v := range s {
		a += v
	}
	return a
}
