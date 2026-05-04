---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Recovery Codes"
description: "Configuring the Recovery Codes second-factor fallback."
summary: "Authelia supports user-managed single-use recovery codes that can be entered at sign-in if the user has lost access to their primary second-factor device."
date: 2026-09-10T21:07:03+10:00
draft: false
images: []
weight: 103400
toc: true
seo:
  title: ""
  description: ""
  canonical: ""
  noindex: false
---

Recovery codes are user-managed single-use codes that act as a fallback at the second-factor step. A user generates a
batch of codes from the _Settings > Two-Factor Authentication_ page, saves them somewhere safe, and can then enter one
of them at sign-in if they have lost access to their primary 2FA device (TOTP authenticator, WebAuthn key, Duo).

## Configuration

```yaml {title="configuration.yml"}
recovery_codes:
  disable: false
```

### disable

{{< confkey type="boolean" default="false" required="no" >}}

Setting this to `true` disables the recovery codes feature for all users: the _Settings_ panel, the sign-in
fallback link, and the related API endpoints all become unavailable.

## Behavior

- Each generation produces 10 codes; each code is 10 alphanumeric characters from an unambiguous uppercase alphabet
  (no `0/O/I/L/1/5`), formatted as `XXXXX-XXXXX` for readability.
- Codes are normalized server-side before lookup (uppercased; whitespace, hyphens, and underscores stripped) so users
  can enter them with or without the visual hyphen.
- Storage is one-way: only an HMAC-SHA256 signature of each code is persisted, peppered with a per-deployment key
  managed alongside the existing one-time-code pepper. A database leak does not yield usable codes.
- On consumption a code is marked consumed (timestamp + remote IP) rather than deleted, so the audit trail is preserved
  for incident review. Regenerating revokes the prior batch the same way.
- Successful recovery-code authentication bumps the session to the _TwoFactor_ level but does NOT mint an elevated
  session. Re-enrolling a primary 2FA device still requires the existing email-based identity verification flow.
- Failed recovery-code attempts are logged through the regulation system under a separate `RecoveryCode` auth type,
  so they do not consume the password or TOTP rate-limit budgets and a recovery-code spray cannot lock the user out
  of normal password sign-in.
- Generation triggers a notification email to the user; a regenerated set logs the new batch event. If the SMTP delivery
  fails, the response surfaces `notification_sent: false` so the UI can warn the user.

## Operator runbook

### Backup and restore

The HMAC pepper for recovery codes lives in the `encryption` table as `hmac_key_rc`, alongside `hmac_key_otc` and
`hmac_key_otp`. A standard storage backup carries it. As long as the master `storage.encryption_key` on the restored
deployment matches the backup, recovery code signatures continue to validate.

If the master encryption key changes (for example, restoring a backup into a new deployment with a different key), all
HMAC peppers in the `encryption` table fail to decrypt at startup. Recovery codes, one-time codes, and any other
HMAC-keyed feature stop validating until you either:

1. Restore the original `storage.encryption_key` and bring the deployment back up; OR
2. Run `authelia storage encryption change-key` to change the master key correctly through the supported flow.

### Pepper rotation

`authelia storage encryption rotate-hmac-key --name rc` rotates the HMAC pepper for recovery codes. The rotation
truncates the `recovery_codes` table because the existing signatures cannot be regenerated without the plaintext
codes. All users will need to regenerate their recovery codes from the Settings page after rotation. Use this command
sparingly and communicate to your users in advance.

### Manual administration

The CLI offers per-user administration:

```shell
authelia storage user recovery-codes status   <username>   # counts and timestamps
authelia storage user recovery-codes list     <username>   # row-level audit (no plaintext)
authelia storage user recovery-codes generate <username>   # forces regenerate; prints codes once
authelia storage user recovery-codes delete   <username>   # soft-revoke all
```

The `generate` subcommand is the supported way to rescue a user who has lost both their primary 2FA device and
their recovery codes: the operator runs it, captures the printed codes, and delivers them to the user through a
trusted channel.

## Metrics

When telemetry is enabled (`telemetry.metrics.enabled: true`) the recovery codes feature exposes the following
Prometheus series:

- `authelia_recovery_codes_generated_total` (counter): cumulative codes generated since process start.
- `authelia_recovery_codes_users_with_codes` (gauge): users with at least one unused code.
- `authelia_recovery_codes_users_low` (gauge): users with 1 or 2 unused codes.
- `authelia_recovery_codes_users_depleted` (gauge): users with zero unused codes but at least one consumed or revoked
  code (i.e. they had codes generated and used or replaced them all).

Recovery code sign-in attempts are also recorded under the existing `authelia_authn_second_factor` counter with the
`type="RecoveryCode"` label.
