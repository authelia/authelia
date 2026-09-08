package utils

import (
	"regexp"
	"strings"
)

const (
	printableUnicodeRegexp = `^[\pL\pM\pN\pP\pS\s]{1,100}$`
	emailRegex             = `^[a-zA-Z0-9+._~!#$%&'*/=?^{|}-]+@[a-zA-Z0-9-.]+\.[a-zA-Z0-9-]+$`
	usernameAndGroupRegex  = `^[a-zA-Z0-9+._\-]{1,100}$`
)

// ValidatePrintableUnicodeString returns true if input is a printable unicode string of up to 100 characters.
func ValidatePrintableUnicodeString(input string) bool {
	var regex = regexp.MustCompile(printableUnicodeRegexp) //nolint:forbidigo

	return regex.MatchString(input)
}

// ValidateEmailString returns true if input is a valid email address.
func ValidateEmailString(input string) bool {
	var regex = regexp.MustCompile(emailRegex)

	return regex.MatchString(input)
}

// ValidateGroups returns true if every group in input is valid, otherwise false and the first invalid group.
func ValidateGroups(input []string) (bool, string) {
	for _, group := range input {
		if !ValidateGroup(group) {
			return false, group
		}
	}

	return true, ""
}

// ValidateGroup returns true if input is a valid group name.
func ValidateGroup(input string) bool {
	var regex = regexp.MustCompile(usernameAndGroupRegex)

	return regex.MatchString(input)
}

// ValidateUsername returns true if input is a valid username or email address.
func ValidateUsername(input string) bool {
	if strings.Contains(input, `@`) {
		return ValidateEmailString(input)
	}

	return ValidatePrintableUnicodeString(input)
}
