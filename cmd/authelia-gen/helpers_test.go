package main

import (
	"net/mail"
	"reflect"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestGetPFlagPath(t *testing.T) {
	testCases := []struct {
		name     string
		have     func(t *testing.T) *pflag.FlagSet
		names    []string
		expected string
		err      string
	}{
		{
			"ShouldFailEmptyFlagSet",
			func(t *testing.T) *pflag.FlagSet {
				return pflag.NewFlagSet("example", pflag.ContinueOnError)
			},
			[]string{"abc", "123"},
			"",
			"failed to lookup flag 'abc': flag accessed but not defined: abc",
		},
		{
			"ShouldFailEmptyFlagNames",
			func(t *testing.T) *pflag.FlagSet {
				return pflag.NewFlagSet("example", pflag.ContinueOnError)
			},
			nil,
			"",
			"no flag names",
		},
		{
			"ShouldLookupFlagNames",
			func(t *testing.T) *pflag.FlagSet {
				flagset := pflag.NewFlagSet("example", pflag.ContinueOnError)

				flagset.String("dir.one", "", "")
				flagset.String("dir.two", "", "")
				flagset.String("file.name", "", "")

				require.NoError(t, flagset.Parse([]string{"--dir.one=abc", "--dir.two=123", "--file.name=path.txt"}))

				return flagset
			},
			[]string{"dir.one", "dir.two", "file.name"},
			"abc/123/path.txt",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, theError := getPFlagPath(tc.have(t), tc.names...)

			if tc.err == "" {
				assert.NoError(t, theError)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.EqualError(t, theError, tc.err)
				assert.Equal(t, "", actual)
			}
		})
	}
}

func TestBuildCSP(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		ruleSets [][]CSPValue
		expected string
	}{
		{
			"ShouldParseDefault",
			codeCSPProductionDefaultSrc,
			[][]CSPValue{
				codeCSPValuesCommon,
				codeCSPValuesProduction,
			},
			"default-src 'self'; frame-src 'none'; object-src 'none'; style-src 'self' 'nonce-%s'; frame-ancestors 'none'; base-uri 'self'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, buildCSP(tc.have, tc.ruleSets...))
		})
	}
}

func TestContainsType(t *testing.T) {
	astring := ""

	testCases := []struct {
		name     string
		have     any
		expected bool
	}{
		{
			"ShouldContainMailAddress",
			mail.Address{},
			true,
		},
		{
			"ShouldContainSchemaAddressPtr",
			&schema.Address{},
			true,
		},
		{
			"ShouldNotContainString",
			astring,
			false,
		},
		{
			"ShouldNotContainStringPtr",
			&astring,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, containsType(reflect.TypeOf(tc.have), decodedTypes))
		})
	}
}

func TestReadTags(t *testing.T) {
	assert.NotPanics(t, func() {
		iReadTags("", reflect.TypeOf(schema.Configuration{}), false, false, false)
	})

	assert.NotPanics(t, func() {
		iReadTags("", reflect.TypeOf(schema.Configuration{}), true, true, false)
	})
}
