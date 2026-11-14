---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Webhooks"
description: "Configuring the Webhooks Settings."
summary: "Authelia can deliver events describing occurrences such as password changes and authentication attempts to external receivers over HTTPS. This section describes how to configure this."
date: 2026-09-13T01:31:45+10:00
draft: false
images: []
weight: 199500
toc: true
aliases:
  - '/c/webhooks'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia can deliver events to one or more external receivers as HTTPS requests. Each event is a [CloudEvents] 1.0
structured mode JSON document describing one occurrence, such as a password change or a failed authentication.

This page covers the configuration options. How events are delivered, the handshake, retries, delivery guarantees,
metrics, the payloads, and signature verification are documented in the [Webhook
Events](../../reference/guides/webhook-events.md) reference guide, which anyone operating a destination or writing a
receiver should read.

Webhooks are disabled when no destinations are configured. Delivery is best effort: an unhealthy or unreachable
receiver never blocks, delays, or fails an authentication or self-service operation.

## Variables

Some of the values within this page can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
webhooks:
  startup_check: false
  destinations:
    - name: 'admin-api'
      address: 'https://admin.{{< sitevar name="domain" nojs="example.com" >}}/hooks/authelia'
      events:
        - 'com.authelia.user.*'
        - 'com.authelia.security.ban.applied'
      timeout: '10 seconds'
      buffer_size: 256
      disable_redaction: false
      signature:
        secret: 'insecure_secret'
        algorithm: 'sha256'
      authentication:
        bearer:
          token: 'insecure_token'
          method: 'header'
          scheme: 'Bearer'
        basic:
          username: 'authelia'
          password: 'insecure_password'
      headers:
        X-Tenant: 'acme'
      validation:
        disable: false
        origin: 'auth.{{< sitevar name="domain" nojs="example.com" >}}'
        force_origin: false
        rate: 0
      batch:
        size: 50
        max_wait: '5 seconds'
        immediate:
          - 'com.authelia.security.*'
      retry:
        attempts: 3
        initial_interval: '1 second'
        maximum_interval: '1 minute'
      tls:
        server_name: 'admin.{{< sitevar name="domain" nojs="example.com" >}}'
        skip_verify: false
        minimum_version: 'TLS1.2'
        maximum_version: 'TLS1.3'
        certificate_chain: |
          -----BEGIN CERTIFICATE-----
          ...
          -----END CERTIFICATE-----
          -----BEGIN CERTIFICATE-----
          ...
          -----END CERTIFICATE-----
        private_key: |
          -----BEGIN PRIVATE KEY-----
          ...
          -----END PRIVATE KEY-----
```

## Options

This section describes the individual configuration options.

### startup_check

{{< confkey type="boolean" default="false" required="no" >}}

Sends one [com.authelia.system.startup_check](../../reference/guides/webhook-events.md#the-startup-check-probe) event to
each destination during startup, so that a misconfigured receiver is discovered before the first real event. A failed
check is logged as a warning and never prevents startup.

### destinations

A list of receivers. Each destination has its own queue, worker, HTTP client, and retry policy, so a slow or broken
receiver does not affect the others.

#### name

{{< confkey type="string" required="yes" >}}

The unique name of this destination. It labels log entries, validation errors, and the
[metrics](../../reference/guides/webhook-events.md#metrics) counter, and is never included in a payload.

#### address

{{< confkey type="string" required="yes" >}}

The URL events are delivered to. This is a plain URL, not the common address syntax.

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
The scheme must be `https` and there is no option to disable this, because payloads describe authentication activity
and can contain credential equivalent values when [disable_redaction](#disable_redaction) is enabled.

For a receiver with a private or self-signed certificate authority use the [tls](#tls) options or the
[certificates_directory](introduction.md#certificates_directory) global option.
{{< /callout >}}

The address must not contain user information (a `username:password@` prefix), a fragment, or an `access_token` query
parameter. Use the [authentication](#authentication) options to present credentials.

Redirects are never followed and the proxy environment variables are honored. See
[Request](../../reference/guides/webhook-events.md#request).

#### events

{{< confkey type="list(string)" required="yes" >}}

The event types this destination receives. Each entry is either the exact name of an event type or a prefix ending in
an asterisk. A single `'*'` selects every type. The types are listed in the
[event catalogue](../../reference/guides/webhook-events.md#event-catalogue).

```yaml {title="configuration.yml"}
webhooks:
  destinations:
    - name: 'admin-api'
      events:
        - 'com.authelia.user.credential.*'
        - 'com.authelia.security.ban.applied'
```

A selector which matches no event type is a startup error, as it is almost always a typo.

Select only the types the receiver acts on, as some can be
[very high volume](../../reference/guides/webhook-events.md#event-volume).

#### timeout

{{< confkey type="string,integer" syntax="duration" default="10 seconds" required="no" >}}

The timeout for a single delivery attempt, covering connection establishment, the TLS handshake, and reading the
response. A value of zero or less applies the default.

#### buffer_size

{{< confkey type="integer" default="256" required="no" >}}

The number of events which may be queued for this destination. A value of `0` applies the default and a negative value
is a startup error. Events are dropped when the queue is full, as described in
[Delivery guarantees](../../reference/guides/webhook-events.md#delivery-guarantees).

#### disable_redaction

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
Enabling this option sends credential equivalent values to the receiver: the One-Time Code issued for a session
elevation, the identity verification link URL, and the revocation link URL.

Anyone holding those values can complete a password reset or a session elevation for the user they were issued to,
without that user's password or second factor. The receiver, and every log, queue, and backup it writes to, becomes as
sensitive as the session secret.
{{< /callout >}}

When disabled these fields are omitted from the payload. See
[Redaction](../../reference/guides/webhook-events.md#redaction) for the field list.

#### signature

The HMAC signature applied to each request body. It is the only thing which proves a payload came from this Authelia
instance. The [authentication](#authentication) options serve a different purpose, identifying Authelia to the
receiver, and both can be configured.

##### secret

{{< confkey type="string" required="no" >}}

The shared secret used to sign each request. When this is empty requests are not signed and the `X-Authelia-Signature`
header is omitted, which is strongly discouraged for any receiver which acts on the events it receives.

It's **strongly recommended** this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string)
with 64 or more characters.

The receiver verifies it as described in
[Verifying the signature](../../reference/guides/webhook-events.md#verifying-the-signature).

##### algorithm

{{< confkey type="string" default="sha256" required="no" >}}

The HMAC hash algorithm, either `sha256` or `sha512`. It is only validated when a [secret](#secret) is configured.

The algorithm is not advertised in the request, so the receiver must be configured with the same value.

#### authentication

The credentials Authelia presents to the receiver. The [CloudEvents HTTP webhook specification] requires every delivery
to carry a token, so exactly one of these options must be configured. Configuring neither or both is a startup error.

- [bearer](#bearer) presents a token. This is the only option which satisfies the specification's token requirement.
- [basic](#basic) presents HTTP Basic credentials, for a receiver which requires them. Basic credentials are not a
  token and do not satisfy that requirement.

These credentials identify Authelia to the receiver but do not prove a payload is genuine. Use [signature](#signature)
for that.

[CloudEvents HTTP webhook specification]: https://github.com/cloudevents/spec/blob/main/cloudevents/http-webhook.md#3-authorization

##### bearer

Presents a token, either in the `Authorization` header or in the `access_token` query parameter.

```yaml {title="configuration.yml"}
webhooks:
  destinations:
    - name: 'admin-api'
      authentication:
        bearer:
          token: 'insecure_token'
          method: 'header'
          scheme: 'Bearer'
```

|   Name   |   Type   | Default  | Required | Sensitive | Description                                                                              |
| :------: | :------: | :------: | :------: | :-------: | :--------------------------------------------------------------------------------------- |
| `token`  | `string` |          |   yes    |    yes    | The token value.                                                                         |
| `method` | `string` | `header` |    no    |    no     | How the token is presented, either `header` or `query`.                                  |
| `scheme` | `string` | `Bearer` |    no    |    no     | The authorization scheme prefixed to the token, separated by a space. Only for `header`. |

With the `header` method the token is sent in the `Authorization` header, prefixed by `scheme`. The combined value must
be a valid HTTP header value.

With the `query` method the token is added to the request address as the `access_token` query parameter, and the
`scheme` option must not be configured. An address is more likely to be logged than a header, so only use this method
when the receiver cannot read the `Authorization` header.

##### basic

Presents HTTP Basic credentials in the `Authorization` header. They are sent with every request and never in answer to
a challenge.

```yaml {title="configuration.yml"}
webhooks:
  destinations:
    - name: 'admin-api'
      authentication:
        basic:
          username: 'authelia'
          password: 'insecure_password'
```

|    Name    |   Type   | Required | Sensitive | Description   |
| :--------: | :------: | :------: | :-------: | :------------ |
| `username` | `string` |   yes    |    no     | The username. |
| `password` | `string` |   yes    |    yes    | The password. |

#### headers

{{< confkey type="dictionary(string)" required="no" >}}

Additional static headers sent with every request to this destination, for example routing or tenancy hints.

```yaml {title="configuration.yml"}
webhooks:
  destinations:
    - name: 'admin-api'
      headers:
        X-Tenant: 'acme'
```

Configuring a reserved header, or an invalid header name or value, is a startup error. The reserved headers are:

- `Authorization`
- `Cache-Control`
- `Connection`
- `Content-Length`
- `Content-Type`
- `Host`
- `Transfer-Encoding`
- `User-Agent`
- Any header beginning with `WebHook-`, in any letter case
- Any header beginning with `X-Authelia-`
- Any header beginning with `ce-`, in any letter case

#### validation

Before any event is delivered Authelia performs the CloudEvents abuse protection handshake, and the receiver confirms
either in its answer or later by requesting a callback URL. The handshake, its outcomes, and the callback are documented
in [Abuse protection](../../reference/guides/webhook-events.md#abuse-protection).

{{< callout context="danger" title="Required" icon="outline/alert-octagon" >}}
A destination which has not confirmed receives no events at all. A receiver must implement the `OPTIONS` response or the
callback, or the destination must set [disable](#disable).
{{< /callout >}}

##### disable

{{< confkey type="boolean" default="false" required="no" >}}

Skips the handshake and delivers events without asking the destination to confirm. Use this only for a receiver you
control which implements neither `OPTIONS` nor the callback. The other `validation` options are neither checked nor used
when this is enabled, and a destination [suspended](../../reference/guides/webhook-events.md#suspension) by a `410`
stays suspended until Authelia is restarted.

##### origin

{{< confkey type="string" required="no" >}}

The origin presented in `WebHook-Request-Origin`. It must be a DNS name such as `auth.example.com`, without a scheme,
port, or path. Any other value is a startup error.

By default this is used only when an origin cannot be derived from the first configured
[session cookie domain](../session/introduction.md#cookies), which is the host of its
[authelia_url](../session/introduction.md#authelia_url) or otherwise its
[domain](../session/introduction.md#domain). Set [force_origin](#force_origin) to present this value regardless.

If no origin can be derived and none is configured, the destination is refused.

##### force_origin

{{< confkey type="boolean" default="false" required="no" >}}

Presents [origin](#origin) even when one could be derived. Use this where receivers know Authelia by a different name,
for example behind a gateway. Enabling it without an `origin` is a startup error.

##### rate

{{< confkey type="integer" required="no" >}}

The delivery rate requested in `WebHook-Request-Rate`, in requests per minute. When this is absent or `0` the header is
not sent. A negative value is a startup error. Authelia paces its requests to the rate the receiver allows, as described
in [Abuse protection](../../reference/guides/webhook-events.md#abuse-protection).

#### batch

Batching delivers several events in one request. It is disabled unless [size](#size) is configured. The request format,
failure handling, and ordering are documented in [Batched
deliveries](../../reference/guides/webhook-events.md#batched-deliveries).

##### size

{{< confkey type="integer" required="no" >}}

The maximum number of events delivered in one request. Batching is disabled when this is absent or `0`. A negative
value, or a value greater than [buffer_size](#buffer_size), is a startup error.

##### max_wait

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The longest a batch waits to fill before it is delivered, measured from its first event. This is the delay batching
adds to an event, unless its type is listed in [immediate](#immediate).

##### immediate

{{< confkey type="list(string)" required="no" >}}

The event types which bypass batching and are delivered on their own as soon as they occur, using the same syntax as
[events](#events). When batching is enabled a selector matching no event type is a startup error. The option is ignored
when batching is disabled.

```yaml {title="configuration.yml"}
webhooks:
  destinations:
    - name: 'admin-api'
      address: 'https://admin.example.com/hooks'
      events:
        - '*'
      batch:
        size: 50
        max_wait: '5 seconds'
        immediate:
          - 'com.authelia.security.*'
```

#### retry

The retry policy for a failed delivery attempt. Which responses are retried, the backoff, and the handling of
`Retry-After` are documented in [Retries and failures](../../reference/guides/webhook-events.md#retries-and-failures).

##### attempts

{{< confkey type="integer" default="3" required="no" >}}

The maximum number of delivery attempts, including the first. A value of `1` disables retrying, a value of `0` applies
the default, and a negative value is a startup error.

##### initial_interval

{{< confkey type="string,integer" syntax="duration" default="1 second" required="no" >}}

The wait before the second attempt, which doubles before each subsequent attempt. See
[Backoff](../../reference/guides/webhook-events.md#backoff).

##### maximum_interval

{{< confkey type="string,integer" syntax="duration" default="1 minute" required="no" >}}

The ceiling applied to the doubling wait. It does not limit a
[Retry-After](../../reference/guides/webhook-events.md#retry-after).

#### tls

{{< confkey type="structure" structure="tls" required="no" >}}

Controls the TLS connection verification parameters for this destination. The minimum version defaults to `TLS1.2`.

By default Authelia uses the system certificate trust for TLS certificate verification and the
[certificates_directory](introduction.md#certificates_directory) global option can be used to augment this.

## Secrets

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `signature.secret`, `authentication.bearer.token`, and `authentication.basic.password` options cannot be supplied
with the [secrets](../methods/secrets.md) `_FILE` environment variables. As with the access control rules, the OpenID
Connect 1.0 clients, and the session cookies, `destinations` is a list of objects, which cannot be configured by
environment variable. See the note under [Environment variables](../methods/secrets.md#environment-variables) and
[ADR2](../../reference/architecture-decision-log/2.md).

A `_FILE` variable for one of these options is ignored, so a destination whose `signature.secret` was moved to one is
not signed at all.
{{< /callout >}}

To keep these values out of the configuration file, use the `fileContent` function of the
[file filters](../methods/files.md#file-filters), which must be enabled:

```yaml {title="configuration.yml"}
webhooks:
  destinations:
    - name: 'admin-api'
      signature:
        secret: '{{ fileContent "/run/secrets/webhook_signature_secret" | trim }}'
```

The `trim` is required because `fileContent` does not strip a trailing newline, and a secret with a trailing newline
produces a signature the receiver cannot reproduce.

[CloudEvents]: https://cloudevents.io/
