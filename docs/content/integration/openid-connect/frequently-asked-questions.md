---
title: "Frequently Asked Questions"
description: "Frequently Asked Questions regarding integrating the Authelia OpenID Connect 1.0 Provider with an OpenID Connect 1.0 Relying Party"
summary: "Frequently Asked Questions regarding integrating the Authelia OpenID Connect 1.0 Provider with an OpenID Connect 1.0 Relying Party."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 690
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Questions

The following section lists individual questions.

### How do I generate a client identifier or client secret?

We strongly recommend the following guidelines for generating a client identifier or client secret:

1. Each client should have a unique identifier and secret pair.
2. Each identifier and secret should be randomly generated.
3. Each identifier and secret should have a length above 40 characters.
4. The secret should be stored in the configuration in a supported hash format. *__Note:__ This does not
   mean you configure the relying party / client application with a hashed secret, the hashed secret should just be used
   for the `client_secret` value in the Authelia client configuration and the relying party / client application should
   have the plain text secret.*
5. Identifiers and Secrets should only have [RFC3986 Unreserved Characters] as some implementations do not appropriately
   encode the identifier or secret when using it to access the token endpoint. See
   [Why does Authelia return an error about the client identifier or client secret being incorrect when they are correct]
   FAQ on this specific issue for more information.

[Why does Authelia return an error about the client identifier or client secret being incorrect when they are correct]: #why-does-authelia-return-an-error-about-the-client-identifier-or-client-secret-being-incorrect-when-they-are-correct

Authelia provides an easy way to perform such actions.

#### Client ID / Identifier

Users can easily generate a client id / identifier by following the [Generating a Random Alphanumeric String] guide. For
example users can perform the below command to generate a client id / identifier with 72 characters
which is printed. This random command also avoids issues with
a relying party / client application encoding the characters correctly as it uses the [RFC3986 Unreserved Characters].

If a different charset is used if the value would be different when URL encoded then it will also print this value
separately.

{{< envTabs "Generate a Random Client ID" >}}
{{< envTab "Docker" >}}
```shell
docker run --rm authelia/authelia:latest authelia crypto rand --length 72 --charset rfc3986
```
{{< /envTab >}}
{{< envTab "Bare-Metal" >}}
```shell
authelia crypto rand --length 72 --charset rfc3986
```
{{< /envTab >}}
{{< /envTabs >}}

[Generating a Random Alphanumeric String]: ../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string

#### Client Secret

Users can easily generate a client secret by following the [Generating a Random Password Hash] guide. For example users
can perform the below command to both generate a client secret with 72 characters which is printed and is to be used
with the relying party and hash it using PBKDF2 which can be stored in the Authelia configuration. This random command
also avoids issues with a relying party / client application encoding the characters correctly as it uses the
[RFC3986 Unreserved Characters].

If a different charset is used if the value would be different when URL encoded then it will also print this value
separately.

{{< envTabs "Generate a Random Client Secret" >}}
{{< envTab "Docker" >}}
```shell
docker run --rm authelia/authelia:latest authelia crypto hash generate pbkdf2 --variant sha512 --random --random.length 72 --random.charset rfc3986
```
{{< /envTab >}}
{{< envTab "Bare-Metal" >}}
```shell
authelia crypto hash generate pbkdf2 --variant sha512 --random --random.length 72 --random.charset rfc3986
```
{{< /envTab >}}
{{< /envTabs >}}

[Generating a Random Password Hash]: ../../reference/guides/generating-secure-values.md#generating-a-random-password-hash

##### Tuning work factors

When hashing the client secrets, Authelia performs the hashing operation to authenticate the client when receiving requests.
This hashing operation takes time by design (the *work* part of the work factor) to hinder an attacker trying to obtain the client secret.
The amount of time taken depends on your hardware and the work factor.

If your client operations time out, you might need to reduce the work factor to a level appropriate for your client and
your hardware's capabilities.

To test the duration of different work factors, you can measure it like this:
`time authelia crypto hash generate pbkdf2 --variant sha512 --iterations 310000 --password insecure_password`.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
You should not use your actual passwords for this test, the time taken should be the same for any reasonable password
length.
{{< /callout >}}

You can read more about password hashing tuning in the
[Passwords reference guide](../../reference/guides/passwords.md#tuning).

##### Plaintext

Authelia *technically* supports storing the plaintext secret in the configuration. This will likely be completely
unavailable in the future as it was a mistake to implement it like this in the first place. While some other OpenID
Connect 1.0 providers operate in this way, it's more often than not that they operating in this way in error. The
current *technical support* for this is only to prevent massive upheaval to users and give them time to migrate.

As per [RFC6819 Section 5.1.4.1.3](https://datatracker.ietf.org/doc/html/rfc6819#section-5.1.4.1.3) the secret should
only be stored by the authorization server as hashes / digests unless there is a very specific specification or protocol
that is implemented by the authorization server which requires access to the secret in the clear to operate properly in
which case the secret should be encrypted and not be stored in plaintext. The most likely long term outcome is that the
client configurations will be stored in the database with the secret both salted and peppered.

Authelia currently does not implement any of the specifications or protocols which require secrets being accessible in
the clear such as most notably the `client_secret_jwt` grant, we will however likely soon implement `client_secret_jwt`.
We are however *__strongly discouraging__* and formally deprecating the use of plaintext client secrets for purposes
outside those required by specifications. We instead recommended that users remove this from their configuration
entirely and use the [How Do I Generate a Client Identifier or Client Secret](#how-do-i-generate-a-client-identifier-or-client-secret) FAQ.

Plaintext is either denoted by the `$plaintext$` prefix where everything after the prefix is the secret. In addition if
the secret does not start with the `$` character it's considered as a plaintext secret for the time being but is
deprecated as is the `$plaintext$` prefix.

### Why does Authelia return an error about the client identifier or client secret being incorrect when they are correct?

When using `client_secret_basic` several implementations of OAuth 2.0 and OpenID Connect 1.0 do not properly URL encode
these values as is absolutely required by the specification before encoding the header value. Both the client id and
client secret must be encoded using the `application/x-www-form-urlencoded` encoding algorithm (i.e. URL encoded) before
being used as the username and password values for the Basic authorization scheme as detailed in
[RFC6749 Section 2.3.1](https://datatracker.ietf.org/doc/html/rfc6749#section-2.3.1).

Authelia enforces this practice. In situations where the client does not conform to the specification we suggest
ensuring the client id and client secret only use the unreserved characters or you URL encode these values yourself.

For these reasons since v4.38.0 Authelia has included a URL encoded value when generating random strings or password
hashes if the URL encoded value differs from the non-encoded value.

### Why isn't my application able to retrieve the token even though I've consented?

The most common cause for this issue is when the affected application can not make requests to the Token [Endpoint].
This becomes obvious when the log level is set to `debug` or `trace` and a presence of requests to the Authorization
[Endpoint] without errors (i.e. returns a success) but an absence of requests made to the Token [Endpoint].

These requests can be identified by looking at the `path` field in the logs, or by messages prefixed with
`Authorization Request` indicating a request to the Authorization [Endpoint] and `Access Request` indicating a request
to the Token [Endpoint]. Therefore if the logs indicate there was an Authorization Request and Access Request this
specific scenario is not applicable to you.

All causes should be clearly logged by the client application, and all errors that do not match this scenario are
clearly logged by Authelia. It's not possible for us to log requests that never occur however.

One potential solution to this is detailed in the [Solution: Configure DNS Appropriately](#configure-dns-appropriately)
section. This section also details how to identity if you're affected.

### Why doesn't the discovery endpoint return the correct issuer and endpoint URL's?

The most common cause for this is if the `X-Forwarded-Proto` and `X-Forwarded-Host` / `Host` headers do not match the
fully qualified URL of the provider. This can be because of requesting from the Authelia port directly i.e. without going
through your proxy or due to a poorly configured proxy.

If you've configured Authelia alongside a proxy and are making a request directly to Authelia you need to perform the
request via the proxy. If you're avoiding the proxy due to a DNS limitation see
[Solution: Configure DNS Appropriately](#configure-dns-appropriately) section.

### Why doesn't the access control configuration work with OpenID Connect 1.0?

The [access control](../../configuration/security/access-control.md) configuration contains several elements which are
not very compatible with OpenID Connect 1.0. They were designed with per-request authorizations in mind. In particular
the [resources](../../configuration/security/access-control.md#resources),
[query](../../configuration/security/access-control.md#query),
[methods](../../configuration/security/access-control.md#methods), and
[networks](../../configuration/security/access-control.md#networks) criteria are very specific to each request and to
some degree so are the [domain](../../configuration/security/access-control.md#domain) and
[domain regex](../../configuration/security/access-control.md#domain_regex) criteria as the token is issued to the client
not a specific domain.

For these reasons we implemented the
[authorization policy](../../configuration/identity-providers/openid-connect/clients.md#authorization_policy) as a direct
option in the client. It's likely in the future that we'll expand this option to encompass the features that work well
with OpenID Connect 1.0 such as the [subject](../../configuration/security/access-control.md#subject) criteria which
reasonably be matched to an individual authorization policy. Because the other criteria are mostly geared towards
per-request authorization these criteria types are fairly unlikely to become part of OpenID Connect 1.0 as there are no
ways to apply these criteria except during the initial authorization request.

See [ADR1](../../reference/architecture-decision-log/1.md) for more information.

### Why isn't the Access Token a JSON Web Token?

The Access Token and it's format is entirely up to Authorization Servers / OpenID Connect 1.0 Providers. The
conventional way the Access Token is presented is as an opaque value which has no meaning. There are quite a few reasons
this is the case, however standards and implementations exist which return the Access Token as a JSON Web Token. This is
not specifically wrong when they do this, just as the standards allow us to decide the value should be opaque it also
allows them to decide not to do that.

The double-edged sword of a JSON Web Token is it's easy to perform a stateless check to see if a JSON Web Token is
"valid" but also very hard to revoke it. A Resource Server may just check the JWT claims and signature to see if it
looks like it *should* still be valid, without performing a stateful check with the Authorization Server frequently
enough. This is why we *__strongly recommend__* that Access Tokens and ID Tokens have a short lifespan.

There also exists standardized mechanisms for Resource Servers and those possessing an opaque Access Token to check the
validity and metadata associated with them. These mechanisms are the Introspection Request and UserInfo Request.

There are some other specific scenarios which would lead to the Access Token being revoked earlier than its original
lifetime which make this desirable. When you combine these with the fact there are standardized mechanisms to have a
similar outcome, it's obvious to us this is the sane default.

For example during a Refresh Flow to the Token Endpoint the previously issued Access Token and Refresh Token should be
transparently revoked. This means as soon as a Refresh Flow is performed either by the authorized party or a malicious
one who's nefariously obtained the Refresh Token, the old Access Token and Refresh Token are effectively useless unless
the Resource Server is caching these results.

In addition as tokens can be manually revoked using the Revocation Endpoint in a scenario where a long lived token was
revoked due to known compromise; the revocation will take place much faster.

Users who still desire or have an application that requires the Access Token is a JWT should configure the
[access_token_signed_response_alg](../../configuration/identity-providers/openid-connect/clients.md#access_token_signed_response_alg)
client configuration option.

### How should I link user accounts to Authelia OpenID Connect 1.0 responses in the application I'm designing?

There are several in-use methodologies for linking user accounts to [OpenID Connect 1.0] Providers. The specification has
a fairly strong opinion about how this is done for various reasons and the supported method by Authelia is the same as
what the specification supports.

Specifically we support using the pairwise comparison of the `iss` and `sub` claim as an anchor to local user accounts.
This combination is a combination that must be unique for any given user identity. Claims such as `email` and
`preferred_username` have no formal guarantees of stability or uniqueness by the [OpenID Connect 1.0] specification or
Authelia.

Several in-use applications including ones that Authelia users frequently use utilize claims such as `email` and
`preferred_username`. However these implementations are in contradiction with the specification. These attributes
realistically should only be used as hints when a user who has not linked their account tries to login or register
their account with [OpenID Connect 1.0]. For example in either scenario an application may prefill the username or email
field of a login or registration form using these claims.

Utilization of these claims could potentially become problematic if we ever implement a feature to change usernames or
email addresses. Therefore we only guarantee the stability of those specific claims and at such a time as we allow
changing of usernames or email addresses the link between those values will remain stable, any other claim is to be
considered fragile.

If interested in the specification you can read the
[Claim Stability and Uniqueness](https://openid.net/specs/openid-connect-core-1_0.html#ClaimStability) section of the
specification.

### Why does Authelia ask for Consent when I've asked for my consent to be remembered or used the implicit consent policy?

There are varying conditions which are controlled by the client when making an Authorization Request where explicit
consent may be required. These areas are mostly governed by the specification, or in situations where the issued token
or tokens have certain properties directly related to Authelia itself such as being used for a Forward Authz Bearer
token.

This effectively means that both the consent mode and remembered consent i.e. pre-configured consent are essentially
preferences should the conditions not require explicit consent and should not be seen as concrete.

In addition it's important to note that the consent mode is merely a suggested default behaviour, especially the
`implicit` consent mode which is not specifically supported by the specification and in many cases will either revert to
`explicit`, silently not perform certain expected actions, or should be expected to return an error the client. This
mode is intended for development and testing, and is heavily discouraged in production.

In addition the consent policy implementation is likely a target for a breaking refactoring before we stabilize the
OpenID Connect 1.0 Provider implementation due to additional understanding of how the specifications require us to
handle Consent Flows.

The following specific and intentional limitations exist:

1. The Authorization Code Flow will require explicit consent if a Refresh Token will be granted to the relying party
   (i.e. if the client is authorized to request and requests the `offline_access` or `offline` scopes).
   1. The specific requirement is that a Refresh Token must not be issued if consent is not requested however the
      implication is that explicit consent is required for Refresh Tokens. In the future we will adjust this so
      the Authorization Code Flow will not mint and grant a Refresh Token unless the relying party explicitly requests
      consent, with an option to automatically require explicit consent instead of silently ignoring the request for
      a Refresh Token.
2. If the client requests the user is prompted to provide consent, the mode will automatically be `explicit` regardless
   of client registration configuration.
3. If the client requests the user is prompted to login again, then the mode will either automatically be
   `explicit` or the flow may also result in a failure that returns an error to the client, regardless of the client
   registration configuration.
4. If the authorization request results in security sensitive tokens such as:
   1. Tokens that can be used as Bearer Tokens which impersonate users for the Authelia Authz endpoints.
   2. Tokens that can be used as Bearer Tokens for the Authelia API, especially for endpoints that allow alteration of
      any security related characteristic of users or Authelia settings.
   3. Any future characteristic deemed appropriate in the future (which will be explicitly documented in this FAQ) such
      as but not limited to:
      1. Minting of Tokens with long lifespans.
      2. Minting of Tokens which have special use cases such as Dynamic Client Registration, Session Management, etc.
      3. In any situation where it's required or recommended by the existing or future specifications including formal
         specifications, security profiles like the Financial-grade API, and best current practices, etc.
      4. Minting of Tokens in any context which may be seen as especially security sensitive by the maintainers either
         through community consultation, identified during security audits or reviews, or identified during normal
         development cycles.
5. If the current flow is otherwise not compatible with implicit consent for any reason; for example:
   1. The Device Authorization Flow which requires interactive consent due to the code functionality.

## Solutions

The following section details solutions for multiple of the questions above.

### Configure DNS Appropriately

In order to make requests to Authelia an application must be able to resolve it. It's important in all instances to
check if the application with the issue can resolve the correct IP address for Authelia between each step of the
process, and this check also can be used to clearly identity if this is the most likely underlying cause for an issue
you're facing.

##### Bare-Metal

1. If you're running an internal DNS server ensure an A record exists for the FQDN of Authelia with the value being the
   IP of the server responsible for handling requests for Authelia.
2. If you're not running an internal DNS server then do check the following:
   1. Ensure the external DNS server(s) have the same A record as described above.
   2. Ensure that that your NAT-hairpin is configured correctly.
   3. If all else fails add a hosts file entry to work around this issue.

##### Docker

1. Ensure both the application with the issue shares a network in common with the proxy container.
2. Ensure an alias for the FQDN of Authelia is present for the proxy container:
   - If using `docker compose` see the
     [network aliases](https://docs.docker.com/compose/compose-file/compose-file-v3/#aliases) documentation
     reference for more information.
   - If using `docker run` see the `--network-alias` option of the [docker run](https://docs.docker.com/engine/reference/commandline/run/)
     reference for more information.

Examples (assuming your Authelia Root URL is `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`):

```yaml {title="compose.yml"}
services:
  application:
    ## Mandatory that the application is on the same network as the proxy.
    networks:
      proxy: {}
  proxy:
    networks:
      ## Mandatory that the proxy is on the same network as the application, and that it has this alias.
      proxy:
        aliases:
          - '{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
  authelia:
    networks:
      proxy: {}
networks:
  proxy:
    ## An external network can be created manually and shared between multiple compose files. This is NOT mandatory.
    external: true
    name: 'proxy-net'
```

```console
docker run -d --name proxy --network proxy --network-alias {{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}} <other proxy arguments>
docker run -d --name application --network proxy <other application arguments>
```

[Endpoint]: ./introduction.md#discoverable-endpoints
[RFC3986 Unreserved Characters]: https://datatracker.ietf.org/doc/html/rfc3986#section-2.3
[OpenID Connect 1.0]: https://openid.net/specs/openid-connect-core-1_0.html
