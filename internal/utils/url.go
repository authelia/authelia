package utils

import (
	"net/url"
	"path"
	"strings"
)

// URLPathFullClean returns a URL path with the query parameters appended (full path) with the path portion parsed
// through path.Clean given a *url.URL.
func URLPathFullClean(u *url.URL) string {
	b := strings.Builder{}

	b.WriteString(path.Clean(u.Path))

	if len(u.Path) != 1 && u.Path[len(u.Path)-1] == '/' {
		b.Write(slashForward)
	}

	if u.RawQuery != "" {
		b.Write(questionMark)
		b.WriteString(u.RawQuery)
	}

	return b.String()
}
