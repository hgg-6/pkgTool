package syncX

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

// TestWriteRWMutexMap 测试写入并发安全的Map
func TestWriteRWMutexMap(t *testing.T) {
	testCases := []struct {
		name                  string
		capacity, maxCapacity int

		key   string
		value int

		wantBool bool
	}{
		{
			name:        "无容量限制添加成功",
			capacity:    0,
			maxCapacity: 0,
			key:         "key1",
			value:       1,

			wantBool: true,
		},
		{
			name:        "有容量限制添加成功",
			capacity:    0,
			maxCapacity: 2,
			key:         "key1",
			value:       1,

			wantBool: true,
		},
		{
			name:        "有容量限制，添加失败，超出最大容量",
			capacity:    0,
			maxCapacity: 1,
			key:         "key1",
			value:       1,

			wantBool: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewRWMutexMap[string, int](tc.capacity, tc.maxCapacity)
			ok := m.Set(tc.key, tc.value)
			assert.True(t, ok)
			ok = m.Set(tc.key+strconv.Itoa(1), tc.value)
			assert.Equal(t, tc.wantBool, ok)
		})
	}
}
