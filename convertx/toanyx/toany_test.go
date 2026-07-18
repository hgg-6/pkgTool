package toanyx

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToAny(t *testing.T) {
	// IntToString：int -> string 走 fmt.Sprint 兜底
	t.Run("IntToString ok", func(t *testing.T) {
		s, ok := ToAny[string](int(123))
		assert.True(t, ok, "转换失败")
		assert.Equal(t, "123", s)
	})

	// StringToInt：string -> int 走 strconv.ParseInt
	t.Run("StringToInt ok", func(t *testing.T) {
		v, ok := ToAny[int]("456")
		assert.True(t, ok, "转换失败")
		assert.Equal(t, 456, v)
	})

	// StringToInt 非法输入应失败
	t.Run("StringToInt invalid", func(t *testing.T) {
		_, ok := ToAny[int]("abc")
		assert.False(t, ok, "非法输入不应转换成功")
	})

	// Float64 -> int
	t.Run("Float64ToInt ok", func(t *testing.T) {
		v, ok := ToAny[int](float64(12.3))
		assert.True(t, ok)
		assert.Equal(t, 12, v)
	})

	// int64 -> uint64（P0-2 回归）
	t.Run("Int64ToUint64 ok", func(t *testing.T) {
		v, ok := ToAny[uint64](int64(789))
		assert.True(t, ok)
		assert.Equal(t, uint64(789), v)
	})

	// json.Number -> uint64（P0-2 回归）
	t.Run("JSONNumberToUint64 ok", func(t *testing.T) {
		v, ok := ToAny[uint64](json.Number("321"))
		assert.True(t, ok)
		assert.Equal(t, uint64(321), v)
	})

	// []map[string]string -> []map[string]any
	t.Run("SliceMapStringToSliceMapAny ok", func(t *testing.T) {
		src := []map[string]string{{"a": "1", "b": "2"}}
		s, ok := ToAny[[]map[string]any](src)
		assert.True(t, ok, "转换失败")
		assert.Len(t, s, 1)
		assert.Equal(t, "1", s[0]["a"])
		assert.Equal(t, "2", s[0]["b"])
	})

	// []map[string]any -> []map[string]any（直接命中）
	t.Run("SliceMapAnyDirect ok", func(t *testing.T) {
		src := []map[string]any{{"a": 1}}
		s, ok := ToAny[[]map[string]any](src)
		assert.True(t, ok)
		assert.Equal(t, 1, s[0]["a"])
	})

	// map[string]string -> map[string]string（直接命中）
	t.Run("MapStringDirect ok", func(t *testing.T) {
		src := map[string]string{"a": "1"}
		s, ok := ToAny[map[string]string](src)
		assert.True(t, ok)
		assert.Equal(t, "1", s["a"])
	})

	// string -> map[string]any 走 JSON 解码
	t.Run("StringJSONToMapAny ok", func(t *testing.T) {
		s, ok := ToAny[map[string]any](`{"a":1,"b":"x"}`)
		assert.True(t, ok)
		assert.Equal(t, 1.0, s["a"])
		assert.Equal(t, "x", s["b"])
	})

	// time.Duration 注册器路径
	t.Run("Int64ToDuration ok", func(t *testing.T) {
		v, ok := ToAny[time.Duration](int64(5_000_000_000))
		assert.True(t, ok)
		assert.Equal(t, 5*time.Second, v)
	})

	// nil 输入应失败
	t.Run("Nil fail", func(t *testing.T) {
		_, ok := ToAny[int](nil)
		assert.False(t, ok)
	})

	// int -> int8 边界（P0-4 修复后会校验，200 超过 int8 范围应失败）
	t.Run("IntToInt8 overflow", func(t *testing.T) {
		_, ok := ToAny[int8](int(200))
		assert.False(t, ok, "200 超出 int8 范围应失败")
	})

	// int -> int8 正常（P0-4 修复后）
	t.Run("IntToInt8 ok", func(t *testing.T) {
		v, ok := ToAny[int8](int(100))
		assert.True(t, ok)
		assert.Equal(t, int8(100), v)
	})

	// int -> int8 负溢出应失败
	t.Run("IntToInt8 negative overflow", func(t *testing.T) {
		_, ok := ToAny[int8](int(-200))
		assert.False(t, ok, "-200 超出 int8 范围应失败")
	})

	// int64 边界值 -> int8
	t.Run("Int64ToInt8 boundary ok", func(t *testing.T) {
		v, ok := ToAny[int8](int64(127))
		assert.True(t, ok)
		assert.Equal(t, int8(127), v)
	})

	// int64 超边界 -> int8 失败
	t.Run("Int64ToInt8 boundary fail", func(t *testing.T) {
		_, ok := ToAny[int8](int64(128))
		assert.False(t, ok, "128 超出 int8 范围应失败")
	})

	// string -> int8 边界检查
	t.Run("StringToInt8 overflow fail", func(t *testing.T) {
		_, ok := ToAny[int8]("200")
		assert.False(t, ok, "200 超出 int8 范围应失败")
	})

	// uint8 -> int8 跨窄类型边界
	t.Run("Uint8ToInt8 overflow fail", func(t *testing.T) {
		_, ok := ToAny[int8](uint8(200))
		assert.False(t, ok, "uint8(200) 超出 int8 范围应失败")
	})
	t.Run("Uint8ToInt8 ok", func(t *testing.T) {
		v, ok := ToAny[int8](uint8(100))
		assert.True(t, ok)
		assert.Equal(t, int8(100), v)
	})

	// int16 -> int8 跨窄类型边界
	t.Run("Int16ToInt8 overflow fail", func(t *testing.T) {
		_, ok := ToAny[int8](int16(300))
		assert.False(t, ok, "int16(300) 超出 int8 范围应失败")
	})

	// 窄源 -> 宽目标应成功
	t.Run("Int8ToInt ok", func(t *testing.T) {
		v, ok := ToAny[int](int8(-5))
		assert.True(t, ok)
		assert.Equal(t, -5, v)
	})
}
