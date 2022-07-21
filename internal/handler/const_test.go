package handler

import (
	"net/url"
	"time"
)

const (
	testInactivity     = time.Second * 10
	testRedirectionURL = "http://www.example.com"
	testUsername       = "john"
)

// MustParseURL is a test func.
func MustParseURL(u string) *url.URL {
	o, err := url.Parse(u)
	if err != nil {
		panic(err)
	}

	return o
}
