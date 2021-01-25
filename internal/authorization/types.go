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

// NewObjectRaw creates a new Object type from a URL and a method header.
func NewObjectRaw(targetURL *url.URL, method []byte) (object Object) {
	return NewObject(targetURL, string(method))
}

// NewObject creates a new Object type from a URL and a method header.
func NewObject(targetURL *url.URL, method string) (object Object) {
	object = Object{
		Scheme: targetURL.Scheme,
		Domain: targetURL.Hostname(),
		Method: method,
	}

	if targetURL.RawQuery == "" {
		object.Path = targetURL.Path
	} else {
		object.Path = targetURL.Path + "?" + targetURL.RawQuery
	}

	return object
}
