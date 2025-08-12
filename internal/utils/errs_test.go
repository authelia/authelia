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
			name:     "ShouldReturnZeroForEmptySlice",
			input:    nil,
			expected: 0,
		},
		{
			name:     "ShouldReturnCorrectLengthForNonEmptySlice",
			input:    []error{errors.New("a"), errors.New("b"), errors.New("c")},
			expected: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := ErrSliceSortAlphabetical(tc.input)
			assert.Equal(t, tc.expected, s.Len())
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
			name:     "ShouldReturnTrueWhenIComesBeforeJAlphabetically",
			input:    []error{errors.New("apple"), errors.New("banana")},
			i:        0,
			j:        1,
			expected: true,
		},
		{
			name:     "ShouldReturnFalseWhenIEqualsJValue",
			input:    []error{errors.New("same"), errors.New("same")},
			i:        0,
			j:        1,
			expected: false,
		},
		{
			name:     "ShouldReturnFalseWhenIComesAfterJAlphabetically",
			input:    []error{errors.New("pear"), errors.New("orange")},
			i:        0,
			j:        1,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := ErrSliceSortAlphabetical(tc.input)
			got := s.Less(tc.i, tc.j)
			assert.Equal(t, tc.expected, got)
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
			name:     "ShouldHandleSwapOnSameIndexWithoutChange",
			input:    []error{errors.New("x"), errors.New("y")},
			i:        1,
			j:        1,
			expected: []string{"x", "y"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpy := append([]error(nil), tc.input...)
			s := ErrSliceSortAlphabetical(cpy)

			s.Swap(tc.i, tc.j)

			got := make([]string, len(s))
			for idx, e := range s {
				got[idx] = e.Error()
			}

			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestErrSliceSortAlphabetical_SortIntegration(t *testing.T) {
	testCases := []struct {
		name     string
		input    []error
		expected []string
	}{
		{
			name:     "ShouldSortAlphabetically",
			input:    []error{errors.New("delta"), errors.New("alpha"), errors.New("charlie"), errors.New("bravo")},
			expected: []string{"alpha", "bravo", "charlie", "delta"},
		},
		{
			name:     "ShouldHandleAlreadySorted",
			input:    []error{errors.New("a"), errors.New("b")},
			expected: []string{"a", "b"},
		},
		{
			name:     "ShouldRespectByteOrderForCaseDifferences",
			input:    []error{errors.New("car"), errors.New("Car")},
			expected: []string{"Car", "car"},
		},
		{
			name:     "ShouldHandleNilSlice",
			input:    nil,
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpy := append([]error(nil), tc.input...)
			s := ErrSliceSortAlphabetical(cpy)
			sort.Sort(s)

			got := make([]string, len(s))
			for i, e := range s {
				got[i] = e.Error()
			}

			assert.Equal(t, tc.expected, got)
		})
	}
}
