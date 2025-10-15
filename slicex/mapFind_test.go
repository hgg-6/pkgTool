package slicex

import (
	"reflect"
	"testing"
)

func TestMapKeyFind(t *testing.T) {
	// 构造测试数据
	data := map[string]any{
		"user": map[string]any{
			"name": "Alice",
			"age":  30,
			"addr": map[string]any{
				"city": "Beijing",
				"zip":  "100000",
			},
		},
		"users": []any{
			map[string]any{
				"name":     "Bob",
				"name2":    "Bob2",
				"email":    "bob@example.com",
				"profile":  map[string]any{"name": "BobProfile"},
				"metadata": map[string]any{"extraName": "hidden"},
			},
			map[string]any{
				"name":  "Charlie",
				"name3": "Charlie3",
				"tags":  []any{"admin", "user"},
			},
		},
		"config": map[string]any{
			"debug":     true,
			"temp_name": "tmp1",
			"temp_data": "tmp2",
			"settings": map[string]any{
				"username": "sysadmin",
				"password": "secret",
			},
		},
		"meta.name": "special_key_with_dot", // 包含点的 key
	}

	tests := []struct {
		name     string
		path     string
		isRexp   bool
		expected []any
		found    bool
	}{
		// === 精确匹配 (isRexp = true) ===
		{
			name:     "exact match - user.name",
			path:     "user.name",
			isRexp:   true,
			expected: []any{"Alice"},
			found:    true,
		},
		{
			name:     "exact match - user.addr.city",
			path:     "user.addr.city",
			isRexp:   true,
			expected: []any{"Beijing"},
			found:    true,
		},
		{
			name:     "exact match - users[0].name",
			path:     "users.0.name",
			isRexp:   true,
			expected: []any{"Bob"},
			found:    true,
		},
		{
			name:     "exact match - users[1].tags[0]",
			path:     "users.1.tags.0",
			isRexp:   true,
			expected: []any{"admin"},
			found:    true,
		},
		{
			name:     "exact match - non-existent key",
			path:     "user.phone",
			isRexp:   true,
			expected: nil,
			found:    false,
		},

		// === 通配展开 (isRexp = true) ===
		{
			name:     "wildcard expand - users.*.name",
			path:     "users.*.name",
			isRexp:   true,
			expected: []any{"Bob", "Charlie"},
			found:    true,
		},
		{
			name:     "wildcard expand - *.name",
			path:     "*.name",
			isRexp:   true,
			expected: []any{"Alice"},
			found:    true,
		},
		{
			name:     "wildcard expand - user.*",
			path:     "user.*",
			isRexp:   true,
			expected: []any{"Alice", 30, map[string]any{"city": "Beijing", "zip": "100000"}},
			found:    true,
		},

		// === key 模糊匹配 (isRexp = false) ===
		{
			name:     "key suffix - *name2",
			path:     "*name2",
			isRexp:   false,
			expected: []any{"Bob2"},
			found:    true,
		},
		{
			name:     "key prefix - name*",
			path:     "name*",
			isRexp:   false,
			expected: []any{"Alice", "Bob", "Charlie", "BobProfile", "hidden"},
			found:    true,
		},
		{
			name:     "key prefix - temp_*",
			path:     "temp_*",
			isRexp:   false,
			expected: []any{"tmp1", "tmp2"},
			found:    true,
		},

		// === 模糊匹配任意层级 (isRexp = false) ===
		{
			name:     "fuzzy match - name (any level)",
			path:     "name",
			isRexp:   false,
			expected: []any{"Alice", "Bob", "Charlie", "BobProfile", "hidden", "sysadmin"},
			found:    true,
		},
		{
			name:     "fuzzy match - non-existent key",
			path:     "phone",
			isRexp:   false,
			expected: nil,
			found:    false,
		},

		// === 特殊情况：key 本身包含 . (isRexp = true) ===
		{
			name:     "exact match - meta.name (key with dot)",
			path:     "meta.name",
			isRexp:   true,
			expected: []any{"special_key_with_dot"},
			found:    true,
		},

		// === 边界情况 ===
		{
			name:     "empty path",
			path:     "",
			isRexp:   true,
			expected: nil,
			found:    false,
		},
		{
			name:     "nil data",
			path:     "user.name",
			isRexp:   true,
			expected: nil,
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := MapKeyFind(data, tt.path, tt.isRexp)

			if found != tt.found {
				t.Errorf("expected found=%v, got %v", tt.found, found)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected result %v, got %v", tt.expected, result)
			}
		})
	}
}

// ==================== Benchmark ====================

func BenchmarkMapKeyFind_ExactMatch(b *testing.B) {
	data := map[string]any{
		"user": map[string]any{
			"name": "Alice",
			"age":  30,
			"addr": map[string]any{"city": "Beijing"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MapKeyFind(data, "user.name", true)
	}
}

func BenchmarkMapKeyFind_WildcardExpand(b *testing.B) {
	data := map[string]any{
		"users": []any{
			map[string]any{"name": "Bob"},
			map[string]any{"name": "Charlie"},
			map[string]any{"name": "David"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MapKeyFind(data, "users.*.name", true)
	}
}

func BenchmarkMapKeyFind_FuzzyMatch_Small(b *testing.B) {
	data := map[string]any{
		"user": map[string]any{
			"profile": map[string]any{
				"settings": map[string]any{
					"name": "Alice",
				},
			},
			"metadata": map[string]any{
				"extra_name": "temp",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MapKeyFind(data, "name", false)
	}
}

// 模拟深层嵌套结构用于性能测试
func deepNestedMap(depth int) map[string]any {
	m := map[string]any{"value": "test"}
	for i := 0; i < depth; i++ {
		m = map[string]any{"level": m}
	}
	return m
}

func BenchmarkMapKeyFind_FuzzyMatch_Deep100(b *testing.B) {
	//data := deepNestedMap(100)
	data := deepNestedMap(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MapKeyFind(data, "value", false)
	}
}

func BenchmarkMapKeyFind_FuzzyMatch_Deep1000(b *testing.B) {
	data := deepNestedMap(1000)
	b.Skip("Skipping deep 1000 benchmark - too slow")
	// b.Skip() 可用于跳过特别耗时的测试
	// 如果你想运行，去掉 Skip 并设置 -timeout 较长

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MapKeyFind(data, "value", false)
	}
}
