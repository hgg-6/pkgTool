package slicex

import (
	"testing"
)

func TestSlicex(t *testing.T) {
	s := make([]int, 5, 5)
	s = []int{1, 3, 5, 6, 7, 4, 2}
	ma := Max[int](s)
	mi := Min[int](s)
	t.Log("max: ", ma)
	t.Log("min: ", mi)

	s = make([]int, 0)
	ma = Max[int](s)
	mi = Min[int](s)
	t.Log("max: ", ma)
	t.Log("min: ", mi)

	numbers := []int{10, 20, 30, 40}
	// 使用 Map：将每个数字转为字符串，格式为 "索引: 数字"
	result := Map[int, float64](numbers, func(idx int, num int) float64 {
		return float64(num) * 1.1
	})
	t.Log(result)

	t.Log(Sum(numbers))
}
