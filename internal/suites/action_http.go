package suites

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"

	"github.com/stretchr/testify/assert"
)

func doHTTPGetQuery(s *SeleniumSuite, url string) []byte {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(s.T(), err)

	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	assert.NoError(s.T(), err)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}
