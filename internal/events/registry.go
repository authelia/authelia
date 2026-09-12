// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"sort"
	"strings"
)

// Descriptor describes a registered event type.
type Descriptor struct {
	// Type is the registered event type name.
	Type string

	// DataSchema is the absolute URL of the published JSON Schema for this type's data.
	DataSchema string

	// New returns an empty payload of the type this descriptor describes.
	New func() Data
}

var registry = map[string]Descriptor{}

// Register adds a Descriptor to the registry. It panics on a duplicate registration, which can only be a programming
// error as registration happens exclusively at init time.
func Register(descriptor Descriptor) {
	if _, ok := registry[descriptor.Type]; ok {
		panic("events: duplicate registration of event type " + descriptor.Type)
	}

	if descriptor.DataSchema == "" {
		descriptor.DataSchema = schemaBaseURL + descriptor.Type + ".json"
	}

	registry[descriptor.Type] = descriptor
}

// Lookup returns the Descriptor for a registered event type.
func Lookup(name string) (descriptor Descriptor, ok bool) {
	descriptor, ok = registry[name]

	return descriptor, ok
}

// IsRegistered returns true when the name is a registered event type.
func IsRegistered(name string) (ok bool) {
	_, ok = registry[name]

	return ok
}

// Registered returns every registered event type name in lexical order.
func Registered() (names []string) {
	names = make([]string, 0, len(registry))

	for name := range registry {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// Match resolves a configuration selector to the registered event type names it selects. A selector is either an exact
// type name or a prefix glob ending in an asterisk. A selector which selects nothing returns an empty slice, which the
// configuration validator treats as an error.
func Match(selector string) (names []string) {
	if !strings.HasSuffix(selector, "*") {
		if IsRegistered(selector) {
			return []string{selector}
		}

		return nil
	}

	prefix := strings.TrimSuffix(selector, "*")

	for name := range registry {
		if strings.HasPrefix(name, prefix) {
			names = append(names, name)
		}
	}

	sort.Strings(names)

	return names
}

// configuration validation runs.
//
//nolint:gochecknoinits // The registry is intentionally populated at init time so that it is complete before
func init() {
	for _, name := range []string{TypeUserPasswordChanged, TypeUserPasswordReset} {
		Register(Descriptor{Type: name, New: func() Data { return &DataUserPassword{Type: name} }})
	}

	for _, name := range []string{
		TypeUserCredentialTOTPAdded, TypeUserCredentialTOTPRemoved,
		TypeUserCredentialWebAuthnAdded, TypeUserCredentialWebAuthnRemoved,
	} {
		Register(Descriptor{Type: name, New: func() Data { return &DataUserCredential{Type: name} }})
	}

	Register(Descriptor{Type: TypeUserIdentityVerificationStarted, New: func() Data { return &DataIdentityVerification{} }})
	Register(Descriptor{Type: TypeUserSessionElevationRequested, New: func() Data { return &DataSessionElevation{} }})
	Register(Descriptor{Type: TypeSystemStartupCheck, New: func() Data { return &DataStartupCheck{} }})

	for _, name := range []string{TypeSecurityAuthenticationSucceeded, TypeSecurityAuthenticationFailed} {
		Register(Descriptor{Type: name, New: func() Data { return &DataAuthentication{Type: name} }})
	}

	for _, name := range []string{TypeSecurityBanApplied, TypeSecurityBanExpired} {
		Register(Descriptor{Type: name, New: func() Data { return &DataBan{Type: name} }})
	}
}
