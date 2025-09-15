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
			src:    int64(1),
			wanRes: "1",
		},
		// ..............
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := ToAny[string](tc.src)
			assert.Equal(t, tc.wanRes, s)
		})
	}
}
