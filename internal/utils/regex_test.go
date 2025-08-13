package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePrintableUnicodeString(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		// Valid cases
		{"ValidAlphanumeric", "abc123", true},
		{"ValidWithSpaces", "hello world", true},
		{"ValidUnicode", "café🌟", true},
		{"ValidPunctuation", "Hello, World!", true},
		{"ValidSymbols", "!@#$%^&*()", true},
		{"ValidTabAndNewline", "hello\tworld\n", true},
		{"SingleCharacter", "a", true},
		{"Exactly100Chars", strings.Repeat("a", 100), true},

		// Invalid cases
		{"EmptyString", "", false},
		{"Over100Chars", strings.Repeat("a", 101), false},
		{"ControlCharacter", "hello\x00world", false},
		{"ControlCharacterTab", "hello\x08world", false}, // backspace
		{"OnlyControlChar", "\x00", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidatePrintableUnicodeString(tc.input)
			assert.Equal(t, tc.shouldPass, result)
		})
	}
}

func TestValidateEmailString(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		{"SimpleEmail", "test@example.com", true},
		{"EmailWithPlus", "user+tag@example.com", true},
		{"EmailWithDot", "first.last@example.com", true},
		{"EmailWithUnderscore", "user_name@example.com", true},
		{"EmailWithDash", "user-name@example.com", true},
		{"EmailWithTilde", "user~name@example.com", true},
		{"EmailWithSpecialChars", "user!#$%&'*/=?^{|}-@example.com", true},
		{"NumbersInLocal", "user123@example.com", true},
		{"NumbersInDomain", "user@example123.com", true},
		{"SubDomain", "user@mail.example.com", true},
		{"DashInDomain", "user@ex-ample.com", true},

		{"EmptyString", "", false},
		{"NoAtSign", "userexample.com", false},
		{"MultipleAtSigns", "user@@example.com", false},
		{"NoLocal", "@example.com", false},
		{"NoDomain", "user@", false},
		{"NoTLD", "user@example", false},
		{"SpaceInEmail", "user @example.com", false},
		{"InvalidCharInLocal", "user@example.com@", false},
		{"EmptyLocal", "@example.com", false},
		{"EmptyDomain", "user@.com", false},
		{"EndWithDot", "user@example.com.", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateEmailString(tc.input)
			assert.Equal(t, tc.shouldPass, result)
		})
	}
}

func TestValidateGroup(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		// Valid groups
		{"SimpleGroup", "mygroup", true},
		{"GroupWithNumbers", "group123", true},
		{"GroupWithDash", "my-group", true},
		{"GroupWithUnderscore", "my_group", true},
		{"GroupWithDot", "my.group", true},
		{"GroupWithPlus", "my+group", true},
		{"SingleChar", "a", true},
		{"NumbersOnly", "123", true},
		{"Exactly100Chars", strings.Repeat("a", 100), true},

		// Invalid groups
		{"EmptyString", "", false},
		{"Over100Chars", strings.Repeat("a", 101), false},
		{"WithSpace", "my group", false},
		{"WithAtSign", "my@group", false},
		{"WithSpecialChars", "my#group", false},
		{"WithSlash", "my/group", false},
		{"WithParens", "my(group)", false},
		{"Unicode", "café", false},
		{"Emoji", "group🌟", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateGroup(tc.input)
			assert.Equal(t, tc.shouldPass, result)
		})
	}
}

func TestValidateGroups(t *testing.T) {
	testCases := []struct {
		name            string
		input           []string
		shouldPass      bool
		expectedInvalid string
	}{
		// Valid cases
		{"EmptySlice", []string{}, true, ""},
		{"SingleValidGroup", []string{"group1"}, true, ""},
		{"MultipleValidGroups", []string{"group1", "group2", "test123"}, true, ""},
		{"ValidGroupsWithSpecialChars", []string{"group-1", "group_2", "group.3"}, true, ""},

		// Invalid cases
		{"SingleInvalidGroup", []string{"invalid group"}, false, "invalid group"},
		{"FirstGroupInvalid", []string{"invalid@group", "validgroup"}, false, "invalid@group"},
		{"LastGroupInvalid", []string{"validgroup", "invalid group"}, false, "invalid group"},
		{"MiddleGroupInvalid", []string{"valid1", "invalid@group", "valid2"}, false, "invalid@group"},
		{"MultipleInvalidGroups", []string{"invalid@1", "invalid 2"}, false, "invalid@1"}, // Returns first invalid
		{"EmptyStringInGroup", []string{"valid", "", "alsovalid"}, false, ""},
		{"TooLongGroup", []string{"valid", strings.Repeat("a", 101)}, false, strings.Repeat("a", 101)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid, invalidGroup := ValidateGroups(tc.input)
			assert.Equal(t, tc.shouldPass, isValid)
			assert.Equal(t, tc.expectedInvalid, invalidGroup)
		})
	}
}

func TestValidateUsername(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		shouldPass bool
	}{
		// Valid usernames (non-email format)
		{"SimpleUsername", "john", true},
		{"UsernameWithNumbers", "user123", true},
		{"UnicodeUsername", "café", true},
		{"UsernameWithSpaces", "john doe", true},
		{"UsernameWithPunctuation", "user-name_test.123", true},
		{"EmojiUsername", "user🌟", true},

		// Valid usernames (email format)
		{"ValidEmail", "user@example.com", true},
		{"EmailWithPlus", "user+tag@example.com", true},
		{"ComplexEmail", "test.user+label@sub.example.com", true},

		// Invalid usernames (non-email format - violates printable unicode rules)
		{"EmptyUsername", "", false},
		{"TooLongUsername", strings.Repeat("a", 101), false},
		{"ControlCharacter", "user\x00name", false},

		// Invalid usernames (email format - violates email rules)
		{"InvalidEmailNoTLD", "user@example", false},
		{"InvalidEmailMultipleAt", "user@@example.com", false},
		{"InvalidEmailNoLocal", "@example.com", false},
		{"InvalidEmailSpaceInDomain", "user@exam ple.com", false},

		// Edge case: @ sign handling
		{"AtSignAlone", "@", false},                          // Treated as email, but invalid
		{"AtSignInMiddle", "user@middle@example.com", false}, // Invalid email
		{"ValidAtInNonEmail", "user@company", false},         // Contains @, so treated as email, but invalid
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateUsername(tc.input)
			assert.Equal(t, tc.shouldPass, result, "Expected %v for input: %s", tc.shouldPass, tc.input)
		})
	}
}
