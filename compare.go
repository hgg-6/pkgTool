package hggTool

// Comparator 用于比较两个对象的大小
//   - src < dst, 返回-1
//   - src = dst, 返回0
//   - src > dst, 返回1
/*
	不要返回任何其它值！
*/
type Comparator[T any] func(src T, dst T) int

// ComparatorRealNumber 用于比较两个对象的大小
//   - src < dst, 返回-1
//   - src = dst, 返回0
//   - src > dst, 返回1
/*
	不要返回任何其它值！
*/
func ComparatorRealNumber[T RealNumber](src T, dst T) int {
	if src < dst {
		return -1
	} else if src == dst {
		return 0
	} else {
		return 1
	}
}

// RealNumber 实数
// 绝大多数情况下，你都应该用这个来表达数字的含义
type RealNumber interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}
