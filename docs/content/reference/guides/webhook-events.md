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

This guide is for people writing a receiver. It documents the wire contract: the envelope, the request, how to verify a
payload is genuine, and what each event type actually contains.

How the sending side is configured is documented in the
[webhooks](../../configuration/miscellaneous/webhooks.md) configuration reference.

## Before you build a receiver

The subsystem has a number of behaviors which are surprising if you meet them for the first time in production. Each is
documented in full below, but they are collected here so that none of them is a surprise:

- [`security.ban.expired` is never emitted](#securitybanexpired-is-never-emitted), despite being a registered type with a
  published schema.
- [`security.authentication.failed` sometimes carries an empty `username`](#an-empty-username-on-a-failure-is-normal),
  including in exactly the credential spraying case you are most likely watching for.
- [`internal_error` is broader than it sounds](#internal_error-is-broad) and is not a clean infrastructure alarm.
- [`user_not_found` is emitted only where the backend says so](#reason), which is not every path where a submitted
  username does not exist.
- [An IP ban names its subject differently depending on which flow hit it](#an-ip-ban-names-its-subject-differently-by-flow).
- [`security.ban.applied` arrives before the failure which caused it](#ordering).
- [The published schemas are permissive, not exact](#schemas): a field being allowed by a schema does not mean the event
  ever carries it.
- [`security.authentication.succeeded` can be very high volume](#event-volume).

## Envelope

Every payload is a [CloudEvents] 1.0 document in structured JSON mode.

```json
{
  "specversion": "1.0",
  "id": "01a095d4-29bd-7a02-98c7-bd3b9c599d3a",
  "type": "user.password.changed",
  "source": "https://auth.example.com",
  "time": "2026-09-12T04:05:06Z",
  "datacontenttype": "application/json",
  "dataschema": "https://www.authelia.com/schemas/webhooks/v1/user.password.changed.json",
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

|     Attribute     |   Type   | Notes                                                                                                             |
| :---------------: | :------: | :---------------------------------------------------------------------------------------------------------------- |
|   `specversion`   | `string` | Always `1.0`.                                                                                                     |
|       `id`        | `string` | A UUIDv7. Generated once per occurrence and stable across delivery retries. See [Idempotency](#idempotency).      |
|      `type`       | `string` | One of the types in the [event catalogue](#event-catalogue).                                                      |
|     `source`      | `string` | The issuer URL of the Authelia instance. See [source](#source).                                                   |
|      `time`       | `string` | RFC 3339 timestamp of the occurrence, in UTC. Not the time of the delivery attempt.                               |
| `datacontenttype` | `string` | Always `application/json`.                                                                                        |
|   `dataschema`    | `string` | Always an `https://www.authelia.com/schemas/webhooks/` URI. See [Schemas](#schemas).                              |
| `autheliaversion` | `string` | A CloudEvents extension attribute carrying the emitting Authelia version.                                         |
|     `subject`     | `string` | The username the occurrence concerns. Omitted when there is no username, which happens more often than you think. |
|      `data`       | `object` | The payload, whose shape is determined by `type`.                                                                 |

The `autheliaversion` extension attribute is spelled without a separator because the CloudEvents specification restricts
extension attribute names to lowercase alphanumeric characters with a maximum length of twenty.

### source

The `source` attribute is the operator's own issuer URL. It is taken from the first configured
[session cookie domain](../../configuration/session/introduction.md#cookies), preferring its
[authelia_url](../../configuration/session/introduction.md#authelia_url) and falling back to `https://` plus its
[domain](../../configuration/session/introduction.md#domain).

A deployment with no session cookie domains configured emits an empty `source`. Do not use `source` to decide whether a
payload is genuine; use the [signature](#verifying-the-signature).

## Abuse protection

Before Authelia delivers any event to a destination it performs the [CloudEvents abuse protection] handshake, so that a
receiver is never sent an unsolicited payload.

[CloudEvents abuse protection]: https://github.com/cloudevents/spec/blob/main/cloudevents/http-webhook.md#4-abuse-protection

```text
OPTIONS /hooks HTTP/1.1
Host: admin.example.com
WebHook-Request-Origin: auth.example.com
WebHook-Request-Rate: *
```

A receiver confirms by answering a 2xx with `WebHook-Allowed-Origin` set to either the requested origin or an asterisk:

```text
HTTP/1.1 200 OK
WebHook-Allowed-Origin: auth.example.com
WebHook-Allowed-Rate: 120
```

|          Header          | Meaning                                                                            |
| :----------------------: | :--------------------------------------------------------------------------------- |
| `WebHook-Request-Origin` | The DNS name of the Authelia instance sending the events.                          |
|  `WebHook-Request-Rate`  | The rate Authelia requests, in requests per minute, or `*` for unrestricted.       |
| `WebHook-Allowed-Origin` | **Required in the response.** The requested origin, or `*`. Anything else refuses. |
|  `WebHook-Allowed-Rate`  | Optional. A ceiling in requests per minute, or `*`. Absent means unrestricted.     |

{{< callout context="danger" title="Required" icon="outline/alert-octagon" >}}
A destination which does not confirm receives _**no events at all**_. A 2xx without `WebHook-Allowed-Origin`, an origin
which is neither the requested one nor `*`, or any non-2xx status, are all refusals. The refusal is logged and Authelia
still starts, because webhook availability must never prevent it running.

An operator may skip the handshake with the destination's
[validation disable](../../configuration/miscellaneous/webhooks.md#disable) option.
{{< /callout >}}

The handshake carries the destination's configured authentication and static headers, so a receiver which authorizes
deliveries can authorize it the same way. It carries no event beyond `X-Authelia-Delivery-Attempt`, so none of the
other `X-Authelia-*` event headers are present. It is performed once per start, not periodically.

A transient failure is retried under the destination's [retry](../../configuration/miscellaneous/webhooks.md#retry)
policy, classified exactly as a delivery is. See [Retries and failures](#retries-and-failures) for the full table: a
`429` honors `Retry-After`, and a `5xx` or a failure which produced no status code at all backs off and is tried
again.

Three refusals are specific to the handshake and are never retried, because no retry can change them: a missing or
mismatched `WebHook-Allowed-Origin`, and a `WebHook-Allowed-Rate` which is invalid or which is absent after a rate was
requested.

Because this runs during startup the whole handshake for one destination is bounded, including its backoff, so a
receiver answering `429` with a long `Retry-After` cannot hold Authelia's start open. Destinations are validated
concurrently, so one slow receiver does not add its delay to the others.

If a receiver answers with `WebHook-Allowed-Rate`, Authelia paces its requests to that ceiling. Retries count against
it, because a retry is another request. It caps requests rather than events, so a
[batched](../../configuration/miscellaneous/webhooks.md#batch) destination carries many more events within the same
rate.

The asynchronous variant of the handshake, `WebHook-Request-Callback`, is not implemented: Authelia has no callback
endpoint for a receiver to confirm against out of band.

## Request

Each delivery is an HTTP `POST` of the envelope to the destination address.

|            Header             | Value                                                                               |
| :---------------------------: | :---------------------------------------------------------------------------------- |
|        `Content-Type`         | `application/cloudevents+json; charset=utf-8`                                       |
|         `User-Agent`          | `Authelia/` followed by the version                                                 |
| `X-Authelia-Webhook-Version`  | The contract version, currently `1`. Matches the `/v1/` segment of the schema URIs. |
|    `X-Authelia-Event-Type`    | The event type, identical to the envelope `type`.                                   |
|     `X-Authelia-Event-Id`     | The event identifier, identical to the envelope `id`.                               |
| `X-Authelia-Delivery-Attempt` | The attempt number of this request, starting at `1`.                                |
|    `X-Authelia-Signature`     | The HMAC signature. Present only when a signature secret is configured.             |

### Batched deliveries

A destination which configures [batching](../../configuration/miscellaneous/webhooks.md#batch) delivers several events
in one request instead. The body is a CloudEvents batched content mode array of the same envelopes, so each member
carries its own `type`, `id` and `dataschema`, and the request differs in three ways:

|            Header            | Value                                                                      |
| :--------------------------: | :------------------------------------------------------------------------- |
|        `Content-Type`        | `application/cloudevents-batch+json; charset=utf-8`                        |
|   `X-Authelia-Event-Count`   | The number of events in the array.                                         |
| `X-Authelia-Event-Type`/`Id` | **Absent.** They describe one event, so a batch carries the count instead. |

Branch on `Content-Type` rather than guessing from the body, and remember that a batching destination always sends an
array, even when it holds a single event. Every other header, the signature construction, and the redaction rules are
identical.

A batch is one request and so succeeds or fails as a unit. There is no way to accept part of one: a `2xx` acknowledges
every event in the array, and a permanent failure drops every event in it. Order is preserved within the array.

{{< callout context="caution" title="Ordering across formats" icon="outline/alert-triangle" >}}
A destination may list event types in its
[immediate](../../configuration/miscellaneous/webhooks.md#immediate) option, which are sent as individual deliveries
rather than waiting for a batch. Such a receiver sees both formats, and an immediate event may arrive **before**
batched events which occurred earlier. Order is guaranteed only within a batch. Sort on the envelope `time` attribute
if a total order matters.
{{< /callout >}}

Any [headers](../../configuration/miscellaneous/webhooks.md#headers) configured for the destination are also sent, as are
the `Authorization` or other header used by its
[authentication](../../configuration/miscellaneous/webhooks.md#authentication) options.

A receiver should respond with a `2xx` status code as soon as it has durably accepted the event, and do its own
processing afterwards. Authelia reads at most 4096 bytes of a response body and discards them; the body is never parsed
and nothing a receiver returns in it has any effect.

## Identifying an Authelia payload

Identification and authentication are different problems and this contract keeps them apart. There are three layers,
usable at different points in a receiver's processing:

1. **Before parsing**, the request headers, in particular `X-Authelia-Webhook-Version` and `User-Agent`. These are
   routing hints. Anyone can send them.
2. **After parsing**, the `dataschema` attribute. The cheapest reliable structural check is that its origin is
   `www.authelia.com` and its path carries the expected `/schemas/webhooks/v1/` prefix. This tells you what shape the
   payload claims to be.
3. **Cryptographically**, the `X-Authelia-Signature` header. This is the only layer which tells you the payload came
   from your Authelia instance.

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
The `dataschema` URI identifies the _**format**_ of a payload. It does not authenticate the _**sender**_. Anybody can
post a document containing that string to your receiver.

A receiver must _**never**_ treat the presence of an `authelia.com` schema URI, the `X-Authelia-*` headers, or the
`User-Agent` as authorization to act. Only the [signature](#verifying-the-signature) does that.
{{< /callout >}}

## Verifying the signature

When a [signature secret](../../configuration/miscellaneous/webhooks.md#secret) is configured, each request carries an
`X-Authelia-Signature` header of the following form:

```text
t=1757640306,v1=1f5a...c9
```

| Element | Meaning                                                                                                                    |
| :-----: | :------------------------------------------------------------------------------------------------------------------------- |
|   `t`   | The Unix timestamp in seconds at which this attempt was signed. Regenerated for each attempt, so it is not the event time. |
|  `v1`   | The lowercase hexadecimal HMAC digest of the signed input, under the signature construction named `v1`.                    |

The signed input is the timestamp, a full stop, then the raw request body bytes exactly as transmitted:

```text
HMAC-SHA256(secret, t + "." + rawBody)
```

The hash function is whichever the destination's
[algorithm](../../configuration/miscellaneous/webhooks.md#algorithm) option selects, either SHA-256 or SHA-512. The
header format is the same for both and the algorithm is not advertised in the request, so a receiver must be told which
one to expect out of band and must use that one when recomputing the digest. A receiver which assumes SHA-256 against a
destination configured for SHA-512 rejects every valid signature.

The timestamp is part of the signed input so that a captured request cannot be replayed indefinitely. The `v1` prefix
versions the construction so that another one can be added later without breaking receivers.

To verify:

1. Compute the HMAC over the **raw body bytes, before any parsing**. Re-serializing the parsed JSON produces a different
   byte sequence and the digest will not match.
2. Compare the digests in **constant time**. A naive string comparison leaks the digest one byte at a time.
3. Reject a request whose `t` is outside a tolerance window. Five minutes is a reasonable default.
4. Reject a request with no signature header when a secret is configured, rather than treating an absent signature as
   an unsigned but acceptable request.

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
        # A malformed header is not a valid signature. Never fall through to the comparison on one.
        return False

    expected = hmac.new(secret.encode(), f"{timestamp}.".encode() + body, digest).hexdigest()

    return hmac.compare_digest(expected, signature)
```

## Idempotency

The envelope `id` is generated once when the occurrence happens and is stable across every delivery attempt of that
event, so it doubles as an idempotency key. `X-Authelia-Event-Id` carries the same value, so a receiver can deduplicate
before it parses the body.

Delivery is best effort with retries. Duplicates are uncommon but they are possible and a receiver must be built to
expect them: one which durably accepts a request and then fails to respond within the destination
[timeout](../../configuration/miscellaneous/webhooks.md#timeout) is retried, and receives an event it has already
processed. Authelia cannot tell that case apart from one where the receiver never got the request.

Deduplicate on `id`, which is stable across every attempt of the same event.

The `id` is a UUIDv7, so its leading bits are the millisecond timestamp of the occurrence. Identifiers therefore sort
in occurrence order, and a receiver which stores them as a primary key gets sequential inserts rather than the random
ones a UUIDv4 would produce. Do not treat that ordering as authoritative: events from separate destinations arrive
independently, and the envelope `time` attribute is the occurrence time a receiver should reason about.

## Retries and failures

Each destination delivers serially and retries according to its own
[retry](../../configuration/miscellaneous/webhooks.md#retry) policy.

The table below governs the [abuse protection](#abuse-protection) handshake as well as deliveries, with one
difference: the handshake runs during startup, so its retries are additionally bounded by a budget which a long
`Retry-After` cannot exceed. A handshake which exhausts its attempts refuses the destination rather than dropping a
single event.

|      Response      | Behavior                                                                                                 |
| :----------------: | :------------------------------------------------------------------------------------------------------- |
|       `2xx`        | Delivered. No further attempts.                                                                          |
|       `429`        | Retried. The `Retry-After` header is honored when present.                                               |
|   `5xx` or more    | Retried using the backoff interval.                                                                      |
|  Transport error   | Retried using the backoff interval. Covers timeouts, connection failures, and TLS verification failures. |
| Any other response | **Permanent failure. The event is dropped and is never retried.**                                        |

The permanent failure case includes every `3xx` and every `4xx` other than `429`. Redirects are deliberately not
followed, so a receiver which answers with a redirect silently loses every event.

`Retry-After` is honored in both of its forms: the delta-seconds form, i.e. an integer number of seconds, and the
HTTP-date form. A date which has already passed, a negative number of seconds, and a value which is neither form are
all ignored, and the normal backoff interval is used instead. The requested wait is capped at
[maximum_interval](../../configuration/miscellaneous/webhooks.md#maximum_interval), so a receiver asking to be left
alone for a day is retried once that ceiling elapses instead. A `Retry-After` affects only the wait before the next
attempt; it does not extend the attempt budget.

The backoff interval starts at
[initial_interval](../../configuration/miscellaneous/webhooks.md#initial_interval), doubles before each subsequent
attempt, is capped at [maximum_interval](../../configuration/miscellaneous/webhooks.md#maximum_interval), and has up to
twenty percent negative jitter applied so that several destinations recovering from one outage do not retry in lockstep.

An event which exhausts its attempts is dropped and logged as an error. Events are never persisted, so nothing is
redelivered after a restart.

## Event volume

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
_**Every**_ authorization request which authenticates, and whose credential check is not served from the authentication
backend's password cache, emits a `security.authentication.succeeded` event. On a deployment where proxies present basic
authentication credentials on every request, that is potentially one webhook per HTTP request to every protected
application.

Scope the destination's [events](../../configuration/miscellaneous/webhooks.md#events) selectors to the types the
receiver actually acts on, rather than subscribing to `'*'` and filtering at the receiver. A destination whose queue
fills up drops events, and the fastest way to fill a queue is to subscribe to the highest volume type without needing
it.
{{< /callout >}}

## Event catalogue

There are thirteen registered event types.

| Type                                 | Notification | Description                                                                     |
| :----------------------------------- | :----------: | :------------------------------------------------------------------------------ |
| `user.password.changed`              |     yes      | A user changed their own password from the settings area.                       |
| `user.password.reset`                |     yes      | A user completed a password reset.                                              |
| `user.credential.totp.added`         |     yes      | A user registered a One-Time Password configuration.                            |
| `user.credential.totp.removed`       |     yes      | A user removed a One-Time Password configuration.                               |
| `user.credential.webauthn.added`     |     yes      | A user registered a WebAuthn credential.                                        |
| `user.credential.webauthn.removed`   |     yes      | A user removed a WebAuthn credential.                                           |
| `user.identity_verification.started` |     yes      | An identity verification was started and its email was sent.                    |
| `user.session.elevation.requested`   |     yes      | A session elevation was requested and its One-Time Code was sent.               |
| `security.authentication.succeeded`  |      no      | An authentication attempt succeeded. See [Event volume](#event-volume).         |
| `security.authentication.failed`     |      no      | An authentication attempt failed.                                               |
| `security.ban.applied`               |      no      | Regulation banned a user or an address.                                         |
| `security.ban.expired`               |      no      | **Never emitted.** See [below](#securitybanexpired-is-never-emitted).           |
| `system.startup_check`               |      no      | A connectivity probe, not an occurrence. See [below](#the-startup-check-probe). |

The eight notification bearing types cover the fixed set of notifications documented in
[Events](../../configuration/notifications/events.md). That page lists six notifications rather than eight, because its
"Second factor registered" and "Second factor removed" rows each fan out to a TOTP type and a WebAuthn type here. A
webhook event is emitted after the notification delivery attempt completes, so the `notification.sent` field is
accurate. When the notifier is disabled the attempt does not occur and the notification is marked
[suppressed](#a-suppressed-notification) instead.

An occurrence produces exactly one event. There is no separate `notification.*` namespace mirroring emails; the fact
that a user was also emailed is carried inside the event as a [notification object](#the-notification-object).

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
A password reset requested for a user which does not exist produces no notification and therefore no
`user.identity_verification.started` event, even though the portal responds identically to the requester. A receiver
cannot use these events to enumerate reset attempts against non-existent users.
{{< /callout >}}

## Common data fields

Every payload carries the same identifying fields.

|     Field      |    Type    | Presence                    | Notes                                                                          |
| :------------: | :--------: | :-------------------------- | :----------------------------------------------------------------------------- |
|   `username`   |  `string`  | Always present, may be `""` | The username the occurrence concerns. Mirrored into the envelope `subject`.    |
| `display_name` |  `string`  | Omitted when unavailable    | Present only on the notification bearing types, which resolve user details.    |
|    `emails`    | `[]string` | Omitted when unavailable    | Present only on the notification bearing types, which resolve user details.    |
|  `remote_ip`   |  `string`  | Omitted when unavailable    | The client address, as determined by the configured remote address resolution. |

The `security.*` types never carry `display_name` or `emails`, because those flows do not resolve user details. They
carry `username` and `remote_ip` only, and `username` is frequently empty.

### Presence versus value

Not every absent value is an absent key, and the difference matters to any receiver which branches on key presence.

These fields are _**always serialized**_, even when their value is the empty string or `false`, and are listed in
`required` by the published schemas:

- `username`, on every event type.
- `stage` and `method`, on the two `security.authentication.*` types.
- `target` and `target_type`, on the two `security.ban.*` types.
- `sent`, on the `notification` object.
- `email`, on each entry of `notification.recipients`.
- Every envelope attribute except `subject`.

Every other field is genuinely omitted from the JSON when it has no value, rather than sent as an empty string.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Test the _**value**_ of `username`, never its presence. `"username" in payload` is true even on the failure events which
name nobody, so a receiver written that way silently accepts an empty subject as a real one. See
[An empty username on a failure is normal](#an-empty-username-on-a-failure-is-normal).
{{< /callout >}}

## The notification object

The eight notification bearing types carry a `notification` object describing the email which the occurrence produced.

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

|    Field     |   Type    | Notes                                                                                     |
| :----------: | :-------: | :---------------------------------------------------------------------------------------- |
|    `sent`    | `boolean` | Whether the notification was successfully handed to the notifier.                         |
| `suppressed` | `boolean` | Present only when `true`. The notifier is disabled, so nothing was handed to it.          |
| `recipients` |  `array`  | The addressees. See below.                                                                |
|   `title`    | `string`  | The notification subject line, which is also the `{{ .Title }}` placeholder in templates. |
|   `error`    | `string`  | The delivery error. Present only when `sent` is `false` and `suppressed` is absent.       |
|   `values`   | `object`  | The template attribute values. See [Notification values](#notification-values).           |

### A suppressed notification

When the [notifier is disabled](../../configuration/notifications/introduction.md#disable) the notification is never
handed to a notifier, so `suppressed` is `true`, `sent` is `false`, and `error` is absent because nothing failed.
Distinguish the two cases by `suppressed` rather than by `sent` alone: `sent: false` with an `error` is a delivery
failure worth alerting on, while `sent: false` with `suppressed: true` is the configured behavior.

The `recipients` and `values` are still present, since they are the addresses which would have been notified and the
content which would have been sent. A deployment which disables the notifier depends on the webhook to carry them, so
note that credential equivalent values such as `one_time_code` and `link_url` remain omitted unless the destination
sets [disable_redaction](../../configuration/miscellaneous/webhooks.md#disable_redaction).

Each entry of `recipients` is an object rather than a bare address so a receiver can attribute an address to a person
without correlating against its own directory.

|     Field      |   Type   | Notes                                                          |
| :------------: | :------: | :------------------------------------------------------------- |
|    `email`     | `string` | Always present. The address the notification was addressed to. |
|   `username`   | `string` | Present when the address belongs to a known user.              |
| `display_name` | `string` | Present when recorded in the authentication backend.           |

For every event in the current catalogue the subject and the sole recipient are the same person, so these values
duplicate the payload's top level fields. They are kept separate so that a future notification addressed to somebody
other than its subject is not a breaking change.

The rendered HTML and text bodies of the email are _**never**_ included. They are produced inside the notifier and are
never handed to the event subsystem at all, so their exclusion is structural rather than a convention.

### Notification values

`values` carries the template attribute values of the email. Which of its fields are populated depends entirely on the
event type, because each emission site fills in only the values its own template uses.

| Type                                 | Populated `values` fields                                    |
| :----------------------------------- | :----------------------------------------------------------- |
| `user.password.changed`              | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `user.password.reset`                | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `user.credential.totp.added`         | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `user.credential.totp.removed`       | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `user.credential.webauthn.added`     | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `user.credential.webauthn.removed`   | `body_prefix`, `body_event`, `body_suffix`, `details`        |
| `user.identity_verification.started` | `domain`, `link_url`\*, `link_text`, `revocation_link_url`\* |
| `user.session.elevation.requested`   | `domain`, `one_time_code`\*, `revocation_link_url`\*         |

\* Redacted by default. See [Redaction](#redaction).

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The table above, not the published schema, is the description of what an event actually carries. `one_time_code`,
`link_url`, and `revocation_link_url` are permitted by _**all eight**_ notification bearing schemas, because they live on
a single shared structure, not because all eight events can produce them. See [Schemas](#schemas).
{{< /callout >}}

### Per-type additional fields

| Type                                 | Field         |   Type   | Notes                                                                                                                                                       |
| :----------------------------------- | :------------ | :------: | :---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `user.credential.*`                  | `description` | `string` | The user supplied description of the credential. Populated for WebAuthn credentials; empty for One-Time Password configurations, which have no description. |
| `user.identity_verification.started` | `action`      | `string` | The action the verification authorizes. The only value currently emitted is `ResetPassword`.                                                                |

## Redaction

Two payload types carry credential equivalent values. These fields are omitted entirely, not nulled and not masked,
unless the destination sets
[disable_redaction](../../configuration/miscellaneous/webhooks.md#disable_redaction).

| Type                                 | Redacted fields                        |
| :----------------------------------- | :------------------------------------- |
| `user.identity_verification.started` | `link_url`, `revocation_link_url`      |
| `user.session.elevation.requested`   | `one_time_code`, `revocation_link_url` |

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
A receiver configured with `disable_redaction` enabled holds values which complete a password reset or a session
elevation for the named user. Such a receiver is as sensitive as the session secret, and so is anything it forwards
those payloads to.
{{< /callout >}}

## Security events

### Authentication

`security.authentication.succeeded` and `security.authentication.failed` add the following fields.

|  Field   |   Type   | Notes                                                 |
| :------: | :------: | :---------------------------------------------------- |
| `stage`  | `string` | `first_factor` or `second_factor`.                    |
| `method` | `string` | `password`, `totp`, `webauthn`, or `duo`.             |
| `reason` | `string` | The failure classification. Present on failures only. |

A passkey authentication is reported as `stage` `first_factor` with `method` `webauthn`, because it replaces the
password entirely while the credential presented is a WebAuthn credential.

#### reason

|         Value         | Meaning                                                                                                      |
| :-------------------: | :----------------------------------------------------------------------------------------------------------- |
| `invalid_credentials` | The user or their authenticator rejected the attempt. A wrong password, a declined push, a failed assertion. |
|       `banned`        | The user or the address was already banned by regulation, so the attempt was refused before it was checked.  |
|   `internal_error`    | Any other error the raising site did not classify. See [internal_error is broad](#internal_error-is-broad).  |
|   `user_not_found`    | The authentication backend reported that the submitted username does not exist. See below.                   |

The `banned` value takes precedence over any other classification so that the reason always agrees with the banned
dimension recorded by the metrics. The full precedence order is `banned`, then a failure the raising site explicitly
marked as a user or authenticator rejection (`invalid_credentials`), then `user_not_found`, then `internal_error`, with
`invalid_credentials` as the fallback for a failure which carries no error at all.

{{< callout context="caution" title="user_not_found does not cover every missing user" icon="outline/alert-triangle" >}}
`user_not_found` is set only when the authentication backend returns its own "user not found" sentinel. It is
_**not**_ inferred from a username simply failing to resolve, so a path which discards the backend error, or where the
backend answers with something else, still classifies as `internal_error` or `invalid_credentials`.

It is emitted on the portal's first factor password sign-in, first factor re-authentication and second factor password
paths, and on the proxy authorization basic credential path. It is _**not**_ restricted to `first_factor`: the backend
returns the same sentinel for a user which is disabled as for one which does not exist, so a user disabled or deleted
while holding a live session fails the password check with `user_not_found` on a `second_factor` event, with the
username populated. It is not emitted on the passkey path, which does not pass the backend error to the classifier, and
it is not emitted when the backend returns no error and no user, which classifies as `invalid_credentials`.

The absence of `user_not_found` on an event therefore does not prove the username exists. Treat it as a positive signal
only.
{{< /callout >}}

An unreachable or failing backend does not produce this sentinel and still classifies as `internal_error`, so watching
`internal_error` as an infrastructure signal is not weakened by this classification.

#### internal_error is broad

`internal_error` is the classification given to any failure which carries an error that the site raising it did not
mark as a user or authenticator rejection. It therefore covers genuine system faults such as an unreachable or failing
authentication backend, but it is _**not**_ exclusively a system fault signal.

The clearly identifiable case of a username which does not exist is carved out as
[`user_not_found`](#reason), so it no longer lands here. What remains is every other error the raising site did not
label: a failing storage or session operation, a regulation lookup failure, a backend which answers with an error that
is not its not-found sentinel, and any future error added to one of these paths without a classification of its own.

Do not build a "our infrastructure is broken" alarm on `internal_error` alone. Correlate it with the presence or absence
of a `username`, with its rate relative to your normal traffic, and with the Authelia logs, which carry the underlying
error.

#### An empty username on a failure is normal

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
`security.authentication.failed` frequently carries an empty `username`, and consequently no envelope `subject`. The
`username` key itself is still present and its value is `""`, so branch on the value and not on the key; see
[Presence versus value](#presence-versus-value). The `remote_ip` field is populated on every real event.
{{< /callout >}}

This happens on two paths:

- **The unknown user first factor path.** When the authentication backend cannot resolve the submitted username, the
  username is never adopted, so the emitted event names nobody. Its `reason` is [`user_not_found`](#reason) when the
  backend reported the username as missing. This is precisely the case an operator watching for credential spraying
  against non-existent accounts cares about, so a receiver which drops or ignores events with no username will miss
  exactly the attack it was built to detect.
- **Most passkey failure paths.** A passkey assertion is evaluated before a user has been resolved, so a failure at
  almost any point before the final credential lookup has no username available to report.

Correlate these events on `remote_ip`, not on `username`.

Authentication _**successes**_ always carry a username, because success implies the user was resolved. The one
exception is the synthetic event produced by the
[startup check](../../configuration/miscellaneous/webhooks.md#startup_check), which carries the username
`startup-check` and _**no**_ `remote_ip`, because it does not originate from a request. It is the only event with a
username and no address.

### Bans

`security.ban.applied` and `security.ban.expired` add the following fields.

|     Field     |   Type   | Notes                                              |
| :-----------: | :------: | :------------------------------------------------- |
|   `target`    | `string` | The banned value, either a username or an address. |
| `target_type` | `string` | `user` or `ip`.                                    |
|   `expires`   | `string` | RFC 3339 timestamp. Present on `applied` only.     |

A `security.ban.applied` event with `target_type` of `ip` carries the address in both `target` and `remote_ip`, and its
`username` is present but empty, because an address is not a user. One with `target_type` of `user` carries the
username in both `target` and `username`, and the address which triggered it in `remote_ip`.

#### The startup check probe

When [startup_check](../../configuration/miscellaneous/webhooks.md#startup_check) is enabled, each destination
receives one `system.startup_check` event at startup to confirm it is reachable and that the credentials are accepted.

It is not an occurrence. No user did anything, and a receiver must not record it as an authentication, a security
signal, or an audit entry. Its `data` object carries `probe` set to `true` so it can be discarded without inspecting
the type. Acknowledge it with a `2xx` like any other delivery.

A failed probe is logged and startup continues regardless, because webhook availability must never prevent Authelia
from starting.

#### security.ban.expired is never emitted

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `security.ban.expired` type is registered, is accepted by the
[events](../../configuration/miscellaneous/webhooks.md#events) configuration option, and has a published schema, but
_**nothing in Authelia ever emits it**_. A destination subscribed to it will never receive anything.
{{< /callout >}}

There is no code path which observes a ban lapsing. Bans expire passively: the expiry is a column, and a ban stops
applying because a SQL predicate stops matching it, with no moment at which any Go code is notified. The one explicit
revocation which does exist lives in the `authelia storage` command line interface, which runs as a separate short lived
process with no dispatcher and therefore no ability to emit anything.

The type is retained in the vocabulary and keeps its published schema so that a receiver written against this contract
does not break if it is implemented later.

**What to do instead:** derive expiry from the `expires` field of `security.ban.applied`. It is an RFC 3339 timestamp of
when the ban stops applying, and it is known at the moment the ban is created. A receiver which wants to track active
bans should schedule its own expiry from that value rather than waiting for a message which will not arrive.

#### An IP ban names its subject differently by flow

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
When an attempt is refused because the _**address**_ is banned, the `username` on the resulting
`security.authentication.failed` event depends on which flow the attempt came through.
{{< /callout >}}

| Flow                                          |    `username`     | `remote_ip` |
| :-------------------------------------------- | :---------------: | :---------: |
| A sign-in at the portal from a banned address | The real username | The address |
| An authorization request banned by address    |       Empty       | The address |

The asymmetry is deliberate. The portal has already resolved the user before regulation refuses the attempt, and a known
user signing in from a banned address is exactly what an operator wants the event to name, so that username is kept.

The proxy authorization path has not resolved a user. All regulation hands it back is the banned value, which for an
address ban is the address itself. Putting an address into the `username` field would cause any receiver correlating on
subject to mint a user account named after an IP address, so no username is given at all. The address is carried in
`remote_ip` either way, which is the field to correlate on.

### Ordering

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
`security.ban.applied` is emitted _**before**_ the `security.authentication.failed` event for the attempt which
triggered the ban. Do not assume the opposite ordering.
{{< /callout >}}

The attempt is recorded with regulation first, and it is that recording which crosses the retry threshold and creates
the ban, so the ban event exists before the failure event is constructed. A receiver processing a batch in order will
therefore see a ban appear for a username or address a moment before it sees the failure which caused it.

A consequence worth knowing is that the failure event which triggered a ban is classified against the ban state as it
was _**before**_ the attempt, so its `reason` is `invalid_credentials` and not `banned`. The `banned` reason appears on
the _**subsequent**_ attempts which the now existing ban refuses.

Regulation only creates bans from first factor failures, so a `security.ban.applied` event is always immediately
followed, within the same attempt, by a `security.authentication.failed` event whose `stage` is `first_factor`.

Events are delivered to a given destination in the order they were emitted, so this ordering is stable rather than a
race. Across destinations there is no ordering relationship at all.

## Schemas

Fifteen [JSON Schema] documents are published: one per event type, one for the envelope, and one for the batched
array a batching destination sends.

```text
https://www.authelia.com/schemas/webhooks/v1/envelope.json
https://www.authelia.com/schemas/webhooks/v1/envelope-batch.json
https://www.authelia.com/schemas/webhooks/v1/<event type>.json
```

The `dataschema` attribute of each envelope is the URI of the schema for its own `data` object. The version segment
matches the `X-Authelia-Webhook-Version` header. A breaking change to a payload requires a new version segment and a
bump of that header.

### OpenAPI

The deliveries are also described in the `webhooks` section of the [OpenAPI](https://www.authelia.com/api/) document
Authelia serves at `/api/openapi.yml`, one entry per event type, with the request headers and the response codes a
receiver is expected to return. Each entry references the published schemas above rather than restating them, so a
generator which resolves remote references produces typed handlers, and one which does not still gets the headers and
the retry semantics.

That section is documented unconditionally. Its presence says nothing about whether this deployment has any
destinations configured.

{{< callout context="caution" title="The schemas are permissive, not exact" icon="outline/alert-triangle" >}}
A published schema describes what a payload is _**allowed**_ to contain, not what a given event type _**does**_ contain.
Validating against it tells you a payload is well formed. It does not tell you which fields to expect.

The clearest example is `one_time_code`, `link_url`, and `revocation_link_url`. These appear in all eight notification
bearing schemas because the notification values are a single shared structure reused by every one of those types, not
because a `user.password.changed` event can ever carry a One-Time Code. It cannot.

Use [Notification values](#notification-values) and [Per-type additional fields](#per-type-additional-fields) to decide
what a given type carries, and treat every field except those documented as always present as optional.
{{< /callout >}}

Most fields are optional in the schemas, because a value which is unavailable is omitted from the JSON rather than sent
as an empty string, so a receiver must tolerate any optional field being missing. The exceptions are the fields listed
under [Presence versus value](#presence-versus-value), which each schema marks `required` and which are always
serialized even when empty.

No schema forbids additional properties. A field added to a payload, and a CloudEvents extension attribute added to the
envelope, are additive changes which ship inside the current version segment, so a receiver must accept a document
containing a member it does not know and must not reject one as invalid on that basis. Validate for the members you
require, and ignore the rest.

[CloudEvents]: https://cloudevents.io/
[JSON Schema]: https://json-schema.org/
