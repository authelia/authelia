package suites

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func doHTTPGetQuery(t *testing.T, url string) []byte {
	client := NewHTTPClient()
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	return body
}
