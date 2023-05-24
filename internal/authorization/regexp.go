package authorization

import (
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/utils"
)

// RegexpGroupStringSubjectMatcher matches the input string against the pattern taking into account Subexp groups.
type RegexpGroupStringSubjectMatcher struct {
	Pattern         regexp.Regexp
	SubexpNameUser  int
	SubexpNameGroup int
}

// IsMatch returns true if the underlying pattern matches the input given the subject.
func (r RegexpGroupStringSubjectMatcher) IsMatch(input string, subject Subject) (match bool) {
	matches := r.Pattern.FindStringSubmatch(input)
	if matches == nil {
		return false
	}

	if subject.IsAnonymous() {
		return true
	}

	if r.SubexpNameUser != -1 && !strings.EqualFold(subject.Username, matches[r.SubexpNameUser]) {
		return false
	}

	if r.SubexpNameGroup != -1 && !utils.IsStringInSliceFold(matches[r.SubexpNameGroup], subject.Groups) {
		return false
	}

	return true
}

// String returns the pattern string.
func (r RegexpGroupStringSubjectMatcher) String() string {
	return r.Pattern.String()
}

// RegexpStringSubjectMatcher just matches the input string against the pattern.
type RegexpStringSubjectMatcher struct {
	Pattern regexp.Regexp
}

// IsMatch returns true if the underlying pattern matches the input.
func (r RegexpStringSubjectMatcher) IsMatch(input string, _ Subject) (match bool) {
	return r.Pattern.MatchString(input)
}

// String returns the pattern string.
func (r RegexpStringSubjectMatcher) String() string {
	return r.Pattern.String()
}
