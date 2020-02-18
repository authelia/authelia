package authorization

import (
	"strings"

	"github.com/authelia/authelia/internal/utils"
)

func isSubjectMatching(subject Subject, subjectRule string) bool {
	// If no subject is provided in the rule, we match any user.
	if subjectRule == "" {
		return true
	}

	if strings.HasPrefix(subjectRule, userPrefix) {
		user := strings.Trim(subjectRule[len(userPrefix):], " ")
		if user == subject.Username {
			return true
		}
	}

	if strings.HasPrefix(subjectRule, groupPrefix) {
		group := strings.Trim(subjectRule[len(groupPrefix):], " ")
		if utils.IsStringInSlice(group, subject.Groups) {
			return true
		}
	}
	return false
}
