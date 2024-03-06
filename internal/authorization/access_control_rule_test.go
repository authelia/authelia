package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccessControlRule_MatchesSubjectExact(t *testing.T) {
	testCases := []struct {
		name     string
		have     *AccessControlRule
		subject  Subject
		expected bool
	}{
		{
			"ShouldNotMatchAnonymous",
			&AccessControlRule{
				Subjects: []AccessControlSubjects{
					{[]SubjectMatcher{schemaSubjectToACLSubject("user:john")}},
				},
			},
			Subject{},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.MatchesSubjectExact(tc.subject))
		})
	}
}
