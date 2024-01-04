package model

import (
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NewAuthorization creates a pointer to an Authorization and allocates it.
func NewAuthorization() *Authorization {
	return &Authorization{}
}

// Authorization represents a decoded authorization header or one to be encoded.
type Authorization struct {
	parsed    bool
	scheme    AuthorizationScheme
	rawscheme string
	value     string
	username  string
	password  string
}

// SchemeRaw returns the raw scheme value.
func (a *Authorization) SchemeRaw() string {
	return a.rawscheme
}

// Scheme returns the scheme enum.
func (a *Authorization) Scheme() AuthorizationScheme {
	return a.scheme
}

// Value returns the value portion of the header.
func (a *Authorization) Value() string {
	return a.value
}

// EncodeHeader encodes a header for use with a HTTP request.
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

// Basic returns the username and password if this was decoded from a header which was set with the basic scheme.
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

// BasicUsername returns the username if this was decoded from a header which was set with the basic scheme.
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

// EncodeBasic encodes the Authorization using a username and password.
func (a *Authorization) EncodeBasic(username, password string) (err error) {
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

// EncodeBearer encodes the Authorization with the value of a bearer token.
func (a *Authorization) EncodeBearer(bearer string) (err error) {
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

// ParseBytes is a helper function to parse the raw header value from bytes into the Authorization.
func (a *Authorization) ParseBytes(raw []byte) (err error) {
	return a.Parse(string(raw))
}

// Parse the raw header value into the Authorization.
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
		a.scheme = AuthorizationSchemeUnknown
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

// NewAuthorizationSchemes parses text values into AuthorizationSchemes case-insensitively.
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

// AuthorizationSchemes is a helper type for a list of AuthorizationScheme.
type AuthorizationSchemes []AuthorizationScheme

// Has returns true if the AuthorizationScheme exists within the AuthorizationSchemes.
func (s AuthorizationSchemes) Has(scheme AuthorizationScheme) bool {
	for _, value := range s {
		if scheme == value {
			return true
		}
	}

	return false
}

// AuthorizationScheme is a low cost representation of an Authorization scheme.
type AuthorizationScheme int

// String returns a string representation of a AuthorizationScheme.
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

const (
	AuthorizationSchemeNone AuthorizationScheme = iota
	AuthorizationSchemeUnknown
	AuthorizationSchemeBasic
	AuthorizationSchemeBearer
)
