package handlers

import (
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	testRequestMethods = []string{
		fasthttp.MethodOptions, fasthttp.MethodHead, fasthttp.MethodGet,
		fasthttp.MethodDelete, fasthttp.MethodPatch, fasthttp.MethodPost,
		fasthttp.MethodPut, fasthttp.MethodConnect, fasthttp.MethodTrace,
	}

	testXHR = map[string]bool{
		testWithoutAccept: false,
		testWithXHRHeader: true,
	}
)

const (
	testXOriginalMethod = "X-Original-Method"
	testXOriginalUrl    = "X-Original-URL"
	testBypass          = "bypass"
	testWithoutAccept   = "WithoutAccept"
	testWithXHRHeader   = "WithXHRHeader"
)

const (
	testBASE32TOTPSecret = "JVHFEUBXJ5CUWN2GGZGDMTKSJNMEQN2YGRJUQM2OKRHECR2QKJGFGRSQJVEVUT2HII2FQSJTKNIVQSCPIJIQ====" //nolint:gosec // Test only credential.
)

const (
	testInactivity           = time.Second * 10
	testRedirectionURLString = "https://www.example.com"
	testUsername             = "john"
	testDisplayName          = "john"
	exampleDotCom            = "example.com"
)

var (
	testRedirectionURL = func() *url.URL {
		u, err := url.ParseRequestURI(testRedirectionURLString)
		if err != nil {
			panic(err)
		}

		return u
	}()
)
