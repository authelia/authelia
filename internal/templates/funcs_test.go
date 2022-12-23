package templates

import (
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFuncGetEnv(t *testing.T) {
	testCases := []struct {
		name     string
		have     map[string]string
		expected map[string]string
	}{
		{"ShouldGetEnv",
			map[string]string{
				"AN_ENV":      "a",
				"ANOTHER_ENV": "b",
			},
			map[string]string{
				"AN_ENV":      "a",
				"ANOTHER_ENV": "b",
			},
		},
		{"ShouldNotGetSecretEnv",
			map[string]string{
				"AUTHELIA_ENV_SECRET": "a",
				"ANOTHER_ENV":         "b",
			},
			map[string]string{
				"AUTHELIA_ENV_SECRET": "",
				"ANOTHER_ENV":         "b",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for key, value := range tc.have {
				assert.NoError(t, os.Setenv(key, value))
			}

			for key, expected := range tc.expected {
				assert.Equal(t, expected, FuncGetEnv(key))
			}

			for key := range tc.have {
				assert.NoError(t, os.Unsetenv(key))
			}
		})
	}
}

func TestFuncExpandEnv(t *testing.T) {
	testCases := []struct {
		name     string
		env      map[string]string
		have     string
		expected string
	}{
		{"ShouldExpandEnv",
			map[string]string{
				"AN_ENV":      "a",
				"ANOTHER_ENV": "b",
			},
			"This is ${AN_ENV} and ${ANOTHER_ENV}",
			"This is a and b",
		},
		{"ShouldNotExpandSecretEnv",
			map[string]string{
				"AUTHELIA_ENV_SECRET":   "a",
				"X_AUTHELIA_ENV_SECRET": "a",
				"ANOTHER_ENV":           "b",
			},
			"This is ${AUTHELIA_ENV_SECRET} and ${ANOTHER_ENV} without ${X_AUTHELIA_ENV_SECRET}",
			"This is  and b without ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for key, value := range tc.env {
				assert.NoError(t, os.Setenv(key, value))
			}

			assert.Equal(t, tc.expected, FuncExpandEnv(tc.have))

			for key := range tc.env {
				assert.NoError(t, os.Unsetenv(key))
			}
		})
	}
}

func TestFuncHashSum(t *testing.T) {
	testCases := []struct {
		name     string
		new      func() hash.Hash
		have     []string
		expected []string
	}{
		{"ShouldHashSHA1", sha1.New, []string{"abc", "123", "authelia"}, []string{"616263da39a3ee5e6b4b0d3255bfef95601890afd80709", "313233da39a3ee5e6b4b0d3255bfef95601890afd80709", "61757468656c6961da39a3ee5e6b4b0d3255bfef95601890afd80709"}},
		{"ShouldHashSHA256", sha256.New, []string{"abc", "123", "authelia"}, []string{"616263e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "313233e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "61757468656c6961e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}},
		{"ShouldHashSHA512", sha512.New, []string{"abc", "123", "authelia"}, []string{"616263cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", "313233cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", "61757468656c6961cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, len(tc.have), len(tc.expected))

			h := FuncHashSum(tc.new)

			for i := 0; i < len(tc.have); i++ {
				assert.Equal(t, tc.expected[i], h(tc.have[i]))
			}
		})
	}
}

func TestFuncStringReplace(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		old, new string
		expected string
	}{
		{"ShouldReplaceSingle", "ABC123", "123", "456", "ABC456"},
		{"ShouldReplaceMultiple", "123ABC123123", "123", "456", "456ABC456456"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringReplace(tc.old, tc.new, tc.have))
		})
	}
}

func TestFuncStringContains(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		substr   string
		expected bool
	}{
		{"ShouldMatchNormal", "abc123", "c12", true},
		{"ShouldNotMatchWrongCase", "abc123", "C12", false},
		{"ShouldNotMatchNotContains", "abc123", "xyz", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringContains(tc.substr, tc.have))
		})
	}
}

func TestFuncStringHasPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		substr   string
		expected bool
	}{
		{"ShouldMatchNormal", "abc123", "abc", true},
		{"ShouldNotMatchWrongCase", "abc123", "ABC", false},
		{"ShouldNotMatchNotPrefix", "abc123", "123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringHasPrefix(tc.substr, tc.have))
		})
	}
}

func TestFuncStringHasSuffix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		substr   string
		expected bool
	}{
		{"ShouldMatchNormal", "abc123xyz", "xyz", true},
		{"ShouldNotMatchWrongCase", "abc123xyz", "XYZ", false},
		{"ShouldNotMatchNotSuffix", "abc123xyz", "123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringHasSuffix(tc.substr, tc.have))
		})
	}
}

func TestFuncStringTrimAll(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		cutset   string
		expected string
	}{
		{"ShouldTrimSuffix", "abc123xyz", "xyz", "abc123"},
		{"ShouldTrimPrefix", "xyzabc123", "xyz", "abc123"},
		{"ShouldNotTrimMiddle", "abcxyz123", "xyz", "abcxyz123"},
		{"ShouldNotTrimWrongCase", "xyzabcxyz123xyz", "XYZ", "xyzabcxyz123xyz"},
		{"ShouldNotTrimWrongChars", "abc123xyz", "456", "abc123xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringTrimAll(tc.cutset, tc.have))
		})
	}
}

func TestFuncStringTrimPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		cutset   string
		expected string
	}{
		{"ShouldNotTrimSuffix", "abc123xyz", "xyz", "abc123xyz"},
		{"ShouldTrimPrefix", "xyzabc123", "xyz", "abc123"},
		{"ShouldNotTrimMiddle", "abcxyz123", "xyz", "abcxyz123"},
		{"ShouldNotTrimWrongCase", "xyzabcxyz123xyz", "XYZ", "xyzabcxyz123xyz"},
		{"ShouldNotTrimWrongChars", "abc123xyz", "456", "abc123xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringTrimPrefix(tc.cutset, tc.have))
		})
	}
}

func TestFuncStringTrimSuffix(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		cutset   string
		expected string
	}{
		{"ShouldTrimSuffix", "abc123xyz", "xyz", "abc123"},
		{"ShouldNotTrimPrefix", "xyzabc123", "xyz", "xyzabc123"},
		{"ShouldNotTrimMiddle", "abcxyz123", "xyz", "abcxyz123"},
		{"ShouldNotTrimWrongCase", "xyzabcxyz123xyz", "XYZ", "xyzabcxyz123xyz"},
		{"ShouldNotTrimWrongChars", "abc123xyz", "456", "abc123xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringTrimSuffix(tc.cutset, tc.have))
		})
	}
}

func TestFuncElemsJoin(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		sep      string
		expected string
	}{
		{"ShouldNotJoinNonElements", "abc123xyz", "xyz", "abc123xyz"},
		{"ShouldJoinStrings", []string{"abc", "123"}, "xyz", "abcxyz123"},
		{"ShouldJoinInts", []int{1, 2, 3}, ",", "1,2,3"},
		{"ShouldJoinBooleans", []bool{true, false, true}, ".", "true.false.true"},
		{"ShouldJoinBytes", [][]byte{[]byte("abc"), []byte("123"), []byte("a")}, "$", "abc$123$a"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncElemsJoin(tc.sep, tc.have))
		})
	}
}

func TestFuncIterate(t *testing.T) {
	testCases := []struct {
		name     string
		have     uint
		expected []uint
	}{
		{"ShouldGiveZeroResults", 0, nil},
		{"ShouldGive10Results", 10, []uint{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncIterate(&tc.have))
		})
	}
}

func TestFuncStringsSplit(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		sep      string
		expected map[string]string
	}{
		{"ShouldSplit", "abc,123,456", ",", map[string]string{"_0": "abc", "_1": "123", "_2": "456"}},
		{"ShouldNotSplit", "abc,123,456", "$", map[string]string{"_0": "abc,123,456"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringSplit(tc.sep, tc.have))
		})
	}
}

func TestFuncStringSplitList(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		sep      string
		expected []string
	}{
		{"ShouldSplit", "abc,123,456", ",", []string{"abc", "123", "456"}},
		{"ShouldNotSplit", "abc,123,456", "$", []string{"abc,123,456"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, FuncStringSplitList(tc.sep, tc.have))
		})
	}
}
