package server

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestGenerateEtag(t *testing.T) {
	testCases := []struct {
		name      string
		payloadA  []byte
		payloadB  []byte
		wantEqual bool
	}{
		{
			name:      "ShouldReturnSameEtagForSamePayload",
			payloadA:  []byte("hello world"),
			payloadB:  []byte("hello world"),
			wantEqual: true,
		},
		{
			name:      "ShouldReturnDifferentEtagForDifferentPayload",
			payloadA:  []byte("hello world"),
			payloadB:  []byte("HELLO WORLD"),
			wantEqual: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			etagA := generateEtag(tc.payloadA)
			etagB := generateEtag(tc.payloadB)

			if tc.wantEqual {
				assert.Equal(t, etagA, etagB, "etags should be equal for identical payloads")
			} else {
				assert.NotEqual(t, etagA, etagB, "etags should differ for different payloads")
			}

			assert.Len(t, etagA, 40, "etag should be 40 characters (sha1 hex)")
		})
	}
}

func TestGetEmbedETags(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "ShouldComputeETagsForEmbeddedLocalesRecursively",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			etags := map[string][]byte{}

			getEmbedETags(locales, "locales", etags)

			assert.Greater(t, len(etags), 0, "expected at least one embedded locale file to have an etag")

			for p, etag := range etags {
				data, err := locales.ReadFile(p)
				assert.NoError(t, err, "should be able to read embedded file %s", p)
				assert.Equal(t, generateEtag(data), etag, "etag for %s should match generateEtag(data)", p)

				break
			}
		})
	}
}

func TestHFSHandleErr(t *testing.T) {
	testCases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "ShouldSet404ForNotExist",
			err:        fs.ErrNotExist,
			wantStatus: fasthttp.StatusNotFound,
		},
		{
			name:       "ShouldSet403ForPermission",
			err:        fs.ErrPermission,
			wantStatus: fasthttp.StatusForbidden,
		},
		{
			name:       "ShouldSet500ForOtherErrors",
			err:        errors.New("some other error"),
			wantStatus: fasthttp.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ctx fasthttp.RequestCtx

			hfsHandleErr(&ctx, tc.err)

			assert.Equal(t, tc.wantStatus, ctx.Response.StatusCode())
		})
	}
}
