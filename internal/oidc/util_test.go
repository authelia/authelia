package oidc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortedSigningAlgs(t *testing.T) {
	testCases := []struct {
		name     string
		have     SortedSigningAlgs
		expected SortedSigningAlgs
	}{
		{
			"ShouldSortRSABeforeECDSA",
			SortedSigningAlgs{SigningAlgECDSAUsingP256AndSHA256, SigningAlgRSAUsingSHA256},
			SortedSigningAlgs{SigningAlgRSAUsingSHA256, SigningAlgECDSAUsingP256AndSHA256},
		},
		{
			"ShouldSortHMACBeforeRSA",
			SortedSigningAlgs{SigningAlgRSAUsingSHA256, SigningAlgHMACUsingSHA256},
			SortedSigningAlgs{SigningAlgHMACUsingSHA256, SigningAlgRSAUsingSHA256},
		},
		{
			"ShouldSortNoneLast",
			SortedSigningAlgs{SigningAlgNone, SigningAlgRSAUsingSHA256, SigningAlgHMACUsingSHA256},
			SortedSigningAlgs{SigningAlgHMACUsingSHA256, SigningAlgRSAUsingSHA256, SigningAlgNone},
		},
		{
			"ShouldHandleAlreadySorted",
			SortedSigningAlgs{SigningAlgHMACUsingSHA256, SigningAlgRSAUsingSHA256},
			SortedSigningAlgs{SigningAlgHMACUsingSHA256, SigningAlgRSAUsingSHA256},
		},
		{
			"ShouldHandleSingleElement",
			SortedSigningAlgs{SigningAlgRSAUsingSHA256},
			SortedSigningAlgs{SigningAlgRSAUsingSHA256},
		},
		{
			"ShouldHandleEmpty",
			SortedSigningAlgs{},
			SortedSigningAlgs{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(tc.have)

			assert.Equal(t, tc.expected, tc.have)
		})
	}
}

func TestIsSigningAlgLess(t *testing.T) {
	testCases := []struct {
		name     string
		i        string
		j        string
		expected bool
	}{
		{"ShouldReturnFalseForEqual", SigningAlgRSAUsingSHA256, SigningAlgRSAUsingSHA256, false},
		{"ShouldReturnFalseRSABeforeHMAC", SigningAlgRSAUsingSHA256, SigningAlgHMACUsingSHA256, false},
		{"ShouldReturnTrueHMACBeforeNone", SigningAlgHMACUsingSHA256, SigningAlgNone, true},
		{"ShouldReturnTrueHMACBeforeRSA", SigningAlgHMACUsingSHA256, SigningAlgRSAUsingSHA512, true},
		{"ShouldReturnTrueHMACBeforeRSAPSS", SigningAlgHMACUsingSHA256, SigningAlgRSAPSSUsingSHA256, true},
		{"ShouldReturnTrueHMACBeforeECDSA", SigningAlgHMACUsingSHA256, SigningAlgECDSAUsingP521AndSHA512, true},
		{"ShouldReturnTrueRSABeforeECDSA", SigningAlgRSAUsingSHA256, SigningAlgECDSAUsingP521AndSHA512, true},
		{"ShouldReturnTrueECDSABeforeUnknown", SigningAlgECDSAUsingP521AndSHA512, "JS121", true},
		{"ShouldReturnFalseUnknownBeforeECDSA", "JS121", SigningAlgECDSAUsingP521AndSHA512, false},
		{"ShouldReturnFalseUnknownPair", "JS121", "TS512", false},
		{"ShouldReturnFalseNoneBeforeRSA", SigningAlgNone, SigningAlgRSAUsingSHA256, false},
		{"ShouldReturnFalseNoneBeforeHMAC", SigningAlgNone, SigningAlgHMACUsingSHA256, false},
		{"ShouldReturnFalseNoneBeforeECDSA", SigningAlgNone, SigningAlgECDSAUsingP256AndSHA256, false},
		{"ShouldReturnFalseRSAPSSBeforeRSA", SigningAlgRSAPSSUsingSHA256, SigningAlgRSAUsingSHA256, false},
		{"ShouldReturnFalseRSAPSSBeforeECDSA", SigningAlgRSAPSSUsingSHA256, SigningAlgECDSAUsingP256AndSHA256, false},
		{"ShouldReturnTrueRSABeforeRSAPSS", SigningAlgRSAUsingSHA256, SigningAlgRSAPSSUsingSHA256, true},
		{"ShouldReturnTrueECDSABeforeRSAPSS", SigningAlgECDSAUsingP256AndSHA256, SigningAlgRSAPSSUsingSHA256, true},
		{"ShouldReturnTrueSamePrefixLexicographic", SigningAlgRSAUsingSHA256, SigningAlgRSAUsingSHA512, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, isSigningAlgLess(tc.i, tc.j))
		})
	}
}

func TestToStringSlice(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		expected []string
	}{
		{
			"ShouldParseStringSlice",
			[]string{"abc", "123"},
			[]string{"abc", "123"},
		},
		{
			"ShouldParseAnySlice",
			[]any{"abc", "123"},
			[]string{"abc", "123"},
		},
		{
			"ShouldParseAnySlice",
			"abc",
			[]string{"abc"},
		},
		{
			"ShouldParseNil",
			nil,
			nil,
		},
		{
			"ShouldParseInt",
			5,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, toStringSlice(tc.have))
		})
	}
}

func TestToTime(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		def      time.Time
		expected time.Time
	}{
		{
			"ShouldParseFloat64",
			float64(123),
			time.Unix(0, 0).UTC(),
			time.Unix(123, 0).UTC(),
		},
		{
			"ShouldParseInt64",
			int64(123),
			time.Unix(0, 0).UTC(),
			time.Unix(123, 0).UTC(),
		},
		{
			"ShouldParseInt",
			123,
			time.Unix(0, 0).UTC(),
			time.Unix(123, 0).UTC(),
		},
		{
			"ShouldParseTime",
			time.Unix(1235, 0).UTC(),
			time.Unix(0, 0).UTC(),
			time.Unix(1235, 0).UTC(),
		},
		{
			"ShouldReturnDefault",
			"abc",
			time.Unix(0, 0).UTC(),
			time.Unix(0, 0).UTC(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, toTime(tc.have, tc.def))
		})
	}
}

func TestGetSectorIdentifierURICache(t *testing.T) {
	testCases := []struct {
		name         string
		setup        func(t *testing.T) (*url.URL, *http.Client, map[string][]string)
		expected     []string
		err          string
		expectCached bool
	}{
		{
			"ShouldReturnFromCache",
			func(t *testing.T) (*url.URL, *http.Client, map[string][]string) {
				u, _ := url.Parse("https://example.com/sector")

				cache := map[string][]string{
					u.String(): {"https://app.example.com/callback"},
				}

				return u, nil, cache
			},
			[]string{"https://app.example.com/callback"},
			"",
			false,
		},
		{
			"ShouldFetchFromServer",
			func(t *testing.T) (*url.URL, *http.Client, map[string][]string) {
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode([]string{"https://app.example.com/callback"})
				}))

				t.Cleanup(srv.Close)

				u, _ := url.Parse(srv.URL)

				return u, srv.Client(), map[string][]string{}
			},
			[]string{"https://app.example.com/callback"},
			"",
			true,
		},
		{
			"ShouldPopulateCache",
			func(t *testing.T) (*url.URL, *http.Client, map[string][]string) {
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode([]string{"https://app.example.com/cb1", "https://app.example.com/cb2"})
				}))

				t.Cleanup(srv.Close)

				u, _ := url.Parse(srv.URL)

				return u, srv.Client(), map[string][]string{}
			},
			[]string{"https://app.example.com/cb1", "https://app.example.com/cb2"},
			"",
			true,
		},
		{
			"ShouldErrOnServerFailure",
			func(t *testing.T) (*url.URL, *http.Client, map[string][]string) {
				u, _ := url.Parse("https://127.0.0.1:1/sector")

				return u, &http.Client{Timeout: 50 * time.Millisecond}, nil
			},
			nil,
			"error occurred making request",
			false,
		},
		{
			"ShouldErrOnInvalidJSON",
			func(t *testing.T) (*url.URL, *http.Client, map[string][]string) {
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("not json"))
				}))

				t.Cleanup(srv.Close)

				u, _ := url.Parse(srv.URL)

				return u, srv.Client(), nil
			},
			nil,
			"error occurred decoding request",
			false,
		},
		{
			"ShouldFetchWithNilCache",
			func(t *testing.T) (*url.URL, *http.Client, map[string][]string) {
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode([]string{"https://app.example.com/callback"})
				}))

				t.Cleanup(srv.Close)

				u, _ := url.Parse(srv.URL)

				return u, srv.Client(), nil
			},
			[]string{"https://app.example.com/callback"},
			"",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sectorURI, client, cache := tc.setup(t)

			ctx := &testClientContext{client: client}

			result, err := getSectorIdentifierURICache(ctx, cache, sectorURI)

			if tc.err != "" {
				assert.ErrorContains(t, err, tc.err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}

			if tc.expectCached && cache != nil {
				cached, ok := cache[sectorURI.String()]

				assert.True(t, ok)
				assert.Equal(t, tc.expected, cached)
			}
		})
	}
}

func TestValidateSectorIdentifierURI(t *testing.T) {
	sectorURI, _ := url.Parse("https://example.com/sector")

	testCases := []struct {
		name         string
		cache        map[string][]string
		redirectURIs []string
		err          string
	}{
		{
			"ShouldSucceedAllRedirectURIsMatch",
			map[string][]string{
				sectorURI.String(): {"https://app.example.com/callback"},
			},
			[]string{"https://app.example.com/callback"},
			"",
		},
		{
			"ShouldSucceedMultipleRedirectURIsAllMatch",
			map[string][]string{
				sectorURI.String(): {"https://app.example.com/cb1", "https://app.example.com/cb2"},
			},
			[]string{"https://app.example.com/cb1", "https://app.example.com/cb2"},
			"",
		},
		{
			"ShouldSucceedNoRedirectURIs",
			map[string][]string{
				sectorURI.String(): {"https://app.example.com/callback"},
			},
			nil,
			"",
		},
		{
			"ShouldErrSingleRedirectURIMismatch",
			map[string][]string{
				sectorURI.String(): {"https://other.example.com/callback"},
			},
			[]string{"https://app.example.com/callback"},
			"error checking redirect_uri 'https://app.example.com/callback' against ''https://other.example.com/callback''",
		},
		{
			"ShouldErrMultipleRedirectURIsMismatch",
			map[string][]string{
				sectorURI.String(): {"https://other.example.com/callback"},
			},
			[]string{"https://app.example.com/cb1", "https://app.example.com/cb2"},
			"error checking redirect_uris ''https://app.example.com/cb1' and 'https://app.example.com/cb2'' against ''https://other.example.com/callback''",
		},
		{
			"ShouldErrPartialMismatch",
			map[string][]string{
				sectorURI.String(): {"https://app.example.com/cb1"},
			},
			[]string{"https://app.example.com/cb1", "https://app.example.com/cb2"},
			"error checking redirect_uri 'https://app.example.com/cb2' against ''https://app.example.com/cb1''",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &testClientContext{}

			err := ValidateSectorIdentifierURI(ctx, tc.cache, sectorURI, tc.redirectURIs)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestFloat64Match(t *testing.T) {
	testCases := []struct {
		name     string
		expected float64
		value    any
		values   []any
		result   bool
	}{
		{
			"ShouldMatchFloat64Value",
			1.0,
			float64(1.0),
			nil,
			true,
		},
		{
			"ShouldMatchInt64Value",
			5.0,
			int64(5),
			nil,
			true,
		},
		{
			"ShouldNotMatchDifferentValue",
			1.0,
			float64(2.0),
			nil,
			false,
		},
		{
			"ShouldMatchInValues",
			3.0,
			nil,
			[]any{float64(1.0), float64(3.0)},
			true,
		},
		{
			"ShouldNotMatchInValuesWhenAbsent",
			5.0,
			nil,
			[]any{float64(1.0), float64(3.0)},
			false,
		},
		{
			"ShouldReturnFalseForNilValueAndEmptyValues",
			1.0,
			nil,
			nil,
			false,
		},
		{
			"ShouldReturnFalseForNonNumericValue",
			1.0,
			"not a number",
			nil,
			false,
		},
		{
			"ShouldReturnFalseForNonNumericValues",
			1.0,
			nil,
			[]any{"not a number"},
			false,
		},
		{
			"ShouldPreferValueOverValues",
			1.0,
			float64(1.0),
			[]any{float64(2.0)},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.result, float64Match(tc.expected, tc.value, tc.values))
		})
	}
}

type testClientContext struct {
	context.Context
	client *http.Client
}

func (c *testClientContext) GetHTTPClient() *http.Client {
	return c.client
}
