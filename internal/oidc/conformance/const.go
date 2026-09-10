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
)

const (
	planConfig           = "conformance-config"
	planBasic            = "conformance-basic"
	planBasicFormPost    = "conformance-basic-form-post"
	planImplicit         = "conformance-implicit"
	planImplicitFormPost = "conformance-implicit-form-post"
	planHybrid           = "conformance-hybrid"
	planHybridFormPost   = "conformance-hybrid-form-post"
)
