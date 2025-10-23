package queueX

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

// 测试查看堆顶元素
func TestPriorityQueuePeek(t *testing.T) {
	testCases := []struct {
		name      string
		less      func(a, b int) bool
		input     []int // 输入
		wanOutput int   // 预期输出
	}{
		{
			name: "小顶堆查看顶堆元素",
			less: func(a, b int) bool { // 告诉堆算法：“在位置 i 和 j 的两个元素，谁应该排在前面？”
				return a < b // a < b 则返回true，对应最小堆
			},
			input:     []int{2, 5, 1, 3, 4},
			wanOutput: 1,
		},
		{
			name: "大顶堆查看顶堆元素",
			less: func(a, b int) bool { // 告诉堆算法：“在位置 i 和 j 的两个元素，谁应该排在前面？”
				return a > b // a > b 则返回true，对应最大堆
			},
			input:     []int{2, 5, 1, 3, 4},
			wanOutput: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pq := NewPriorityQueue[int](tc.less, 0)
			for _, v := range tc.input {
				pq.Enqueue(v)
			}
			v, ok := pq.Peek()
			assert.True(t, ok)
			assert.Equal(t, v, tc.wanOutput)
			log.Println(v)
		})
	}

}

// 测试入队
func TestPriorityQueueEnqueue(t *testing.T) {
	testCases := []struct {
		name      string
		less      func(a, b int) bool
		input     []int // 输入
		wanOutput int   // 预期输出
	}{
		{
			name: "小顶堆入队",
			less: func(a, b int) bool { // 告诉堆算法：“在位置 i 和 j 的两个元素，谁应该排在前面？”
				return a < b // a < b 则返回true，对应最小堆
			},
			input:     []int{2, 5, 1, 3, 4},
			wanOutput: 1,
		},
		{
			name: "大顶堆入队",
			less: func(a, b int) bool { // 告诉堆算法：“在位置 i 和 j 的两个元素，谁应该排在前面？”
				return a > b // a > b 则返回true，对应最大堆
			},
			input:     []int{2, 5, 1, 3, 4},
			wanOutput: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pq := NewPriorityQueue[int](tc.less, 0)
			for _, v := range tc.input {
				pq.Enqueue(v)
			}
			v, ok := pq.Peek()
			assert.True(t, ok)
			assert.Equal(t, v, tc.wanOutput)
		})
	}
}

// 测试出队
func TestPriorityQueueDequeue(t *testing.T) {
	testCases := []struct {
		name       string
		less       func(a, b int) bool
		input      []int // 输入
		wanOutput  int   // 预期顶堆出队元素输出
		wanOutput1 int   // 预期堆顶出队后堆顶元素
	}{
		{
			name: "小顶堆出队",
			less: func(a, b int) bool { // 告诉堆算法：“在位置 i 和 j 的两个元素，谁应该排在前面？”
				return a < b // a < b 则返回true，对应最小堆
			},
			input:      []int{2, 5, 1, 3, 4},
			wanOutput:  1,
			wanOutput1: 2,
		},
		{
			name: "大顶堆出队",
			less: func(a, b int) bool { // 告诉堆算法：“在位置 i 和 j 的两个元素，谁应该排在前面？”
				return a > b // a > b 则返回true，对应最大堆
			},
			input:      []int{2, 5, 1, 3, 4},
			wanOutput:  5,
			wanOutput1: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pq := NewPriorityQueue[int](tc.less, 0)
			for _, v := range tc.input {
				pq.Enqueue(v)
			}
			ot, ok := pq.Dequeue()
			assert.True(t, ok)
			assert.Equal(t, ot, tc.wanOutput)
			v, ok := pq.Peek()
			assert.True(t, ok)
			assert.Equal(t, v, tc.wanOutput1)
		})
	}
}

func TestPriorityQueueEnqueues(t *testing.T) {
	testCases := []struct {
		name      string
		less      func(a, b int) bool
		input     []int // 输入
		wanOutput int   // 预期输出
	}{
		{
			name: "小顶堆入队",
			less: func(a, b int) bool { // 告诉堆算法：“在位置 i 和 j 的两个元素，谁应该排在前面？”
				return a < b // a < b 则返回true，对应最小堆
			},
			input:     []int{2, 5, 1, 3, 4},
			wanOutput: 1,
		},
		{
			name: "大顶堆入队",
			less: func(a, b int) bool { // 告诉堆算法：“在位置 i 和 j 的两个元素，谁应该排在前面？”
				return a > b // a > b 则返回true，对应最大堆
			},
			input:     []int{2, 5, 1, 3, 4},
			wanOutput: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pq := NewPriorityQueue[int](tc.less, 0)
			pq.EnqueueBatch(tc.input)
			v, ok := pq.Peek()
			assert.True(t, ok)
			assert.Equal(t, v, tc.wanOutput)
		})
	}
}
