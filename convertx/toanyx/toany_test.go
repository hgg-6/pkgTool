package toanyx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToAny(t *testing.T) {
	testCases := []struct {
		name string
		src  any

		wanRes any
	}{
		{
			name:   "IntToString ok",
			src:    map[string]string{"a": "1", "b": "2"},
			wanRes: map[string]string{"a": "1", "b": "2"},
		},
		// ...............
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, ok := ToAny[[]map[string]any](tc.src)
			if ok {
				assert.Equal(t, tc.wanRes, s)
				t.Log(s)
			}
		})
	}
}
