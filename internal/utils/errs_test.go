package utils

import (
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrSliceSortAlphabetical_Len(t *testing.T) {
	testCases := []struct {
		name     string
		input    []error
		expected int
	}{
		{
			"ShouldReturnZeroForEmptySlice",
			nil,
			0,
		},
		{
			"ShouldReturnCorrectLengthForNonEmptySlice",
			[]error{errors.New("a"), errors.New("b"), errors.New("c")},
			3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ErrSliceSortAlphabetical(tc.input).Len())
		})
	}
}

func TestErrSliceSortAlphabetical_Less(t *testing.T) {
	testCases := []struct {
		name     string
		input    []error
		i, j     int
		expected bool
	}{
		{
			"ShouldReturnTrueWhenIComesBeforeJAlphabetically",
			[]error{errors.New("apple"), errors.New("banana")},
			0,
			1,
			true,
		},
		{
			"ShouldReturnFalseWhenIEqualsJValue",
			[]error{errors.New("same"), errors.New("same")},
			0,
			1,
			false,
		},
		{
			"ShouldReturnFalseWhenIComesAfterJAlphabetically",
			[]error{errors.New("pear"), errors.New("orange")},
			0,
			1,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ErrSliceSortAlphabetical(tc.input).Less(tc.i, tc.j))
		})
	}
}

func TestErrSliceSortAlphabetical_Swap(t *testing.T) {
	testCases := []struct {
		name     string
		input    []error
		i, j     int
		expected []string
	}{
		{
			name:     "ShouldSwapElementsAtIndices",
			input:    []error{errors.New("a"), errors.New("b"), errors.New("c")},
			i:        0,
			j:        2,
			expected: []string{"c", "b", "a"},
		},
		{
			"ShouldHandleSwapOnSameIndexWithoutChange",
			[]error{errors.New("x"), errors.New("y")},
			1,
			1,
			[]string{"x", "y"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpy := append([]error(nil), tc.input...)
			have := ErrSliceSortAlphabetical(cpy)

			have.Swap(tc.i, tc.j)

			actual := make([]string, len(have))
			for i, e := range have {
				actual[i] = e.Error()
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestErrSliceSortAlphabetical_SortIntegration(t *testing.T) {
	testCases := []struct {
		name     string
		have     []error
		expected []string
	}{
		{
			"ShouldSortAlphabetically",
			[]error{errors.New("delta"), errors.New("alpha"), errors.New("charlie"), errors.New("bravo")},
			[]string{"alpha", "bravo", "charlie", "delta"},
		},
		{
			"ShouldHandleAlreadySorted",
			[]error{errors.New("a"), errors.New("b")},
			[]string{"a", "b"},
		},
		{
			"ShouldRespectByteOrderForCaseDifferences",
			[]error{errors.New("car"), errors.New("Car")},
			[]string{"Car", "car"},
		},
		{
			"ShouldHandleNilSlice",
			nil,
			[]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			have := ErrSliceSortAlphabetical(tc.have)
			sort.Sort(have)

			actual := make([]string, len(have))
			for i, err := range have {
				actual[i] = err.Error()
			}

			assert.Equal(t, tc.expected, actual)
		})
	}
}
