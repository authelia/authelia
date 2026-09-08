package authorization

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

// AccessControlNetworks represents the networks criteria of an access control rule.
type AccessControlNetworks []*net.IPNet

// IsMatch returns true if the given subject matches these networks.
func (a AccessControlNetworks) IsMatch(subject Subject) bool {
	if len(a) == 0 {
		return true
	}

	for _, network := range a {
		if network.Contains(subject.IP) {
			return true
		}
	}

	return false
}

// SubjectMatcher is a matcher that takes a subject.
type SubjectMatcher interface {
	IsMatch(subject Subject) (match bool)
}

// StringSubjectMatcher is a matcher that takes an input string and subject.
type StringSubjectMatcher interface {
	IsMatch(input string, subject Subject) (match bool)
}

// SubjectObjectMatcher is a matcher that takes both a subject and an object.
type SubjectObjectMatcher interface {
	IsMatch(subject Subject, object Object) (match bool)
}

// ObjectMatcher is a matcher that takes an object.
type ObjectMatcher interface {
	IsMatch(object Object) (match bool)
}

// Subject represents the identity of a user for the purposes of ACL matching.
type Subject struct {
	Username string
	Groups   []string
	ClientID string
	IP       net.IP
}

// String returns a string representation of the Subject.
func (s Subject) String() string {
	return fmt.Sprintf("username=%s groups=%s ip=%s", s.Username, strings.Join(s.Groups, ","), s.IP.String())
}

// IsAnonymous returns true if the Subject username and groups are empty.
func (s Subject) IsAnonymous() bool {
	return s.Username == "" && len(s.Groups) == 0 && s.ClientID == ""
}

// Object represents a protected object for the purposes of ACL matching.
type Object struct {
	URL *url.URL

	Domain string
	Path   string
	Method string
}

// String is a string representation of the Object.
func (o Object) String() string {
	if o.URL == nil {
		return ""
	}

	return o.URL.String()
}

// NewObjectMethodURLOrSchemeHostPath creates a new [Object] from the raw method and raw URL, falling back to the raw
// scheme, host, and path when the raw URL is absent.
func NewObjectMethodURLOrSchemeHostPath(rawMethod, rawURL, rawScheme, rawHost, rawPath []byte) (object *Object, err error) {
	if len(rawURL) > 0 {
		return NewObjectMethodURL(rawMethod, rawURL)
	}

	return NewObjectMethodSchemeHostPath(rawMethod, rawScheme, rawHost, rawPath)
}

// NewObjectMethodSchemeHostPath creates a new [Object] from the raw method and the raw scheme, host, and path which are
// joined to form the request URL.
func NewObjectMethodSchemeHostPath(rawMethod, rawScheme, rawHost, rawPath []byte) (object *Object, err error) {
	if len(rawScheme) == 0 {
		return nil, fmt.Errorf("missing scheme value")
	}

	if len(rawHost) == 0 {
		return nil, fmt.Errorf("missing host value")
	}

	rawURL := utils.BytesJoin(rawScheme, sepSchemeHost, rawHost, rawPath)

	return NewObjectMethodURL(rawMethod, rawURL)
}

// NewObjectMethodURL creates a new [Object] from the raw method and raw URL, validating the method and applying request
// URI normalization to the URL.
func NewObjectMethodURL(rawMethod, rawURL []byte) (object *Object, err error) {
	var objectURL *url.URL

	if hasInvalidMethodCharacters(rawMethod) {
		return nil, fmt.Errorf("method header with value '%s' has invalid characters", rawMethod)
	}

	method := string(rawMethod)

	if objectURL, err = url.ParseRequestURI(string(rawURL)); err != nil {
		return nil, fmt.Errorf("error occurred parsing object url: %w", err)
	}

	o := NewObject(objectURL, method)

	return &o, nil
}

// NewObjectRaw creates a new Object type from a URL and a method header.
func NewObjectRaw(targetURL *url.URL, method []byte) (object Object) {
	return NewObject(targetURL, string(method))
}

// NewObject creates a new Object type from a URL and a method header.
func NewObject(targetURL *url.URL, method string) (object Object) {
	targetURL.Scheme = strings.ToLower(targetURL.Scheme)
	targetURL.Host = strings.ToLower(targetURL.Host)

	return Object{
		URL:    targetURL,
		Domain: targetURL.Hostname(),
		Path:   utils.URLPathFullClean(targetURL),
		Method: strings.ToUpper(method),
	}
}

// RuleMatchResult describes how well a rule matched a subject/object combo.
type RuleMatchResult struct {
	Rule *AccessControlRule

	Skipped bool

	MatchDomain        bool
	MatchResources     bool
	MatchQuery         bool
	MatchMethods       bool
	MatchNetworks      bool
	MatchSubjects      bool
	MatchSubjectsExact bool
}

// IsMatch returns true if all the criteria matched.
func (r RuleMatchResult) IsMatch() (match bool) {
	return r.MatchDomain && r.MatchResources && r.MatchQuery && r.MatchMethods && r.MatchNetworks && r.MatchSubjectsExact
}

// IsPotentialMatch returns true if the rule is potentially a match.
func (r RuleMatchResult) IsPotentialMatch() (match bool) {
	return r.MatchDomain && r.MatchResources && r.MatchQuery && r.MatchMethods && r.MatchNetworks && r.MatchSubjects && !r.MatchSubjectsExact
}

func hasInvalidMethodCharacters(v []byte) bool {
	for _, c := range v {
		if (c < 0x41 || c > 0x5A) && (c < 0x61 || c > 0x7A) {
			return true
		}
	}

	return false
}
