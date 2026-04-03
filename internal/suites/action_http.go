package suites

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func doHTTPGetQuery(t *testing.T, url string) []byte {
	t.Helper()

	client := NewHTTPClient()
	req, err := http.NewRequest(fasthttp.MethodGet, url, nil)
	assert.NoError(t, err)

	req.Header.Add(fasthttp.HeaderAccept, "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	return body
}
