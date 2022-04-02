package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBoolCountLessThanN(t *testing.T) {
	type h struct {
		n    int
		v    bool
		vals []bool
	}

	testCases := []struct {
		have h
		want bool
	}{
		{
			have: h{
				n:    1,
				v:    true,
				vals: []bool{true, false, false},
			},
			want: true,
		},
		{
			have: h{
				n:    1,
				v:    true,
				vals: []bool{true, true, false},
			},
			want: false,
		},
		{
			have: h{
				n:    2,
				v:    true,
				vals: []bool{true, true, false},
			},
			want: true,
		},
		{
			have: h{
				n:    2,
				v:    true,
				vals: []bool{true, true, true},
			},
			want: false,
		},
		{
			have: h{
				n:    300,
				v:    true,
				vals: []bool{true, true, true},
			},
			want: true,
		},
		{
			have: h{
				n:    300,
				v:    false,
				vals: []bool{true, true, true},
			},
			want: true,
		},
		{
			have: h{
				n:    2,
				v:    false,
				vals: []bool{false, false, false},
			},
			want: false,
		},
		{
			have: h{
				n:    1,
				v:    false,
				vals: []bool{false, true, true},
			},
			want: true,
		},
		{
			have: h{
				n: 20,
				v: false,
				vals: []bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
					true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
					true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
					true, true, true, true, true},
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		countTrue := 0
		countFalse := 0

		for _, v := range tc.have.vals {
			if v {
				countTrue++
			} else {
				countFalse++
			}
		}

		t.Run(fmt.Sprintf("%d %t true(%d)-false(%d)/should be %t", tc.have.n, tc.have.v, countTrue, countFalse, tc.want), func(t *testing.T) {
			assert.Equal(t, tc.want, IsBoolCountLessThanN(tc.have.n, tc.have.v, tc.have.vals...))
		})
	}
}
