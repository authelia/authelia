package suites

import (
	"io"
	"net/http"

	"github.com/valyala/fasthttp"
)

func doHTTPGetQuery(url string) (body []byte, err error) {
	client := NewHTTPClient()

	req, err := http.NewRequest(fasthttp.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(fasthttp.HeaderAccept, "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
