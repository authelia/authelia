package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func PathWithQueryFromURI(uri *url.URL) (path string) {
	if len(uri.RawQuery) == 0 {
		return uri.EscapedPath()
	}

	return uri.EscapedPath() + "?" + uri.RawQuery
}

// ParseNormalizedRequestURI applies RFC 3986 §6.2 normalization to a request URI so that the resulting *url.URL can be
// safely compared against access control policies. It applies syntax-based normalization (§6.2.2) as well as the
// scheme-based normalization (§6.2.3) for the well known schemes (http, https, ws, wss).
func ParseNormalizedRequestURI(input string) (uri *url.URL, err error) {
	if uri, err = url.ParseRequestURI(input); err != nil {
		return nil, err
	}

	return NormalizedRequestURI(uri)
}

func NormalizedRequestURI(uri *url.URL) (out *url.URL, err error) {
	if uri.Opaque != "" {
		return nil, fmt.Errorf("cannot normalize opaque URI %q", uri.String())
	}

	// Case Normalization (§6.2.2.1): the scheme is case-insensitive and is normalized to lowercase. The host is left
	// with its original case at this point as doing so would alter the case-sensitive path matching which may be relied
	// on and case folding of the host is instead handled at comparison time by default. This gives users a clear
	// migration path if necessary.
	uri.Scheme = strings.ToLower(uri.Scheme)

	// Percent Encoding Normalization (§6.2.2.2) is handled for the Host, RawPath, Path, and RawQuery. The hex digits of
	// retained percent-encoded octets are uppercased (§6.2.2.1) and octets representing unreserved characters are
	// decoded.
	uri.Host = normalizeRequestURIEscaping(uri.Host)
	uri.RawPath = normalizeRequestURIEscaping(uri.EscapedPath())

	// Empty Path Segment Normalization: runs of unencoded path separators are collapsed into a single separator so that
	// empty path segments cannot be leveraged to evade access control comparison. This is performed before
	// remove_dot_segments below so a dot-segment split by empty segments (e.g. '/a/..//b') resolves as expected.
	uri.RawPath = normalizeRequestURIMergeSlashes(uri.RawPath)

	if uri.Path, err = url.PathUnescape(uri.RawPath); err != nil {
		return nil, err
	}

	uri.RawQuery = normalizeRequestURIEscaping(uri.RawQuery)

	// Scheme-Based Normalization (§6.2.3): remove an empty or default port for the well known schemes.
	uri.Host = normalizeRequestURIHostPort(uri.Scheme, uri.Host)

	// Path Segment Normalization (§6.2.2.3) is handled using ResolveReference which handles the strictest version of
	// remove_dot_segments (§5.2.4).
	uri = new(url.URL).ResolveReference(uri)

	// Scheme-Based Normalization (§6.2.3): an empty path on a URI with an authority is equivalent to '/'.
	if uri.Host != "" && uri.Path == "" {
		uri.Path = "/"
	}

	return uri, nil
}

func normalizeRequestURIHostPort(scheme, host string) string {
	i := strings.LastIndex(host, ":")

	if i < 0 {
		return host
	}

	if j := strings.LastIndex(host, "]"); j > i {
		return host
	}

	hostname, port := host[:i], host[i+1:]

	if port == "" || normalizeRequestURIIsSchemeDefaultPort(scheme, port) {
		return hostname
	}

	return host
}

func normalizeRequestURIIsSchemeDefaultPort(scheme, port string) bool {
	switch scheme {
	case http, ws:
		return port == "80"
	case https, wss:
		return port == "443"
	default:
		return false
	}
}

// normalizeRequestURIMergeSlashes collapses runs of consecutive unencoded path separators ('/') into a single
// separator. Percent-encoded slashes ("%2F") are intentionally left untouched as they represent a literal slash
// character within a segment rather than a separator between segments.
func normalizeRequestURIMergeSlashes(s string) string {
	if !strings.Contains(s, "//") {
		return s
	}

	var builder strings.Builder

	builder.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '/' && i+1 < len(s) && s[i+1] == '/' {
			continue
		}

		builder.WriteByte(s[i])
	}

	return builder.String()
}

// normalizeRequestURIEscaping uppercases percent-encoding hex digits (§6.2.2.1) and decodes percent-encoded octets that
// correspond to unreserved characters (§6.2.2.2). Reserved octets stay encoded.
func normalizeRequestURIEscaping(s string) string {
	var builder strings.Builder

	builder.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) && normalizeRequestURIIsHex(s[i+1]) && normalizeRequestURIIsHex(s[i+2]) {
			octet := normalizeRequestURIUnhex(s[i+1])<<4 | normalizeRequestURIUnhex(s[i+2])

			if normalizeRequestURIIsUnreserved(octet) {
				builder.WriteByte(octet)
			} else {
				builder.WriteByte('%')
				builder.WriteByte(normalizeRequestURIUpperHex(s[i+1]))
				builder.WriteByte(normalizeRequestURIUpperHex(s[i+2]))
			}

			i += 2
		} else {
			builder.WriteByte(s[i])
		}
	}

	return builder.String()
}

func normalizeRequestURIIsUnreserved(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' ||
		'0' <= c && c <= '9' || c == '-' || c == '.' || c == '_' || c == '~'
}

func normalizeRequestURIIsHex(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

func normalizeRequestURIUnhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	default:
		return c - 'A' + 10
	}
}

func normalizeRequestURIUpperHex(c byte) byte {
	if 'a' <= c && c <= 'f' {
		return c - 'a' + 'A'
	}

	return c
}

func URLPathWithQueryClean(uri *url.URL) (path string) {
	base := &url.URL{}

	lengthPath := len(uri.Path)

	hasPath := lengthPath > 0
	hasQuery := len(uri.RawQuery) > 0

	switch {
	case hasPath && !hasQuery:
		return base.ResolveReference(uri).EscapedPath()
	case hasPath:
		return base.ResolveReference(uri).EscapedPath() + "?" + uri.RawQuery
	case hasQuery && uri.Path[lengthPath-1] == '/':
		return base.ResolveReference(uri).EscapedPath() + "/?" + uri.RawQuery
	case hasQuery:
		return base.ResolveReference(uri).EscapedPath() + "?" + uri.RawQuery
	default:
		return base.ResolveReference(uri).EscapedPath()
	}
}

// IsURISafeRedirection returns true if the URI passes the IsURISecure and HasURIDomainSuffix, i.e. if the scheme is
// secure and the given URI has a hostname that is either exactly equal to the given domain or if it has a suffix of the
// domain prefixed with a period.
func IsURISafeRedirection(uri *url.URL, domain string) bool {
	return IsURISecure(uri) && HasURIDomainSuffix(uri, domain)
}

// IsURISecure returns true if the URI has a secure schemes (https or wss).
func IsURISecure(uri *url.URL) bool {
	switch uri.Scheme {
	case https, wss:
		return true
	default:
		return false
	}
}

// HasURIDomainSuffix returns true if the URI hostname is equal to the domain suffix or if it has a suffix of the domain
// suffix prefixed with a period.
func HasURIDomainSuffix(uri *url.URL, domainSuffix string) bool {
	return HasDomainSuffix(uri.Hostname(), domainSuffix)
}

// HasDomainSuffix returns true if the URI hostname is equal to the domain or if it has a suffix of the domain
// prefixed with a period.
func HasDomainSuffix(domain, domainSuffix string) bool {
	if domainSuffix == "" {
		return false
	}

	if strings.EqualFold(domain, domainSuffix) {
		return true
	}

	if (strings.HasPrefix(domainSuffix, period) && StringHasSuffixFold(domain, domainSuffix)) || StringHasSuffixFold(domain, period+domainSuffix) {
		return true
	}

	return false
}

// EqualURLs returns true if the two *url.URL values are effectively equal taking into consideration web normalization.
func EqualURLs(first, second *url.URL) bool {
	if first == nil && second == nil {
		return true
	} else if first == nil || second == nil {
		return false
	}

	if !strings.EqualFold(first.Scheme, second.Scheme) {
		return false
	}

	if !strings.EqualFold(first.Host, second.Host) {
		return false
	}

	if first.Path != second.Path {
		return false
	}

	if first.RawQuery != second.RawQuery {
		return false
	}

	if first.Fragment != second.Fragment {
		return false
	}

	if first.RawFragment != second.RawFragment {
		return false
	}

	return true
}

// IsURLInSlice returns true if the needle url.URL is in the []url.URL haystack.
func IsURLInSlice(needle *url.URL, haystack []*url.URL) (has bool) {
	for i := 0; i < len(haystack); i++ {
		if EqualURLs(needle, haystack[i]) {
			return true
		}
	}

	return false
}
