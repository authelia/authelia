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

func ValidatePrintableUnicodeString(input string) bool {
	var regex = regexp.MustCompile(printableUnicodeRegexp) //nolint:forbidigo

	return regex.MatchString(input)
}

func ValidateEmailString(input string) bool {
	var regex = regexp.MustCompile(emailRegex)

	return regex.MatchString(input)
}

func ValidateGroups(input []string) (bool, string) {
	for _, group := range input {
		if !ValidateGroup(group) {
			return false, group
		}
	}

	return true, ""
}

func ValidateGroup(input string) bool {
	var regex = regexp.MustCompile(usernameAndGroupRegex)

	return regex.MatchString(input)
}

func ValidateUsername(input string) bool {
	if strings.Contains(input, `@`) {
		return ValidateEmailString(input)
	}

	return ValidatePrintableUnicodeString(input)
}
