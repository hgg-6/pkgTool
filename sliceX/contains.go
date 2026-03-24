package sliceX

import (
	"fmt"
	"hash/fnv"
)

// EqualFunc 比较两个元素是否相等的函数类型
type EqualFunc[T any] func(src, dst T) bool

func hashOf[T any](v T) uint64 {
	h := fnv.New64a()
	fmt.Fprintf(h, "%v", v)
	return h.Sum64()
}

func groupByHash[T any](slice []T, equal EqualFunc[T]) map[uint64][]T {
	groups := make(map[uint64][]T)
	for _, v := range slice {
		h := hashOf(v)
		found := false
		for _, existing := range groups[h] {
			if equal(existing, v) {
				found = true
				break
			}
		}
		if !found {
			groups[h] = append(groups[h], v)
		}
	}
	return groups
}

// Contains 判断 src 里面是否存在 dst
func Contains[T comparable](src []T, dst T) bool {
	return ContainsFunc[T](src, func(src T) bool {
		return src == dst
	})
}

// ContainsFunc 判断 src 里面是否存在 dst
// 你应该优先使用 Contains
func ContainsFunc[T any](src []T, equal func(src T) bool) bool {
	// 遍历调用equal函数进行判断
	for _, v := range src {
		if equal(v) {
			return true
		}
	}
	return false
}

// ContainsAny 判断 src 里面是否存在 dst 中的任何一个元素
func ContainsAny[T comparable](src, dst []T) bool {
	srcMap := toMap[T](src)
	for _, v := range dst {
		if _, exist := srcMap[v]; exist {
			return true
		}
	}
	return false
}

// ContainsAnyFunc 判断 src 里面是否存在 dst 中的任何一个元素
// 你应该优先使用 ContainsAny
func ContainsAnyFunc[T any](src, dst []T, equal EqualFunc[T]) bool {
	srcGroups := groupByHash(src, equal)
	for _, dv := range dst {
		h := hashOf(dv)
		if bucket, ok := srcGroups[h]; ok {
			for _, sv := range bucket {
				if equal(sv, dv) {
					return true
				}
			}
		}
	}
	return false
}

// ContainsAll 判断 src 里面是否存在 dst 中的所有元素
func ContainsAll[T comparable](src, dst []T) bool {
	srcMap := toMap[T](src)
	for _, v := range dst {
		if _, exist := srcMap[v]; !exist {
			return false
		}
	}
	return true
}

// ContainsAllFunc 判断 src 里面是否存在 dst 中的所有元素
// 你应该优先使用 ContainsAll
func ContainsAllFunc[T any](src, dst []T, equal EqualFunc[T]) bool {
	srcGroups := groupByHash(src, equal)
	for _, dv := range dst {
		h := hashOf(dv)
		bucket, ok := srcGroups[h]
		if !ok {
			return false
		}
		found := false
		for _, sv := range bucket {
			if equal(sv, dv) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
