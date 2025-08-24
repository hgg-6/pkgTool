package jwtx

import (
	"fmt"
	"strings"
)

// MapKeyFunctional 支持：【注意rexp参数是如果key本身就有【.】，而非作为正则，true为正则，false为非正则，不可多传[]bool】
//   - 精确: user.name【1速度,最优】
//   - 通配展开: users.*.name【2速度,仅次精确】
//   - key 模糊: users.*name2, users.name*【2速度,仅次精确】
//   - 模糊匹配: name (任意层级)【3速度,仅用于小数据，大数据影响性能，eg: 1000层map】
func MapKeyFunctional(data map[string]any, path string, isRexp bool) ([]any, bool) {
	if data == nil || path == "" {
		return nil, false
	}

	//// 判断模式
	//if strings.ContainsRune(path, '.') {
	//	return queryPath(data, strings.Split(path, "."))
	//} else {
	//	// 单段路径
	//	return fuzzyOrWildcardQuery(data, path)
	//}

	// 判断模式
	if isRexp {
		return queryPath(data, strings.Split(path, "."))
	} else {
		// 单段路径
		return fuzzyOrWildcardQuery(data, path)
	}
}

// queryPath 处理含 . 的路径
func queryPath(data map[string]any, keys []string) ([]any, bool) {
	result := queryRecursive([]any{data}, keys, 0)
	return result, len(result) > 0
}

func queryRecursive(current []any, keys []string, index int) []any {
	if index >= len(keys) || len(current) == 0 {
		return nil
	}

	key := keys[index]
	var next []any

	for _, item := range current {
		if item == nil {
			continue
		}

		switch v := item.(type) {
		case map[string]any:
			if key == "*" {
				// 展开所有 value
				for _, val := range v {
					next = append(next, val)
				}
			} else if strings.HasPrefix(key, "*") {
				// 匹配 key 以 suffix 结尾
				suffix := key[1:]
				for k, val := range v {
					if strings.HasSuffix(k, suffix) {
						next = append(next, val)
					}
				}
			} else if strings.HasSuffix(key, "*") {
				// 匹配 key 以 prefix 开头
				prefix := key[:len(key)-1]
				for k, val := range v {
					if strings.HasPrefix(k, prefix) {
						next = append(next, val)
					}
				}
			} else {
				// 精确匹配
				if val, exists := v[key]; exists {
					next = append(next, val)
				}
			}
		case []any:
			if key == "*" {
				next = append(next, v...)
			} else {
				var idx int
				if _, err := fmt.Sscanf(key, "%d", &idx); err == nil && idx >= 0 && idx < len(v) {
					next = append(next, v[idx])
				}
			}
		default:
			// 基本类型，无法再展开
		}
	}

	if index == len(keys)-1 {
		return next
	} else if len(next) > 0 {
		return queryRecursive(next, keys, index+1)
	}
	return nil
}

// fuzzyOrWildcardQuery 单段路径：模糊或 key 通配
func fuzzyOrWildcardQuery(data map[string]any, path string) ([]any, bool) {
	var result []any

	if strings.HasPrefix(path, "*") {
		suffix := path[1:]
		findBySuffix(data, suffix, &result)
	} else if strings.HasSuffix(path, "*") {
		prefix := path[:len(path)-1]
		findByPrefix(data, prefix, &result)
	} else {
		// 普通模糊匹配
		findAllByKey(data, path, &result)
	}

	return result, len(result) > 0
}

// findBySuffix 查找所有 key 以 suffix 结尾的值
func findBySuffix(data map[string]any, suffix string, result *[]any) {
	var walk func(any)
	walk = func(item any) {
		if item == nil {
			return
		}
		switch v := item.(type) {
		case map[string]any:
			for k, val := range v {
				if strings.HasSuffix(k, suffix) {
					*result = append(*result, val)
				}
				walk(val)
			}
		case []any:
			for _, elem := range v {
				walk(elem)
			}
		}
	}
	walk(data)
}

// findByPrefix 查找所有 key 以 prefix 开头的值
func findByPrefix(data map[string]any, prefix string, result *[]any) {
	var walk func(any)
	walk = func(item any) {
		if item == nil {
			return
		}
		switch v := item.(type) {
		case map[string]any:
			for k, val := range v {
				if strings.HasPrefix(k, prefix) {
					*result = append(*result, val)
				}
				walk(val)
			}
		case []any:
			for _, elem := range v {
				walk(elem)
			}
		}
	}
	walk(data)
}

// findAllByKey 模糊匹配 key（原功能）
func findAllByKey(data map[string]any, target string, result *[]any) {
	var walk func(any)
	walk = func(item any) {
		if item == nil {
			return
		}
		switch v := item.(type) {
		case map[string]any:
			if val, exists := v[target]; exists {
				*result = append(*result, val)
			}
			for _, val := range v {
				walk(val)
			}
		case []any:
			for _, elem := range v {
				walk(elem)
			}
		}
	}
	walk(data)
}
