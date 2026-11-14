// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Package conformance provides the OpenID Connect 1.0 conformance suite plan and client definitions Authelia is
// certified against. It is shared by the authelia-gen conformance generator and the OIDCConformance integration suite
// so that a plan's client identifiers, secrets and redirect URIs cannot drift from the clients Authelia is configured
// with.
package conformance

const (
	// NameConfig is the builder name of the Config OP profile.
	NameConfig = "config"

	// NameBasic is the builder name of the Basic OP profile.
	NameBasic = "basic"

	// NameBasicFormPost is the builder name of the Basic OP profile using the form post response mode.
	NameBasicFormPost = "basic-form-post"

	// NameHybrid is the builder name of the Hybrid OP profile.
	NameHybrid = "hybrid"

	// NameHybridFormPost is the builder name of the Hybrid OP profile using the form post response mode.
	NameHybridFormPost = "hybrid-form-post"

	// NameImplicit is the builder name of the Implicit OP profile.
	NameImplicit = "implicit"

	// NameImplicitFormPost is the builder name of the Implicit OP profile using the form post response mode.
	NameImplicitFormPost = "implicit-form-post"

	// NameRelyingPartyBasic is the builder name of the Basic RP profile.
	NameRelyingPartyBasic = "rp-basic"

	// NameRelyingPartyBasicFormPost is the builder name of the Basic RP profile using the form post response mode.
	NameRelyingPartyBasicFormPost = "rp-basic-form-post"

	// NameRelyingPartyConfig is the prefix of the builder names of the Config RP profile, which has a builder for each
	// of its modules.
	NameRelyingPartyConfig = "rp-config"
)

var (
	// RelyingPartyConfigModules are the modules of the Config RP profile in the order of its plan.
	RelyingPartyConfigModules = []RelyingPartyModule{
		{Suffix: "discovery", Name: "oidcc-client-test-discovery-openid-config"},
		{Suffix: "jwks", Name: "oidcc-client-test-discovery-jwks-uri-keys"},
		{Suffix: "issuer", Name: "oidcc-client-test-discovery-issuer-mismatch"},
		{Suffix: "sig-none", Name: "oidcc-client-test-idtoken-sig-none"},
		{Suffix: "signing", Name: "oidcc-client-test-signing-key-rotation-just-before-signing"},
		{Suffix: "rotation", Name: "oidcc-client-test-signing-key-rotation"},
	}
)
