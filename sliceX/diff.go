package sliceX

// DiffSet 差集，只支持 comparable 类型
// 已去重
// 并且返回值的顺序是不确定的
func DiffSet[T comparable](src, dst []T) []T {
	srcMap := toMap[T](src)
	for _, val := range dst {
		delete(srcMap, val)
	}

	var ret = make([]T, 0, len(srcMap))
	for key := range srcMap {
		ret = append(ret, key)
	}

	return ret
}

// DiffSetFunc 差集，已去重
// 你应该优先使用 DiffSet
func DiffSetFunc[T any](src, dst []T, equal equalFunc[T]) []T {
	var ret = make([]T, 0, len(src))
	for _, val := range src {
		if !ContainsFunc[T](dst, func(src T) bool {
			return equal(src, val)
		}) {
			ret = append(ret, val)
		}
	}
	return deduplicateFunc[T](ret, equal)
}
