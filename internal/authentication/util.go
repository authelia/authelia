package authentication

import "net/url"

func stringURL(uri *url.URL) string {
	if uri == nil {
		return ""
	}

	return uri.String()
}
