package model

import (
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NewAuthorization returns a new empty Authorization.
func NewAuthorization() *Authorization {
	return &Authorization{}
}

// Authorization represents a parsed Authorization or Proxy-Authorization header value.
type Authorization struct {
	parsed    bool
	scheme    AuthorizationScheme
	rawscheme string
	value     string
	username  string
	password  string
}

// SchemeRaw returns the scheme exactly as it appeared in the parsed header value.
func (a *Authorization) SchemeRaw() string {
	return a.rawscheme
}

// Scheme returns the parsed AuthorizationScheme.
func (a *Authorization) Scheme() AuthorizationScheme {
	return a.scheme
}

// Value returns the value portion of the parsed header value.
func (a *Authorization) Value() string {
	return a.value
}

// EncodeHeader returns the header value representation of this Authorization.
func (a *Authorization) EncodeHeader() string {
	if !a.parsed {
		return ""
	}

	switch a.scheme {
	case AuthorizationSchemeNone:
		return ""
	case AuthorizationSchemeBasic, AuthorizationSchemeBearer:
		return fmt.Sprintf("%s %s", cases.Title(language.English).String(a.scheme.String()), a.value)
	default:
		return ""
	}
}

// Basic returns the username and password provided the parsed scheme is the basic scheme.
func (a *Authorization) Basic() (username, password string) {
	if !a.parsed {
		return "", ""
	}

	switch a.scheme {
	case AuthorizationSchemeBasic:
		return a.username, a.password
	default:
		return "", ""
	}
}

// BasicUsername returns the username provided the parsed scheme is the basic scheme.
func (a *Authorization) BasicUsername() (username string) {
	if !a.parsed {
		return ""
	}

	switch a.scheme {
	case AuthorizationSchemeBasic:
		return a.username
	default:
		return ""
	}
}

// ParseBasic parses the given username and password as the basic scheme.
func (a *Authorization) ParseBasic(username, password string) (err error) {
	if a.parsed {
		return fmt.Errorf("invalid state: this scheme has already performed a parse action")
	}

	switch {
	case len(username) == 0:
		return fmt.Errorf("invalid value: username must not be empty")
	case strings.Contains(username, ":"):
		return fmt.Errorf("invalid value: username must not contain the ':' character")
	case len(password) == 0:
		return fmt.Errorf("invalid value: password must not be empty")
	}

	a.parsed = true

	a.username, a.password, a.scheme, a.rawscheme = username, password, AuthorizationSchemeBasic, AuthorizationSchemeBasic.String()

	a.value = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))

	return nil
}

// ParseBearer parses the given token as the bearer scheme.
func (a *Authorization) ParseBearer(bearer string) (err error) {
	if a.parsed {
		return fmt.Errorf("invalid state: this scheme has already performed a parse action")
	}

	if err = a.validateSchemeBearerValue(bearer); err != nil {
		return err
	}

	a.parsed = true

	a.value, a.scheme, a.rawscheme = bearer, AuthorizationSchemeBearer, AuthorizationSchemeBearer.String()

	return nil
}

// Parse parses the given raw header value.
func (a *Authorization) Parse(raw string) (err error) {
	if a.parsed {
		return fmt.Errorf("invalid state: this scheme has already performed a parse action")
	}

	if len(raw) == 0 {
		return fmt.Errorf("invalid value: the value provided to be parsed was empty")
	}

	scheme, value, found := strings.Cut(raw, " ")

	if !found {
		return fmt.Errorf("invalid scheme: the scheme is missing")
	}

	switch s := strings.ToLower(scheme); s {
	case AuthorizationSchemeBasic.String():
		if err = a.parseSchemeBasic(value); err != nil {
			return err
		}

		a.scheme = AuthorizationSchemeBasic
	case AuthorizationSchemeBearer.String():
		if err = a.parseSchemeBearer(value); err != nil {
			return err
		}

		a.scheme = AuthorizationSchemeBearer
	default:
		return fmt.Errorf("invalid scheme: scheme with name '%s' is unknown", s)
	}

	a.parsed = true

	a.rawscheme = scheme
	a.value = value

	return nil
}

func (a *Authorization) parseSchemeBasic(value string) (err error) {
	var decoded []byte

	if decoded, err = base64.StdEncoding.DecodeString(value); err != nil {
		return fmt.Errorf("invalid value: failed to parse base64 basic scheme value: %w", err)
	}

	username, password, found := strings.Cut(string(decoded), ":")

	if !found {
		return fmt.Errorf("invalid value: failed to find the username password separator in the decoded basic scheme value")
	}

	if len(username) == 0 {
		return fmt.Errorf("invalid value: failed to find the username in the decoded basic value as it was empty")
	}

	if len(password) == 0 {
		return fmt.Errorf("invalid value: failed to find the password in the decoded basic value as it was empty")
	}

	a.username, a.password = username, password

	return nil
}

func (a *Authorization) parseSchemeBearer(value string) (err error) {
	return a.validateSchemeBearerValue(value)
}

func (a *Authorization) validateSchemeBearerValue(bearer string) (err error) {
	switch {
	case len(bearer) == 0:
		return fmt.Errorf("invalid value: bearer scheme value must not be empty")
	case !reToken64.MatchString(bearer):
		return fmt.Errorf("invalid value: bearer scheme value must only contain characters noted in RFC6750 2.1")
	default:
		return nil
	}
}

// ParseBytes parses the given raw header value as a byte slice.
func (a *Authorization) ParseBytes(raw []byte) (err error) {
	return a.Parse(string(raw))
}

// NewAuthorizationSchemes returns the AuthorizationSchemes for the given raw scheme names.
func NewAuthorizationSchemes(schemes ...string) AuthorizationSchemes {
	var s AuthorizationSchemes

	for _, raw := range schemes {
		switch strings.ToLower(raw) {
		case AuthorizationSchemeBasic.String():
			s = append(s, AuthorizationSchemeBasic)
		case AuthorizationSchemeBearer.String():
			s = append(s, AuthorizationSchemeBearer)
		}
	}

	return s
}

// AuthorizationSchemes represents a group of AuthorizationScheme values.
type AuthorizationSchemes []AuthorizationScheme

// Has returns true if the given AuthorizationScheme is in this group.
func (s AuthorizationSchemes) Has(scheme AuthorizationScheme) bool {
	for _, value := range s {
		if scheme == value {
			return true
		}
	}

	return false
}

// AuthorizationScheme represents the scheme of an Authorization header value.
type AuthorizationScheme int

// String returns the string representation of this AuthorizationScheme.
func (s AuthorizationScheme) String() string {
	switch s {
	case AuthorizationSchemeBasic:
		return "basic"
	case AuthorizationSchemeBearer:
		return "bearer"
	default:
		return ""
	}
}

// AuthorizationScheme values.
const (
	AuthorizationSchemeNone AuthorizationScheme = iota
	AuthorizationSchemeBasic
	AuthorizationSchemeBearer
)
