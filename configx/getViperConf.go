package configx

import (
	"fmt"
	"strconv"
)

// Get 从 ConfigIn 安全获取指定类型的配置值
//   - 利用泛型约束，支持自动类型转换（如 float64 → int, string → bool 等）
//   - 如项目中有多个配置文件读取，需传入 fileName 文件名 参数指定,例如: Get[int](cfg, "port", "app.yaml")
func Get[T any](cfg ConfigIn, key string, fileName ...string) T {
	if len(fileName) == 0 {
		raw := cfg.Get(key)
		return convertToType[T](raw, key)
	}
	raw := cfg.Get(key, fileName[0])
	return convertToType[T](raw, key)
}

//// MustGet 获取配置，失败则 panic（适合启动阶段校验）
//func MustGet[T any](cfg ConfigIn, key string) T {
//	raw := cfg.Get(key)
//	if raw == nil {
//		panic(fmt.Sprintf("config key '%s' is not set", key))
//	}
//	v := convertToType[T](raw, key)
//	var zero T
//	if v == zero && fmt.Sprintf("%v", raw) != fmt.Sprintf("%v", zero) {
//		panic(fmt.Sprintf("config key '%s': cannot convert %T to %T", key, raw, zero))
//	}
//	return v
//}
//
//// GetOrDefault 获取配置，失败返回默认值
//func GetOrDefault[T any](cfg ConfigIn, key string, defaultValue T) T {
//	raw := cfg.Get(key)
//	if raw == nil {
//		return defaultValue
//	}
//	v := convertToType[T](raw, key)
//	var zero T
//	if v == zero && fmt.Sprintf("%v", raw) != fmt.Sprintf("%v", zero) {
//		return defaultValue
//	}
//	return v
//}

// convertToType 核心转换逻辑
func convertToType[T any](raw any, key string) T {
	var zero T

	if raw == nil {
		return zero
	}

	// 尝试直接类型断言
	if v, ok := raw.(T); ok {
		return v
	}

	// 类型不匹配，尝试常见兼容转换
	switch any(zero).(type) {
	case int:
		switch val := raw.(type) {
		case float64:
			return any(int(val)).(T)
		case int64:
			return any(int(val)).(T)
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				return any(i).(T)
			}
		}
	case int64:
		switch val := raw.(type) {
		case int:
			return any(int64(val)).(T)
		case float64:
			return any(int64(val)).(T)
		case string:
			if i, err := strconv.ParseInt(val, 10, 64); err == nil {
				return any(i).(T)
			}
		}
	case float64:
		switch val := raw.(type) {
		case int:
			return any(float64(val)).(T)
		case int64:
			return any(float64(val)).(T)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return any(f).(T)
			}
		}
	case string:
		return any(fmt.Sprintf("%v", raw)).(T) // 任何类型都能转字符串
	case bool:
		switch val := raw.(type) {
		case string:
			if val == "true" || val == "1" || val == "on" || val == "yes" {
				return any(true).(T)
			} else if val == "false" || val == "0" || val == "off" || val == "no" {
				return any(false).(T)
			}
		case int, int64:
			if fmt.Sprintf("%v", val) == "0" {
				return any(false).(T)
			} else {
				return any(true).(T)
			}
		}
	case []string:
		if slice, ok := raw.([]any); ok {
			result := make([]string, len(slice))
			for i, item := range slice {
				result[i] = fmt.Sprintf("%v", item)
			}
			return any(result).(T)
		}
		if slice, ok := raw.([]string); ok {
			return any(slice).(T)
		}
	}

	// 转换失败，返回零值
	return zero
}
