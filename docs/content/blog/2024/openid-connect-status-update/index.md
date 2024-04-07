---
title: "OpenID Connect 1.0 and SAML 2.0 Status Updates"
description: "The OpenID Connect 1.0 implementation is nearing practical completion so it can be considered non-beta, meaning soon we can start working on SAML 2.0."
summary: "The OpenID Connect 1.0 implementation is nearing practical completion so it can be considered non-beta, meaning soon we can start working on SAML 2.0."
date: 2024-04-07T00:55:09+10:00
draft: false
weight: 50
categories: ["News", "Community Updates"]
tags: ["community-updates", "openid-connect"]
contributors: ["James Elliott"]
pinned: false
homepage: false
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

It should be no surprise to most people reading this that OAuth 2.0 and OpenID Connect 1.0 are very vast specifications
which encompasses many different specifications that must be taken into account for various features. Similarly SAML 2.0
is very complicated. We have taken a deliberately cautious and strategic approach to implementation.

Specifically we've offered our OAuth 2.0 and OpenID Connect 1.0 implementation as a beta feature for some time now. This
has allowed us to relatively easily implement the relevant specifications without making too many mistakes and made it
easy when we made those mistakes to fix them.

This article discusses both the short, medium, and long term plans for OAuth 2.0 / OpenID Connect 1.0 and the long term
plans for SAML 2.0.


## OpenID Connect 1.0

There are likely to be a couple of upcoming breaking changes to OpenID Connect 1.0 as we make adjustments to rectify
some issues. It should be noted that while these changes are *technically* breaking changes our
[versioning policy](../../../policies/versioning.md#exceptions) has an explicit exception at the current time for
OpenID Connect 1.0.

There are also quite a few exciting features that we think will entice more users to utilize these features.

We are fairly likely to pick a point somewhere within the features outlined below where we consider OpenID Connect 1.0
as no longer being in a beta. We may from time to time introduce new specification elements which will be introduced as
a beta but it's likely they'll remain a beta for much less time than the original implementation.

### Claims Parameter

Currently we do not support the claims parameter at all. We intend to support this in the very near future. This
Authorization Request parameter allows a relying party / client to request specific claims to be returned either via the
ID Token or User Info endpoint, as well as optional claims. This means a client doesn't need to request an overly broad
scope which has far more information than they need, they can just request the specific claims relevant to them provided
they're authorized to request the scope the specific claim belongs to.

This will also introduce a fix which prevents User Info claims from being returned in the ID Token by default, they will
have to either be explicitly requested or the client registration will have to be configured to explicitly return them.
As per the [Requesting Claims via Scope Values](https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims)
section of the OpenID Connect 1.0 specification we will still return the claims granted via scopes on the User Info
endpoint.

The specific benefit here is the ID Token can be thrown around more freely to other parties to verify the Authorization
took place without giving them access to potentially private information the client did not need to share.

Some client registrations or client configurations may need to be updated to adapt to this specific element. For example
if the client configuration can be explicitly configured to obtain claims from the User Info endpoint but it's currently
configured to explicitly obtain the claims via the ID Token, this will need to be reversed.

Some preliminary work has already started on this feature and it's working as intended, we're just waiting to release
it.

### Complete Standard Claims

There are several standard claims we've yet to implement that should be relatively easily. Combining this with the
[Claims Parameter](#claims-parameter) is incredibly desirable as we can add a whole heap of additional information that
clients may wish to leverage but at the same time narrow these claims down to make requests / processing more lite.

Some preliminary work has already started on this feature and it's working as intended, we're just waiting to release
it.

### Custom Claims Mapping

Similar to [Complete Standard Claims](#complete-standard-claims) adding the ability for users to make custom claims and
scope mappings from their chosen backend is incredibly desirable for us. This is a fairly large feature however as
you have to take into account the fact transformation of the attribute may be necessary depending on the circumstances.

For example if you'd like to return a new bool claim but you want to return that based on a list of strings, that will
be very hard to do without some form of transformation. We're currently looking at
[Common Expression Language](https://github.com/google/cel-spec) though still undecided. It has an
[AST implementation in Go](https://github.com/google/cel-spec) which looks very promising.

Some preliminary work has already started on this feature.

### Prompt and Max Age Parameters

We partially support the `prompt` parameter currently, specifically if the `prompt` value `none` is requested. However
we specifically need to add an element where if the `prompt` parameter is set to `login` and the user did not
authenticate as part of the authorization, we should make them authenticate again. We currently return an error in this
situation but we will very soon implement this feature.

Similarly, the `max_age` parameter should make the user authenticates again if a length of time since the last
authentication exceeds the value provided (in seconds), but instead return an error.

To make it clear if it isn't, the returning of an error *prevents* login, and in both scenarios we return an error.

Some preliminary work has already started on this feature.

### Device Authorization Grant

The [OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628) is probably a flow you're
somewhat familiar with. A user attempts to login to an app via a TV or other embedded device; and a link, QR Code,
or both are displayed to the user with a one time code. The user visits the link on any device they want, either via
manual entry or via the QR Code, and enters the code into their chosen device and finishes the authorization and consent
process.

This is a feature we're likely to implement sooner rather than later.

Some preliminary work has already started on this feature.

### Issuer and Client Configuration in SQL

As part of the [Dashboard / Control Panel and CLI for Administrators](../../../roadmap/active/dashboard-control-panel-and-cli-for-admins.md)
we intend to allow dynamically adding both issuers and client registrations to the storage solution. In addition even
if the Administrator UI is disabled long term it makes more sense for most people to store the registrations in the
database so you don't have to restart Authelia to add another.

Don't worry this will be optional, however some features will not be available if you don't use this method. We will
offer both a declarative in-configuration option for users, as well as a way to easily either preload (i.e. one time)
or overwrite the current configuration using an environment variable such as `X_AUTHELIA_OPENID_CONNECT_ISSUER_PATHS`
and `X_AUTHELIA_OPENID_CONNECT_CLIENTS_PATHS`.

### Dynamic Client Registration

One of the key benefits of storing the [Issuer and Client Configuration in SQL](#issuer-and-client-configuration-in-sql)
is allowing the use of [OAuth 2.0 Dynamic Client Registration Protocol](https://datatracker.ietf.org/doc/html/rfc7591)
and [OpenID Connect Registration 1.0](https://openid.net/specs/openid-connect-registration-1_0.html) secured by an
OAuth 2.0 flow.

This is a longer term goal for us as this can be used in various scenarios to programmatically register clients.

### Token Exchange

[OAuth 2.0 Token Exchange](https://datatracker.ietf.org/doc/html/rfc8693) is on our radar as a nice to have feature
which we will likely implement. This specification allows a client to exchange a foreign token for a token issued by the
Authelia issuer which can be useful in certain scenarios and has been requested by a user.

## SAML 2.0

Unfortunately the only thing happening with SAML 2.0 currently is the investigation into the specification itself and
the available libraries we have to work with. This should put us in a reasonable position to implement it when the time
comes, which is likely to occur as a beta after OpenID Connect 1.0 leaves its beta; this is due to the sheer complexity
of the specifications.
