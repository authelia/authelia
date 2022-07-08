package utils

import (
	"net/url"
	"path"
)

// URLPathFullClean returns a URL path with the query parameters appended (full path) with the path portion parsed
// through path.Clean given a *url.URL.
func URLPathFullClean(u *url.URL) (p string) {
	switch len(u.RawQuery) {
	case 0:
		return path.Clean(u.Path)
	default:
		return path.Clean(u.Path) + "?" + u.RawQuery
	}
}
