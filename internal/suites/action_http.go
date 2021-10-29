package suites

import (
	"io"
	"net/http"
	"testing"

	"github.com/matryer/is"
)

func doHTTPGetQuery(t *testing.T, url string) []byte {
	is := is.New(t)
	client := NewHTTPClient()
	req, err := http.NewRequest("GET", url, nil)
	is.NoErr(err)

	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	is.NoErr(err)

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	return body
}
