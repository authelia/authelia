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

This page covers configuration. The payloads, request headers, signature verification, and the behavior of each event
type are documented in the [Webhook Events](../../reference/guides/webhook-events.md) reference guide, which anyone
writing a receiver should read.

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
each destination during startup, so that a misconfigured receiver is discovered before the first real event.

The probe is sent regardless of the destination's [events](#events) selectors. It is not sent to a destination which
refused the [validation](#validation) handshake or has not yet confirmed it. A failed check is logged as a warning and
never prevents startup.

### destinations

A list of receivers. Each destination has its own queue, worker, HTTP client, and retry policy, so a slow or broken
receiver does not affect the others.

#### name

{{< confkey type="string" required="yes" >}}

The unique name of this destination. It labels log entries, validation errors, and the [metrics](#metrics) counter, and
is never included in a payload.

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

Redirects are never followed, as a redirect would resend the destination's credentials to another host. A `3xx`
response is a permanent delivery failure and the event is dropped.

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

{{< callout context="caution" title="Event Volume" icon="outline/alert-triangle" >}}
Every authorization request which authenticates with basic credentials, and whose credential check is not served from
the authentication backend's password cache, emits a `com.authelia.security.authentication.succeeded` event. Where
proxies present basic credentials on each request this can be one event per HTTP request to every protected
application.

Select only the types the receiver acts on. For a security feed
`com.authelia.security.authentication.failed` and `com.authelia.security.ban.applied` are usually sufficient.
{{< /callout >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `com.authelia.security.ban.expired` type is accepted by this option but is
[never emitted](../../reference/guides/webhook-events.md#comautheliasecuritybanexpired-is-never-emitted).
{{< /callout >}}

#### timeout

{{< confkey type="string,integer" syntax="duration" default="10 seconds" required="no" >}}

The timeout for a single delivery attempt, covering connection establishment, the TLS handshake, and reading the
response. A value of zero or less applies the default.

#### buffer_size

{{< confkey type="integer" default="256" required="no" >}}

The number of events which may be queued for this destination. A value of `0` applies the default and a negative value
is a startup error.

Enqueuing never blocks. When the queue is full the event is dropped, an error is logged, and the [metrics](#metrics)
counter records a `dropped` outcome.

A destination delivers serially and waits during retry backoff, so a failing receiver is not consuming its queue. Size
the buffer with that in mind, or reduce [attempts](#attempts) and [maximum_interval](#maximum_interval) for a high
volume destination.

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

These credentials identify Authelia to the receiver. They do not prove a payload is genuine, because anything which can
read a request can replay them. Use [signature](#signature) for that.

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

With the `query` method the token is added to the request address as the `access_token` query parameter, after any
query parameters the [address](#address) already has, and `Cache-Control: no-store` is sent with every request. The
`scheme` option must not be configured. An address is more likely to be logged than a header, by a proxy or by the
receiver's access log, so only use this method when the receiver cannot read the `Authorization` header.

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

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The [CloudEvents HTTP webhook specification] requires a token, and HTTP Basic credentials are not one. A destination
which uses this option does not satisfy that requirement. Use the [bearer](#bearer) option where conformance with the
specification matters.
{{< /callout >}}

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
- `Content-Type`
- `User-Agent`
- Any header beginning with `X-Authelia-`
- Any header beginning with `ce-`, in any letter case

#### validation

Before any event is delivered Authelia performs the [CloudEvents abuse protection] handshake: an `OPTIONS` request to
the destination's address carrying `WebHook-Request-Origin` and, when a [rate](#rate) is configured,
`WebHook-Request-Rate`. The receiver confirms by answering with a `WebHook-Allowed-Origin` header which is either the
same origin or an asterisk.

A receiver may instead answer with a 2xx and no `WebHook-Allowed-Origin`, and confirm later by requesting the URL
Authelia sends in `WebHook-Request-Callback`. See
[Confirming by callback](../../reference/guides/webhook-events.md#confirming-by-callback).

[CloudEvents abuse protection]: https://github.com/cloudevents/spec/blob/main/cloudevents/http-webhook.md#4-abuse-protection

{{< callout context="danger" title="Required" icon="outline/alert-octagon" >}}
A destination which has not confirmed receives no events at all, and events which occur in the meantime are not
recorded in the [metrics](#metrics). A receiver must implement the `OPTIONS` response or the callback, or the
destination must set [disable](#disable).
{{< /callout >}}

The destination's [authentication](#authentication) and [headers](#headers) are sent with the handshake.

The outcome of the handshake is one of:

- **Confirmed.** The destination receives events.
- **Refused.** The response permits a different origin, or its `WebHook-Allowed-Rate` is invalid or missing when a rate
  was requested. The refusal is logged, is not retried, and the destination receives no events until Authelia is
  restarted.
- **Held.** Any other outcome, including a response with no `WebHook-Allowed-Origin` and a handshake which exhausts its
  retries. The destination receives no events until it confirms by callback. A destination which confirmed by callback
  on an earlier start is not held, as the confirmation is kept in storage.

A `429`, `503`, or other `5xx` response, or a failure which produced no response, is retried under the destination's
[retry](#retry) policy before the outcome is decided.

Startup continues whatever the outcome. The handshake for each destination is limited to thirty seconds in total and
destinations are validated concurrently. It is performed once per start, and repeated hourly only for a destination
which was suspended by answering a delivery with `410`, as described in
[Retries and failures](../../reference/guides/webhook-events.md#retries-and-failures).

##### disable

{{< confkey type="boolean" default="false" required="no" >}}

Skips the handshake and delivers events without asking the destination to confirm. Use this only for a receiver you
control which implements neither `OPTIONS` nor the callback. The other `validation` options are neither checked nor
used when this is enabled.

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
not sent. A negative value is a startup error.

When a rate is requested the receiver must answer with `WebHook-Allowed-Rate`, either a positive integer or an
asterisk for no limit. A response which omits it is a refusal. Authelia paces its requests to the rate the receiver
allows, which may differ from the one requested, and a receiver may send the header even when no rate was requested.

The limit applies to requests, not events. Retries count against it, and a [batch](#batch) counts as one request.

#### batch

Batching delivers several events in one request. It is disabled unless [size](#size) is configured.

A batch is sent in the CloudEvents batched content mode: the media type is `application/cloudevents-batch+json` and the
body is a JSON array of envelopes, even when it holds one event. The per event `ce-` headers are not sent and
`X-Authelia-Event-Count` carries the number of events.

A batching destination still receives single event requests for the types listed in [immediate](#immediate) and for
the [startup_check](#startup_check) probe, so its receiver must handle both formats.

A batch succeeds or fails as a unit. A permanent failure drops every event in it and a retry resends all of them. A
receiver should order events by the envelope `time` attribute, not by their position in the array.

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

{{< callout context="caution" title="Ordering" icon="outline/alert-triangle" >}}
An immediate event can arrive before batched events which occurred earlier. A receiver which needs events in order
must sort on the envelope `time` attribute.
{{< /callout >}}

#### retry

The retry policy for a failed delivery attempt. Which responses are retried, and the handling of `Retry-After`, are
documented in [Retries and failures](../../reference/guides/webhook-events.md#retries-and-failures).

##### attempts

{{< confkey type="integer" default="3" required="no" >}}

The maximum number of delivery attempts, including the first. A value of `1` disables retrying, a value of `0` applies
the default, and a negative value is a startup error.

##### initial_interval

{{< confkey type="string,integer" syntax="duration" default="1 second" required="no" >}}

The wait before the second attempt. The wait doubles before each subsequent attempt, and each wait is shortened by a
random amount of up to twenty percent so that destinations recovering from the same outage do not retry together.

##### maximum_interval

{{< confkey type="string,integer" syntax="duration" default="1 minute" required="no" >}}

The ceiling applied to the doubling wait.

It does not limit a `Retry-After` header returned with a `429` or `503` response, which is honored in full. That wait
applies to the destination as a whole, so later events wait as well.

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

## Security

| Concern                    | Behavior                                                                                       |
| :------------------------- | :--------------------------------------------------------------------------------------------- |
| Transport                  | HTTPS only. TLS 1.2 minimum by default.                                                        |
| Redirects                  | Never followed, so credentials are not resent to another host. A `3xx` is a permanent failure. |
| Payload authenticity       | Only the HMAC [signature](#signature) proves a payload came from Authelia.                     |
| Receiver authenticity      | The [tls](#tls) verification parameters. `skip_verify` disables this.                          |
| Credential equivalent data | Omitted unless [disable_redaction](#disable_redaction) is enabled.                             |
| Response bodies            | At most 4096 bytes are read and discarded. The body is never parsed.                           |

## Delivery guarantees

Delivery is best effort with retries, and queues are held in memory only. An event is dropped when:

- the destination's buffer is full,
- it fails every attempt,
- it receives a permanent failure response, or
- Authelia stops or crashes before it is delivered.

On shutdown the queued events have five seconds to drain. That period is shared by every destination, and when it
elapses the delivery in progress is cancelled and the remainder are dropped with an error logged.

A receiver may see the same event twice, so deduplicate on the envelope `id`, which is the same on every attempt. See
[Idempotency](../../reference/guides/webhook-events.md#idempotency). A destination which does not batch is sent its
events in the order they were emitted.

Do not use webhooks as the system of record for anything you cannot afford to lose. They are a notification and
automation mechanism, not an audit log.

## Metrics

When [telemetry](../../reference/guides/metrics.md) is enabled, delivery outcomes are recorded against the
`authelia_webhook_delivery` counter with the labels `destination`, `event`, and `outcome`. The `outcome` label is one of
`delivered`, `retried`, or `dropped`.

An attempt which fails and will be retried records `retried`, so one event may record several `retried` outcomes before
it records `delivered` or `dropped`.

Events which occur while a destination is refused, held, or suspended are not queued and are not counted.

[CloudEvents]: https://cloudevents.io/
