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
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia can deliver a description of each of a fixed set of occurrences to one or more external receivers as an HTTPS
request. Each occurrence produces exactly one event, encoded as a [CloudEvents] 1.0 structured mode JSON document.

This page describes how the subsystem is configured. The payload contract, the request headers, the signature
verification procedure, and the behavior of each individual event type are documented in the
[Webhook Events](../../reference/guides/webhook-events.md) reference guide, which anyone writing a receiver should read.

Webhooks are disabled when no destinations are configured. Delivery is deliberately best effort: an unhealthy or
unreachable receiver never blocks, delays, or fails an authentication or self-service operation.

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
        - 'user.*'
        - 'security.ban.applied'
      timeout: '10 seconds'
      buffer_size: 256
      disable_redaction: false
      signature:
        secret: 'insecure_secret'
        algorithm: 'sha256'
      authentication:
        bearer:
          token: 'insecure_token'
          header: 'Authorization'
          scheme: 'Bearer'
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
          - 'security.*'
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

Performs a single delivery attempt against every destination during startup so that a misconfigured receiver is
discovered immediately rather than at the time of the first real occurrence.

A failed check is logged as a warning and _**never**_ prevents startup, because webhook availability must not prevent
Authelia from authenticating users.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The startup check delivers a synthetic
[system.startup_check](../../reference/guides/webhook-events.md#the-startup-check-probe) event whose `data` object
carries `probe` set to `true`, and it delivers it to _**every**_ destination regardless of the [events](#events)
selectors that destination configures. A receiver which only subscribes to `user.*` will still see this one event.

It does not describe an occurrence: no user did anything, and a receiver must not record it as an authentication, a
security signal, or an audit entry. Acknowledge it with a `2xx` and discard it.
{{< /callout >}}

### destinations

A list of receivers. Each destination has its own queue, worker, HTTP client, and retry policy, so a slow or broken
receiver does not affect any other receiver.

#### name

{{< confkey type="string" required="yes" >}}

The unique name of this destination. It labels log entries, validation errors, and the
[metrics](#metrics) counter. It is never included in a payload.

Two destinations may not share a name.

#### address

{{< confkey type="string" required="yes" >}}

The URL events are delivered to. Unlike most other `address` options this is a plain URL rather than the common address
syntax, as it names an HTTP endpoint rather than a connector.

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
The scheme _**must**_ be `https` and there is _**no**_ option to disable this. Unlike the
[SMTP notifier](../notifications/smtp.md#disable_require_tls) there is no `disable_require_tls` equivalent, because
payloads describe authentication activity and may, when [disable_redaction](#disable_redaction) is enabled, contain
credential equivalent values.

A receiver using a private or self-signed certificate authority must be accommodated with the [tls](#tls) options or the
[certificates_directory](introduction.md#certificates_directory) global option, not by downgrading the transport.
{{< /callout >}}

The address must not contain user information, i.e. a `username:password@` prefix before the host, or a fragment. Use the
[authentication](#authentication) options to present credentials instead.

Redirect responses are never followed. A redirect would resend the destination's credentials to a host which has not been
vetted, so a `3xx` response is treated as a permanent delivery failure and the event is dropped.

#### events

{{< confkey type="list(string)" required="yes" >}}

The event types this destination receives. Each entry is either the exact name of a registered event type, or a prefix
glob which ends in an asterisk. A single `'*'` selects every type.

```yaml {title="configuration.yml"}
webhooks:
  destinations:
    - name: 'admin-api'
      events:
        - 'user.credential.*'
        - 'security.ban.applied'
```

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
A selector which matches no registered event type is a _**startup error**_, not a warning. This is deliberate: a selector
which matches nothing is almost always a typo, and silently ignoring it would produce a destination that appears healthy
while receiving nothing.
{{< /callout >}}

{{< callout context="caution" title="Event Volume" icon="outline/alert-triangle" >}}
Do not subscribe to everything without considering the volume. _**Every**_ authorization request which authenticates,
and whose credential check is not served from the authentication backend's password cache, emits a
`security.authentication.succeeded` event. On a deployment where proxies present basic authentication credentials on
each request this is a very high volume feed, potentially one event per HTTP request to every protected application.

Scope the selectors to the types the receiver actually acts on. If you want a security feed rather than an audit feed,
`security.authentication.failed` and `security.ban.applied` are usually what you want, and
`security.authentication.succeeded` is usually not.
{{< /callout >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `security.ban.expired` type is registered, is accepted by this option, and has a published schema, but it is
_**never emitted**_. Selecting it is valid configuration which delivers nothing. See
[security.ban.expired is never emitted](../../reference/guides/webhook-events.md#securitybanexpired-is-never-emitted) for
the reason and for what to do instead.
{{< /callout >}}

The full list of types is documented in the
[event catalogue](../../reference/guides/webhook-events.md#event-catalogue).

#### timeout

{{< confkey type="string,integer" syntax="duration" default="10 seconds" required="no" >}}

The timeout applied to an individual delivery attempt, covering the entire request including connection establishment,
the TLS handshake, and reading the response.

A value of zero or less is treated as unset and the default is applied.

#### buffer_size

{{< confkey type="integer" default="256" required="no" >}}

The number of events which may be queued for this destination at once. A value of `0` is treated as unset and the
default is applied; a negative value is a startup error.

Enqueuing never blocks. When the queue is full the event is dropped, an error is logged, and the
[metrics](#metrics) counter records a `dropped` outcome. This is the mechanism which guarantees that a slow receiver can
never slow down Authelia itself.

A destination delivers serially and sleeps during retry backoff, so a receiver which is failing and being retried is also
not consuming its queue. Size the buffer with that in mind, or reduce [retry.attempts](#attempts) and
[retry.maximum_interval](#maximum_interval) for a high volume destination.

#### disable_redaction

{{< confkey type="boolean" default="false" required="no" >}}

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
Enabling this option causes credential equivalent values to be transmitted to the receiver, specifically the One-Time
Code issued for a session elevation, the identity verification link URL, and the revocation link URL.

A party holding those values can complete a password reset or a session elevation _**for any user**_, without any
knowledge of that user's password or second factor. Enabling this option therefore makes the receiver, the transport, and
every log, queue, and backup the receiver writes to, _**exactly as sensitive as the session secret**_. Treat a
compromise of such a receiver as a full compromise of every account which received a notification while it was
compromised.

Leave this disabled unless you have a concrete requirement and the receiver is held to the same standard as Authelia
itself.
{{< /callout >}}

When this option is disabled, which is the default, the sensitive fields are omitted from the payload entirely rather
than nulled or masked. See [Redaction](../../reference/guides/webhook-events.md#redaction) for the exact field list.

#### signature

The HMAC signature applied to each payload. The signature proves a payload originated from this Authelia instance, and
is the _**only**_ thing which does so. It is distinct from [authentication](#authentication), which proves Authelia's
identity to the receiver; both may be configured and they solve different problems.

##### secret

{{< confkey type="string" required="no" >}}

The shared secret used to sign each payload. Signing is disabled and the `X-Authelia-Signature` header is omitted when
this is empty, which is _**strongly discouraged**_ for any receiver which acts on the events it receives.

It's **strongly recommended** this is a
[Random Alphanumeric String](../../reference/guides/generating-secure-values.md#generating-a-random-alphanumeric-string)
with 64 or more characters.

The receiver side of this value, and the verification procedure it must implement, is documented in
[Verifying the signature](../../reference/guides/webhook-events.md#verifying-the-signature).

##### algorithm

{{< confkey type="string" default="sha256" required="no" >}}

The HMAC hash algorithm. Must be either `sha256` or `sha512`.

The header format is identical for both, and the `v1=` element is a lowercase hexadecimal digest whose length is the only
externally visible difference. A receiver must be configured with the algorithm out of band; it is not advertised in the
request.

#### authentication

The credentials Authelia presents to the receiver. The `bearer` and `basic` options are mutually exclusive and
configuring both is a startup error.

These credentials prove Authelia's identity to the receiver. They do not prove the payload is genuine, because anything
which can read a request can replay the header it carries. Use [signature](#signature) for that.

##### bearer

Presents a token in a header.

```yaml {title="configuration.yml"}
webhooks:
  destinations:
    - name: 'admin-api'
      authentication:
        bearer:
          token: 'insecure_token'
          header: 'Authorization'
          scheme: 'Bearer'
```

|   Name   |   Type   |     Default     | Required | Sensitive | Description                                                                                                     |
| :------: | :------: | :-------------: | :------: | :-------: | :-------------------------------------------------------------------------------------------------------------- |
| `token`  | `string` |                 |   yes    |    yes    | The token value.                                                                                                |
| `header` | `string` | `Authorization` |    no    |    no     | The header the token is sent in.                                                                                |
| `scheme` | `string` |    `Bearer`     |    no    |    no     | The scheme prefixed to the token, separated by a space. Configure as an empty string for a bare API key header. |

The `scheme` default of `Bearer` is only applied when `header` is `Authorization`. A custom header such as `X-API-Key`
therefore sends the bare token unless a scheme is explicitly configured.

##### basic

Presents HTTP Basic credentials in the `Authorization` header.

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

Additional static headers included with every request to this destination. Useful for routing or tenancy hints a
receiver needs before it parses the body.

```yaml {title="configuration.yml"}
webhooks:
  destinations:
    - name: 'admin-api'
      headers:
        X-Tenant: 'acme'
```

A destination may not set a reserved header, and configuring one is a startup error rather than a silently ignored
entry. The reserved headers are:

- `Content-Type`
- `User-Agent`
- Any header beginning with `X-Authelia-`
- The header occupied by this destination's own [authentication](#authentication) block

#### validation

Before any event is delivered Authelia performs the [CloudEvents abuse protection] handshake against the destination:
an `OPTIONS` request to its address carrying `WebHook-Request-Origin` and `WebHook-Request-Rate`, which the receiver
must answer with a 2xx and a `WebHook-Allowed-Origin` of either the same origin or an asterisk.

[CloudEvents abuse protection]: https://github.com/cloudevents/spec/blob/main/cloudevents/http-webhook.md#4-abuse-protection

{{< callout context="danger" title="Required" icon="outline/alert-octagon" >}}
A destination which does not confirm the handshake receives _**no events at all**_. A receiver must implement the
`OPTIONS` response, or the destination must set [disable](#disable).
{{< /callout >}}

The destination's [authentication](#authentication) and [headers](#headers) are sent with the handshake, so a receiver
which authorizes deliveries can authorize it the same way. No event is carried, so no `X-Authelia-*` event headers are
present.

A transient failure is retried under this destination's [retry](#retry) policy, the same as a delivery: a `429` honors
`Retry-After`, and a `5xx` or a failure which produced no status code backs off and is tried again. Anything a retry
cannot change is refused immediately, which means any other `4xx` or a response which does not permit the origin.

A refusal is logged and the destination is skipped. Startup continues regardless, because webhook availability must
never prevent Authelia running. The handshake is performed once per start, not periodically, it is bounded so that a
long `Retry-After` cannot hold startup open, and destinations are validated concurrently so one slow receiver does not
delay the others.

##### disable

{{< confkey type="boolean" default="false" required="no" >}}

Skips the handshake and delivers events without asking the destination to confirm that it accepts them. Use this only
for a receiver you control which does not implement `OPTIONS`.

##### origin

{{< confkey type="string" required="no" >}}

The origin presented in `WebHook-Request-Origin`.

By default this is used only when an origin cannot be derived from the configuration, which is the host of the
[Authelia URL](../session/introduction.md#authelia_url) of the first configured session cookie domain. The derived
value is preferred because it names the instance which actually sends the events. Set [force_origin](#force_origin) to
present this value regardless.

If no origin can be derived and none is configured, the handshake fails and the destination receives no events.

##### force_origin

{{< confkey type="boolean" default="false" required="no" >}}

Presents [origin](#origin) even when one could be derived from the configuration. Use this where Authelia is reached
under a different name than the one it knows itself by, behind a gateway for example. Enabling it without an `origin`
is a startup error.

##### rate

{{< confkey type="integer" required="no" >}}

The delivery rate requested in `WebHook-Request-Rate`, in requests per minute. An unrestricted rate is requested when
this is absent or `0`.

A receiver may answer with `WebHook-Allowed-Rate` to cap the rate regardless of what was requested. Authelia then
paces its requests to that ceiling, including retries, since a retry is another request. Note that it caps _requests_
rather than events, so a [batching](#batch) destination carries far more events within the same rate. Pacing is
interrupted by shutdown, so it never extends the drain beyond its grace period.

#### batch

Batching accumulates events and delivers several in a single request. It is disabled unless [size](#size) is
configured, and it is configured per destination, so one receiver may batch while another does not.

A batching destination sends the CloudEvents batched content mode: `Content-Type` is
`application/cloudevents-batch+json` and the body is a JSON array of the same envelopes, so each member carries its own
`type`, `id` and `dataschema`. The per event `X-Authelia-Event-Type` and `X-Authelia-Event-Id` headers are replaced by
`X-Authelia-Event-Count`. A batching destination always sends an array, even when the batch holds one event, so a
receiver has a single format to implement.

A batch is one request, so it succeeds or fails as a unit. There is no way to accept part of a batch: a permanent
failure drops every event it carried, and a retry resends all of them. Order is preserved within a batch.

##### size

{{< confkey type="integer" required="no" >}}

The number of events delivered in a single request. Batching is disabled when this is absent or `0`. A negative value
is a startup error, as is a value greater than [buffer_size](#buffer_size), which could never be reached.

##### max_wait

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The longest an event waits for its batch to fill before the batch is delivered regardless of how full it is. The window
runs from the first event of the batch, so no event is ever held longer than this no matter how slowly the batch fills.

This is the latency batching adds. A destination which receives a ban or a failed authentication will see it up to this
long after it occurred, unless the type is listed in [immediate](#immediate).

##### immediate

{{< confkey type="list(string)" required="no" >}}

The event types which bypass batching and are delivered on their own as soon as they occur, either exact names or
prefix globs ending in an asterisk, matching the syntax of [events](#events). A selector matching no known event type
is a startup error.

Use this for occurrences an operator should see without delay:

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
          - 'security.*'
```

{{< callout context="caution" title="Ordering" icon="outline/alert-triangle" >}}
A destination which lists types here emits two wire formats, and an immediate event may overtake batched events which
occurred before it, because it does not wait for the batch to flush. Order is guaranteed only within a batch. A
receiver which needs a total order must sort on the envelope `time` attribute rather than on arrival.
{{< /callout >}}

#### retry

The retry policy applied to a failed delivery attempt for this destination. The classification of a response, and the
handling of `Retry-After`, are documented in
[Retries](../../reference/guides/webhook-events.md#retries-and-failures).

##### attempts

{{< confkey type="integer" default="3" required="no" >}}

The maximum number of delivery attempts _**including**_ the first. A value of `1` disables retrying. A negative value is
a startup error, and a value of `0` is treated as unset so the default is applied.

##### initial_interval

{{< confkey type="string,integer" syntax="duration" default="1 second" required="no" >}}

The backoff interval before the second attempt. The interval doubles before each subsequent attempt.

Up to twenty percent negative jitter is applied to each wait so that multiple destinations recovering from the same
outage do not retry in lockstep.

##### maximum_interval

{{< confkey type="string,integer" syntax="duration" default="1 minute" required="no" >}}

The ceiling applied to the doubling backoff interval.

It is also the ceiling applied to a `Retry-After` header returned with a `429 Too Many Requests` response. Both forms of
that header are honored, a count of seconds and an HTTP date, but neither can make Authelia wait longer than this
value.

#### tls

{{< confkey type="structure" structure="tls" required="no" >}}

Controls the TLS connection verification parameters for this destination. The minimum version defaults to `TLS1.2`.

By default Authelia uses the system certificate trust for TLS certificate verification and the
[certificates_directory](introduction.md#certificates_directory) global option can be used to augment this.

## Secrets

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `signature.secret`, `authentication.bearer.token`, and `authentication.basic.password` options _**cannot**_ be
supplied with the [secrets](../methods/secrets.md) `_FILE` environment variable mechanism, and no
`AUTHELIA_WEBHOOKS_DESTINATIONS_..._FILE` variable exists.

This is not specific to webhooks. `destinations` is a list of objects, and no section of the configuration which is a
list of objects can be configured by an environment variable or a secret. The same limitation applies to the access
control rules, the OpenID Connect 1.0 clients, and the session cookies. See the note under
[Environment variables](../methods/secrets.md#environment-variables) and
[ADR2](../../reference/architecture-decision-log/2.md).

A destination whose `signature.secret` was "moved" to a `_FILE` variable is not signed at all, because the variable is
simply ignored. Verify that the `X-Authelia-Signature` header is present at the receiver after any such change.
{{< /callout >}}

To keep these values out of the configuration file itself, use the
[file filters](../methods/files.md#file-filters) `fileContent` template function, which reads the value from a file at
load time:

```yaml {title="configuration.yml"}
webhooks:
  destinations:
    - name: 'admin-api'
      signature:
        secret: '{{ fileContent "/run/secrets/webhook_signature_secret" | trim }}'
```

The `trim` is important: unlike the `_FILE` mechanism, `fileContent` does not strip a trailing newline, and a secret
with a stray newline produces a signature the receiver cannot reproduce.

File filters are not enabled by default; see [File Filters](../methods/files.md#file-filters) for how to enable them.

## Security

A short summary of the decisions this subsystem makes on your behalf, and the ones it leaves to you.

| Concern                    | Behavior                                                                                                                          |
| :------------------------- | :-------------------------------------------------------------------------------------------------------------------------------- |
| Transport                  | HTTPS only, no escape hatch. TLS 1.2 minimum by default.                                                                          |
| Redirects                  | Never followed, so credentials are never resent to an unvetted host. A `3xx` is a permanent failure.                              |
| Payload authenticity       | Only the HMAC [signature](#signature) proves a payload came from Authelia. Nothing else in the request does.                      |
| Receiver authenticity      | The [tls](#tls) verification parameters. `skip_verify` disables this.                                                             |
| Credential equivalent data | Omitted unless [disable_redaction](#disable_redaction) is enabled. See the warning on that option before enabling it.             |
| Response bodies            | Read up to 4096 bytes and discarded, so a hostile or broken receiver cannot exhaust memory. The body is never parsed or acted on. |

## Delivery guarantees

Delivery is best effort with retries, and queues are held in memory only. A receiver may therefore see the same event
twice, so deduplicate on the envelope `id`, which is stable across every attempt.

- An event which cannot be queued because the buffer is full is dropped.
- An event which fails every attempt is dropped.
- An event which receives a permanent failure response is dropped.
- Events queued but not yet delivered when Authelia shuts down are given a five second grace period to drain, and the
  remainder are dropped with an error logged. The grace period is shared by every destination rather than granted to
  each of them in turn, and when it elapses the delivery in progress is cancelled and any pending retry backoff is
  abandoned, so that a slow or unresponsive receiver cannot delay the shutdown of the rest of Authelia.
- Events are not persisted, so a restart or a crash loses whatever was queued.

Do not use webhooks as the system of record for anything you cannot afford to lose. They are a notification and
automation mechanism, not an audit log. Each destination preserves the order in which events were emitted, and the event
identifier is stable across retries so a receiver can deduplicate; see
[Idempotency](../../reference/guides/webhook-events.md#idempotency).

## Metrics

When [telemetry](../../reference/guides/metrics.md) is enabled, delivery outcomes are recorded against the
`authelia_webhook_delivery` counter with the labels `destination`, `event`, and `outcome`. The `outcome` label is one of
`delivered`, `retried`, or `dropped`.

An attempt which fails and will be retried records `retried`, so a single event may record several `retried` outcomes
before it records either `delivered` or `dropped`.

[CloudEvents]: https://cloudevents.io/
