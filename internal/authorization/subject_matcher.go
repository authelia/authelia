package authorization

import (
	"strings"

	"github.com/authelia/authelia/internal/utils"
)

func isSubjectMatching(subject Subject, subjectRule []string) bool {
	for _, ruleSubject := range subjectRule {
		// If no subject is provided in the rule, we match any user.
		if ruleSubject == "" {
			continue
		}

		if strings.HasPrefix(ruleSubject, userPrefix) {
			user := strings.Trim(ruleSubject[len(userPrefix):], " ")
			if user == subject.Username {
				continue
			}
		}

		if strings.HasPrefix(ruleSubject, groupPrefix) {
			group := strings.Trim(ruleSubject[len(groupPrefix):], " ")
			if utils.IsStringInSlice(group, subject.Groups) {
				continue
			}
		}

		return false
	}

	return true
}
