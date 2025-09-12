package configx

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ConfigValue 读取配置项，泛型约束支持的类型
type configValue interface {
	// 基础类型
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
	~float32 | ~float64 |
	~string | ~bool |
	// 时间类型
	time.Time |
	// 切片类型
	~[]string | ~[]int | ~[]int64 | ~[]float64 |
	// map 类型
	~map[string]string | ~map[string]any
}

// GetUnmarshalStruct 从配置文件读取到的值，反序列化为结构体
//   - key是配置项的 key，如 "mysql.port"
//   - rawVal 存储转换结果，读取结果存入结构体，要传指针
//   - 如项目中有多个配置文件读取，需传入 fileName 文件名 参数指定,例如: Get[int](cfg, "port", "app.yaml")
func GetUnmarshalStruct(cfg ConfigIn, key string, rawVal any, fileName ...string) error {
	return cfg.GetUnmarshalKey(key, rawVal, fileName...)
}

// Get 从 ConfigIn 安全获取指定类型的配置值
//   - 利用泛型约束，支持自动类型转换指定返回值（如 float64 → int, string → bool 等）
//   - 如项目中有多个配置文件读取，需传入 fileName 文件名 参数指定,例如: Get[int](cfg, "port", "app.yaml")
func Get[T configValue](cfg ConfigIn, key string, fileName ...string) T {
	raw := cfg.Get(key, fileName...)
	return convertToType[T](raw, key)
}

// convertToType 核心转换逻辑
func convertToType[T configValue](raw any, key string) T {
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

	// ========== 整数类型 ==========
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
	case int8:
		switch val := raw.(type) {
		case float64:
			return any(int8(val)).(T)
		case int64:
			return any(int8(val)).(T)
		case int:
			return any(int8(val)).(T)
		case string:
			if i, err := strconv.ParseInt(val, 10, 8); err == nil {
				return any(int8(i)).(T)
			}
		}
	case int16:
		switch val := raw.(type) {
		case float64:
			return any(int16(val)).(T)
		case int64:
			return any(int16(val)).(T)
		case int:
			return any(int16(val)).(T)
		case string:
			if i, err := strconv.ParseInt(val, 10, 16); err == nil {
				return any(int16(i)).(T)
			}
		}
	case int32:
		switch val := raw.(type) {
		case float64:
			return any(int32(val)).(T)
		case int64:
			return any(int32(val)).(T)
		case int:
			return any(int32(val)).(T)
		case string:
			if i, err := strconv.ParseInt(val, 10, 32); err == nil {
				return any(int32(i)).(T)
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

	// ========== 无符号整数类型 ==========
	case uint:
		switch val := raw.(type) {
		case float64:
			return any(uint(val)).(T)
		case int64:
			if val >= 0 {
				return any(uint(val)).(T)
			}
		case string:
			if u, err := strconv.ParseUint(val, 10, 64); err == nil {
				return any(uint(u)).(T)
			}
		}
	case uint8:
		switch val := raw.(type) {
		case float64:
			return any(uint8(val)).(T)
		case int64:
			if val >= 0 && val <= 255 {
				return any(uint8(val)).(T)
			}
		case uint:
			return any(uint8(val)).(T)
		case string:
			if u, err := strconv.ParseUint(val, 10, 8); err == nil {
				return any(uint8(u)).(T)
			}
		}
	case uint16:
		switch val := raw.(type) {
		case float64:
			return any(uint16(val)).(T)
		case int64:
			if val >= 0 && val <= 65535 {
				return any(uint16(val)).(T)
			}
		case uint:
			return any(uint16(val)).(T)
		case string:
			if u, err := strconv.ParseUint(val, 10, 16); err == nil {
				return any(uint16(u)).(T)
			}
		}
	case uint32:
		switch val := raw.(type) {
		case float64:
			return any(uint32(val)).(T)
		case int64:
			if val >= 0 && val <= 0xFFFFFFFF {
				return any(uint32(val)).(T)
			}
		case uint:
			return any(uint32(val)).(T)
		case string:
			if u, err := strconv.ParseUint(val, 10, 32); err == nil {
				return any(uint32(u)).(T)
			}
		}
	case uint64:
		switch val := raw.(type) {
		case int64:
			if val >= 0 {
				return any(uint64(val)).(T)
			}
		case float64:
			if val >= 0 {
				return any(uint64(val)).(T)
			}
		case uint:
			return any(uint64(val)).(T)
		case string:
			if u, err := strconv.ParseUint(val, 10, 64); err == nil {
				return any(u).(T)
			}
		}

	// ========== 浮点数类型 ==========
	case float32:
		switch val := raw.(type) {
		case int, int64:
			return any(float32(val.(int64))).(T)
		case float64:
			return any(float32(val)).(T)
		case string:
			if f, err := strconv.ParseFloat(val, 32); err == nil {
				return any(float32(f)).(T)
			}
		}
	case float64:
		switch val := raw.(type) {
		case int, int64:
			return any(float64(val.(int64))).(T)
		case float32:
			return any(float64(val)).(T)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return any(f).(T)
			}
		}

	// ========== 字符串、布尔 ==========
	case string:
		return any(fmt.Sprintf("%v", raw)).(T)
	case bool:
		switch val := raw.(type) {
		case string:
			switch strings.ToLower(val) {
			case "true", "1", "on", "yes":
				return any(true).(T)
			case "false", "0", "off", "no":
				return any(false).(T)
			}
		case int, int64:
			return any(val != 0).(T)
		case float64:
			return any(val != 0).(T)
		case uint, uint64:
			return any(val != 0).(T)
		}

	// ========== 切片类型 ==========
	case []string:
		var result []string
		switch val := raw.(type) {
		case []string:
			result = val
		case []any:
			result = make([]string, len(val))
			for i, item := range val {
				result[i] = fmt.Sprintf("%v", item)
			}
		case string:
			if len(val) > 0 {
				parts := strings.Split(val, ",")
				result = make([]string, len(parts))
				for i, p := range parts {
					result[i] = strings.TrimSpace(p)
				}
			} else {
				result = []string{}
			}
		default:
			if bs, err := json.Marshal(raw); err == nil {
				var arr []string
				if json.Unmarshal(bs, &arr) == nil {
					result = arr
				}
			}
		}
		if result != nil {
			return any(result).(T)
		}

	case []int:
		var result []int
		switch val := raw.(type) {
		case []any:
			result = make([]int, len(val))
			for i, item := range val {
				switch v := item.(type) {
				case float64:
					result[i] = int(v)
				case int64:
					result[i] = int(v)
				case string:
					if n, err := strconv.Atoi(v); err == nil {
						result[i] = n
					}
				}
			}
		case []int:
			result = val
		case string:
			if err := json.Unmarshal([]byte(val), &result); err != nil {
				result = []int{}
			}
		}
		return any(result).(T)

	case []int64:
		var result []int64
		switch val := raw.(type) {
		case []any:
			result = make([]int64, len(val))
			for i, item := range val {
				switch v := item.(type) {
				case float64:
					result[i] = int64(v)
				case int:
					result[i] = int64(v)
				case string:
					if n, err := strconv.ParseInt(v, 10, 64); err == nil {
						result[i] = n
					}
				}
			}
		case []int64:
			result = val
		case string:
			if err := json.Unmarshal([]byte(val), &result); err != nil {
				result = []int64{}
			}
		}
		return any(result).(T)

	case []float64:
		var result []float64
		switch val := raw.(type) {
		case []any:
			result = make([]float64, len(val))
			for i, item := range val {
				switch v := item.(type) {
				case int, int64:
					result[i] = float64(v.(int64))
				case float32:
					result[i] = float64(v)
				case string:
					if f, err := strconv.ParseFloat(v, 64); err == nil {
						result[i] = f
					}
				}
			}
		case []float64:
			result = val
		case string:
			if err := json.Unmarshal([]byte(val), &result); err != nil {
				result = []float64{}
			}
		}
		return any(result).(T)

	// ========== 时间、Map 类型 ==========
	case time.Time:
		switch val := raw.(type) {
		case string:
			for _, layout := range []string{
				time.RFC3339,
				"2006-01-02 15:04:05",
				"2006-01-02",
				time.ANSIC,
				time.UnixDate,
				time.RFC822,
				time.RFC1123,
			} {
				if t, err := time.Parse(layout, val); err == nil {
					return any(t).(T)
				}
			}
		case int:
			return any(time.Unix(int64(val), 0)).(T)
		case int8:
			return any(time.Unix(int64(val), 0)).(T)
		case int16:
			return any(time.Unix(int64(val), 0)).(T)
		case int32:
			return any(time.Unix(int64(val), 0)).(T)
		case int64:
			return any(time.Unix(val, 0)).(T)
		case float32:
			return any(time.Unix(int64(val), 0)).(T)
		case float64:
			return any(time.Unix(int64(val), 0)).(T)
		case uint:
			return any(time.Unix(int64(val), 0)).(T)
		case uint8:
			return any(time.Unix(int64(val), 0)).(T)
		case uint16:
			return any(time.Unix(int64(val), 0)).(T)
		case uint32:
			return any(time.Unix(int64(val), 0)).(T)
		case uint64:
			return any(time.Unix(int64(val), 0)).(T)
		}
	case time.Duration:
		switch val := raw.(type) {
		case string:
			if d, err := time.ParseDuration(val); err == nil {
				return any(d).(T)
			}
		case int:
			return any(time.Duration(val)).(T)
		case int8:
			return any(time.Duration(val)).(T)
		case int16:
			return any(time.Duration(val)).(T)
		case int32:
			return any(time.Duration(val)).(T)
		case int64:
			return any(time.Duration(val)).(T)
		case float32:
			return any(time.Duration(val)).(T)
		case float64:
			return any(time.Duration(val)).(T)
		case uint:
			return any(time.Duration(val)).(T)
		case uint8:
			return any(time.Duration(val)).(T)
		case uint16:
			return any(time.Duration(val)).(T)
		case uint32:
			return any(time.Duration(val)).(T)
		case uint64:
			return any(time.Duration(val)).(T)
		}

	// ========== 映射类型 ==========
	case map[string]string:
		result := make(map[string]string)
		switch val := raw.(type) {
		case map[string]any:
			for k, v := range val {
				result[k] = fmt.Sprintf("%v", v)
			}
			return any(result).(T)
		case map[string]string:
			return any(val).(T)
		}

	case map[string]any:
		if m, ok := raw.(map[string]any); ok {
			return any(m).(T)
		}
	}

	// 转换失败，返回零值（或 panic）
	// panic(fmt.Sprintf("config key '%s': cannot convert %T to %T", key, raw, zero))
	return zero
}
