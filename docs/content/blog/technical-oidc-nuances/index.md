---
title: "Technical: OpenID Connect 1.0 Nuances"
description: "This is a commentary on several troubling trends in the security world, as well as an explainer on some fundamental OpenID Connect 1.0 concepts."
summary: "This is a commentary on several troubling trends in the security world, as well as an explainer on some fundamental OpenID Connect 1.0 concepts."
date: 2025-05-10T17:45:53+10:00
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

The intended audience of this article are those curious about some of the design choices Authelia has made recently
surrounding claims, as well as those trying to navigate implementation of the specifications. I suspect some who are
just interested in the topic may also find this beneficial. I have seen some interest from parties within the Open
Source Community as well as the more specific Authelia Community for information like this, so hopefully it helps
satisfy some of this.

Because of the technical nature we're intentionally not including it on the homepage; but we will probably publish more
articles like this in the future in the same [category](categories/technical).

Learning about [OpenID Connect 1.0] and the associated specifications has been quite a journey. There is a lot to
read and even more to understand. This journey has taken years even with preexisting knowledge; and the following
article represents the elements of the specifications that were not only the more time-consuming ones to grasp entirely,
but also what I now consider the most important ones; at least in the way I understand them today.

While writing this article it is becoming clearer and clearer to me that even though this article looks long it will
only take most interested people about 20 minutes to read. I find this frustratingly ironic. I can only hope that these
learnings help current and future [OpenID Connect 1.0] Providers and Relying Parties save time in understanding what I
now consider a wonderfully beautiful framework and technology.

The decision to write this article mainly comes the fact that I have found myself recently and regularly discussing a
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

When these specifications were first envisioned, there existed no specific format for the _Access Token_[^1]. The format
was entirely up to the implementer. Effectively an _Access Token_[^1] was completely opaque and meaningless to the party
that was utilizing it. This concept is important, as it is the foundation of the intent of the _Access Token_[^1]. In
contrast the _ID Token_[^2] is very strictly a _JSON Web Token_[^3].

Some confusion has developed over the years surrounding the purpose of these tokens. You see they are meant to only be
understood and verified by the _Authorization Server_[^4] or integrated _Resource Server_[^5]. They can be completely opaque.

However as several providers started using the _JSON Web Token_[^3] or _JWT_[^3] for these tokens, and the
[JSON Web Token Profile for OAuth 2.0 Access Tokens] was ratified, people are increasingly using these to determine the
identity of a user. The problem is this is not the intent behind these tokens. In fact for other reasons that will be
discussed in the next section of this article it is actually a harmful decision to make.

The intended purpose of [OpenID Connect 1.0] was to solve this particular issue with OAuth 2.0 as well as a few others.
How it solves this particular issue is via the _ID Token_[^2]. These tokens are designed to carry information that will
uniquely identify a user.

This becomes more visible when you look at the final audience (`aud` claim) of a _ID Token_[^2] vs an
_Access Token_[^1] using the _JSON Web Token_[^3] format. According to
[RFC7519 Section 4.1.3](https://www.rfc-editor.org/rfc/rfc7519.html#section-4.1.3) the audience identifies the
recipients that the _JWT_[^3] is intended for.

The below examples are the example _ID Token_[^2] and _Access Token_[^1] which are effectively valid in content for an
_Authorization Code Flow_[^6] with the following parameters:

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

As opposed to the _Authorization Code Flow_[^6] where the `sub` should normally be the Resource Owner's subject identifier,
the Client Credentials Flow has an interesting effect on the _Access Token_[^1] where the `sub` should normally be the
Client ID. While you could technically try to use this to validate the identity of the end-user this is not the intended
purpose behind this _Access Token_[^1] format. In fact it's not even guaranteed by any normal area of the specification
that this will be the case.

If you take a look the audience of the _Access Token_[^1] is not the Client ID, this is because it is not the intended
recipient and should not use it to validate the identity of a user. You will also note the audience in the _ID Token_[^2] is
the Client ID, it is required to be the Client ID, though it optionally can have additional values. This clearly
expresses that the Client is the intended recipient.

The short version of this is that the _Access Token_[^1] is meant to be understood by the Authorization Server for uses at
the _User Information Endpoint_[^7], _Introspection Endpoint_[^8], _Revocation Endpoint_[^9], and various other endpoints it
decides to implement; or the endpoints of a Resource Server with deep understanding over how to validate the token. The
_ID Token_[^2] is intended to be used by the Relying Party. It's used as a means of a verifiable proof the user is a unique
individual, or at the very least has granted access to their account (under normal circumstances).

Another interesting difference between the _ID Token_[^2] and _Access Token_[^1] above is that the _Access Token_[^1] has a `scope`
claim, but the _ID Token_[^2] does not. This may seem like a mistake but I assure you that it isn't. There is no need for
the _ID Token_[^2] to have a `scope` claim, as the token has a very clear intention; sharing user identity, and it's only
used for this purpose. The only clearly expressed intention behind an _Access Token_[^1] is in the name of the token type;
it's used to _access_ things. This is why the concepts of _Scope_[^10] and _Audience_[^11] exist; rather than the token having
blanket access, they limit the access to very specific endpoints and actions in a fairly explicitly expressed way.

This concept of what kinds of things these tokens can access specifically in [OpenID Connect 1.0] is going to be touched
on later, specifically when discussing claims availability.

----

## Claim Stability and Uniqueness: The effect on Identity Binding

There are various [Claims] available to implementers in most [OpenID Connect 1.0] Provider implementations. These [Claims]
vary from email addresses, to usernames, to readable names. A troubling trend that I have seen in both Enterprise and
Open Source projects is that they use whatever claim they feel like to bind identities together. In fact many don't even
use them to perform any kind of binding at all, they just decide because the username or email matches that the user is
signed in.

This is not how the specification indicates this should be done. In fact rather than just leaving it up to everyone to
decide how to handle this [OpenID Connect 1.0] clearly spells out that these [Claims] must not be used for this purpose.
Instead it directs our attention the `sub` and `iss` [Claims] being the only
[Stable and Unique](https://openid.net/specs/openid-connect-core-1_0.html#ClaimStability) [Claims] that an end-user can
cleanly be identified by.

The intent and gravity of neglecting this element is very clear as soon as you realize that providers may allow users to
change their email address or username. These values are meant to be anchored to a user, never changing, regardless of
what other values change. Not only for security, but for the user experience to be seamless. Just because they change
their email they should not be prevented from signing into an application; and this would be the case if you do not use
the intended [Claims].

This is linked heavily to the first concept because the _Access Token_[^1] may in some way identify the end-user, but it's
not required to. In fact the _Access Token_[^1] may not even be a _JWT_[^3], it's up to the [OpenID Connect 1.0] Provider how
they deliver it, as long as they understand it.

So some astute readers may be thinking. Why do these [Claims] exist in the first place then? Well it comes down to three
primary functions. Obviously these functions are not an exhaustive list, but it should be enough to explain the
principles.

The first function of these [Claims] is to provide helpful information or hints to the Relying Party during a
Registration Flow, or in some instances an Identity Binding Flow (i.e. binding the `sub` and `iss` claims to an identity
that already exists) when they're not already logged in. For example they may prefill a form.

The second function of these [Claims] is to provide the Relying Party with information about a Resource Owner with an
already bound identity that they may not want to store themselves; such as they may perform a Authorization Flow to
temporarily obtain the Resource Owners address information
for a purchase.

The third function of these [Claims] is to provide the Relying Party a way to obtain updated details about a Resource
Owner who already has a bound identity. For example updating their contact details. This neaty ties into the next
concept.

----

## Claims Availability

There are various ways to request and grant [Claims] in the [OpenID Connect 1.0] specification. Many assume that the
[Claims] are either exclusively available in the _ID Token_[^2], will always be in the _ID Token_[^2], or even worse as we've
previously found out in the _Access Token_[^1]. The assumption stems from the fact certain scopes grant the Relying Party
access to certain sets of [Claims]. But what does the specification really intend?

Well if we dive into the [Claims] that are standard in the _ID Token_[^2], it has a very minimal set of [Claims] by default.
It can contain more, and in some scenarios it should contain a lot of [Claims].

### Scope Parameter

Scopes seem to be mostly be assumed to always mint an _ID Token_[^2] with all the scopes relevant claims. Why is this
assumption mostly flawed?

The section on [Requesting Claims using Scope Values] has a crucial passage regarding this:

> The Claims requested by the profile, email, address, and phone scope values are returned from the UserInfo Endpoint,
> as described in Section 5.3.2, when a response_type value is used that results in an Access Token being issued.

These scopes are very clearly not intended to include the [Claims] in the _ID Token_[^2]. In fact the
_User Information Endpoint_[^7] is meant to return them. The specification does not specifically prevent a
[OpenID Connect 1.0] Provider from returning them in the _ID Token_[^2], but it strongly suggests you shouldn't do
this normally.

There's another crucial sentence directly after the first one.

> However, when no Access Token is issued (which is the case for the response_type value id_token), the resulting Claims
> are returned in the ID Token.

This seems to solidify this point. When the _Implicit Flow_[^12] is used, as the _Implicit Flow_[^12] is the only flow
that does not result in an _Access Token_[^1]; the [OpenID Connect 1.0] Provider should mint an _ID Token_[^2] that's
populated with the claims normally accessible at the _User Information Endpoint_[^7]. In fact it's only in one variation of the
_Implicit Flow_[^12], when the `response_type` parameter is only `id_token`. In this instance the _ID Token_[^2] is
categorically required to be populated with every one of the [Claims] the _Access Token_[^1] would normally be able to
access at the _User Information Endpoint_[^7].

It's amazing how this neatly ties back into the use case for the _Access Token_[^1] at the very start of this article.
It's intended for use cases like accessing specific API's at the Authorization Server. If you then consider the fact the
_ID Token_[^2] is a static snapshot of the unique identity of a _Resource Owner_[^13] as described in the above concept
surrounding Claim Stability and Uniqueness, and then extrapolate that the request to the _User Information Endpoint_[^7] could
realistically have the most up-to-date information about the user; I feel like it all just makes complete sense.

The question is if a Relying Party is not intending on using the _Access Token_[^1] to access the
_User Information Endpoint_[^7] why are they not using the _Implicit Flow_[^12] with the _Response Mode_[^14] `form_post`
since this flow is intended specifically to identify the user. Since they will not be issued an _Access Token_[^1], and
only the _ID Token_[^2], and it's performed over `form_post`, and the _ID Token_[^2] is signed and potentially encrypted
multiple times; with the exception of the lack of client authentication most of the issues surrounding this flow
mitigated by only returning the _ID Token_[^2] via `form_post`. That being said I think the security concerns are
definitely a reasonable explanation; but the privacy and security benefits of using the _User Information Endpoint_[^7]
are too strong an argument to not implement this properly.

What do I mean by privacy and security benefits? Well the _User Information Endpoint_[^7] requires an active
_Access Token_[^1] to obtain the claims. This means even if the _ID Token_[^2] does not need to contain privacy
sensitive information, and if _Access Token_[^1] is compromised it can be revoked. In addition security conscious
_Authorization Servers_[^4] will automatically mint a new _Access Token_[^1] and revoke the old one during the
_Refresh Token Flow_[^15].


### Claims Parameter

To cement the above point there is actually another means by which clients using **_any_** flow other than the
_Implicit Flow_[^12] which only returns an _ID Token_[^2] can obtain additional claims in the _ID Token_[^2].

This is done via the [Claims Parameter]. The [Claims Parameter] allows requesting specific claims be present in the
_ID Token_[^2]. In addition it has the added benefit of allowing granular requests for specific claims rather than
granting an entire _Scope_[^10] which may have many useless claims to the Relying Party.

This clearly has several impactful elements to security, privacy, and usability. This is also the parameter that is used
to indicate elements which are optional to consent to. I suspect most users have seen these dialogs which ask the user
what properties they want to allow the Relying Party to be able to access, even if they were not fully conscious of it.

The advantage of using the claims parameter makes it quite desirable. You can specifically request the `openid` scope,
and request the specific claims you need access to, and you can request that these claims are made available either
at the _User Information Endpoint_[^7] which is better for security and privacy, or in the _ID Token_[^2]. It's just too
powerful not to use it.

## Footnotes

[^1]: An Access Token is traditionally an opaque token which is described in
[Section 1.4 of RFC6749](https://datatracker.ietf.org/doc/html/rfc6749#section-1.4) or sometimes a _JSON Web Token_[^3]
which is either completely proprietary or described in [RFC9068](https://datatracker.ietf.org/doc/html/rfc9068). The
Access Token is used to access resources.

[^2]: An ID Token is a _JSON Web Token_[^3] which is described in
[ID Token Section of OpenID Connect 1.0 Core](https://openid.net/specs/openid-connect-core-1_0.html). The ID Token is
intentionally designed to identify a unique user and has a strictly defined format and contents.

[^3]: A JSON Web Token or JWT has multiple serialization forms, in our usages of them currently we use the compact
serialization and this is what we're referring to. The JSON Web Token (when signed, not encrypted) itself consists of
three distinct parts, the header which contains important metadata about the token format, the body which contains the
claims, and the signature which is a cryptographic hash of the other two parts. Both the header and the body are
minimized JSON objects which are Base64 URL encoded. JSON Web Tokens are described in detail in
[RFC7519](https://datatracker.ietf.org/doc/html/rfc7519).

[^4]: The Authorization Server has the role of authorizing access to resources held on the _Resource Server_[^5].

[^5]: The Resource Server has the role of holding resources which must be granted access to by the _Resource Owner_[^13]
which is validated by the _Authorization Server_[^4].

[^6]: The Authorization Code Flow is a flow which returns an Authorization Code in the Authorization Response. This
Authorization Code is short lived, and is exchanged at the Token Endpoint along with any client authentication
requirements for the minted tokens.

[^7]: The User Information Endpoint is an OAuth 2.0 secured endpoint that has information about a resource owner
commonly referred to as a user. You can read more about the User Information Endpoint by reading the
[UserInfo Endpoint Section of OpenID Connect 1.0 Core](https://openid.net/specs/openid-connect-core-1_0.html#UserInfo).

[^8]: The Introspection Endpoint is used to obtain information about an Access Token regardless of the format. You
can read more about Token Introspection in [RFC7662](https://datatracker.ietf.org/doc/html/rfc7662).

[^9]: The Revocation Endpoint is used to revoke an Access Token and/or _Refresh Token_[^16]. You
can read more about Token Introspection in [RFC7009](https://datatracker.ietf.org/doc/html/rfc7009).

[^10]: The Scope defines specific actions a token is permitted to do or information they have access to. In
[OpenID Connect 1.0] the most common use for them is to define as set of claims the token can access. You can read more
about Scopes in the [OAuth Defining Scopes Explainer](https://www.oauth.com/oauth2-servers/scope/defining-scopes/).

[^11]: The Audience describes the intended recipient that the token (especially a _JSON Web Token_[^3]). This is
represented as an array of strings which either contain URI's or some other uniquely identifiable name for the
recipient. This is stored in the `aud` claim of _JSON Web Token_[^3] if applicable.

[^12]: The Implicit Flow is a flow which directly returns the requested tokens within the Authorization Response. It does
not exchange any short-lived code for the requested tokens. This flow is traditionally less secure due to the fact no
client authentication occurs in order to obtain the tokens. The Implicit Flow is described in detail in the
[Implicit Flow Section of OpenID Connect 1.0 Core](https://openid.net/specs/openid-connect-core-1_0.html#ImplicitFlowAuth).

[^13]: The Resource Owner is the owner of the information being requested, this is typically the user, but also can be
the _Relying Party_[^17].

[^14]: The Response Mode describes how the Authorization Endpoint responds to Authorization Requests. There are three
primary modes; `query` which redirects the _Resource Owner_[^13] with the response in the URI query parameters,
`fragment` which redirects the _Resource Owner_[^13] with the response in the URI fragment parameters, and `form_post`
which performs a `POST` request with the response in the request body.

[^15]: The Refresh Token Flow is a flow which exchanges a _Refresh Token_[^16] at the Token Endpoint for new tokens with
the same security characteristics or narrowed security characteristics.

[^16]: The Refresh Token is typically a completely opaque token which can be used to reissue other tokens.

[^17]: The Relying Party is the party which relies on the OpenID Connect 1.0 Provider to process the Authorization Flow
and the _Resource Owner_[^13] to grant access to the information.

[OpenID Connect 1.0]: https://openid.net/specs/openid-connect-core-1_0.html
[Requesting Claims using Scope Values]: https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
[Claims]: https://openid.net/specs/openid-connect-core-1_0.html#Claims
[Claims Parameter]: https://openid.net/specs/openid-connect-core-1_0.html#ClaimsParameter
[JSON Web Token Profile for OAuth 2.0 Access Tokens]: https://datatracker.ietf.org/doc/html/rfc9068
