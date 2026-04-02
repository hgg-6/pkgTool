package sliceX

import "slices"

/*
	此方法会返回一个新的切片，不会对原切片做修改，需修改原切片方式直接用标准库就行了
*/

// ReverseSlice 对传入的切片进行逆序，返回一个新的切片，不修改原切片
func ReverseSlice[T NumberOrString](s []T) []T {
	out := slices.Clone(s)
	slices.Reverse(out)
	return out
}

// SortAsc 对传入的切片进行从小到大排序，返回一个新的切片，不修改原切片
func SortAsc[T NumberOrString](s []T) []T {
	out := slices.Clone(s)
	slices.Sort(out)
	return out
}

// SortDesc 对传入的切片进行从大到小排序，返回一个新的切片，不修改原切片
func SortDesc[T NumberOrString](s []T) []T {
	out := slices.Clone(s)
	slices.Sort(out)
	slices.Reverse(out)
	return out
}
