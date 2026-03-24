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
func DiffSetFunc[T any](src, dst []T, equal EqualFunc[T]) []T {
	// 分组去重
	srcGroups := groupByHash(src, equal)
	dstGroups := groupByHash(dst, equal)

	// 收集结果
	var ret []T
	for h, srcVals := range srcGroups {
		dstVals, ok := dstGroups[h]
		if !ok {
			// 该哈希下 dst 中没有元素，srcVals 全部加入结果
			ret = append(ret, srcVals...)
			continue
		}
		// 对于每个 src 元素，检查是否在 dst 中出现
		for _, sv := range srcVals {
			found := false
			for _, dv := range dstVals {
				if equal(sv, dv) {
					found = true
					break
				}
			}
			if !found {
				ret = append(ret, sv)
			}
		}
	}
	return ret
}
