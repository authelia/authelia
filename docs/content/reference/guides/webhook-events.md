---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Webhook Events"
description: "A reference guide on the webhook event payloads Authelia delivers"
summary: "This section contains reference documentation for the events Authelia delivers to webhook receivers, the envelope which carries them, and how a receiver verifies and interprets them."
date: 2026-09-13T01:31:45+10:00
draft: false
images: []
weight: 220
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

This guide is for people writing a receiver. It documents the envelope, the request, how to verify a payload, and what
each event type contains. The sending side is documented in the
[webhooks](../../configuration/miscellaneous/webhooks.md) configuration reference.

## Before you build a receiver

Several behaviors are easy to get wrong. Each is documented in full below:

- A destination receives nothing until it completes the [abuse protection](#abuse-protection) handshake.
- [`com.authelia.security.ban.expired` is never emitted](#comautheliasecuritybanexpired-is-never-emitted).
- [`com.authelia.security.authentication.failed` often carries an empty
  `username`](#an-empty-username-on-a-failure-is-normal), including for attempts against users which do not exist.
- [`internal_error` is broad](#internal_error-is-broad). With the LDAP backend it includes a wrong password.
- [`user_not_found`](#reason) is only set where the authentication backend reports it.
- [An attempt refused because the address is banned names the
  user](#an-attempt-from-a-banned-address-still-names-the-user), not the address.
- [`com.authelia.security.ban.applied` arrives before the failure which caused it](#ordering).
- [The published schemas are permissive](#schemas): a field a schema allows is not necessarily one the event carries.
- [`com.authelia.security.authentication.succeeded` can be very high volume](#event-volume).

## Envelope

Every payload is a [CloudEvents] 1.0 document in structured JSON mode.

```json
{
  "specversion": "1.0",
  "id": "01a095d4-29bd-7a02-98c7-bd3b9c599d3a",
  "type": "com.authelia.user.password.changed",
  "source": "https://auth.example.com",
  "time": "2026-09-12T04:05:06.123456789Z",
  "datacontenttype": "application/json",
  "dataschema": "https://www.authelia.com/schemas/webhooks/v1/com.authelia.user.password.changed.json",
  "autheliaversion": "4.39.0",
  "subject": "john",
  "data": {
    "username": "john",
    "display_name": "John Smith",
    "emails": ["john@example.com"],
    "remote_ip": "198.51.100.12",
    "notification": {
      "sent": true,
      "recipients": [
        {
          "email": "john@example.com",
          "username": "john",
          "display_name": "John Smith"
        }
      ],
      "title": "Password changed successfully",
      "values": {
        "body_prefix": "your",
        "body_event": "Password Change",
        "body_suffix": "was successful.",
        "details": {
          "Action": "Password Change"
        }
      }
    }
  }
}
```

|     Attribute     |   Type   | Notes                                                                                         |
| :---------------: | :------: | :-------------------------------------------------------------------------------------------- |
|   `specversion`   | `string` | Always `1.0`.                                                                                 |
|       `id`        | `string` | A UUIDv7, the same on every delivery attempt. See [Idempotency](#idempotency).                |
|      `type`       | `string` | One of the types in the [event catalogue](#event-catalogue).                                  |
|     `source`      | `string` | The URL of the Authelia instance. See [source](#source).                                      |
|      `time`       | `string` | RFC 3339 timestamp of the occurrence in UTC, not of the delivery attempt.                     |
| `datacontenttype` | `string` | Always `application/json`.                                                                    |
|   `dataschema`    | `string` | An `https://www.authelia.com/schemas/webhooks/` URI. See [Schemas](#schemas).                 |
| `autheliaversion` | `string` | A CloudEvents extension attribute carrying the Authelia version.                              |
|     `subject`     | `string` | What the occurrence concerns. Omitted when there is nothing to name. See [subject](#subject). |
|      `data`       | `object` | The payload, whose shape is determined by `type`.                                             |

### subject

The `subject` attribute lets a receiver or an intermediary filter and route events without parsing `data`. It is
omitted when it has no value and is never an empty string.

| Event types                              | `subject`                                                                 |
| :--------------------------------------- | :------------------------------------------------------------------------ |
| `com.authelia.user.*`                    | The username.                                                             |
| `com.authelia.security.authentication.*` | The username. Omitted when no user was resolved.                          |
| `com.authelia.security.ban.*`            | The banned value: the username for a user ban, the address for an IP ban. |
| `com.authelia.system.startup_check`      | Omitted.                                                                  |

### source

The `source` attribute is taken from the first configured
[session cookie domain](../../configuration/session/introduction.md#cookies): its
[authelia_url](../../configuration/session/introduction.md#authelia_url), or otherwise `https://` followed by its
[domain](../../configuration/session/introduction.md#domain).

Do not use `source` to decide whether a payload is genuine. Use the [signature](#verifying-the-signature).

## Abuse protection

Before Authelia delivers any event to a destination it performs the [CloudEvents abuse protection] handshake:

[CloudEvents abuse protection]: https://github.com/cloudevents/spec/blob/main/cloudevents/http-webhook.md#4-abuse-protection

```text
OPTIONS /hooks HTTP/1.1
Host: admin.example.com
WebHook-Request-Origin: auth.example.com
WebHook-Request-Rate: 120
WebHook-Request-Callback: https://auth.example.com/api/webhooks/confirm?id=admin-api&key=...
```

A receiver confirms by answering with `WebHook-Allowed-Origin`:

```text
HTTP/1.1 200 OK
WebHook-Allowed-Origin: auth.example.com
WebHook-Allowed-Rate: 120
```

|           Header           | Meaning                                                                                              |
| :------------------------: | :--------------------------------------------------------------------------------------------------- |
|  `WebHook-Request-Origin`  | The DNS name of the Authelia instance sending the events.                                            |
|   `WebHook-Request-Rate`   | The rate Authelia requests, in requests per minute. Omitted when no rate is configured.              |
| `WebHook-Request-Callback` | The URL a receiver requests to confirm later. See [Confirming by callback](#confirming-by-callback). |
|  `WebHook-Allowed-Origin`  | The requested origin, compared without regard to case, or `*`. Any other value is a refusal.         |
|   `WebHook-Allowed-Rate`   | A limit in requests per minute, or `*`. Required when `WebHook-Request-Rate` was sent.               |

The request carries the destination's configured authentication and static headers, and
`X-Authelia-Delivery-Attempt`. It carries no event, so there are no `ce-` headers.

{{< callout context="danger" title="Required" icon="outline/alert-octagon" >}}
A destination which has not confirmed receives no events at all, and events which occur in the meantime are not queued
for it. An operator may skip the handshake with the destination's
[validation disable](../../configuration/miscellaneous/webhooks.md#disable) option.
{{< /callout >}}

The response is handled as follows:

| Response                                                                  | Outcome                                                      |
| :------------------------------------------------------------------------ | :----------------------------------------------------------- |
| `429`, `503`, or another `5xx`, or no response                            | Retried. See [Retries and failures](#retries-and-failures).  |
| `WebHook-Allowed-Origin` permits the origin, with any other status        | Confirmed.                                                   |
| `WebHook-Allowed-Origin` names a different origin                         | Refused.                                                     |
| `WebHook-Allowed-Rate` is invalid, or is absent when a rate was requested | Refused.                                                     |
| No `WebHook-Allowed-Origin`, with any other status                        | Held until [confirmed by callback](#confirming-by-callback). |
| The retries are exhausted                                                 | Held until [confirmed by callback](#confirming-by-callback). |

Consent is carried by the headers, so a 2xx alone confirms nothing. A refusal is logged and is never retried, and a
refused destination receives no events until Authelia is restarted. In every case Authelia itself still starts.

The handshake is performed once per start. It is limited to thirty seconds per destination, including retries, and
destinations are validated concurrently. It is repeated only to recheck a destination
[suspended by a `410`](#retries-and-failures).

When a receiver answers with a numeric `WebHook-Allowed-Rate`, Authelia paces its requests to that limit. The limit
applies to requests, not events: a retry counts against it and a [batch](#batched-deliveries) counts once.

### Confirming by callback

A receiver which cannot decide immediately answers the handshake with a 2xx and no `WebHook-Allowed-Origin`. The
destination is then held. It receives no events, including the [startup check probe](#the-startup-check-probe).

A destination is also held when the handshake gets no usable answer, for example a receiver which does not implement
`OPTIONS` and answers `405`. This is logged as an error, since the receiver may never have seen the callback URL.

The receiver grants permission by requesting the URL from `WebHook-Request-Callback` with a `GET` or a `POST`. An
administrator who finds the URL in the receiver's logs can therefore open it in a browser. The destination receives
events from that moment.

```text
POST /api/webhooks/confirm?id=admin-api&key=... HTTP/1.1
Host: auth.example.com
WebHook-Allowed-Rate: 120
```

The request may carry `WebHook-Allowed-Rate`. When it does not, or the value is invalid, the rate Authelia requested
applies. Authelia answers `200` when permission is granted or was already granted, and `404` when the destination is
unknown, the key is wrong, the destination refused the handshake, or the destination disables validation.

The URL is built from the [source](#source), names the destination in `id`, and carries a `key`. Treat it as a secret,
as anyone who holds it can grant the destination permission.

The `key` is an HMAC of the destination's name and address, under a key which Authelia generates once and keeps in its
storage. It is therefore:

- different for every destination,
- unchanged by a restart, and
- the same on every Authelia instance which shares the storage backend.

Changing a destination's name or [address](../../configuration/miscellaneous/webhooks.md#address) changes its `key`.

A confirmation is recorded in storage against the destination's name and address, so a destination which has confirmed
stays confirmed when Authelia restarts. It no longer applies when:

- the destination's name or address is changed,
- the destination refuses the handshake, or
- the destination answers a delivery with `410`.

Where several Authelia instances share one storage backend, a callback confirms the instance which receives it
immediately. The others check storage once a minute while a destination is held.

## Request

Each delivery is an HTTP `POST` of the envelope to the destination address.

|            Header             | Value                                                                               |
| :---------------------------: | :---------------------------------------------------------------------------------- |
|        `Content-Type`         | `application/cloudevents+json; charset=utf-8`                                       |
|         `User-Agent`          | `Authelia/` followed by the version                                                 |
| `X-Authelia-Webhook-Version`  | The contract version, currently `1`. Matches the `/v1/` segment of the schema URIs. |
| `X-Authelia-Delivery-Attempt` | The attempt number of this request, starting at `1`.                                |
|    `X-Authelia-Signature`     | The HMAC signature. Present only when a signature secret is configured.             |
|   `WebHook-Request-Origin`    | The origin presented in the handshake. Present when an origin is known.             |
|       `ce-specversion`        | The envelope `specversion`.                                                         |
|            `ce-id`            | The envelope `id`.                                                                  |
|           `ce-type`           | The envelope `type`.                                                                |
|          `ce-source`          | The envelope `source`.                                                              |
|           `ce-time`           | The envelope `time`.                                                                |
|        `ce-dataschema`        | The envelope `dataschema`.                                                          |
|     `ce-autheliaversion`      | The envelope `autheliaversion`.                                                     |
|         `ce-subject`          | The envelope `subject`. Absent when the envelope has none.                          |

The `ce-` headers are the metadata headers of the [CloudEvents HTTP protocol binding]. They repeat the envelope
attributes so a receiver can route or deduplicate a delivery before it parses the body.

- HTTP header names are case insensitive, so they may arrive as `Ce-Id`.
- The values are percent-encoded: a space, a double quote, a percent sign, and every character outside printable ASCII
  is sent as the `%XX` form of its UTF-8 bytes. Decode a value once before using it.
- There is no `ce-datacontenttype` header, as the binding forbids it.
- The [signature](#verifying-the-signature) covers the body only, so a receiver which verifies it should take these
  values from the body.

[CloudEvents HTTP protocol binding]: https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/bindings/http-protocol-binding.md#323-metadata-headers

Any [headers](../../configuration/miscellaneous/webhooks.md#headers) configured for the destination are also sent, as
are the credentials configured by its [authentication](../../configuration/miscellaneous/webhooks.md#authentication)
options:

- A token, either in the `Authorization` header or in the `access_token` query parameter of the address. A request
  which carries it in the address also carries `Cache-Control: no-store`.
- HTTP Basic credentials in the `Authorization` header, sent with every request and never in answer to a challenge.
  Basic credentials are not a token, so a destination which uses them does not satisfy the token requirement of the
  [CloudEvents] HTTP webhook specification.

A receiver should respond with a `2xx` as soon as it has durably accepted the event, and do its own processing
afterwards. Authelia reads at most 4096 bytes of a response body and discards them.

### Batched deliveries

A destination which configures [batching](../../configuration/miscellaneous/webhooks.md#batch) is sent several events
in one request. The body is a CloudEvents batched content mode array of the same envelopes, and the request differs in
three ways:

|          Header          | Value                                               |
| :----------------------: | :-------------------------------------------------- |
|      `Content-Type`      | `application/cloudevents-batch+json; charset=utf-8` |
| `X-Authelia-Event-Count` | The number of events in the array.                  |
|          `ce-*`          | Absent.                                             |

A batch is always an array, even when it holds one event. A batching destination is still sent single event requests
for the types in its [immediate](../../configuration/miscellaneous/webhooks.md#immediate) option and for the
[startup check probe](#the-startup-check-probe), so branch on `Content-Type`.

A batch succeeds or fails as a unit: a `2xx` acknowledges every event in the array and a permanent failure drops every
event in it.

Authelia writes the array in the order the events were emitted, but the [CloudEvents] specification defines a batch as
unordered and intermediaries may split or merge batches. An immediate event can also arrive before batched events
which occurred earlier. Order events by the envelope `time` attribute.

## Identifying an Authelia payload

1. **Before parsing**, the `X-Authelia-Webhook-Version` and `User-Agent` headers are routing hints. Anyone can send
   them.
2. **After parsing**, the `dataschema` attribute identifies the format the payload claims to have. Check that its host
   is `www.authelia.com` and its path starts with `/schemas/webhooks/v1/`.
3. **Cryptographically**, the `X-Authelia-Signature` header is the only thing which shows the payload came from your
   Authelia instance.

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
The `dataschema` URI, the `X-Authelia-*` headers, and the `User-Agent` can be sent by anybody. A receiver must only act
on a payload whose [signature](#verifying-the-signature) it has verified.
{{< /callout >}}

## Verifying the signature

When a [signature secret](../../configuration/miscellaneous/webhooks.md#secret) is configured, each request carries an
`X-Authelia-Signature` header:

```text
t=1757640306,v1=1f5a...c9
```

| Element | Meaning                                                                                   |
| :-----: | :---------------------------------------------------------------------------------------- |
|   `t`   | The Unix timestamp in seconds at which this attempt was signed. It is not the event time. |
|  `v1`   | The lowercase hexadecimal HMAC digest of the signed input.                                |

The signed input is the timestamp, a full stop, then the raw request body:

```text
HMAC-SHA256(secret, t + "." + rawBody)
```

The hash function is SHA-256 or SHA-512, as selected by the destination's
[algorithm](../../configuration/miscellaneous/webhooks.md#algorithm) option. It is not advertised in the request, so
the receiver must be configured with the same value.

To verify:

1. Compute the HMAC over the raw body bytes, before any parsing. Re-serialized JSON produces a different digest.
2. Compare the digests in constant time.
3. Reject a request whose `t` is outside a tolerance window. Five minutes is a reasonable default.
4. Reject a request with no signature header.

```python
import hashlib
import hmac
import time

DIGESTS = {"sha256": hashlib.sha256, "sha512": hashlib.sha512}


def verify(secret: str, header: str, body: bytes, algorithm: str = "sha256", tolerance: int = 300) -> bool:
    digest = DIGESTS.get(algorithm)

    if digest is None:
        return False

    try:
        parts = dict(part.split("=", 1) for part in header.split(","))
        timestamp, signature = parts["t"], parts["v1"]

        if abs(time.time() - int(timestamp)) > tolerance:
            return False
    except (AttributeError, KeyError, TypeError, ValueError):
        return False

    expected = hmac.new(secret.encode(), f"{timestamp}.".encode() + body, digest).hexdigest()

    return hmac.compare_digest(expected.encode(), signature.encode())
```

## Idempotency

The envelope `id` is generated once per occurrence and is the same on every delivery attempt. The `ce-id` header
carries the same value, so a receiver can deduplicate before it parses the body.

Duplicates are uncommon but possible. A receiver which accepts a request and then fails to respond within the
destination [timeout](../../configuration/miscellaneous/webhooks.md#timeout) is sent the event again.

The `id` is a UUIDv7, so identifiers sort in approximate order of occurrence. Use the envelope `time` attribute where
the order matters.

## Retries and failures

Each destination delivers serially and retries according to its own
[retry](../../configuration/miscellaneous/webhooks.md#retry) policy.

|      Response      | Behavior                                                                                                 |
| :----------------: | :------------------------------------------------------------------------------------------------------- |
|       `2xx`        | Delivered.                                                                                               |
|   `429` or `503`   | Retried. The `Retry-After` header is honored when present.                                               |
|       `410`        | The event is dropped and the destination is suspended.                                                   |
|    Other `5xx`     | Retried using the backoff interval.                                                                      |
|  Transport error   | Retried using the backoff interval. Covers timeouts, connection failures, and TLS verification failures. |
| Any other response | Permanent failure. The event is dropped.                                                                 |

Every `3xx` and every other `4xx` is a permanent failure. Redirects are not followed, so a receiver which answers with
a redirect loses every event.

An event which exhausts its attempts is dropped and logged as an error. Events are never persisted, so nothing is
redelivered after a restart.

### Suspension

A `410` means the receiver has been retired. The destination is suspended, its recorded confirmation is removed, and
events which occur while it is suspended are dropped.

A suspended destination is rechecked once an hour by repeating the [abuse protection](#abuse-protection) handshake. It
receives events again when a recheck is answered with a permitting `WebHook-Allowed-Origin`. Requesting the callback
URL does not resume a suspended destination.

A destination which [disables validation](../../configuration/miscellaneous/webhooks.md#disable) has no handshake to
repeat, so its suspension is lifted after an hour and the next event is delivered to it.

### Retry-After

`Retry-After` is honored as either a number of seconds or an HTTP date. A date in the past, a negative number, and any
other value are ignored and the backoff interval is used.

The wait is not capped at [maximum_interval](../../configuration/miscellaneous/webhooks.md#maximum_interval), so a
receiver asking to be left alone for a day receives no deliveries for a day. That includes later events, which are held
in the destination's buffer and dropped once it is full. A `Retry-After` does not add attempts.

### Backoff

The backoff interval starts at [initial_interval](../../configuration/miscellaneous/webhooks.md#initial_interval),
doubles before each subsequent attempt, and is capped at
[maximum_interval](../../configuration/miscellaneous/webhooks.md#maximum_interval). Each wait is shortened by a random
amount of up to twenty percent.

## Event volume

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Every authorization request which authenticates with basic credentials, and whose credential check is not served from
the authentication backend's password cache, emits a `com.authelia.security.authentication.succeeded` event. Where
proxies present basic credentials on every request, that can be one event per HTTP request to every protected
application.

Limit the destination's [events](../../configuration/miscellaneous/webhooks.md#events) selectors to the types the
receiver acts on. A destination whose queue fills drops events.
{{< /callout >}}

## Event catalogue

| Type                                              | Notification | Description                                                                      |
| :------------------------------------------------ | :----------: | :------------------------------------------------------------------------------- |
| `com.authelia.user.password.changed`              |     yes      | A user changed their own password from the settings area.                        |
| `com.authelia.user.password.reset`                |     yes      | A user completed a password reset.                                               |
| `com.authelia.user.credential.totp.added`         |     yes      | A user registered a One-Time Password configuration.                             |
| `com.authelia.user.credential.totp.removed`       |     yes      | A user removed a One-Time Password configuration.                                |
| `com.authelia.user.credential.webauthn.added`     |     yes      | A user registered a WebAuthn credential.                                         |
| `com.authelia.user.credential.webauthn.removed`   |     yes      | A user removed a WebAuthn credential.                                            |
| `com.authelia.user.identity_verification.started` |     yes      | An identity verification was started.                                            |
| `com.authelia.user.session.elevation.requested`   |     yes      | A session elevation was requested.                                               |
| `com.authelia.security.authentication.succeeded`  |      no      | An authentication attempt succeeded. See [Event volume](#event-volume).          |
| `com.authelia.security.authentication.failed`     |      no      | An authentication attempt failed.                                                |
| `com.authelia.security.ban.applied`               |      no      | Regulation banned a user or an address.                                          |
| `com.authelia.security.ban.expired`               |      no      | **Never emitted.** See [below](#comautheliasecuritybanexpired-is-never-emitted). |
| `com.authelia.system.startup_check`               |      no      | A connectivity probe. See [below](#the-startup-check-probe).                     |

The eight notification types correspond to the notifications documented in
[Events](../../configuration/notifications/events.md). That page lists six, because its "Second factor registered" and
"Second factor removed" rows each map to a TOTP type and a WebAuthn type here.

An occurrence produces one event. The email which accompanies it is described inside the event by the
[notification object](#the-notification-object), and the event is emitted after the attempt to send that email, whether
or not it succeeded.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
A password reset requested for a user which does not exist, or which has no email address, produces no
`com.authelia.user.identity_verification.started` event.
{{< /callout >}}

## Common data fields

|     Field      |    Type    | Presence                    | Notes                                                           |
| :------------: | :--------: | :-------------------------- | :-------------------------------------------------------------- |
|   `username`   |  `string`  | Always present, may be `""` | The username the occurrence concerns.                           |
| `display_name` |  `string`  | Omitted when unavailable    | Only on the notification types.                                 |
|    `emails`    | `[]string` | Omitted when unavailable    | Only on the notification types.                                 |
|  `remote_ip`   |  `string`  | Omitted when unavailable    | The client address, as determined by remote address resolution. |

Of these fields the `com.authelia.security.*` types carry `username` and `remote_ip` only, and `username` may be empty.
The fields each of them adds are documented under [Authentication](#authentication) and [Bans](#bans).

### Presence versus value

These fields are always serialized, even when the value is `""` or `false`, and are listed in `required` by the
published schemas:

- `username`, on every event type.
- `stage` and `method`, on the `com.authelia.security.authentication.*` types.
- `target` and `target_type`, on the `com.authelia.security.ban.*` types.
- `probe`, on `com.authelia.system.startup_check`.
- `sent`, on the `notification` object.
- `email`, on each entry of `notification.recipients`.
- Every envelope attribute except `subject`.

Every other field is omitted from the JSON when it has no value.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Test the value of `username`, not its presence. The key is present with the value `""` on the failure events which
name nobody. See [An empty username on a failure is normal](#an-empty-username-on-a-failure-is-normal).
{{< /callout >}}

## The notification object

The eight notification types carry a `notification` object describing the email for the occurrence.

```json
"notification": {
  "sent": true,
  "recipients": [
    {
      "email": "john@example.com",
      "username": "john",
      "display_name": "John Smith"
    }
  ],
  "title": "Password changed successfully",
  "values": {}
}
```

|    Field     |   Type    | Notes                                                                                      |
| :----------: | :-------: | :----------------------------------------------------------------------------------------- |
|    `sent`    | `boolean` | Whether the notification was successfully handed to the notifier.                          |
| `suppressed` | `boolean` | Present only when `true`. The notifier is disabled, so nothing was handed to it.           |
| `recipients` |  `array`  | The addressees. See below.                                                                 |
|   `title`    | `string`  | The notification subject line, which is also the `{{ .Title }}` placeholder in templates.  |
|   `error`    | `string`  | Why the notification was not sent. Present only when `sent` is `false` and not suppressed. |
|   `values`   | `object`  | The template attribute values. See [Notification values](#notification-values).            |

The `error` is a delivery error, a failure to look up the user, or the user having no email address. When the user
could not be looked up `recipients` and `values` are absent, and when the user has no email address `recipients` is
absent.

A `com.authelia.user.password.changed` event has no `notification` object at all when the password was changed but the
session could not be saved afterwards.

The rendered HTML and text bodies of the email are never included.

### Recipients

|     Field      |   Type   | Notes                                                          |
| :------------: | :------: | :------------------------------------------------------------- |
|    `email`     | `string` | Always present. The address the notification was addressed to. |
|   `username`   | `string` | Present when the address belongs to a known user.              |
| `display_name` | `string` | Present when recorded in the authentication backend.           |

For every current event type the sole recipient is the user the event concerns, so these values repeat the payload's
top level fields.

### A suppressed notification

When the [notifier is disabled](../../configuration/notifications/introduction.md#disable) `suppressed` is `true`,
`sent` is `false`, and `error` is absent. A `sent` of `false` with an `error` is a failure worth alerting on, and with
`suppressed` it is the configured behavior.

The `recipients` and `values` are present as they would be for a sent notification. Credential equivalent values such
as `one_time_code` and `link_url` are still omitted unless the destination sets
[disable_redaction](../../configuration/miscellaneous/webhooks.md#disable_redaction).

### Notification values

The populated fields of `values` depend on the event type.

| Type                                              | Populated `values` fields                                    |
| :------------------------------------------------ | :----------------------------------------------------------- |
| `com.authelia.user.password.changed`              | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `com.authelia.user.password.reset`                | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `com.authelia.user.credential.totp.added`         | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `com.authelia.user.credential.totp.removed`       | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `com.authelia.user.credential.webauthn.added`     | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `com.authelia.user.credential.webauthn.removed`   | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `com.authelia.user.identity_verification.started` | `domain`, `link_url`\*, `link_text`, `revocation_link_url`\* |
| `com.authelia.user.session.elevation.requested`   | `domain`, `one_time_code`\*, `revocation_link_url`\*         |

\* Redacted by default. See [Redaction](#redaction).

The published schemas permit every one of these fields on all eight types, because the types share one `values`
structure. This table describes what each type carries.

### Per-type additional fields

| Type                                              | Field         |   Type   | Notes                                                                                                 |
| :------------------------------------------------ | :------------ | :------: | :---------------------------------------------------------------------------------------------------- |
| `com.authelia.user.credential.*`                  | `description` | `string` | The user supplied description of a WebAuthn credential. Omitted for One-Time Password configurations. |
| `com.authelia.user.identity_verification.started` | `action`      | `string` | The action the verification authorizes. The only value currently emitted is `ResetPassword`.          |

## Redaction

Two event types carry credential equivalent values. These fields are omitted, not nulled or masked, unless the
destination sets [disable_redaction](../../configuration/miscellaneous/webhooks.md#disable_redaction).

| Type                                              | Redacted fields                        |
| :------------------------------------------------ | :------------------------------------- |
| `com.authelia.user.identity_verification.started` | `link_url`, `revocation_link_url`      |
| `com.authelia.user.session.elevation.requested`   | `one_time_code`, `revocation_link_url` |

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
A receiver of unredacted payloads holds values which complete a password reset or a session elevation for the named
user. It is as sensitive as the session secret, and so is anything it forwards those payloads to.
{{< /callout >}}

## Security events

### Authentication

`com.authelia.security.authentication.succeeded` and `com.authelia.security.authentication.failed` add the following
fields.

|  Field   |   Type   | Notes                                                 |
| :------: | :------: | :---------------------------------------------------- |
| `stage`  | `string` | `first_factor` or `second_factor`.                    |
| `method` | `string` | `password`, `totp`, `webauthn`, or `duo`.             |
| `reason` | `string` | The failure classification. Present on failures only. |

A passkey authentication is reported with a `stage` of `first_factor` and a `method` of `webauthn`.

A success always carries a username.

#### reason

The first matching row applies.

|         Value         | Meaning                                                                                                    |
| :-------------------: | :--------------------------------------------------------------------------------------------------------- |
|       `banned`        | The user or the address was already banned by regulation.                                                  |
| `invalid_credentials` | The user's authenticator rejected the attempt: a declined push or a failed WebAuthn assertion.             |
|   `user_not_found`    | The authentication backend reported that the user does not exist or is disabled.                           |
|   `internal_error`    | Any other error. See [internal_error is broad](#internal_error-is-broad).                                  |
| `invalid_credentials` | The attempt failed without an error: a wrong One-Time Password, or a wrong password with the file backend. |

`user_not_found` is set only when the authentication backend returns its "user not found" error, which it also returns
for a disabled user. It occurs:

- on the first factor password, re-authentication, and second factor password paths of the portal,
- on the basic credential path of proxy authorization, and
- on the passkey path, when the user who owns the presented credential no longer exists or is disabled.

A user disabled or deleted while holding a session therefore produces `user_not_found` on a `second_factor` event with
the username populated. Because `banned` takes precedence, an attempt against a missing user from a banned address is
`banned`.

The absence of `user_not_found` does not show that the username exists.

#### internal_error is broad

`internal_error` is given to any failure with an error which was not classified as a rejection or as a missing user. It
covers system faults, such as an unreachable authentication backend or a failing storage or session operation, but it
is not limited to them.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
With the LDAP backend a wrong password is reported as `internal_error`, because the backend returns the rejected bind
as an error. Only the file backend reports a wrong password as `invalid_credentials`.
{{< /callout >}}

Some faults produce no event at all: a failed regulation lookup, a connection pool timeout in the authentication
backend, and on the proxy authorization path any failure to look up the user other than the user not being found.

Do not alert on `internal_error` alone. Correlate it with the Authelia logs, which carry the underlying error.

#### An empty username on a failure is normal

`com.authelia.security.authentication.failed` often carries a `username` of `""` and no envelope `subject`. This
happens on two paths:

- **A first factor attempt for a user the backend cannot resolve.** This is what credential spraying against
  non-existent accounts looks like, so a receiver which ignores events with no username misses it.
- **Passkey failures.** Every passkey failure carries an empty username, except a refusal because the user is banned
  and a failure to regenerate the session.

Correlate these events on `remote_ip`.

### Bans

`com.authelia.security.ban.applied` and `com.authelia.security.ban.expired` add the following fields.

|     Field     |   Type   | Notes                                              |
| :-----------: | :------: | :------------------------------------------------- |
|   `target`    | `string` | The banned value, either a username or an address. |
| `target_type` | `string` | `user` or `ip`.                                    |
|   `expires`   | `string` | RFC 3339 timestamp. Present on `applied` only.     |

| `target_type` | `target`     | `username`   | `remote_ip`                         |
| :-----------: | :----------- | :----------- | :---------------------------------- |
|     `ip`      | The address  | `""`         | The address                         |
|    `user`     | The username | The username | The address which triggered the ban |

The envelope [subject](#subject) is the `target`.

#### com.authelia.security.ban.expired is never emitted

The type is registered, is accepted by the [events](../../configuration/miscellaneous/webhooks.md#events) option, and
has a published schema, but nothing emits it. Bans expire passively when their expiry time passes, and a ban revoked
with the `authelia storage` command is revoked by a separate process which cannot emit events. The type is kept so that
it can be implemented later without changing the contract.

To track active bans, schedule the expiry from the `expires` field of `com.authelia.security.ban.applied`.

#### An attempt from a banned address still names the user

When an attempt is refused because the address is banned, the `username` on the
`com.authelia.security.authentication.failed` event is the user who made the attempt, and the address is in
`remote_ip`. The address is never placed in `username`.

When the user cannot be resolved the `username` is empty, as described under
[An empty username on a failure is normal](#an-empty-username-on-a-failure-is-normal).

### Ordering

`com.authelia.security.ban.applied` is emitted before the `com.authelia.security.authentication.failed` event for the
attempt which triggered the ban, because recording the attempt is what creates the ban.

- Only first factor password failures create bans, so the failure which follows has a `stage` of `first_factor` and a
  `method` of `password`.
- One attempt can create both an address ban and a user ban, in that order, before its failure event.
- The `reason` of that failure is never `banned`. It is whatever the attempt would otherwise be, usually
  `invalid_credentials`, or `user_not_found` when an address is banned for attempts against users which do not exist.
  The `banned` reason appears on later attempts.

A destination which does not batch is sent its events in the order they were emitted. For a batching destination see
[Batched deliveries](#batched-deliveries). There is no ordering between destinations.

## The startup check probe

When [startup_check](../../configuration/miscellaneous/webhooks.md#startup_check) is enabled, each confirmed
destination is sent one `com.authelia.system.startup_check` event at startup to check that it is reachable and accepts
the credentials. It is always sent as a single event request.

It does not describe an occurrence, so a receiver must not record it as an authentication, a security signal, or an
audit entry. Its `data` object carries `probe` set to `true`. Acknowledge it with a `2xx` and discard it.

## Schemas

Fifteen [JSON Schema] documents are published: one per event type, one for the envelope, and one for the batched array.

```text
https://www.authelia.com/schemas/webhooks/v1/envelope.json
https://www.authelia.com/schemas/webhooks/v1/envelope-batch.json
https://www.authelia.com/schemas/webhooks/v1/<event type>.json
```

The `dataschema` attribute of each envelope is the URI of the schema for its `data` object. The version segment matches
the `X-Authelia-Webhook-Version` header, and a breaking change to a payload requires a new version.

A schema describes what a payload is allowed to contain, which is more than a given event type carries. Use
[Notification values](#notification-values) and [Per-type additional fields](#per-type-additional-fields) to decide
what to expect, and treat every field as optional except those listed under
[Presence versus value](#presence-versus-value).

No schema forbids additional properties. New payload fields and new envelope extension attributes are added within the
current version, so a receiver must accept members it does not know.

### OpenAPI

The deliveries are also described in the `webhooks` section of the [OpenAPI](https://www.authelia.com/api/) document
Authelia serves at `/api/openapi.yml`. It has one entry per event type, plus `validation` for the abuse protection
handshake and `batch` for batched deliveries, each with the request headers and the response codes a receiver is
expected to return. The payload schemas are embedded as components.

The section is always present, whether or not any destinations are configured.

[CloudEvents]: https://cloudevents.io/
[JSON Schema]: https://json-schema.org/
