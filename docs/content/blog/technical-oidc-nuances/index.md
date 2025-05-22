---
title: "Technical: OpenID Connect 1.0 Nuances"
description: "This is a commentary on several troubling trends in the security world, as well as an explainer on some fundamental OpenID Connect 1.0 concepts."
summary: "This is a commentary on several troubling trends in the security world, as well as an explainer on some fundamental OpenID Connect 1.0 concepts."
date: 2025-05-10T07:49:15+00:00
draft: false
weight: 50
categories: ["Technical", "OpenID Connect 1.0"]
tags: ["technical", "specifications"]
contributors: ["James Elliott"]
pinned: false
homepage: false
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

----

The intended audience of this article is those curious about some of the design choices Authelia has recently made
surrounding _Claims_[^1], as well as those trying to navigate implementation of the specifications. I suspect some who are
just interested in the topic may also find this beneficial. I have seen some interest from parties within the Open
Source Community as well as the more specific Authelia Community for information like this, so hopefully it helps
satisfy some of this.

Because of the technical nature, I've intentionally not including it on the homepage, but we will probably publish more
articles like this in the future in the same [category](categories/technical).

Learning about [OpenID Connect 1.0] and the associated specifications has been quite a journey. There is a lot to
read and even more to understand. This journey has taken years even with preexisting knowledge; and the following
article represents the elements of the specifications that were not only the most time-consuming to accurately grasp,
but also what I now consider the most important elements, at least in the way I understand them today.

While writing this article it is becoming clearer and clearer to me that even though this article looks long it will
only take most interested people about 20 minutes to read. I find this frustratingly ironic. I can only hope that these
learnings help current and future [OpenID Connect 1.0] Providers and Relying Parties save time in understanding what I
now consider a wonderfully beautiful framework and technology.

The decision to write this article mainly comes from the fact that I have found myself recently and regularly discussing a
few topics associated with [OpenID Connect 1.0] and in particular some of the nuances that don't seem well understood.
I fail to see today why these concepts are hard to understand, they are very clearly detailed in the relevant
specifications; I can only summarize at least from my anecdotal experience no one appears to discuss them in detail.

I suspect most people will find this topic rather boring. However, those of you who find technical articles interesting
will probably really enjoy this, especially if your understanding of the [OpenID Connect 1.0] specifications is
non-existent or just starting to develop.

Several of these concepts are heavily linked together and in my opinion once you realize how each of these concepts
interacts with the others, there is a sort of orchestral design behind each of them; it's almost like the specification
designers actually put a lot of thought into this.

----

## Access Tokens and ID Tokens

When these specifications were first envisioned, there existed no specific format for the _Access Token_[^2]. The format
was entirely up to the implementer. Effectively an _Access Token_[^2] was completely opaque and meaningless to the party
that was utilizing it. This concept is important, as it is the foundation of the intent of the _Access Token_[^2], in
contrast the _ID Token_[^3] is strictly a _JSON Web Token_[^4].

Some confusion has developed over the years surrounding the purpose of these tokens. You see they are meant to have
meaning and to be verified only by the _Authorization Server_[^5] or integrated _Resource Server_[^6]. This is partially
because they can be completely opaque, and a couple of other reasons we'll get into. It's true that these tokens can be
Introspected regardless of if they are opaque or a _JSON Web Token_[^4], but this is typically the role of the
_Resource Server_[^6].

However as the [JSON Web Token Profile for OAuth 2.0 Access Tokens] was ratified several providers started using the
_JWT_[^4] for _Access Tokens_[^2], and people are increasingly using these to determine the identity of a user.
The problem is this is not the intent behind these tokens. In fact, for other reasons that will be discussed in the next
section of this article it is actually a harmful decision to make.

The intended purpose of [OpenID Connect 1.0] was to solve this particular issue with OAuth 2.0 as well as a few others.
How it solves this particular issue is via the _ID Token_[^3]. These tokens are designed to carry information that will
uniquely identify a user.

This becomes more visible when you look at the final audience (`aud` _Claim_[^1]) of a _ID Token_[^3] vs an
_Access Token_[^2] using the _JSON Web Token_[^4] format. According to
[RFC7519 Section 4.1.3](https://www.rfc-editor.org/rfc/rfc7519.html#section-4.1.3) the audience identifies the
recipients that the _JWT_[^4] is intended for.

The below examples are the example _ID Token_[^3] and _Access Token_[^2] which are effectively valid in content for an
_Authorization Code Flow_[^7] with the following parameters:

1. Client ID: `K2LQE4XRC54N7C2F5ZLF`
2. Authorized Audience: `https://auth.example.com/api/oidc/introspection`
3. Scopes: `openid profile email`

```json {title="ID Token"}
{
  "jti": "91de5882-ff69-46b6-b13b-165199f3191f",
  "iss": "https://auth.example.com",
  "sub": "d2fdc83d-d7ad-4ced-81d8-0bb87db4a127",
  "aud": "K2LQE4XRC54N7C2F5ZLF",
  "exp": 1745755215,
  "iat": 1745755000
}
```

```json {title="Access Token"}
{
  "jti": "f30450c1-a60c-43ab-b855-e670f84ba45a",
  "iss": "https://auth.example.com",
  "sub": "d2fdc83d-d7ad-4ced-81d8-0bb87db4a127",
  "aud": "https://auth.example.com/api/oidc/introspection",
  "exp": 1745755215,
  "iat": 1745755000,
  "client_id": "K2LQE4XRC54N7C2F5ZLF",
  "scope": "openid profile email"
}
```

As opposed to the _Authorization Code Flow_[^7] where the `sub` should normally be the _Resource Owners_[^8] subject
identifier, the Client Credentials Flow has an interesting effect on the _Access Token_[^2] where the `sub` should
normally be the Client ID. While you could technically try to use this to validate the identity of the _End-User_[^9],
this is not the intended purpose behind this _Access Token_[^2] format. In fact, it's not even guaranteed by any normal
area of the specification that this will be the case.

If you take a look the audience of the _Access Token_[^2] is not the Client ID, this is because it is not the intended
recipient and should not use it to validate the identity of a user. You will also note the audience in the _ID Token_[^3] is
the Client ID, it is required to be the Client ID, though it optionally can have additional values. This clearly
expresses that the Client is the intended recipient.

The short version of this is that the _Access Token_[^2] is meant to be understood by the Authorization Server for uses at
the _User Information Endpoint_[^10], _Introspection Endpoint_[^11], _Revocation Endpoint_[^12], and various other endpoints it
decides to implement; or the endpoints of a Resource Server with deep understanding over how to validate the token. The
_ID Token_[^3] is intended to be used by the _Relying Party_[^13]. It's used as a means of a verifiable proof the user
is a unique individual, or at the very least has granted access to their account (under normal circumstances).

Another interesting difference between the _ID Token_[^3] and _Access Token_[^2] above is that the _Access Token_[^2]
has a `scope` _Claim_[^1], but the _ID Token_[^3] does not. This may seem like a mistake but I assure you that it isn't.
There is no need for the _ID Token_[^3] to have a `scope` _Claim_[^1], as the token has a very clear intention; sharing
user identity, and it's only used for this purpose. The only clearly expressed intention behind an _Access Token_[^2]
is in the name of the token type; it's used to _access_ things. This is why the concepts of
_Scope_[^14] and _Audience_[^15] exist; rather than the token having blanket access, they limit the access to very
specific endpoints and actions in a fairly explicitly expressed way.

This concept of what kinds of things these tokens can access specifically in [OpenID Connect 1.0] is going to be touched
on later, specifically when discussing _Claims_[^1] availability.

----

## Claim Stability and Uniqueness: The effect on Identity Binding

There are various _Claims_[^1] available to implementers in most [OpenID Connect 1.0] Provider implementations. These
_Claims_[^1] vary from email addresses, to usernames, to readable names. A troubling trend that I have seen in both Enterprise and
Open Source projects is that they use whatever _Claim_[^1] they feel like to bind identities together. In fact many don't even
use them to perform any kind of binding at all, they just decide because the username or email matches that the user is
signed in.

This is not how the specification indicates this should be done. In fact rather than just leaving it up to everyone to
decide how to handle this [OpenID Connect 1.0] clearly spells out that these _Claims_[^1] must not be used for this purpose.
Instead it directs our attention the `sub` and `iss` _Claims_[^1] being the only
[Stable and Unique](https://openid.net/specs/openid-connect-core-1_0.html#ClaimStability) _Claims_[^1] that an _End-User_[^9] can
cleanly be identified by.

The [Claim Stability and Uniqueness] Section of [OpenID Connect 1.0] reads:

> The sub (subject) and iss (issuer) Claims from the ID Token, used together, are the only Claims that an RP can rely
> upon as a stable identifier for the End-User, since the sub Claim MUST be locally unique and never reassigned within
> the Issuer for a particular End-User, as described in Section 2. Therefore, the only guaranteed unique identifier for
> a given End-User is the combination of the iss Claim and the sub Claim.
>
> All other Claims carry no such guarantees across different issuers in terms of stability over time or uniqueness
> across users, and Issuers are permitted to apply local restrictions and policies. For instance, an Issuer MAY re-use
> an email Claim Value across different End-Users at different points in time, and the claimed email address for a given
> End-User MAY change over time. Therefore, other Claims such as email, phone_number, preferred_username, and name MUST
> NOT be used as unique identifiers for the End-User, whether obtained from the ID Token or the UserInfo Endpoint.

The intent and gravity of neglecting this element is very clear as soon as you realize that providers may allow users to
change their email address or username. TThese values are intended to be persistent identifiers for a user, never
changing, regardless of what other values change. Not only for security, but for the user experience to be seamless.
Just because they change their email they should not be prevented from signing into an application; and this would be
the case if you do not use the intended _Claims_[^1].

This is linked heavily to the first concept because the _Access Token_[^2] may in some way identify the _End-User_[^9], but it's
not required to. In fact the _Access Token_[^2] may not even be a _JWT_[^4], it's up to the [OpenID Connect 1.0] Provider how
they deliver it, as long as they understand it.

So some astute readers may be thinking. Why do these _Claims_[^1] exist in the first place then? Well it comes down to three
primary functions. Obviously these functions are not an exhaustive list, but it should be enough to explain the
principles.

The first function of these _Claims_[^1] is to provide helpful information or hints to the _Relying Party_[^13] during a
Registration Flow, or in some instances an Identity Binding Flow (i.e. binding the `sub` and `iss` _Claims_[^1] to an
identity that already exists) when they're not already logged in. For example they may prefill a form.

The second function of these _Claims_[^1] is to provide the _Relying Party_[^13] with information about a _Resource Owner_[^8]
with an already bound identity that they may not want to store themselves; such as they may perform a Authorization Flow
to temporarily obtain the _Resource Owners_[^8] address information for a purchase.

The third function of these _Claims_[^1] is to provide the _Relying Party_[^13] a way to obtain updated details about a
_Resource Owner_[^8] who already has a bound identity. For example updating their contact details. This leads naturally
into the next concept.

----

## Claims Availability

There are various ways to request and grant _Claims_[^1] in the [OpenID Connect 1.0] specification. Many assume that the
_Claims_[^1] are either exclusively available in the _ID Token_[^3], will always be in the _ID Token_[^3], or even worse
as we've previously found out in the _Access Token_[^2]. The assumption stems from the fact certain scopes grant the
_Relying Party_[^13] access to certain sets of _Claims_[^1]. But what does the specification really intend?

> An identity token represents the outcome of an authentication process. It contains at a bare minimum an identifier for
> the user (called the sub aka subject claim) and information about how and when the user authenticated. It can contain
> additional identity data.

This comes directly from the
[OpenID Foundations "How OpenID Connect Works" article](https://openid.net/developers/how-connect-works/). This clearly
indicates the _ID Token_[^3] contains a bare minimum information to identify a user as well as information about _how_
and _when_ the user authenticated, and that it **_can_** contain additional identity data.

If we dive into the _Claims_[^1] that are standard in the _ID Token_[^3], it has a very minimal set of _Claims_[^1]
by default, and it's clear the ways in which a _Relying Party_[^13] may request additional _Claims_[^1] and in which
scenarios it must contain additional _Claims_[^1].

### Scope Parameter

Scopes seem to be mostly be assumed to always mint an _ID Token_[^3] with all the scopes relevant _Claims_[^1]. Why is
this assumption flawed?

The section on [Requesting Claims using Scope Values] has a crucial passage regarding this:

> The Claims requested by the profile, email, address, and phone scope values are returned from the UserInfo Endpoint,
> as described in Section 5.3.2, when a response_type value is used that results in an Access Token being issued.

These scopes are very clearly not intended to include the _Claims_[^1] in the _ID Token_[^3]. In fact the
_User Information Endpoint_[^10] is meant to return them. The specification does not specifically prevent a
[OpenID Connect 1.0] Provider from returning them in the _ID Token_[^3], but it strongly suggests you shouldn't do
this normally.

There's another crucial sentence directly after the first one.

> However, when no Access Token is issued (which is the case for the response_type value id_token), the resulting Claims
> are returned in the ID Token.

This seems to solidify this point. When the _Implicit Flow_[^16] is used, as the _Implicit Flow_[^16] is the only flow
that potentially will not result in an _Access Token_[^2]; the [OpenID Connect 1.0] Provider should mint an
_ID Token_[^3] that's populated with the _Claims_[^1] normally accessible at the _User Information Endpoint_[^10]. In fact
it's only in one variation of the _Implicit Flow_[^16], when the `response_type` parameter is only `id_token`. In this
instance the _ID Token_[^3] is categorically required to be populated with every one of the _Claims_[^1] the
_Access Token_[^2] would normally be able to access at the _User Information Endpoint_[^10].

It's amazing how this neatly ties back into the use case for the _Access Token_[^2] at the very start of this article.
It's intended for use cases like accessing specific API's at the Authorization Server. If you then consider the fact the
_ID Token_[^3] is a static snapshot of the unique identity of a _Resource Owner_[^8] as described in the above concept
surrounding Claim Stability and Uniqueness, and then extrapolate that the request to the
_User Information Endpoint_[^10] could realistically have the most up-to-date information about the user; I feel like
it all just makes complete sense.

The question is if a _Relying Party_[^13] is not intending on using the _Access Token_[^2] to access the
_User Information Endpoint_[^10] why are they not using the _Implicit Flow_[^16] with the _Response Mode_[^17] `form_post`
since this flow is intended specifically to identify the user. Since they will not be issued an _Access Token_[^2], and
only the _ID Token_[^3], and it's performed over `form_post`, and the _ID Token_[^3] is signed and potentially encrypted
multiple times; with the exception of the lack of client authentication most of the issues surrounding this flow
mitigated by only returning the _ID Token_[^3] via `form_post`. That being said I think the security concerns are
definitely a reasonable explanation; but the privacy and security benefits of using the _User Information Endpoint_[^10]
are too strong an argument to not implement this properly.

What do I mean by privacy and security benefits? Well the _User Information Endpoint_[^10] requires an active
_Access Token_[^2] to obtain the _Claims_[^1]. This means even if the _ID Token_[^3] does not need to contain privacy
sensitive information. An added benefit of this approach is that the _Access Token_[^2] can be revoked, and security
conscious _Authorization Servers_[^5] will automatically mint a new _Refresh Token_[^19] at the same time it mints a
new _Access Token_[^2] during the _Refresh Token Flow_[^18] which effectively rotates both of these tokens as the old
ones should be intentioanlly revoked.


### Claims Parameter

To cement the above point there is actually another means by which clients using **_any_** flow other than the
_Implicit Flow_[^16] which only returns an _ID Token_[^3] can obtain additional _Claims_[^1] in the _ID Token_[^3].

This is done via the [Claims Parameter]. The [Claims Parameter] allows requesting specific _Claims_[^1] be present in the
_ID Token_[^3]. In addition it has the added benefit of allowing granular requests for specific _Claims_[^1] rather than
granting an entire _Scope_[^14] which may have many useless _Claims_[^1] to the _Relying Party_[^13].

This clearly has several impactful elements to security, privacy, and usability. This is also the parameter that is used
to indicate elements which are optional to consent to. I suspect most users have seen these dialogs which ask the user
what properties they want to allow the _Relying Party_[^13] to be able to access, even if they were not fully conscious of it.

The advantages of using the _Claims_[^1] parameter makes it quite a desirable feature. You can specifically request the
`openid` scope, and request the specific _Claims_[^1] you need access to, and you can request that these _Claims_[^1] are made
available either at the _User Information Endpoint_[^10] which is better for security and privacy, or in the
_ID Token_[^3]. It's just too powerful not to use it.

## Footnotes

[^1]: A Claim is some information that has been asserted about an Entity such as a _End-User_[^9] or
_Resource Owner_[^8]. This is also specifically defined in the [OpenID Connect 1.0 Core Terminology] as well as the
[Claims] section which describes them in detail.

[^2]: An Access Token is traditionally an opaque token which is described in
[Section 1.4 of RFC6749](https://datatracker.ietf.org/doc/html/rfc6749#section-1.4) or sometimes a _JSON Web Token_[^4]
which is either completely proprietary or described in [RFC9068](https://datatracker.ietf.org/doc/html/rfc9068). The
Access Token is used to access resources.

[^3]: An ID Token otherwise known as an Identity Token is a _JSON Web Token_[^4] which is described in
[ID Token Section of OpenID Connect 1.0 Core](https://openid.net/specs/openid-connect-core-1_0.html). The ID Token is
intentionally designed to identify a unique user and has a strictly defined format and contents.
This is also specifically defined in the [OpenID Connect 1.0 Core Terminology].

[^4]: A JSON Web Token or JWT has multiple serialization forms, in our usages of them currently we use the compact
serialization and this is what we're referring to. The JSON Web Token (when signed, not encrypted) itself consists of
three distinct parts, the header which contains important metadata about the token format, the body which contains the
_Claims_[^1], and the signature which is a cryptographic hash of the other two parts. Both the header and the body are
minimized JSON objects which are Base64 URL encoded. JSON Web Tokens are described in detail in
[RFC7519](https://datatracker.ietf.org/doc/html/rfc7519).

[^5]: The Authorization Server has the role of authorizing access to resources held on the _Resource Server_[^6].

[^6]: The Resource Server has the role of holding resources which must be granted access to by the _Resource Owner_[^8]
which is validated by the _Authorization Server_[^5].

[^7]: The Authorization Code Flow is a flow which returns an Authorization Code in the Authorization Response. This
Authorization Code is short lived, and is exchanged at the Token Endpoint along with any client authentication
requirements for the minted tokens.

[^8]: The Resource Owner is the owner of the information being requested, this is typically the user, but also can be
the _Relying Party_[^13].

[^9]: The End-User is the human participant. This is a type of _Resource Owner_[^8]. This is also specifically defined
in the [OpenID Connect 1.0 Core Terminology].

[^10]: The User Information Endpoint is an OAuth 2.0 secured endpoint that has information about a _Resource Owner_[^8]
commonly referred to as a user. You can read more about the User Information Endpoint by reading the
[UserInfo Endpoint Section of OpenID Connect 1.0 Core](https://openid.net/specs/openid-connect-core-1_0.html#UserInfo).

[^11]: The Introspection Endpoint is used to obtain information about an Access Token regardless of the format. You
can read more about Token Introspection in [RFC7662](https://datatracker.ietf.org/doc/html/rfc7662).

[^12]: The Revocation Endpoint is used to revoke an Access Token and/or _Refresh Token_[^19]. You
can read more about Token Introspection in [RFC7009](https://datatracker.ietf.org/doc/html/rfc7009).

[^13]: The Relying Party is the party which relies on the OpenID Connect 1.0 Provider to process the Authorization Flow
and the _Resource Owner_[^8] to grant access to the information.

[^14]: The Scope defines specific actions a token is permitted to do or information they have access to. In
[OpenID Connect 1.0] the most common use for them is to define as set of _Claims_[^1] the token can access. You can read more
about Scopes in the [OAuth Defining Scopes Explainer](https://www.oauth.com/oauth2-servers/scope/defining-scopes/).

[^15]: The Audience describes the intended recipient that the token (especially a _JSON Web Token_[^4]). This is
represented as an array of strings which either contain URI's or some other uniquely identifiable name for the
recipient. This is stored in the `aud` _Claim_[^1] of _JSON Web Token_[^4] if applicable.

[^16]: The Implicit Flow is a flow which directly returns the requested tokens within the Authorization Response. It does
not exchange any short-lived code for the requested tokens. This flow is traditionally less secure due to the fact no
client authentication occurs in order to obtain the tokens. The Implicit Flow is described in detail in the
[Implicit Flow Section of OpenID Connect 1.0 Core](https://openid.net/specs/openid-connect-core-1_0.html#ImplicitFlowAuth).
This is also specifically defined in the [OpenID Connect 1.0 Core Terminology].

[^17]: The Response Mode describes how the Authorization Endpoint responds to Authorization Requests. There are three
primary modes; `query` which redirects the _Resource Owner_[^8] with the response in the URI query parameters,
`fragment` which redirects the _Resource Owner_[^8] with the response in the URI fragment parameters, and `form_post`
which performs a `POST` request with the response in the request body.

[^18]: The Refresh Token Flow is a flow which exchanges a _Refresh Token_[^19] at the Token Endpoint for new tokens with
the same security characteristics or narrowed security characteristics.

[^19]: The Refresh Token is typically a  completely opaque token which can be used to reissue other tokens.

[OpenID Connect 1.0]: https://openid.net/specs/openid-connect-core-1_0.html
[Requesting Claims using Scope Values]: https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
[Claims]: https://openid.net/specs/openid-connect-core-1_0.html#Claims
[Claims Parameter]: https://openid.net/specs/openid-connect-core-1_0.html#ClaimsParameter
[JSON Web Token Profile for OAuth 2.0 Access Tokens]: https://datatracker.ietf.org/doc/html/rfc9068
[OpenID Connect 1.0 Core Terminology]: https://openid.net/specs/openid-connect-core-1_0.html#Terminology
[Claim Stability and Uniqueness]: https://openid.net/specs/openid-connect-core-1_0.html#ClaimStability
