package authorization

import (
	"github.com/authelia/authelia/internal/utils"
)

// AccessControlSubject abstracts an ACL subject of type `group:` or `user:`.
type AccessControlSubject interface {
	IsMatch(subject Subject) (match bool)
}

// AccessControlSubjects represents an ACL subject.
type AccessControlSubjects struct {
	Subjects []AccessControlSubject
}

// AddSubject appends the ACL subject based on a subject rule string.
func (acs *AccessControlSubjects) AddSubject(subjectRule string) {
	subject := schemaSubjectToACLSubject(subjectRule)

	if subject != nil {
		acs.Subjects = append(acs.Subjects, subject)
	}
}

// IsMatch returns true if the ACL subjects match the subject properties.
func (acs AccessControlSubjects) IsMatch(subject Subject) (match bool) {
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

// IsMatch returns true if the ACL User name matches the subject username.
func (acu AccessControlUser) IsMatch(subject Subject) (match bool) {
	return subject.Username == acu.Name
}

// AccessControlGroup represents an ACL subject of type `group:`.
type AccessControlGroup struct {
	Name string
}

// IsMatch returns true if the ACL Group name matches one of the subjects group names.
func (acg AccessControlGroup) IsMatch(subject Subject) (match bool) {
	return utils.IsStringInSlice(acg.Name, subject.Groups)
}
