package handlers

import (
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
	testXOriginalUrl    = "X-Original-Url"
	testBypass          = "bypass"
	testWithoutAccept   = "WithoutAccept"
	testWithXHRHeader   = "WithXHRHeader"
)

const (
	testInactivity     = time.Second * 10
	testRedirectionURL = "http://redirection.local"
	testUsername       = "john"
	exampleDotCom      = "example.com"
)
