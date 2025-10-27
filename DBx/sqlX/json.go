package sqlX

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JsonColumn 是一个支持 NULL 的泛型 JSON 列包装器
// 类似 sql.NullString，但用于任意可 JSON 序列化的类型
type JsonColumn[T any] struct {
	Val   T    // 存储实际数据
	Valid bool // 表示该字段在数据库中是否为非 NULL
}

// Value 实现 driver.Valuer 接口
// 将 Go 值转换为数据库存储值（[]byte 或 nil）
func (j JsonColumn[T]) Value() (driver.Value, error) {
	if !j.Valid {
		return nil, nil
	}
	return json.Marshal(j.Val)
}

// Scan 实现 sql.Scanner 接口
// 从数据库读取值（[]byte, string, nil）并反序列化到 j.Val
func (j *JsonColumn[T]) Scan(src any) error {
	if src == nil {
		// 数据库值为 NULL：重置为零值，Valid = false
		var zero T
		j.Val = zero
		j.Valid = false
		return nil
	}

	var bs []byte
	switch v := src.(type) {
	case []byte:
		bs = v
	case string:
		bs = []byte(v)
	default:
		return fmt.Errorf("JsonColumn.Scan: unsupported src type %T", src)
	}

	if err := json.Unmarshal(bs, &j.Val); err != nil {
		return fmt.Errorf("JsonColumn.Scan: failed to unmarshal JSON: %w", err)
	}
	j.Valid = true
	return nil
}
