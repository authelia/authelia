package authorization

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// Subject subject who to check access control for.
type Subject struct {
	Username string
	Groups   []string
	IP       net.IP
}

func (s Subject) String() string {
	return fmt.Sprintf("username=%s groups=%s ip=%s", s.Username, strings.Join(s.Groups, ","), s.IP.String())
}

// Object object to check access control for.
type Object struct {
	Scheme string
	Domain string
	Path   string
	Method string
}

func (o Object) String() string {
	return fmt.Sprintf("%s://%s%s", o.Scheme, o.Domain, o.Path)
}

func NewObject(targetUrl *url.URL, method []byte) (object Object) {
	object = Object{
		Scheme: targetUrl.Scheme,
		Domain: targetUrl.Hostname(),
		Method: string(method),
	}

	if targetUrl.RawQuery == "" {
		object.Path = targetUrl.Path
	} else {
		object.Path = targetUrl.Path + "?" + targetUrl.RawQuery
	}

	return object
}