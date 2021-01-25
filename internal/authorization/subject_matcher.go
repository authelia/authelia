package authorization

import (
	"strings"

	"github.com/authelia/authelia/internal/utils"
)

// isSubjectMatching checks if a subject matches any subject entry in the rule.
func isSubjectMatching(subject Subject, subjectRules [][]string) bool {
	if len(subjectRules) == 0 {
		return true
	}

	for _, subjectRule := range subjectRules {
		if isSingleSubjectMatching(subject, subjectRule) {
			return true
		}
	}

	return false
}

// isSubjectMatching checks if a subject matches a single one in the rule.
func isSingleSubjectMatching(subject Subject, subjectRule []string) bool {
	for _, ruleSubject := range subjectRule {
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
