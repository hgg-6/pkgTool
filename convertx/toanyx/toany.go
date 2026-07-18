package toanyx

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// dstTypeValue 泛型约束
type dstTypeValue interface {
	// 基础类型
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 |
		~string | ~bool |
		// 时间类型
		time.Time |
		//time.Duration |
		// 切片类型
		~[]string | ~[]int | ~[]int64 | ~[]float64 |
		// map 类型
		~map[string]string | ~map[string]any |
		// 新增切片 of map
		~[]map[string]any
}

// ToAny 安全转换，返回值 + 是否成功
func ToAny[T dstTypeValue](v any) (T, bool) {
	return convertToType[T](v)
}

// 转换器函数类型
type converterFunc[T any] func(any) (T, bool)

// 转换器注册表（支持扩展）
var converters = make(map[string]any)

// 注册转换器（内部使用）
func registerConverter[T any](fn converterFunc[T]) {
	var zero T
	typeName := fmt.Sprintf("%T", zero)
	converters[typeName] = fn
}

// 获取转换器
func getConverter[T any]() converterFunc[T] {
	var zero T
	typeName := fmt.Sprintf("%T", zero)
	if fn, ok := converters[typeName]; ok {
		return fn.(converterFunc[T])
	}
	return nil
}

// 初始化注册常用转换器
func init() {
	// ========== 基础类型 ==========
	registerConverter(convertInt[int])
	registerConverter(convertInt[int8])
	registerConverter(convertInt[int16])
	registerConverter(convertInt[int32])
	registerConverter(convertInt[int64])

	registerConverter(convertUint[uint])
	registerConverter(convertUint[uint8])
	registerConverter(convertUint[uint16])
	registerConverter(convertUint[uint32])
	registerConverter(convertUint[uint64])

	registerConverter(convertFloat[float32])
	registerConverter(convertFloat[float64])

	registerConverter(convertString)
	registerConverter(convertBool)

	// ========== 时间类型 ==========
	registerConverter(convertTime)
	registerConverter(convertDuration)

	// ========== 切片类型 ==========
	registerConverter(convertSliceString)
	registerConverter(convertSliceInt)
	registerConverter(convertSliceInt64)
	registerConverter(convertSliceFloat64)
	registerConverter(convertSliceByte)

	// ========== Map 类型 ==========
	registerConverter(convertMapStringString)
	registerConverter(convertMapStringAny)

	// ========== Slice of Map ==========
	registerConverter(convertSliceMapStringAny)
}

// ========== 通用转换函数 ==========

// 整数通用转换，带目标类型边界检查（超出范围返回 ok=false，避免静默截断/溢出）。
func convertInt[T ~int | ~int8 | ~int16 | ~int32 | ~int64](src any) (T, bool) {
	minT, maxT := int64Range[T]()
	inRange := func(i int64) bool { return i >= minT && i <= maxT }
	// uInRange 用于无符号源：maxT>=0 时（有符号目标）用 uint64 比较。
	uInRange := func(u uint64) bool { return maxT >= 0 && u <= uint64(maxT) }
	switch v := src.(type) {
	case T:
		return v, true
	case int:
		if inRange(int64(v)) {
			return T(v), true
		}
	case int8:
		if inRange(int64(v)) {
			return T(v), true
		}
	case int16:
		if inRange(int64(v)) {
			return T(v), true
		}
	case int32:
		if inRange(int64(v)) {
			return T(v), true
		}
	case int64:
		if inRange(v) {
			return T(v), true
		}
	case uint:
		if uInRange(uint64(v)) {
			return T(v), true
		}
	case uint8:
		if uInRange(uint64(v)) {
			return T(v), true
		}
	case uint16:
		if uInRange(uint64(v)) {
			return T(v), true
		}
	case uint32:
		if uInRange(uint64(v)) {
			return T(v), true
		}
	case uint64:
		if uInRange(v) {
			return T(v), true
		}
	case float32:
		// 截断转 int64，仅做边界检查。
		if fv := int64(v); inRange(fv) {
			return T(fv), true
		}
	case float64:
		if fv := int64(v); inRange(fv) {
			return T(fv), true
		}
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil && inRange(i) {
			return T(i), true
		}
	case json.Number:
		if i, err := v.Int64(); err == nil && inRange(i) {
			return T(i), true
		}
	}
	return *new(T), false
}

// int64Range 返回 T 能表示的最小/最大 int64 值。
// 用 any(zero).(type) 做类型分发，避免 ^T(0) 在大位宽无符号类型上的转换陷阱。
func int64Range[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() (min, max int64) {
	var zero T
	switch any(zero).(type) {
	case int8:
		return math.MinInt8, math.MaxInt8
	case int16:
		return math.MinInt16, math.MaxInt16
	case int32:
		return math.MinInt32, math.MaxInt32
	default: // int / int64
		return math.MinInt64, math.MaxInt64
	}
}

// 无符号整数通用转换
func convertUint[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](src any) (T, bool) {
	// maxT 是 T 的最大值转换为 uint64 表示。
	// 注意：不能用 int64(^T(0))，当 T=uint64 时 ^T(0)=MaxUint64 强转 int64 会溢出为 -1，
	// 导致后续上界判断恒为 false。统一用 uint64 作为比较基准。
	maxT := uint64(^T(0))
	switch v := src.(type) {
	case T:
		return v, true
	case int:
		if v >= 0 {
			return T(v), true
		}
	case int8:
		if v >= 0 {
			return T(v), true
		}
	case int16:
		if v >= 0 {
			return T(v), true
		}
	case int32:
		if v >= 0 {
			return T(v), true
		}
	case int64:
		if v >= 0 && uint64(v) <= maxT {
			return T(v), true
		}
	case uint:
		return T(v), true
	case uint8:
		return T(v), true
	case uint16:
		return T(v), true
	case uint32:
		return T(v), true
	case uint64:
		if v <= maxT {
			return T(v), true
		}
	case float32:
		if v >= 0 && uint64(v) <= maxT {
			return T(v), true
		}
	case float64:
		if v >= 0 && uint64(v) <= maxT {
			return T(v), true
		}
	case string:
		if u, err := strconv.ParseUint(v, 10, 64); err == nil && u <= maxT {
			return T(u), true
		}
	case json.Number:
		if u, err := v.Int64(); err == nil && u >= 0 && uint64(u) <= maxT {
			return T(u), true
		}
	}
	return *new(T), false
}

// 浮点数通用转换
func convertFloat[T ~float32 | ~float64](src any) (T, bool) {
	switch v := src.(type) {
	case T:
		return v, true
	case int, int8, int16, int32, int64:
		return T(reflectToInt64(v)), true
	case uint, uint8, uint16, uint32, uint64:
		return T(reflectToUint64(v)), true
	case float32:
		return T(v), true
	case float64:
		return T(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return T(f), true
		}
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return T(f), true
		}
	}
	return *new(T), false
}

// reflectToInt64 辅助函数（避免重复）
func reflectToInt64(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	default:
		return 0
	}
}

func reflectToUint64(v any) uint64 {
	switch x := v.(type) {
	case uint:
		return uint64(x)
	case uint8:
		return uint64(x)
	case uint16:
		return uint64(x)
	case uint32:
		return uint64(x)
	case uint64:
		return x
	default:
		return 0
	}
}

// 字符串转换
func convertString(src any) (string, bool) {
	switch v := src.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	case fmt.Stringer:
		return v.String(), true
	case error:
		return v.Error(), true
	default:
		return fmt.Sprintf("%v", src), true // 总是成功
	}
}

// 布尔转换
func convertBool(src any) (bool, bool) {
	switch v := src.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "on", "yes", "y":
			return true, true
		case "false", "0", "off", "no", "n":
			return false, true
		}
	case int, int8, int16, int32, int64:
		return reflectToInt64(v) != 0, true
	case uint, uint8, uint16, uint32, uint64:
		return reflectToUint64(v) != 0, true
	case float32, float64:
		return v != 0, true
	}
	return false, false
}

// 时间转换
func convertTime(src any) (time.Time, bool) {
	switch v := src.(type) {
	case time.Time:
		return v, true
	case string:
		for _, layout := range []string{
			time.RFC3339, time.RFC3339Nano,
			"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02",
			time.ANSIC, time.UnixDate, time.RFC822, time.RFC1123,
		} {
			if t, err := time.Parse(layout, v); err == nil {
				return t, true
			}
		}
	case int, int8, int16, int32, int64:
		return time.Unix(reflectToInt64(v), 0), true
	case uint, uint8, uint16, uint32, uint64:
		return time.Unix(int64(reflectToUint64(v)), 0), true
	case float32:
		return time.Unix(int64(v), 0), true
	case float64:
		return time.Unix(int64(v), 0), true
	}
	return time.Time{}, false
}

// Duration 转换
func convertDuration(src any) (time.Duration, bool) {
	switch v := src.(type) {
	case time.Duration:
		return v, true
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d, true
		}
	case int, int8, int16, int32, int64:
		return time.Duration(reflectToInt64(v)), true
	case uint, uint8, uint16, uint32, uint64:
		return time.Duration(reflectToUint64(v)), true
	case float32:
		return time.Duration(v), true
	case float64:
		return time.Duration(v), true
	}
	return 0, false
}

// ========== 切片转换 ==========

func convertSliceString(src any) ([]string, bool) {
	switch v := src.(type) {
	case []string:
		return v, true
	case []any:
		res := make([]string, len(v))
		for i, item := range v {
			if s, ok := convertString(item); ok {
				res[i] = s
			} else {
				res[i] = fmt.Sprintf("%v", item)
			}
		}
		return res, true
	case string:
		if len(strings.TrimSpace(v)) == 0 {
			return []string{}, true
		}
		parts := strings.Split(v, ",")
		res := make([]string, len(parts))
		for i, p := range parts {
			res[i] = strings.TrimSpace(p)
		}
		return res, true
	default:
		if bs, err := json.Marshal(src); err == nil {
			var arr []string
			if json.Unmarshal(bs, &arr) == nil {
				return arr, true
			}
		}
	}
	return nil, false
}

func convertSliceInt(src any) ([]int, bool) {
	switch v := src.(type) {
	case []int:
		return v, true
	case []any:
		res := make([]int, len(v))
		for i, item := range v {
			if n, ok := convertInt[int](item); ok {
				res[i] = n
			} else {
				return nil, false
			}
		}
		return res, true
	case string:
		var arr []int
		if err := json.Unmarshal([]byte(v), &arr); err == nil {
			return arr, true
		}
	}
	return nil, false
}

func convertSliceInt64(src any) ([]int64, bool) {
	switch v := src.(type) {
	case []int64:
		return v, true
	case []any:
		res := make([]int64, len(v))
		for i, item := range v {
			if n, ok := convertInt[int64](item); ok {
				res[i] = n
			} else {
				return nil, false
			}
		}
		return res, true
	case string:
		var arr []int64
		if err := json.Unmarshal([]byte(v), &arr); err == nil {
			return arr, true
		}
	}
	return nil, false
}

func convertSliceFloat64(src any) ([]float64, bool) {
	switch v := src.(type) {
	case []float64:
		return v, true
	case []any:
		res := make([]float64, len(v))
		for i, item := range v {
			if f, ok := convertFloat[float64](item); ok {
				res[i] = f
			} else {
				return nil, false
			}
		}
		return res, true
	case string:
		var arr []float64
		if err := json.Unmarshal([]byte(v), &arr); err == nil {
			return arr, true
		}
	}
	return nil, false
}

func convertSliceByte(src any) ([]byte, bool) {
	switch v := src.(type) {
	case []byte:
		return v, true
	case string:
		return []byte(v), true
	}
	return nil, false
}

// ========== Map 转换 ==========

func convertMapStringString(src any) (map[string]string, bool) {
	switch v := src.(type) {
	case map[string]string:
		return v, true
	case map[string]any:
		res := make(map[string]string, len(v))
		for k, val := range v {
			if s, ok := convertString(val); ok {
				res[k] = s
			} else {
				res[k] = fmt.Sprintf("%v", val)
			}
		}
		return res, true
	case string:
		var m map[string]string
		if err := json.Unmarshal([]byte(v), &m); err == nil {
			return m, true
		}
	}
	return nil, false
}

func convertMapStringAny(src any) (map[string]any, bool) {
	if m, ok := src.(map[string]any); ok {
		return m, true
	}
	if s, ok := src.(string); ok {
		var m map[string]any
		if err := json.Unmarshal([]byte(s), &m); err == nil {
			return m, true
		}
	}
	return nil, false
}

// ========== []map 切片转换 ==========

// convertSliceMapStringAny 将 any 转为 []map[string]any，不使用反射
func convertSliceMapStringAny(src any) ([]map[string]any, bool) {
	switch v := src.(type) {
	case []map[string]any:
		return v, true // 直接命中，最快路径

	case []any:
		// 尝试转换每个元素为 map[string]any
		res := make([]map[string]any, len(v))
		for i, item := range v {
			if m, ok := convertMapStringAny(item); ok {
				res[i] = m
			} else {
				return nil, false // 任一元素失败，整体失败
			}
		}
		return res, true

	case string:
		// 尝试 JSON 解码
		var arr []map[string]any
		if err := json.Unmarshal([]byte(v), &arr); err == nil {
			return arr, true
		}

		// 尝试是否是单个 map（兼容场景）
		var singleMap map[string]any
		if err := json.Unmarshal([]byte(v), &singleMap); err == nil {
			return []map[string]any{singleMap}, true
		}

	case []map[string]string:
		// 转换 []map[string]string → []map[string]any
		res := make([]map[string]any, len(v))
		for i, m := range v {
			newMap := make(map[string]any, len(m))
			for k, val := range m {
				newMap[k] = val
			}
			res[i] = newMap
		}
		return res, true
	}

	return nil, false
}

// ========== 核心调度函数 ==========

func convertToType[T dstTypeValue](src any) (T, bool) {
	if src == nil {
		return *new(T), false
	}

	// 1. 直接类型匹配（最快）
	if v, ok := src.(T); ok {
		return v, true
	}

	// 2. 查找注册的转换器
	if converter := getConverter[T](); converter != nil {
		if result, ok := converter(src); ok {
			return result, true
		}
	}

	// 3. 兜底：fmt.Sprint（仅对 string 类型安全）
	var zero T
	if _, ok := any(zero).(string); ok {
		return any(fmt.Sprintf("%v", src)).(T), true
	}

	return *new(T), false
}
