---
title: "OpenID Connect 1.0 Provider"
description: "Authelia OpenID Connect 1.0 Provider Implementation"
summary: "The OpenID Connect 1.0 Provider role is a very useful but complex feature to enhance interoperability of Authelia with other products."
date: 2025-03-23T19:03:40+11:00
draft: false
images: []
weight: 320
toc: true
aliases:
  - /r/openid-connect
  - /docs/roadmap/oidc.html
  - /roadmap/active/openid-connect
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

We have decided to implement [OAuth 2.0] and [OpenID Connect 1.0] as a beta feature. While it's relatively stable there
may inevitably be the occasional breaking change as we carefully implement each aspect of the relevant specifications.
It's suggested to use a bit more caution with this feature than most features, we do however greatly appreciate your
feedback. [OpenID Connect 1.0] and it's related endpoints are not enabled by default unless you explicitly configure the
[OpenID Connect 1.0 Provider Configuration] and [OpenID Connect 1.0 Registered Clients] sections.

As [OpenID Connect 1.0] is fairly complex (the [OpenID Connect 1.0] Provider role especially so) it's intentional that
it is both a beta and that the implemented features are part of a thoughtful roadmap. Items that are not immediately
obvious as required (i.e. bug fixes or spec features), will likely be discussed in team meetings or on GitHub issues
before being added to the list. We want to implement this feature in a very thoughtful way in order to avoid security
issues.

## OpenID Certified™

Authelia is [OpenID Certified™] to conform to the [OpenID Connect™ protocol].

{{< figure src="/images/oid-certification.jpg" class="center" process="resize 200x" >}}

For more information read the
[Integration Documentation](../../integration/openid-connect/introduction.md#openid-certified).

## Stages

This section represents the stages involved in implementation of this feature. The stages are either in order of
implementation due to there being an underlying requirement to implement them in this order, or in a rough order due to
how important or difficult to implement they are.

### Beta 1

{{< roadmap-status stage="complete" version="v4.29.0" >}}

Feature List:

* [User Consent](https://openid.net/specs/openid-connect-core-1_0.html#Consent)
* [Authorization Code Flow](https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowSteps)
* [OpenID Connect Discovery 1.0](https://openid.net/specs/openid-connect-discovery-1_0.html)
* [RS256 Signature Strategy](https://datatracker.ietf.org/doc/html/rfc7518#section-3.1)
* Per Client Scope/Grant Type/Response Type Restriction
* Per Client Authorization Policy (1FA/2FA)
* Per Client List of Valid Redirection URI's
* [Confidential Client Type](https://datatracker.ietf.org/doc/html/rfc6749#section-2.1)

### Beta 2

{{< roadmap-status stage="complete" version="v4.30.0" >}}

Feature List:

* [Userinfo Endpoint](https://openid.net/specs/openid-connect-core-1_0.html#UserInfo)
* Parameter Entropy
* Token/Code Lifespan
* Client Debug Messages
* Client Audience
* [Public Client Type](https://datatracker.ietf.org/doc/html/rfc6749#section-2.1)

### Beta 3

{{< roadmap-status stage="complete" version="v4.34.0" >}}

Feature List:

* [RFC7636: Proof Key for Code Exchange (PKCE)] for Authorization Code Flow
* Claims:
  * `preferred_username` - sending the username in this claim instead of the `sub` claim.

### Beta 4

{{< roadmap-status stage="complete" version="v4.35.0" >}}

Feature List:

* Persistent Storage
  * Tokens
  * Auditable Information
  * Subject to User Mapping
* Opaque [RFC4122] UUID v4's for subject identifiers
* Support for Pairwise and Plain subject identifier types as per [OpenID Connect Core 1.0 (Subject Identifier Types)]
  * Utilize the pairwise example method 3 as per [OpenID Connect Core 1.0 (Pairwise Identifier Algorithm)]
* Claims:
  * `sub` - replace username with opaque random [RFC4122] UUID v4
  * `amr` - authentication method references as per [RFC8176]
  * `azp` - authorized party as per [OpenID Connect Core 1.0 (ID Token)]
  * `client_id` - the Client ID as per [RFC8693 Section 4.3]
* [Cross Origin Resource Sharing] (CORS):
  * Automatically allow all cross-origin requests to the discovery endpoints
  * Automatically allow all cross-origin requests to the JSON Web Keys endpoint
  * Optionally allow cross-origin requests to the other endpoints individually

### Beta 5

{{< roadmap-status stage="complete" version="v4.37.0" >}}

Feature List:

* [JWK's backed by X509 Certificate Chains](https://datatracker.ietf.org/doc/html/rfc7517#section-4.7)
* Hashed Client Secrets
* Per-Client [Consent](https://openid.net/specs/openid-connect-core-1_0.html#Consent) Mode:
  * Explicit:
    * The default
    * Always asks for end-user consent
  * Implicit:
    * Not expressly standards compliant
    * Never asks for end-user consent
    * Not compatible with the `consent` prompt type
  * Pre-Configured:
    * Allows users to save consent sessions for a duration configured by the administrator
    * Operates nearly identically to the explicit consent mode

### Beta 6

{{< roadmap-status stage="complete" version="v4.38.0" >}}

* [RFC9068: JSON Web Token (JWT) Profile for OAuth 2.0 Access Tokens]
* [RFC9126: OAuth 2.0 Pushed Authorization Requests]
* [RFC9207: OAuth 2.0 Authorization Server Issuer Identification]
* [RFC6750: OAuth 2.0 Bearer Token Usage]
* [JWT Secured Authorization Response Mode for OAuth 2.0 (JARM)]
* [JWT Response for OAuth Token Inspection]
* [RFC7523: JSON Web Token (JWT) Profile for OAuth 2.0 Client Authentication and Authorization Grants]:
  * Client Auth Method `client_secret_jwt`
  * Client Auth Method `private_key_jwt`
* Per-Client [RFC7636: Proof Key for Code Exchange (PKCE)] Policy
* Per-Client Token Lifespans
* [OAuth 2.0 Client Credentials Grant](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4)
* Multiple Issuer JWKs:
  * `RS256`, `RS384`, `RS512`
  * `PS256`, `PS384`, `PS512`
  * `ES256`, `ES384`, `ES512`
* [Client RBAC: Users and Groups](#client-rbac-users-and-groups):
  * Policies can be mapped to individual clients and reused
  * Match criteria is only subjects as this is the only effective thing that is deterministic during the life of an
    authorization

### Beta 7

{{< roadmap-status stage="complete" version="v4.39.0" >}}

{{< callout context="danger" title="Important Notes" icon="outline/alert-octagon" >}}
This version will contain one or more breaking changes per our
[Versioning Policy](../../policies/versioning.md#experimental-features).
{{< /callout >}}

Breaking Changes:

* Default ID Token Claims (related to Claims Handling)
* Removal of Legacy Endpoints

Feature List:

* Prompt Handling
* Display Handling
* Claims Handling
  * Only include the required claims within the minted [ID Token]
  * Allow clients to use the claims parameter to request the claims which are included in the minted [ID Token]
  * Allow a user definable claims policy to allow access to custom claims and scopes
* Attribute Mapping
  * Custom Claims
  * Claims Policies
  * Expression Based Claims
* Custom Scopes
* [RFC8628: OAuth 2.0 Device Authorization Grant]
* [JSON Web Encryption](https://datatracker.ietf.org/doc/html/rfc7516)
* [Client RBAC: Networks](#client-rbac-networks)

See [OpenID Connect Core 1.0 (Mandatory to Implement Features for All OpenID Providers)].

### Beta 8

{{< roadmap-status stage="in-progress" version="v4.40.0" >}}

{{< callout context="danger" title="Important Notes" icon="outline/alert-octagon" >}}
This version will contain one or more breaking changes per our
[Versioning Policy](../../policies/versioning.md#experimental-features).
{{< /callout >}}

Breaking Changes:

* Removal of plaintext passwords except in cases that explicitly require them (HMAC-based client auth methods)
* Rework of Consent Policy

Feature List:

* In-Storage Configuration:
  * [JSON Web Key Rotation](https://openid.net/specs/openid-connect-messages-1_0-20.html#rotate.sig.keys)
  * [Multi-Issuer Configuration](#multi-issuer-configuration) (require one per Issuer URL)
  * Dynamic Client Registration:
    * Specification Based with Special Opaque Tokens (authelia_dcrt_*) as per
      [OpenID Connect Dynamic Client Registration 1.0], [RFC7591: OAuth 2.0 Dynamic Client Registration Protocol], and
      [RFC7592: OAuth 2.0 Dynamic Client Registration Management Protocol]
    * Via the CLI
    * Import from YAML:
      * Manual method
      * Bootstrap method:
        * Defaults to one time only
        * Can optionally override the database configuration
    * Salt (random) and/or Peppered (storage encryption) Client Credentials
* Subject Sectoring:
  * Pairwise Pseudonymous Identifier (PPID)
  * Sector Identifier validation.

Potential Feature List:

* Injecting Bearer JSON Web Tokens into Requests (backend authentication)
* Revoke Tokens on User Logout or Expiration

### Beta 9

{{< roadmap-status >}}

Feature List:

* [OpenID Connect Session Management 1.0](#openid-connect-session-management-10)
* [OpenID Connect Back-Channel Logout 1.0](#openid-connect-back-channel-logout-10)
* [OpenID Connect Front-Channel Logout 1.0](#openid-connect-front-channel-logout-10)
* [OpenID Connect RP-Initiated Logout 1.0](#openid-connect-rp-initiated-logout-10)
* [OpenID Connect Client-Initiated Backchannel Authentication Flow 1.0 (CIBA)](#openid-connect-client-initiated-backchannel-authentication-flow-10-ciba)

Potential Feature List:

* Injecting Bearer JSON Web Tokens into Requests (backend authentication)
* Revoke Tokens on User Logout or Expiration

### General Availability

{{< roadmap-status >}}

This stage will signify official stability guarantees surrounding this implemented feature.

### Miscellaneous

This stage lists features which individually do not fit into a specific stage and may or may not be implemented.

#### Multi-Issuer Configuration

{{< callout context="danger" title="Important Notes" icon="outline/alert-octagon" >}}
This will be a planned breaking-change as per our
[Versioning Policy](../../policies/versioning.md#experimental-features).
{{< /callout >}}

The initial design of our [OpenID Connect 1.0] implementation was before
[Multi-Domain Protection](multi-domain-protection.md) was considered. It's important for the future of Authelia that we
carefully consider the implications of this and force users to configure a issuer per domain they wish to serve
[OpenID Connect 1.0] from and each of these are completely separate logical units.

#### OAuth 2.0 Authorization Server Metadata

{{< roadmap-status stage="complete" version="v4.34.0" >}}

For more information see the [RFC8414: OAuth 2.0 Authorization Server Metadata] specification.

#### OAuth 2.0 Token Exchange

{{< roadmap-status >}}

For more information see the [RFC8693: OAuth 2.0 Token Exchange] specification.

#### OAuth 2.0 Dynamic Client Registration Protocol

{{< roadmap-status >}}

For more information see the [OAuth 2.0] website for the [RFC7591: OAuth 2.0 Dynamic Client Registration Protocol]
specification; and see both
[OAuth 2.0 Client Registration Management Protocol](#oauth-20-dynamic-client-registration-management-protocol) and
[OpenID Connect Dynamic Client Registration 1.0](#openid-connect-dynamic-client-registration-10).

See also [Beta 8](#beta-8).

#### OAuth 2.0 Dynamic Client Registration Management Protocol

{{< roadmap-status >}}

For more information see the [OAuth 2.0] website for the
[RFC7592: OAuth 2.0 Dynamic Client Registration Management Protocol] specification; and see both
[OAuth 2.0 Client Registration Protocol](#oauth-20-dynamic-client-registration-protocol) and
[OpenID Connect Dynamic Client Registration 1.0](#openid-connect-dynamic-client-registration-10).

See also [Beta 8](#beta-8).

#### OpenID Connect Dynamic Client Registration 1.0

{{< roadmap-status >}}

For more information see the [OpenID Connect 1.0] website for the [OpenID Connect Dynamic Client Registration 1.0]
specification; and see both
[OAuth 2.0 Client Registration Protocol](#oauth-20-dynamic-client-registration-protocol) and
[OAuth 2.0 Client Registration Management Protocol](#oauth-20-dynamic-client-registration-management-protocol).

See also [Beta 8](#beta-8).

#### OpenID Connect Session Management 1.0

{{< roadmap-status >}}

For more information see the [OpenID Connect 1.0] website for the [OpenID Connect Session Management 1.0] specification.

See also [Beta 9](#beta-9).

#### OpenID Connect Back-Channel Logout 1.0

{{< roadmap-status >}}

For more information see the [OpenID Connect 1.0] website for the [OpenID Connect Back-Channel Logout 1.0]
specification.

Should be implemented at a similar time to [Dynamic Client Registration](#openid-connect-dynamic-client-registration-10).

See also [Beta 9](#beta-9).

#### OpenID Connect Front-Channel Logout 1.0

{{< roadmap-status >}}

For more information see the [OpenID Connect 1.0] website for the [OpenID Connect Front-Channel Logout 1.0]
specification.

Should be implemented at the same time, or just after
[OpenID Connect Dynamic Client Registration 1.0](#openid-connect-dynamic-client-registration-10).

See also [Beta 9](#beta-9).

#### OpenID Connect RP-Initiated Logout 1.0

{{< roadmap-status >}}

See the [OpenID Connect 1.0] website for the [OpenID Connect RP-Initiated Logout 1.0] specification.

See also [Beta 9](#beta-9).

#### OpenID Connect Client-Initiated Backchannel Authentication Flow 1.0 (CIBA)

{{< roadmap-status >}}

See the [OpenID Connect 1.0] website for the [OpenID Connect Client-Initiated Backchannel Authentication Flow 1.0]
(CIBA) specification.

See also [Beta 9](#beta-9).

#### OpenID Connect 1.0 FAPI 2.0 Security Profile

{{< roadmap-status stage="in-progress" >}}

This profile is a suite of security focused features and settings which comply with several financial requirements in
various jurisdictions. While we're not expressly targeting these financial institutions the security profile itself
has many security-enhancing measures which everyone can benefit from.

See the [OpenID Connect 1.0] website for the [FAPI 2.0 Security Profile] specification, and the
[FAPI 2.0 Attacker Model].

#### End-User Scope Grants

{{< roadmap-status >}}

Allow users to choose which scopes they grant. It may be better to just allow optional claims and to avoid implementing
this feature all together.

#### Client RBAC: Users and Groups

{{< roadmap-status stage="complete" version="v4.38.0" >}}

See also [Beta 6](#beta-6) and [Client RBAC: Networks](#client-rbac-networks).

Allow the creation of custom authorization policies for [OpenID Connect 1.0]. Allow the policies to contain either users
or groups and an effective authorization policy applied to them from either `one_factor`, `two_factor`, or `deny`.

Allow these policies to be configured on one or more clients.

#### Client RBAC: Networks

{{< roadmap-status stage="complete" version="v4.39.0" >}}

See also [Beta 7](#beta-7) and [Client RBAC: Users and Groups](#client-rbac-users-and-groups).

Allow enhancing the existing custom authorization policies to include networks.

#### Preferred Username Claim

{{< roadmap-status stage="complete" version="v4.33.2" >}}

The `preferred_username` claim was missing and was fixed.

[ID Token]: https://openid.net/specs/openid-connect-core-1_0.html#IDToken
[Cross Origin Resource Sharing]: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS

[RFC8176]: https://datatracker.ietf.org/doc/html/rfc8176
[RFC8693 Section 4.3]: https://datatracker.ietf.org/doc/html/rfc8693/#section-4.3
[RFC4122]: https://datatracker.ietf.org/doc/html/rfc4122

[OpenID Connect 1.0 Provider Configuration]: ../../configuration/identity-providers/openid-connect/provider.md
[OpenID Connect 1.0 Registered Clients]: ../../configuration/identity-providers/openid-connect/clients.md

[OAuth 2.0]: https://oauth.net/2/
[OpenID Connect 1.0]: https://openid.net/connect/
[OpenID Connect Dynamic Client Registration 1.0]: https://openid.net/specs/openid-connect-registration-1_0.html
[OpenID Connect Session Management 1.0]: https://openid.net/specs/openid-connect-session-1_0.html
[OpenID Connect Back-Channel Logout 1.0]: https://openid.net/specs/openid-connect-backchannel-1_0.html
[OpenID Connect Front-Channel Logout 1.0]: https://openid.net/specs/openid-connect-frontchannel-1_0.html
[OpenID Connect RP-Initiated Logout 1.0]: https://openid.net/specs/openid-connect-rpinitiated-1_0.html
[OpenID Connect Client-Initiated Backchannel Authentication Flow 1.0]: https://openid.net/specs/openid-client-initiated-backchannel-authentication-core-1_0.html
[OpenID Connect Core 1.0 (ID Token)]: https://openid.net/specs/openid-connect-core-1_0.html#IDToken
[OpenID Connect Core 1.0 (Subject Identifier Types)]: https://openid.net/specs/openid-connect-core-1_0.html#SubjectIDTypes
[OpenID Connect Core 1.0 (Pairwise Identifier Algorithm)]: https://openid.net/specs/openid-connect-core-1_0.html#PairwiseAlg
[OpenID Connect Core 1.0 (Mandatory to Implement Features for All OpenID Providers)]: https://openid.net/specs/openid-connect-core-1_0.html#ServerMTI

[FAPI 2.0 Security Profile]: https://openid.net/specs/fapi-2_0-security-02.html
[FAPI 2.0 Attacker Model]: https://openid.net/specs/fapi-attacker-model-2_0-final.html

[RFC7636: Proof Key for Code Exchange (PKCE)]: https://datatracker.ietf.org/doc/html/rfc7636
[RFC7523: JSON Web Token (JWT) Profile for OAuth 2.0 Client Authentication and Authorization Grants]: https://datatracker.ietf.org/doc/html/rfc7523
[RFC9126: OAuth 2.0 Pushed Authorization Requests]: https://datatracker.ietf.org/doc/html/rfc9126
[RFC8414: OAuth 2.0 Authorization Server Metadata]: https://datatracker.ietf.org/doc/html/rfc8414
[RFC9207: OAuth 2.0 Authorization Server Issuer Identification]: https://datatracker.ietf.org/doc/html/rfc9207
[RFC6750: OAuth 2.0 Bearer Token Usage]: https://datatracker.ietf.org/doc/html/rfc6750
[RFC9068: JSON Web Token (JWT) Profile for OAuth 2.0 Access Tokens]: https://datatracker.ietf.org/doc/html/rfc9068
[RFC8628: OAuth 2.0 Device Authorization Grant]: https://datatracker.ietf.org/doc/html/rfc8628
[RFC7591: OAuth 2.0 Dynamic Client Registration Protocol]: https://datatracker.ietf.org/doc/html/rfc7591
[RFC7592: OAuth 2.0 Dynamic Client Registration Management Protocol]: https://datatracker.ietf.org/doc/html/rfc7592
[RFC8693: OAuth 2.0 Token Exchange]: https://datatracker.ietf.org/doc/html/rfc8693
[JWT Secured Authorization Response Mode for OAuth 2.0 (JARM)]: https://openid.net/specs/oauth-v2-jarm.html
[JWT Response for OAuth Token Inspection]: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-jwt-introspection-response
[OpenID Certified™]: https://openid.net/certification/
[OpenID Connect™ protocol]: https://openid.net/developers/how-connect-works/
