// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"
)

// URLAmbiguity is a bitmask describing the classes of ambiguity present in a URL. A URL is ambiguous when it contains
// components which different implementations (i.e. a proxy and a backend) may interpret differently, such that each
// could have a different understanding of which resource was requested.
type URLAmbiguity uint

const (
	// URLAmbiguityUserInfo is a URL containing a user part, including an empty one i.e. 'https://user@example.com/',
	// which RFC9110 deprecates for the http and https schemes and is commonly used to disguise the host.
	URLAmbiguityUserInfo URLAmbiguity = 1 << iota

	// URLAmbiguityEncodedDotSegment is a path containing a dot segment formed with encoding i.e. '%2e%2e', '%2f..%2f',
	// '/%2e.%2f' or '..%2f'.
	URLAmbiguityEncodedDotSegment

	// URLAmbiguityParameterDotSegment is a path containing a dot or empty segment with a parameter i.e. '/..;/',
	// '/.;/', '/;/' or '/%2e%2e;x/'.
	URLAmbiguityParameterDotSegment

	// URLAmbiguityDoubleEncodedDotSegment is a path which forms an encoded dot segment only after being decoded a
	// second time i.e. '%252e%252e%252f'.
	URLAmbiguityDoubleEncodedDotSegment

	// URLAmbiguityBackslash is a path containing a literal backslash or a percent-encoded backslash (%5C) which some
	// implementations treat as a separator.
	URLAmbiguityBackslash

	// URLAmbiguityControlCharacter is a path or query which contains or decodes to an ASCII control character i.e. %00.
	URLAmbiguityControlCharacter

	// URLAmbiguityInvalidUTF8 is a path which decodes to an invalid UTF-8 sequence i.e. overlong encodings like %C0%AE.
	URLAmbiguityInvalidUTF8

	// URLAmbiguityFragment is a URL containing a '#'. A request URI never contains a fragment, so the character is
	// retained in the path or query by [url.ParseRequestURI] whereas other implementations may treat it as a delimiter.
	URLAmbiguityFragment

	// URLAmbiguityEncodedSeparator is a path containing a percent-encoded forward slash (%2F).
	URLAmbiguityEncodedSeparator

	// URLAmbiguityEncodedPercent is a path containing a percent-encoded percent sign (%25) which is susceptible to
	// being decoded more than once.
	URLAmbiguityEncodedPercent

	// URLAmbiguityDotSegment is a path containing a literal '.' or '..' segment.
	URLAmbiguityDotSegment

	// URLAmbiguityEmptySegment is a path containing an empty segment i.e. '//'.
	URLAmbiguityEmptySegment

	// URLAmbiguityPathParameter is a path containing a literal ';' which some implementations treat as the start of
	// segment parameters and strip from the segment i.e. '/admin;x/'. An encoded ';' (%3B) is instead covered by
	// URLAmbiguityEncodedNonUnreserved.
	URLAmbiguityPathParameter

	// URLAmbiguityInvalidQueryEncoding is a query containing a malformed percent-encoding.
	URLAmbiguityInvalidQueryEncoding

	// URLAmbiguityEncodedNonUnreserved is a path containing the percent-encoding of a character which is not
	// unreserved i.e. '%2F', '%25', '%3B', '%20' or '%C3%A9'. It never applies to the query.
	URLAmbiguityEncodedNonUnreserved

	// URLAmbiguityEncodedUnreserved is a path containing the percent-encoding of an unreserved character i.e. '%41',
	// '%7E' or '%2E'. It never applies to the query.
	URLAmbiguityEncodedUnreserved

	// URLAmbiguityEmptyDotSegment is a path where a dot segment follows an empty segment such that removing dot segments
	// per RFC 3986 and merging empty segments before removing dot segments resolve to different paths i.e.
	// '/public//../admin' is '/public/admin' per RFC 3986 and '/admin' when empty segments are merged first.
	URLAmbiguityEmptyDotSegment

	// URLAmbiguityDefault is the set of URLAmbiguity rejected by a URIParser unless configured otherwise.
	URLAmbiguityDefault = URLAmbiguityUserInfo | URLAmbiguityEncodedDotSegment | URLAmbiguityParameterDotSegment |
		URLAmbiguityDoubleEncodedDotSegment | URLAmbiguityBackslash | URLAmbiguityControlCharacter |
		URLAmbiguityInvalidUTF8 | URLAmbiguityFragment | URLAmbiguityPathParameter | URLAmbiguityEmptyDotSegment

	// URLAmbiguityAll is every URLAmbiguity.
	URLAmbiguityAll = URLAmbiguityDefault | URLAmbiguityEncodedSeparator | URLAmbiguityEncodedPercent |
		URLAmbiguityDotSegment | URLAmbiguityEmptySegment | URLAmbiguityInvalidQueryEncoding |
		URLAmbiguityEncodedNonUnreserved | URLAmbiguityEncodedUnreserved
)

var urlAmbiguityDefinitions = []struct {
	ambiguity   URLAmbiguity
	name        string
	description string
}{
	{URLAmbiguityUserInfo, "user_info", "url contains userinfo"},
	{URLAmbiguityEncodedDotSegment, "path_encoded_dot_segment", "path contains an encoded dot segment"},
	{URLAmbiguityParameterDotSegment, "path_parameter_dot_segment", "path contains a dot segment with a parameter"},
	{URLAmbiguityDoubleEncodedDotSegment, "path_double_encoded_dot_segment", "path contains a double encoded dot segment"},
	{URLAmbiguityBackslash, "path_backslash", "path contains a backslash"},
	{URLAmbiguityControlCharacter, "control_character", "url contains a control character"},
	{URLAmbiguityInvalidUTF8, "path_invalid_utf8", "path contains invalid utf-8"},
	{URLAmbiguityFragment, "fragment", "url contains a fragment delimiter"},
	{URLAmbiguityEncodedSeparator, "path_encoded_separator", "path contains a percent-encoded path separator"},
	{URLAmbiguityEncodedPercent, "path_encoded_percent", "path contains a percent-encoded percent sign"},
	{URLAmbiguityDotSegment, "path_dot_segment", "path contains a dot segment"},
	{URLAmbiguityEmptySegment, "path_empty_segment", "path contains an empty segment"},
	{URLAmbiguityPathParameter, "path_parameter", "path contains a segment parameter delimiter"},
	{URLAmbiguityInvalidQueryEncoding, "query_invalid_encoding", "query contains a malformed percent-encoding"},
	{URLAmbiguityEncodedNonUnreserved, "path_encoded_non_unreserved", "path contains an encoded character which is not unreserved"},
	{URLAmbiguityEncodedUnreserved, "path_encoded_unreserved", "path contains an encoded unreserved character"},
	{URLAmbiguityEmptyDotSegment, "path_empty_dot_segment", "path contains a dot segment following an empty segment"},
}

// Has returns true if all the provided ambiguities are present in this URLAmbiguity.
func (a URLAmbiguity) Has(ambiguity URLAmbiguity) bool {
	return a&ambiguity == ambiguity
}

// String returns a comma separated description of each ambiguity present.
func (a URLAmbiguity) String() string {
	descriptions := make([]string, 0, len(urlAmbiguityDefinitions))

	for _, d := range urlAmbiguityDefinitions {
		if a.Has(d.ambiguity) {
			descriptions = append(descriptions, d.description)
		}
	}

	return strings.Join(descriptions, ", ")
}

// URLAmbiguityFromName returns the URLAmbiguity for the configuration name provided.
func URLAmbiguityFromName(name string) (ambiguity URLAmbiguity, ok bool) {
	for _, d := range urlAmbiguityDefinitions {
		if d.name == name {
			return d.ambiguity, true
		}
	}

	return 0, false
}

// URLAmbiguityNames returns every configuration name for a URLAmbiguity.
func URLAmbiguityNames() []string {
	names := make([]string, len(urlAmbiguityDefinitions))

	for i, d := range urlAmbiguityDefinitions {
		names[i] = d.name
	}

	return names
}

// URIParser parses request URIs, rejecting those with a URLAmbiguity and producing normalized forms of the rest. It is
// immutable once created and safe for concurrent use.
type URIParser struct {
	reject URLAmbiguity
}

// URIParserOption configures a URIParser during creation.
type URIParserOption func(*URIParser)

// WithURLRejectAmbiguity adds the provided ambiguities to the set rejected by the URIParser. Options apply in order.
func WithURLRejectAmbiguity(ambiguity URLAmbiguity) URIParserOption {
	return func(p *URIParser) {
		p.reject |= ambiguity
	}
}

// WithURLAllowAmbiguity removes the provided ambiguities from the set rejected by the URIParser. Options apply in
// order.
func WithURLAllowAmbiguity(ambiguity URLAmbiguity) URIParserOption {
	return func(p *URIParser) {
		p.reject &^= ambiguity
	}
}

var defaultURIParser = NewURIParser()

// NewURIParser creates a URIParser which rejects URLAmbiguityDefault modified by the provided options.
func NewURIParser(opts ...URIParserOption) *URIParser {
	p := &URIParser{reject: URLAmbiguityDefault}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// DefaultURIParser returns a shared URIParser with the default options.
func DefaultURIParser() *URIParser {
	return defaultURIParser
}

// NewURIParserFromNames creates a URIParser which allows the ambiguity classes named in allow and rejects the ambiguity
// classes named in reject, see URLAmbiguityFromName. Unknown names are ignored.
func NewURIParserFromNames(allow, reject []string) *URIParser {
	opts := make([]URIParserOption, 0, len(allow)+len(reject))

	for _, name := range allow {
		if ambiguity, ok := URLAmbiguityFromName(name); ok {
			opts = append(opts, WithURLAllowAmbiguity(ambiguity))
		}
	}

	for _, name := range reject {
		if ambiguity, ok := URLAmbiguityFromName(name); ok {
			opts = append(opts, WithURLRejectAmbiguity(ambiguity))
		}
	}

	return NewURIParser(opts...)
}

// ParsedURI is the result of URIParser.Parse.
type ParsedURI struct {
	// URL is the *[url.URL] from [url.ParseRequestURI] with the raw path intact.
	URL *url.URL

	// AccessPath is the decoded and cleaned path with the query appended, see URIAccessPath.
	AccessPath string

	// Canonical is the normalized URI, see URICanonical.
	Canonical string
}

// ParseRequestURI creates a URIParser with the provided options, uses it to parse the input, and discards it.
func ParseRequestURI(input string, opts ...URIParserOption) (parsed *ParsedURI, err error) {
	return NewURIParser(opts...).Parse(input)
}

// Parse parses the input with [url.ParseRequestURI]. The input must be either an absolute path or an absolute URI with
// an absolute or empty path, otherwise an error wrapping ErrURLNotAbsolutePath is returned. An error wrapping
// ErrURLAmbiguous is returned if the input contains any ambiguity this URIParser rejects. Any userinfo is removed from
// the input text of returned errors.
func (p *URIParser) Parse(input string) (parsed *ParsedURI, err error) {
	var u *url.URL

	if u, err = url.ParseRequestURI(input); err != nil {
		var uerr *url.Error

		if errors.As(err, &uerr) {
			if strings.IndexByte(input, '@') != -1 {
				return nil, &url.Error{Op: uerr.Op, URL: URIRemoveUserInfo(uerr.URL), Err: errors.New("invalid url")}
			}

			return nil, uerr
		}

		return nil, err
	}

	if u.Opaque != "" || (u.Path != "" && u.Path[0] != '/') {
		return nil, fmt.Errorf("%w: '%s'", ErrURLNotAbsolutePath, URIRemoveUserInfo(input))
	}

	if ambiguity := urlAmbiguity(input, u) & p.reject; ambiguity != 0 {
		return nil, fmt.Errorf("%w: %s", ErrURLAmbiguous, ambiguity)
	}

	return &ParsedURI{URL: u, AccessPath: URIAccessPath(u), Canonical: URICanonical(u)}, nil
}

// URICanonical returns the normalized form of the URI following RFC 3986 sections 6.2.2 and 6.2.3. The scheme and host
// are lowercased, the default port is removed, an empty path becomes '/', percent-encodings are uppercased, encoded
// unreserved characters are decoded, dot segments are removed per RFC 3986 section 5.2.4, empty segments and a trailing
// slash are retained, and an empty query is dropped. Encoded reserved characters remain encoded, and the userinfo is
// dropped. A URI without a host results in only the path and query.
func URICanonical(u *url.URL) string {
	b := &strings.Builder{}

	if hostname := u.Hostname(); hostname != "" {
		scheme := strings.ToLower(u.Scheme)

		b.WriteString(scheme)
		b.WriteString("://")

		if strings.IndexByte(hostname, ':') != -1 {
			host, zone, hasZone := strings.Cut(hostname, "%")

			b.WriteString("[" + strings.ToLower(host))

			if hasZone {
				b.WriteString("%25" + strings.ReplaceAll(zone, "%", "%25"))
			}

			b.WriteString("]")
		} else {
			b.WriteString(strings.ToLower(hostname))
		}

		if port := u.Port(); port != "" && port != urlSchemeDefaultPort(scheme) {
			b.WriteString(":" + port)
		}
	}

	b.WriteString(urlPathCanonical(urlPathRaw(u)))

	if len(u.RawQuery) != 0 {
		b.WriteString("?" + urlQueryCanonical(u.RawQuery))
	}

	return b.String()
}

func urlSchemeDefaultPort(scheme string) string {
	switch scheme {
	case "https", "wss":
		return "443"
	case "http", "ws":
		return "80"
	default:
		return ""
	}
}

// URIRemoveUserInfo removes the userinfo from the raw input text without parsing it, removing everything up to the last
// '@' other than the scheme and authority marker. Inputs which start with a '/' are returned unchanged.
func URIRemoveUserInfo(input string) string {
	at := strings.LastIndexByte(input, '@')
	if at == -1 || strings.HasPrefix(input, "/") {
		return input
	}

	prefix := urlSchemePrefix(input)

	if len(prefix) > at {
		prefix = ""
	}

	return prefix + input[at+1:]
}

func urlSchemePrefix(input string) string {
	for i := 0; i < len(input); i++ {
		c := input[i]

		switch {
		case c == ':' && i > 0:
			if strings.HasPrefix(input[i+1:], "//") {
				return input[:i+3]
			}

			return input[:i+1]
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z':
		case i > 0 && (('0' <= c && c <= '9') || c == '+' || c == '-' || c == '.'):
		default:
			return ""
		}
	}

	return ""
}

func urlPathRaw(u *url.URL) string {
	if u.RawPath != "" {
		return u.RawPath
	}

	return u.EscapedPath()
}

func urlAmbiguity(input string, u *url.URL) (ambiguity URLAmbiguity) {
	raw := urlPathRaw(u)

	ambiguity = urlAmbiguityPathRaw(raw) | urlAmbiguityPathDecoded(u.Path) | urlAmbiguityDotSegments(raw) |
		urlAmbiguityEmptyDotSegment(raw) | urlAmbiguityEmptyDotSegment(u.Path) | urlAmbiguityQuery(u.RawQuery)

	if u.User != nil {
		ambiguity |= URLAmbiguityUserInfo
	}

	if strings.IndexByte(input, '#') != -1 {
		ambiguity |= URLAmbiguityFragment
	}

	return ambiguity
}

func urlAmbiguityPathRaw(raw string) (ambiguity URLAmbiguity) {
	for i := 0; i < len(raw); i++ {
		switch c := raw[i]; {
		case c == '\\':
			ambiguity |= URLAmbiguityBackslash
		case c == ';':
			ambiguity |= URLAmbiguityPathParameter
		case c == '%' && i+2 < len(raw) && isHex(raw[i+1]) && isHex(raw[i+2]):
			d := unhex(raw[i+1])<<4 | unhex(raw[i+2])

			switch {
			case isUnreserved(d):
				ambiguity |= URLAmbiguityEncodedUnreserved
			default:
				ambiguity |= URLAmbiguityEncodedNonUnreserved
			}

			switch d {
			case '/':
				ambiguity |= URLAmbiguityEncodedSeparator
			case '\\':
				ambiguity |= URLAmbiguityBackslash
			case '%':
				ambiguity |= URLAmbiguityEncodedPercent
			}

			i += 2
		}
	}

	return ambiguity
}

func urlAmbiguityPathDecoded(decoded string) (ambiguity URLAmbiguity) {
	for i := 0; i < len(decoded); i++ {
		switch c := decoded[i]; {
		case c == '\\':
			ambiguity |= URLAmbiguityBackslash
		case isControl(c):
			ambiguity |= URLAmbiguityControlCharacter
		}
	}

	if !utf8.ValidString(decoded) {
		ambiguity |= URLAmbiguityInvalidUTF8
	}

	segments := strings.Split(decoded, "/")

	for i, segment := range segments {
		if segment == "" && i != 0 && i != len(segments)-1 {
			ambiguity |= URLAmbiguityEmptySegment
		}
	}

	return ambiguity
}

func urlAmbiguityDotSegments(raw string) (ambiguity URLAmbiguity) {
	for _, segment := range strings.Split(raw, "/") {
		switch segment {
		case ".", "..":
			ambiguity |= URLAmbiguityDotSegment
		}

		encoded := strings.IndexByte(segment, '%') != -1
		decoded := urlDecodeTolerant(segment)

		ambiguity |= urlAmbiguityDotSegmentPieces(decoded, encoded)

		if strings.Contains(segment, "%25") {
			if urlAmbiguityDotSegmentPieces(urlDecodeTolerant(decoded), true)&URLAmbiguityEncodedDotSegment != 0 {
				ambiguity |= URLAmbiguityDoubleEncodedDotSegment
			}
		}
	}

	return ambiguity
}

func urlAmbiguityDotSegmentPieces(decoded string, encoded bool) (ambiguity URLAmbiguity) {
	for _, piece := range strings.FieldsFunc(decoded, func(r rune) bool { return r == '/' || r == '\\' }) {
		stripped, _, hasParam := strings.Cut(piece, ";")
		isDot := stripped == "." || stripped == ".."

		if isDot && encoded {
			ambiguity |= URLAmbiguityEncodedDotSegment
		}

		if hasParam && (isDot || stripped == "") {
			ambiguity |= URLAmbiguityParameterDotSegment
		}
	}

	return ambiguity
}

func urlDecodeTolerant(s string) string {
	if strings.IndexByte(s, '%') == -1 {
		return s
	}

	b := make([]byte, 0, len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
			b = append(b, unhex(s[i+1])<<4|unhex(s[i+2]))

			i += 2

			continue
		}

		b = append(b, s[i])
	}

	return string(b)
}

func urlAmbiguityQuery(query string) (ambiguity URLAmbiguity) {
	for i := 0; i < len(query); i++ {
		switch c := query[i]; {
		case c == '%' && i+2 < len(query) && isHex(query[i+1]) && isHex(query[i+2]):
			if isControl(unhex(query[i+1])<<4 | unhex(query[i+2])) {
				ambiguity |= URLAmbiguityControlCharacter
			}

			i += 2
		case c == '%':
			ambiguity |= URLAmbiguityInvalidQueryEncoding
		case isControl(c):
			ambiguity |= URLAmbiguityControlCharacter
		}
	}

	return ambiguity
}

func urlPathCanonical(raw string) string {
	segments := strings.Split(raw, "/")

	for i := 1; i < len(segments); i++ {
		segments[i] = urlSegmentCanonical(segments[i])
	}

	return urlRemoveDotSegments(strings.Join(segments, "/"))
}

func urlRemoveDotSegments(p string) string {
	segments := strings.Split(p, "/")
	stack := make([]string, 0, len(segments))

	var trailing bool

	for i, segment := range segments {
		if i == 0 {
			continue
		}

		last := i == len(segments)-1

		switch segment {
		case ".":
			trailing = last
		case "..":
			if len(stack) != 0 {
				stack = stack[:len(stack)-1]
			}

			trailing = last
		default:
			stack = append(stack, segment)
		}
	}

	b := &strings.Builder{}

	b.Grow(len(p) + 1)

	for _, segment := range stack {
		b.WriteByte('/')
		b.WriteString(segment)
	}

	if trailing || len(stack) == 0 {
		b.WriteByte('/')
	}

	return b.String()
}

func urlAmbiguityEmptyDotSegment(raw string) (ambiguity URLAmbiguity) {
	if !strings.Contains(raw, "//") {
		return 0
	}

	if urlMergeEmptySegments(urlPathCanonical(raw)) != urlPathCanonical(urlMergeEmptySegments(raw)) {
		return URLAmbiguityEmptyDotSegment
	}

	return 0
}

func urlMergeEmptySegments(p string) string {
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}

	return p
}

func urlSegmentCanonical(segment string) string {
	b := &strings.Builder{}

	b.Grow(len(segment))

	for i := 0; i < len(segment); i++ {
		switch c := segment[i]; {
		case c == '%' && i+2 < len(segment) && isHex(segment[i+1]) && isHex(segment[i+2]):
			if d := unhex(segment[i+1])<<4 | unhex(segment[i+2]); isUnreserved(d) {
				b.WriteByte(d)
			} else {
				writePercentEncoded(b, d)
			}

			i += 2
		case c != '%' && isPChar(c):
			b.WriteByte(c)
		default:
			writePercentEncoded(b, c)
		}
	}

	return b.String()
}

func urlQueryCanonical(query string) string {
	b := &strings.Builder{}

	b.Grow(len(query))

	for i := 0; i < len(query); i++ {
		switch c := query[i]; {
		case c == '%' && i+2 < len(query) && isHex(query[i+1]) && isHex(query[i+2]):
			if d := unhex(query[i+1])<<4 | unhex(query[i+2]); isUnreserved(d) {
				b.WriteByte(d)
			} else {
				writePercentEncoded(b, d)
			}

			i += 2
		case c == '/' || c == '?' || (c != '%' && isPChar(c)):
			b.WriteByte(c)
		default:
			writePercentEncoded(b, c)
		}
	}

	return b.String()
}

func writePercentEncoded(b *strings.Builder, c byte) {
	const upperhex = "0123456789ABCDEF"

	b.WriteByte('%')
	b.WriteByte(upperhex[c>>4])
	b.WriteByte(upperhex[c&15])
}

func isUnreserved(c byte) bool {
	switch {
	case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9':
		return true
	}

	switch c {
	case '-', '.', '_', '~':
		return true
	}

	return false
}

func isPChar(c byte) bool {
	if isUnreserved(c) {
		return true
	}

	switch c {
	case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '@':
		return true
	}

	return false
}

func isControl(c byte) bool {
	return c < 0x20 || c == 0x7f
}

func isHex(c byte) bool {
	switch {
	case '0' <= c && c <= '9', 'a' <= c && c <= 'f', 'A' <= c && c <= 'F':
		return true
	}

	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	default:
		return c - 'A' + 10
	}
}
