package authorization

import (
	"github.com/authelia/authelia/v4/internal/utils"
)

// AccessControlSubjects represents an ACL subject.
type AccessControlSubjects struct {
	Subjects []SubjectMatcher
}

// AddSubject appends to the AccessControlSubjects based on a subject rule string.
func (acs *AccessControlSubjects) AddSubject(subjectRule string) {
	subject := schemaSubjectToACLSubject(subjectRule)

	if subject != nil {
		acs.Subjects = append(acs.Subjects, subject)
	}
}

// IsMatch returns true if the ACL subjects match the subject properties.
func (acs *AccessControlSubjects) IsMatch(subject Subject) (match bool) {
	for _, rule := range acs.Subjects {
		if !rule.IsMatch(subject) {
			return false
		}
	}

	return true
}

// AccessControlUser represents an ACL subject of type `user:`.
type AccessControlUser struct {
	Name string
}

// IsMatch returns true if the AccessControlUser name matches the Subject username.
func (acu AccessControlUser) IsMatch(subject Subject) (match bool) {
	return subject.Username == acu.Name
}

// AccessControlGroup represents an ACL subject of type `group:`.
type AccessControlGroup struct {
	Name string
}

// IsMatch returns true if the AccessControlGroup name matches one of the groups of the Subject.
func (acg AccessControlGroup) IsMatch(subject Subject) (match bool) {
	return utils.IsStringInSlice(acg.Name, subject.Groups)
}

// AccessControlClient represents an ACL subject of type `oauth2:client:`.
type AccessControlClient struct {
	Provider string
	ID       string
}

// IsMatch returns true if the AccessControlClient name matches one of the groups of the Subject.
func (acg AccessControlClient) IsMatch(subject Subject) (match bool) {
	return acg.ID == subject.ClientID
}
