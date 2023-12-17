package authorization

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

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
	return s.Username == "" && len(s.Groups) == 0
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
	return o.URL.String()
}

// NewObjectRaw creates a new Object type from a URL and a method header.
func NewObjectRaw(targetURL *url.URL, method []byte) (object Object) {
	return NewObject(targetURL, string(method))
}

// NewObject creates a new Object type from a URL and a method header.
func NewObject(targetURL *url.URL, method string) (object Object) {
	return Object{
		URL:    targetURL,
		Domain: targetURL.Hostname(),
		Path:   utils.URLPathFullClean(targetURL),
		Method: method,
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
	return r.MatchDomain && r.MatchResources && r.MatchMethods && r.MatchNetworks && r.MatchSubjectsExact
}

// IsPotentialMatch returns true if the rule is potentially a match.
func (r RuleMatchResult) IsPotentialMatch() (match bool) {
	return r.MatchDomain && r.MatchResources && r.MatchMethods && r.MatchNetworks && r.MatchSubjects && !r.MatchSubjectsExact
}
