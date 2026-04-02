package sliceX

import (
	"reflect"
	"testing"
)

func TestReverseSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"正常逆序", []int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
		{"单个元素", []int{1}, []int{1}},
		{"空切片", []int{}, []int{}},
		{"两个元素", []int{1, 2}, []int{2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReverseSlice(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReverseSlice() = %v, want %v", got, tt.want)
			}
			// 验证不修改原切片
			if len(tt.input) > 1 && reflect.DeepEqual(tt.input, got) {
				t.Error("ReverseSlice() 修改了原切片")
			}
		})
	}

	// string 类型
	t.Run("string类型逆序", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		want := []string{"c", "b", "a"}
		got := ReverseSlice(input)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ReverseSlice() = %v, want %v", got, want)
		}
	})
}

func TestSortAsc(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"无序排序", []int{3, 1, 4, 1, 5, 9, 2, 6}, []int{1, 1, 2, 3, 4, 5, 6, 9}},
		{"已排序", []int{1, 2, 3}, []int{1, 2, 3}},
		{"逆序输入", []int{3, 2, 1}, []int{1, 2, 3}},
		{"单个元素", []int{42}, []int{42}},
		{"空切片", []int{}, []int{}},
		{"重复元素", []int{5, 5, 5}, []int{5, 5, 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortAsc(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortAsc() = %v, want %v", got, tt.want)
			}
		})
	}

	// float64 类型
	t.Run("float64类型升序", func(t *testing.T) {
		input := []float64{3.14, 1.41, 2.71}
		want := []float64{1.41, 2.71, 3.14}
		got := SortAsc(input)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SortAsc() = %v, want %v", got, want)
		}
	})

	// string 类型
	t.Run("string类型升序", func(t *testing.T) {
		input := []string{"banana", "apple", "cherry"}
		want := []string{"apple", "banana", "cherry"}
		got := SortAsc(input)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SortAsc() = %v, want %v", got, want)
		}
	})
}

func TestSortDesc(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"无序排序", []int{3, 1, 4, 1, 5, 9, 2, 6}, []int{9, 6, 5, 4, 3, 2, 1, 1}},
		{"已排序升序", []int{1, 2, 3}, []int{3, 2, 1}},
		{"已排序降序", []int{3, 2, 1}, []int{3, 2, 1}},
		{"单个元素", []int{42}, []int{42}},
		{"空切片", []int{}, []int{}},
		{"负数排序", []int{-1, -3, 2, 0}, []int{2, 0, -1, -3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortDesc(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortDesc() = %v, want %v", got, tt.want)
			}
		})
	}

	// string 类型
	t.Run("string类型降序", func(t *testing.T) {
		input := []string{"apple", "cherry", "banana"}
		want := []string{"cherry", "banana", "apple"}
		got := SortDesc(input)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SortDesc() = %v, want %v", got, want)
		}
	})
}
