---
title: "Protecting SSH and Console Logins with pam_authelia"
description: "Install and configure pam_authelia to delegate PAM-based authentication (SSH, login, sudo) to Authelia, with 1FA, 2FA, and OAuth2 Device Authorization flows."
summary: "End-to-end guide for installing the pam_authelia PAM module, wiring sshd through it, and configuring each supported flow: password only, TOTP, Duo push, and the RFC 8628 Device Authorization grant."
date: 2026-04-15T10:00:00+11:00
draft: false
images: []
weight: 560
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

# Introduction

[pam_authelia] is a [PAM] module that lets anything authenticating through PAM (most commonly [OpenSSH], but also `login`, `su`, `sudo`, and any other PAM consumer) delegate credential verification and two-factor challenges to an [Authelia] server over its existing HTTP API. Nothing on the server side needs to change: [pam_authelia] calls the same `/api/firstfactor`, `/api/user/info`, `/api/secondfactor/totp`, `/api/secondfactor/duo`, `/api/oidc/device-authorization`, and `/api/oidc/token` endpoints that the Authelia web portal uses.

The module ships as two artifacts that cooperate over a stdin/stdout pipe protocol:

- `pam_authelia.so`: a small C shim loaded into the PAM consumer's process (e.g. `sshd`). It handles the PAM conversation function (`pam_conv`) for prompting the user and securely wiping credentials from memory, and delegates everything else to the Go helper.
- `pam_authelia`: a Go helper binary that handles every HTTPS request to Authelia, parses responses, orchestrates the 2FA flow, and renders QR codes for the OAuth2 Device Authorization grant.

The split exists because a CGO-based single-binary PAM module would pull the Go runtime into every `sshd` preauth child, and because a clean process boundary simplifies fork safety and credential zeroisation. Operators do not need to understand the protocol to use the module.

This guide walks through installing [pam_authelia], wiring `sshd` through it, and configuring each supported flow.

# Assumptions

This guide makes the following assumptions:

- [Authelia] is already set up, running, and reachable over HTTPS from the host where you plan to install [pam_authelia]. If you are using a self-signed certificate the CA certificate must be available on that host (see the [`ca-cert`](#ca-cert) option).
- The user you plan to authenticate already exists in Authelia's authentication backend and has TOTP, Duo, and/or a Device Authorization OIDC client enrolled, depending on which 2FA flow you plan to use.
- You have `root` on the host running the PAM consumer and can edit `/etc/pam.d/*` and `/etc/ssh/sshd_config` and reload `sshd`.

# How it works

The flow for a single SSH login looks like this:

```text
sshd ──▶ PAM ──▶ pam_authelia.so ──fork+exec──▶ pam_authelia (Go) ──HTTPS──▶ Authelia
  ▲                     │                              │
  │  pam_conv prompts ──┘                              │
  └─────────────────────────────── SUCCESS / FAILURE ──┘
```

1. `sshd` accepts the TCP connection and hands the authentication over to PAM per its `/etc/pam.d/sshd` stack.
2. PAM loads `pam_authelia.so`, which forks the Go helper (`pam_authelia`) with the PAM module options passed as CLI flags.
3. Username and password are sent to the Go helper over a pipe. For `auth-level=2FA` the password is taken from `PAM_AUTHTOK` (set by a preceding `pam_unix` entry) so the user is never prompted by [pam_authelia] for it.
4. The Go helper calls Authelia (`/api/firstfactor`, then optionally `/api/user/info` and one of the `/api/secondfactor/*` or `/api/oidc/*` endpoints) and writes prompt/info/success/failure commands back to the C shim, which surfaces them to the SSH client through `pam_conv`.
5. For the Device Authorization flow the helper additionally runs OIDC discovery against the configured Authelia URL and calls `/userinfo` with the issued access token, to verify the approved identity matches the Linux username. See [Device Authorization identity binding](#device-authorization-identity-binding) for the full check list.
6. On success the shim returns `PAM_SUCCESS` to `sshd`; on failure it returns `PAM_AUTH_ERR`.

The Go helper never writes credentials to logs, zeroes them from memory after use, and enforces HTTPS-only communication with Authelia.

# Installation

[pam_authelia] is released from [github.com/authelia/pam](https://github.com/authelia/pam) as `.deb`, glibc tarball, and musl tarball artifacts for `amd64`, `arm`, and `arm64`. Checksums, SBOMs, and GPG signatures are published alongside every release.

Regardless of which channel you use, two files get installed:

|               File               |                    Destination                    |
|:--------------------------------:|:--------------------------------------------------:|
|         `pam_authelia`           |                `/usr/bin/pam_authelia`             |
|        `pam_authelia.so`         | `/lib/security/pam_authelia.so` (distro-dependent) |

The PAM module directory differs between distributions (Debian uses `/lib/x86_64-linux-gnu/security/`, Alpine uses `/lib/security/`, Arch uses `/usr/lib/security/`), so if you are installing manually you may need to locate the directory containing `pam_unix.so` and install `pam_authelia.so` alongside it. The packaged installation methods below handle this automatically.

## Debian and Ubuntu

The preferred method is to install from the [Authelia APT repository], which publishes signed packages for both Authelia itself and [pam_authelia]. If you have already added the repository to your system you can install with a single command:

```bash
sudo apt update && sudo apt install pam_authelia
```

If you have not added the repository yet, follow the [APT Repository setup steps][Authelia APT repository] first and then run the command above. The repository is signed with Authelia's release key and handles upgrades automatically, so you do not need to download `.deb` files manually.

Alternatively, you can download a specific `.deb` release directly from [github.com/authelia/pam](https://github.com/authelia/pam) and install it with `apt install ./<file>.deb`; useful for pinning a particular version or for hosts that cannot reach the APT repository.

## Arch Linux

Three community packages are maintained in the [Arch Linux AUR](https://aur.archlinux.org/), covering every preference:

|      Package       |                           What it installs                            |                When to pick it                 |
|:------------------:|:----------------------------------------------------------------------:|:-----------------------------------------------:|
|  `pam_authelia`    |   Builds from the latest tagged release tarball on your build host    |       Preferred for reproducible builds        |
| `pam_authelia-bin` |         Installs the prebuilt upstream binary artifact as-is          |    Fastest install, no local toolchain needed    |
| `pam_authelia-git` |          Tracks the `master` branch of [github.com/authelia/pam](https://github.com/authelia/pam)       | Following unreleased changes or testing patches |

Install whichever suits you using your AUR helper of choice, for example with [`paru`](https://github.com/Morganamilo/paru):

```bash
paru -S pam_authelia-bin
```

or with [`yay`](https://github.com/Jguer/yay):

```bash
yay -S pam_authelia-bin
```

All three packages install the same file layout, so the PAM configuration examples later in this guide apply unchanged.

## Alpine Linux

Download the `-musl` tarball from the [github.com/authelia/pam](https://github.com/authelia/pam) releases page and extract the two files into place:

```bash
curl -LO https://github.com/authelia/pam/releases/latest/download/pam_authelia-v0.1.0-linux-amd64-musl.tar.gz
tar -xzf pam_authelia-v0.1.0-linux-amd64-musl.tar.gz
sudo install -m 0755 pam_authelia /usr/bin/pam_authelia
sudo install -m 0644 pam_authelia.so /lib/security/pam_authelia.so
```

## Other Linux distributions (generic tarball)

For glibc distributions without a native package, use the `glibc` tarball instead of `-musl`:

```bash
curl -LO https://github.com/authelia/pam/releases/latest/download/pam_authelia-v0.1.0-linux-amd64.tar.gz
tar -xzf pam_authelia-v0.1.0-linux-amd64.tar.gz
sudo install -m 0755 pam_authelia /usr/bin/pam_authelia
sudo install -m 0644 pam_authelia.so "$(dirname "$(find /lib /usr/lib -name pam_unix.so -print -quit)")/pam_authelia.so"
```

The `find | dirname` dance picks up whichever PAM module directory your distribution uses by locating the well-known `pam_unix.so` file.

## Building from source

If none of the packaged install channels above fit your platform, both artifacts can be built directly from the [github.com/authelia/pam](https://github.com/authelia/pam) repository. You will need:

- [Go] `1.26` or newer
- `gcc` (or any C11-capable compiler that understands `-fstack-protector-strong` and `-D_FORTIFY_SOURCE=3`)
- `make`
- `libpam` development headers: `libpam0g-dev` on Debian/Ubuntu, `linux-pam-dev` on Alpine, `pam` is included in the base system on Arch

Clone the repository:

```bash
git clone https://github.com/authelia/pam.git
cd pam
```

Build the Go helper binary. The flags match Authelia's own release build: `-trimpath` strips local paths from the binary, `-ldflags '-s -w'` strips the symbol table and DWARF debug information for a smaller binary. `CGO_ENABLED=0` is deliberate; the Go helper does not link against libc and is safe to build as a static binary:

```bash
CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o pam_authelia ./cmd/pam_authelia
```

Build the C shim. The `shim/Makefile` handles the hardening flags for you (`-fstack-protector-strong`, `-D_FORTIFY_SOURCE=3`, full RELRO, `-z now`, `-fPIC`, `-fno-plt` on Linux) and detects `.so` vs `.dylib` based on the host platform:

```bash
make -C shim
```

Install both artifacts. The `pam_authelia.so` destination depends on your distribution; use `find` to locate the directory that already contains `pam_unix.so`:

```bash
sudo install -m 0755 pam_authelia /usr/bin/pam_authelia
sudo install -m 0644 shim/pam_authelia.so \
    "$(dirname "$(find /lib /usr/lib -name pam_unix.so -print -quit)")/pam_authelia.so"
```

Confirm both files are in place and have the expected modes before adding `pam_authelia.so` to your PAM stack:

```bash
ls -l /usr/bin/pam_authelia
ls -l "$(dirname "$(find /lib /usr/lib -name pam_unix.so -print -quit)")/pam_authelia.so"
```

# SSH server prerequisites

[pam_authelia] uses the PAM keyboard-interactive conversation to prompt for passwords and 2FA codes, so `sshd` must be configured to use PAM and to permit keyboard-interactive authentication. The minimum `/etc/ssh/sshd_config` looks like this:

```text {title="/etc/ssh/sshd_config"}
UsePAM yes
KbdInteractiveAuthentication yes
PasswordAuthentication no
AuthenticationMethods keyboard-interactive
```

If you plan to use the [OAuth2 Device Authorization flow](#configuring-authelia-for-the-device-authorization-flow) you may want to consider `LoginGraceTime`. The default of 2 minutes is usually enough for users to scan the QR code and approve on their phone, but if you see logins timing out while users are still mid-approval you can raise it:

```text {title="/etc/ssh/sshd_config"}
LoginGraceTime 5m
```

Reload `sshd` after editing the file:

```bash
sudo systemctl reload sshd
```

# PAM module options

The following options can be supplied to `pam_authelia.so` in any PAM stack file (commonly `/etc/pam.d/sshd`). Every option is a `key=value` pair except for the boolean `debug` flag which takes no value. Option names are case-sensitive and use kebab-case.

The `required` badge on each option below uses one of three values, matching the convention used elsewhere in the Authelia documentation:

- __`yes`__: the option must be set; the module will refuse to authenticate without it.
- __`no`__: optional; the shown default applies when the option is omitted.
- __`situational`__: required only under specific configurations (for example [`oauth2-client-id`](#oauth2-client-id) is required when [`method-priority`](#method-priority) contains `device_authorization`, otherwise it is ignored).

## Options

### url

{{< confkey type="string" required="yes" >}}

The URL of the Authelia server. Must use the `https://` scheme. This is the base URL the Go helper uses for every API call, for example POSTing to `/api/firstfactor`.

__Example:__

```text {title="/etc/pam.d/sshd"}
auth required pam_authelia.so url=https://auth.example.com auth-level=1FA+2FA
```

### auth-level

{{< confkey type="string" default="1FA+2FA" required="no" >}}

The authentication level to enforce. Must be one of `1FA`, `2FA`, or `1FA+2FA` (case-sensitive). Each level is described in full under [Authentication flows](#authentication-flows):

- `1FA`: password only, validated against Authelia's first-factor endpoint.
- `2FA`: the password is read from `PAM_AUTHTOK` (set by a preceding module such as `pam_unix.so`), Authelia is queried silently for 1FA, and the user is prompted for the second factor.
- `1FA+2FA`: the user is prompted for a password, and upon success is prompted for the second factor.

### cookie-name

{{< confkey type="string" default="authelia_session" required="no" >}}

The name of the session cookie Authelia issues on successful 1FA. Must match the server-side [`session.cookies[].name`](../../../configuration/session/introduction.md#name) value in your Authelia configuration. Only change this if your Authelia deployment has a non-default session cookie name.

### ca-cert

{{< confkey type="string" required="no" >}}

Path to a custom CA certificate (PEM-encoded) used to verify Authelia's TLS certificate. Defaults to the system trust store. Use this when Authelia is served behind a private CA:

```text {title="/etc/pam.d/sshd"}
auth required pam_authelia.so url=https://auth.internal ca-cert=/etc/ssl/certs/internal-ca.pem
```

The file must be readable by the user `sshd` drops privileges to during authentication (typically `root` for the PAM preauth child).

### timeout

{{< confkey type="integer" default="60" required="no" >}}

Upper bound in seconds on the entire PAM exchange, including time spent waiting for user input and for Authelia responses. When the timeout fires the C shim kills the Go helper and returns `PAM_AUTH_ERR` to `sshd`.

The default of 60 seconds is comfortable for password and TOTP flows and usually sufficient for the [Device Authorization flow](#configuring-authelia-for-the-device-authorization-flow) too, especially if users approve on a phone they already have to hand.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
If you see `sshd` aborting Device Authorization logins before the Go helper has finished polling (for example because users take longer than 60 seconds to find their phone before even starting the approval), raise this option on the `pam_authelia.so` line in `/etc/pam.d/sshd`:

```text {title="/etc/pam.d/sshd"}
auth required pam_authelia.so url=https://auth.example.com timeout=300 \
    method-priority=device_authorization oauth2-client-id=pam-authelia
```

This option only governs the PAM-side exchange. It is unrelated to Authelia's own device code expiry (configured server-side via [`identity_providers.oidc.lifespans.device_code`](../../../configuration/identity-providers/openid-connect/provider.md#device_code)). If your users hit `device authorization token expired` that's an Authelia-side timeout and raising this PAM option will not help. See [Troubleshooting](#troubleshooting) for how to handle that case.
{{< /callout >}}

### binary

{{< confkey type="string" default="/usr/bin/pam_authelia" required="no" >}}

Absolute path to the `pam_authelia` Go helper binary. Override only if you installed the binary somewhere non-standard (for example when building from source and installing under `/opt/` or `/usr/local/bin/`).

### method-priority

{{< confkey type="string" required="no" >}}

A comma-separated list of 2FA method identifiers the module should try, in order. Valid entries are `totp`, `mobile_push`, `device_authorization`, and the special `user` keyword. The first entry whose method is usable for the current user is selected; if none match, authentication fails.

When this option is omitted the module uses whichever 2FA method Authelia has stored as the user's preference. See [Method priority and the `user` entry](#method-priority-and-the-user-entry) for worked examples.

### oauth2-client-id

{{< confkey type="string" required="situational" >}}

OAuth2 client ID for the Device Authorization grant. __Required__ when [`method-priority`](#method-priority) contains `device_authorization`; ignored otherwise. Must match a client configured on the Authelia side with `grant_types: ['urn:ietf:params:oauth:grant-type:device_code']`. See [Configuring Authelia for the Device Authorization flow](#configuring-authelia-for-the-device-authorization-flow) for the server-side setup.

### oauth2-client-secret

{{< confkey type="string" required="situational" >}}

OAuth2 client secret. __Required__ when the client referenced by [`oauth2-client-id`](#oauth2-client-id) is a confidential client (i.e. configured with `token_endpoint_auth_method: 'client_secret_post'` on the Authelia side). Omit this for public clients.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
This secret appears in cleartext in `/etc/pam.d/*` and will be visible to anyone with read access to those files. On most distributions `/etc/pam.d/*` is already `0644` and owned by `root`, but verify that the file is not world-readable if you consider the device-flow client secret sensitive, or use a public client (no secret) instead.
{{< /callout >}}

### oauth2-scope

{{< confkey type="string" default="openid,authelia.pam" required="no" >}}

Comma-separated OAuth2 scopes to request on the Device Authorization endpoint. The module normalizes the comma-separated form into the space-separated form required by [RFC 6749] before sending the HTTP request. Only relevant when the Device Authorization flow is enabled.

Both `openid` and `authelia.pam` are **mandatory** and are enforced at config parse time; the Go helper refuses to start if either is missing. `openid` is required so Authelia issues an ID token the helper can verify; `authelia.pam` is the custom scope that grants the `authelia.pam.username` claim used to bind the issued token to the Linux username the PAM module is authenticating. You can append additional scopes (for example `openid,authelia.pam,email`) without breaking this contract, but you cannot drop either of the two required ones. See [Device Authorization identity binding](#device-authorization-identity-binding) for the full rationale and the server-side `claims_policies` and custom-scope configuration that produces the claim.

### debug

{{< confkey type="boolean" default="false" required="no" >}}

A boolean flag with no value; its presence enables debug logging. Diagnostic lines are written to `stderr`, which `sshd` captures in its journal (see [Troubleshooting](#troubleshooting) for the exact log format and how to read it).

# Authentication flows

[pam_authelia] supports three authentication levels, controlled by the [`auth-level`](#auth-level) option, and four 2FA methods, controlled by [`method-priority`](#method-priority). The three levels are:

## `1FA`: password only

The user is prompted for a password. The module POSTs `{"username": "...", "password": "..."}` to `/api/firstfactor` and grants the login on HTTP 200 with `status: OK`. No second factor is ever attempted; this mode is only useful when Authelia is acting as a centralized password store and you do not want two-factor enforcement on PAM logins.

## `2FA`: password from PAM stack, then second factor

The password is taken from `PAM_AUTHTOK`, which must be populated by a preceding module such as `pam_unix.so`:

```text {title="/etc/pam.d/sshd"}
auth required pam_unix.so
auth required pam_authelia.so url=https://auth.example.com auth-level=2FA
```

[pam_authelia] silently POSTs the same credentials to `/api/firstfactor` (the user is not re-prompted), and upon success prompts for the second factor. This mode is useful when local Unix passwords are the source of truth and Authelia is only consulted for the second factor.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Because the password captured by `pam_unix.so` is forwarded verbatim to Authelia's first-factor endpoint, the user's __local Unix password must match their Authelia password__. If the two drift out of sync the silent 1FA call to Authelia will fail and the login will be rejected even though `pam_unix.so` already accepted the password. Operators running this mode should either provision the same password in both places when the account is created, or use `1FA+2FA` (described below) instead, which prompts the user once and validates only against Authelia.
{{< /callout >}}

## `1FA+2FA`: password then second factor

The most common deployment. The user is prompted for their password, the module validates it against `/api/firstfactor`, and then prompts for the second factor:

```text {title="/etc/pam.d/sshd"}
auth required pam_authelia.so url=https://auth.example.com auth-level=1FA+2FA
```

## Second factor method selection

For `2FA` and `1FA+2FA`, the module fetches `/api/user/info` to discover which 2FA methods the user has enrolled, then picks one according to [`method-priority`](#method-priority):

|        Method          |                   Endpoint                   |                          User interaction                          |
|:----------------------:|:---------------------------------------------:|:-------------------------------------------------------------------:|
|          TOTP          |         `POST /api/secondfactor/totp`         |              Types a 6- or 8-digit code from an authenticator app           |
|        Duo push        |         `POST /api/secondfactor/duo`          |                      Approves the push on their phone                      |
| Device Authorization   | `POST /api/oidc/device-authorization` (setup) | Scans the QR code or visits the verification URL, completes Authelia login (1FA and 2FA if required), approves the consent prompt, presses Enter |
|                        |         `POST /api/oidc/token` (poll)         |                                                                     |

{{< callout context="caution" title="WebAuthn over SSH" icon="outline/alert-triangle" >}}
[pam_authelia](https://github.com/authelia/pam) cannot drive the direct `/api/secondfactor/webauthn` flow, because FIDO2 authenticators need USB or NFC access to the client host and the SSH keyboard-interactive channel cannot pass an authenticator ceremony through. Behavior depends on what else the user has enrolled and on the `method-priority` setting:

- __WebAuthn is the user's Authelia preference but they also have TOTP or Duo enrolled__, and [`method-priority`](#method-priority) is unset or contains `user` (the default), [pam_authelia](https://github.com/authelia/pam) automatically falls through to TOTP, then Duo, then Device Authorization, and authenticates via the first usable method. No operator intervention needed.
- __WebAuthn is the user's only enrolled method__: the direct 2FA path fails, but WebAuthn still works via the [Device Authorization flow](#configuring-authelia-for-the-device-authorization-flow). At the verification URL the user is logging in through a real browser at the Authelia portal, so WebAuthn (and any other 2FA method Authelia supports) works normally there. Configure `method-priority=device_authorization` or `method-priority=device_authorization,user` on the PAM stack.
- __[`method-priority`](#method-priority) is set to an explicit list that excludes both `user` and `device_authorization`__ (for example `method-priority=totp` when the user has only WebAuthn enrolled): authentication fails with `no usable 2FA method for this user`. Either enroll an additional method on the Authelia side or widen the priority list so the module can fall through.
{{< /callout >}}

# Method priority and the `user` entry

When [`method-priority`](#method-priority) is omitted, [pam_authelia] uses whichever 2FA method the user has marked as preferred in Authelia. For most deployments this is the right behavior. For cases where you want the PAM stack to enforce a specific 2FA flow regardless of the user's preference (for example, "always use the Device Authorization flow on servers in this fleet"), use an explicit priority list.

A priority list is a comma-separated list of method identifiers. The module walks the list top-to-bottom and uses the first one that resolves to a usable method for the current user. Valid entries are:

- `totp`: use TOTP if the user has it enrolled.
- `mobile_push`: use a Duo push if the user has Duo enrolled.
- `device_authorization`: use the OAuth2 Device Authorization grant. Requires [`oauth2-client-id`](#oauth2-client-id) to be set.
- `user`: a special entry that resolves to the user's Authelia preference at runtime. If that preference is WebAuthn (unsupported over SSH) or empty, the module falls back through TOTP, Duo, and Device Authorization in that order.

Worked examples:

|               Priority list              |                                            Behavior                                            |
|:-----------------------------------------:|:------------------------------------------------------------------------------------------------:|
|                  `totp`                   |              Always TOTP; fail if the user has not enrolled TOTP.              |
|          `totp,mobile_push,user`          |        Prefer TOTP, then Duo push, then fall back to whatever Authelia stores as the preference.       |
|        `device_authorization,user`        |         Prefer the Device Authorization flow, fall back to the user's stored preference.          |
|                  `user`                   |           Always respect the user's Authelia preference (identical to the default behavior).          |

# Example PAM configurations

The following examples all target `/etc/pam.d/sshd`. Replace the `url=` hostname with your Authelia deployment. Each example is self-contained and can be used unmodified (except for `url=`).

## 1FA only

```text {title="/etc/pam.d/sshd"}
auth required pam_authelia.so url=https://auth.example.com auth-level=1FA
account required pam_permit.so
session required pam_permit.so
```

## 1FA+2FA (recommended default)

```text {title="/etc/pam.d/sshd"}
auth required pam_authelia.so url=https://auth.example.com auth-level=1FA+2FA
account required pam_permit.so
session required pam_permit.so
```

## 2FA only (local password + Authelia second factor)

```text {title="/etc/pam.d/sshd"}
auth required pam_unix.so
auth required pam_authelia.so url=https://auth.example.com auth-level=2FA
account required pam_permit.so
session required pam_permit.so
```

## Device Authorization flow

```text {title="/etc/pam.d/sshd"}
auth required pam_authelia.so url=https://auth.example.com \
    auth-level=1FA+2FA \
    method-priority=device_authorization,user \
    oauth2-client-id=pam-authelia \
    oauth2-client-secret=hashed-secret-here \
    oauth2-scope=openid,authelia.pam \
    timeout=300
account required pam_permit.so
session required pam_permit.so
```

The `timeout=300` value gives the user five minutes to approve on their phone before the PAM exchange is torn down. Both scopes (`openid` and `authelia.pam`) are required; see [`oauth2-scope`](#oauth2-scope) for why, and [Configuring Authelia for the Device Authorization flow](#configuring-authelia-for-the-device-authorization-flow) for the matching server-side setup.

## Custom CA (self-signed Authelia)

```text {title="/etc/pam.d/sshd"}
auth required pam_authelia.so url=https://auth.internal \
    auth-level=1FA+2FA \
    ca-cert=/etc/ssl/certs/internal-ca.pem
account required pam_permit.so
session required pam_permit.so
```

# Configuring Authelia for the Device Authorization flow

The Device Authorization flow is the only 2FA method that requires additional server-side configuration. Beyond registering an OIDC client with the `urn:ietf:params:oauth:grant-type:device_code` grant type, you must also define a [claims policy](../../../configuration/identity-providers/openid-connect/provider.md#claims_policies) that emits the `authelia.pam.username` claim and a custom [OIDC scope](../../../configuration/identity-providers/openid-connect/provider.md#scopes) (`authelia.pam`) that grants it. [pam_authelia] uses the claim to bind the issued token to the Linux username it is authenticating. Without these two pieces the device flow aborts at the new identity-binding check with `claim "authelia.pam.username" missing from userinfo response`. See [Device Authorization identity binding](#device-authorization-identity-binding) for the full rationale.

Add the following blocks to your Authelia configuration:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    claims_policies:
      pam:
        custom_claims:
          authelia.pam.username:
            attribute: 'username'
    scopes:
      authelia.pam:
        claims:
          - 'authelia.pam.username'
    clients:
      - client_id: 'pam-authelia'
        client_name: 'pam_authelia device flow'
        client_secret: '$pbkdf2-sha512$310000$...'
        public: false
        authorization_policy: 'two_factor'
        grant_types:
          - 'urn:ietf:params:oauth:grant-type:device_code'
        token_endpoint_auth_method: 'client_secret_post'
        claims_policy: 'pam'
        scopes:
          - 'openid'
          - 'authelia.pam'
```

What each block does:

- __`claims_policies.pam`__: a reusable policy named `pam` with a single custom claim, `authelia.pam.username`, whose value is sourced from the backend's `username` attribute. If your backend's raw username doesn't match the Linux account verbatim (different case, an `@realm` suffix, etc.) anchor the claim at a derived attribute instead; see [Case sensitivity and username normalization](#case-sensitivity-and-username-normalization).
- __`scopes.authelia.pam`__: a custom OIDC scope named `authelia.pam` that grants the `authelia.pam.username` claim. The name is arbitrary but must match whatever you send via the PAM [`oauth2-scope`](#oauth2-scope) option; `authelia.pam` is the default the Go helper expects.
- __Client-level `claims_policy: 'pam'`__: attaches the `pam` claims policy to this specific client so the custom claim is actually emitted when a token is issued.
- __Client-level `scopes: ['openid', 'authelia.pam']`__: the client is only allowed to request these two scopes, matching the PAM module's defaults.

Additional notes:

- The `client_secret` value in Authelia's configuration is a hashed representation; use [`authelia crypto hash generate`](../../../reference/guides/generating-secure-values.md) to produce one from a random plaintext secret. The cleartext value is what you pass to [pam_authelia] via [`oauth2-client-secret`](#oauth2-client-secret).
- `authorization_policy: 'two_factor'` forces the device approval itself to require 2FA in Authelia, which effectively gives you two-factor over SSH via the one 2FA challenge the user performs on their phone.
- If you prefer a public client (no secret in `/etc/pam.d/*`), set `public: true` on the Authelia side, omit `client_secret` there, and omit [`oauth2-client-secret`](#oauth2-client-secret) in the PAM config. The `claims_policy` + `scopes` fields still apply.

Once the server-side configuration is in place, reload Authelia and configure the PAM stack as shown in the [Device Authorization flow](#device-authorization-flow) example.

# Device Authorization identity binding

The Device Authorization flow has a subtle trust gap that [pam_authelia] closes explicitly. The OAuth2 token endpoint has no notion of "which local Linux account asked for this code". If left unchecked, any Authelia account holder who scans a displayed QR code can approve the flow with *their own* credentials, and the token endpoint will issue a valid access token. Without an identity check, the PAM module would accept that token as proof of authentication and let the approver log in as the *requesting* Linux user. [pam_authelia] prevents this by verifying the issued token against the PAM username before returning success.

## How the check works

After [pam_authelia]'s poll of `/api/oidc/token` returns an access token and ID token, the Go helper runs the following verification steps in order, and fails closed if any step fails:

1. __OIDC discovery__ against the configured Authelia URL, fetching the JWKs document used to verify signatures.
2. __ID token verification__: signature against the discovery-supplied JWKs, issuer, audience (must equal [`oauth2-client-id`](#oauth2-client-id)), and expiry.
3. __Userinfo request__: calls `/userinfo` under Bearer authentication with the access token.
4. __Token substitution defense__: asserts that `userinfo.sub == id_token.sub`. A mismatch here indicates someone swapped an unrelated access token in for one issued during this device flow.
5. __Username binding__: looks up the `authelia.pam.username` claim in the userinfo response and **case-sensitively** compares it to the Linux username the PAM shim passed to the Go helper on stdin. Missing claim, wrong type, empty value, or any difference fails the login.

On success the helper writes `device identity verified: claim "authelia.pam.username" == pam username "<user>"` to the debug log and returns `PAM_SUCCESS`. On failure it writes a diagnostic line (for example `authelia identity "jane" does not match pam username "john"`) to stderr and returns `PAM_AUTH_ERR`. There is no partial-success path.

## Case sensitivity and username normalization

The comparison is case-sensitive because Linux usernames are. If your Authelia identity store holds usernames in a shape that doesn't match the Linux account verbatim (mixed case, an `@realm` suffix, an email, etc.) you must normalize on the Authelia side before the claim is emitted. The cleanest path is to define a derived [user attribute](../../../configuration/definitions/user-attributes.md) via Authelia's expression engine and anchor the claim at that derived attribute instead of the raw `username`:

```yaml {title="configuration.yml"}
definitions:
  user_attributes:
    pam_username:
      expression: 'username.lowerAscii()'

identity_providers:
  oidc:
    claims_policies:
      pam:
        custom_claims:
          authelia.pam.username:
            attribute: 'pam_username'
```

Strip an `@domain` suffix the same way:

```yaml {title="configuration.yml"}
definitions:
  user_attributes:
    pam_username:
      expression: 'username.split("@")[0].lowerAscii()'
```

Any expression supported by Authelia's [user attributes engine](../../../configuration/definitions/user-attributes.md) works; substitute a per-user override, concatenate fields, or combine with other attributes. As long as the resolved value equals the local Linux account name verbatim, the bind succeeds.

# Verification

Test each configured flow with a plain `ssh` command. Use a dedicated non-root user that exists in Authelia; if you lock yourself out of your only administrator account you will need out-of-band console access to recover.

## 1FA

```bash
ssh john@server.example.com
```

You will be prompted for `john`'s password. Enter the password you configured in Authelia. A successful login should reach the shell prompt within a second.

## 1FA+2FA with TOTP

```bash
ssh john@server.example.com
```

You will be prompted for the password, then for a TOTP code. Enter the 6- or 8-digit code from your authenticator app. A successful login should land you at the shell prompt.

## Device Authorization flow

```bash
ssh john@server.example.com
```

A QR code will be rendered in your terminal along with the verification URL and user code. Scan the QR code on your phone (or visit the URL on any browser that can reach Authelia), complete the Authelia consent flow, and press Enter in the SSH session. [pam_authelia] will poll the token endpoint and return you to the shell once Authelia confirms the approval.

## Inspecting the debug log

With [`debug`](#debug) enabled in the PAM config, a successful 1FA+2FA login with TOTP produces log lines similar to:

```text
pam_authelia: POST https://auth.example.com/api/firstfactor
pam_authelia: response status=200 status_field="OK"
pam_authelia: user info method="totp" has_totp=true has_webauthn=false has_duo=false
pam_authelia: selected "totp" (from priority entry "totp")
pam_authelia: POST https://auth.example.com/api/secondfactor/totp
pam_authelia: response status=200 status_field="OK"
```

On systemd-based distributions these lines are captured by the journal and can be read with:

```bash
sudo journalctl -u ssh -t pam_authelia --since '5 minutes ago'
```

# Troubleshooting

## `Authentication failed` with no useful logs

If the SSH client reports `Authentication failed` and the server journal only shows sshd's `PAM: Authentication failure for user` line with nothing from [pam_authelia], the [`debug`](#debug) flag is not enabled. Add `debug` to the `pam_authelia.so` line in `/etc/pam.d/sshd`, reproduce the login, and re-check the journal.

## `ssh: unable to authenticate, attempted methods [none keyboard-interactive], no supported methods remain`

This message from the SSH client means PAM returned `PAM_AUTH_ERR` early. Check the journal for [pam_authelia] lines; common causes are:

- The Go helper binary is missing or at a non-default path. Confirm `/usr/bin/pam_authelia` exists and is executable, or set [`binary`](#binary) to the correct path.
- The CA certificate path in [`ca-cert`](#ca-cert) is wrong or unreadable; look for `failed to read CA certificate` on stderr.
- `sshd_config` is missing `UsePAM yes` or `KbdInteractiveAuthentication yes`.

## `device authorization response status=401` in the debug log

The `oauth2-client-id` or `oauth2-client-secret` does not match what Authelia has configured for the Device Authorization client, or the client is configured as public on the Authelia side but the PAM config is passing a secret (or vice versa). Reconcile the two sides. The quoted string above is the literal log line the Go helper writes; the lowercase form is intentional, since it matches what you will see in `journalctl`.

## `claim "authelia.pam.username" missing from userinfo response`

Authelia is not emitting the `authelia.pam.username` claim on the userinfo endpoint, so [pam_authelia]'s identity check fails closed. Check that:

1. A [claims policy](../../../configuration/identity-providers/openid-connect/provider.md#claims_policies) exists with `custom_claims.authelia.pam.username` defined, anchored at the right backend attribute.
2. A [custom scope](../../../configuration/identity-providers/openid-connect/provider.md#scopes) named `authelia.pam` exists and its `claims` list grants `authelia.pam.username`.
3. The OIDC client used by [pam_authelia] has `claims_policy: 'pam'` (or whatever you named the policy) attached and has `authelia.pam` in its `scopes` list.
4. The PAM [`oauth2-scope`](#oauth2-scope) option includes `authelia.pam` so the scope is actually requested at device-auth time.

See [Configuring Authelia for the Device Authorization flow](#configuring-authelia-for-the-device-authorization-flow) for the full working YAML.

## `authelia identity "..." does not match pam username "..."`

The user who approved the device flow in the browser is not the same user the PAM module is authenticating. This is [pam_authelia]'s confused-deputy defense working as intended; it means someone other than the SSH-requesting user attempted to approve the flow, or there is a legitimate case or realm-suffix mismatch between the Linux and Authelia usernames. If it's the latter, normalize the username on the Authelia side via a [derived user attribute](../../../configuration/definitions/user-attributes.md) and anchor the `authelia.pam.username` claim at the derived attribute; see [Case sensitivity and username normalization](#case-sensitivity-and-username-normalization).

## `id token verification failed: ...`

The issued ID token did not verify against the Authelia-advertised JWKs. Common causes:

- The [`oauth2-client-id`](#oauth2-client-id) in the PAM config does not match the `audience` of the token Authelia issued (usually because you changed the client ID on one side without the other).
- Authelia's JWKs were rotated while a cached discovery document in the Go helper still references the old ones; restart `sshd` (or whatever PAM consumer caches the helper process) after rotating keys.
- The PAM [`url`](#url) points at a reverse proxy that rewrites the issuer URL on the way through; the `iss` claim Authelia writes won't match the discovery document the helper fetches. Put [pam_authelia] in front of Authelia's canonical public URL, not an internal host name.

## `userinfo request failed`

The access token was rejected by `/userinfo`, usually because the custom `authelia.pam` scope wasn't actually granted at device-auth time. Double-check that the scope name in the PAM [`oauth2-scope`](#oauth2-scope) matches the server-side `identity_providers.oidc.scopes` entry exactly, and that the client's `scopes` list includes it.

## `--oauth2-scope must include openid` / `--oauth2-scope must include authelia.pam`

Config validation errors from [pam_authelia] when an operator passes an explicit [`oauth2-scope`](#oauth2-scope) that drops one of the two mandatory scopes. Restore both; the defaults already include them, so the simplest fix is usually to remove the explicit `oauth2-scope=` entirely from `/etc/pam.d/sshd`.

## `response status=429` in the debug log

Authelia's [regulation](../../../configuration/security/regulation.md) rate-limited the request. Wait for the regulation window to elapse, or tune `regulation.max_retries` and `regulation.find_time` on the Authelia side. [pam_authelia] does not retry rate-limited requests automatically.

## Device Authorization flow fails with `device authorization token expired`

This error comes from Authelia's token endpoint and means the server-side device code lifetime elapsed before the user approved the flow on their phone. The Go helper is still alive at this point (it is simply relaying the `expired_token` response Authelia sent back), so raising the PAM [`timeout`](#timeout) option will not help and neither will raising `sshd`'s `LoginGraceTime`.

There are two real fixes:

1. Approve faster on your phone. The default Authelia device code lifetime is 10 minutes, which is usually plenty.
2. If 10 minutes genuinely is not enough for your users, raise Authelia's device code lifetime server-side by setting [`identity_providers.oidc.lifespans.device_code`](../../../configuration/identity-providers/openid-connect/provider.md#device_code), either globally or on a per-client custom lifespan attached to your Device Authorization client.

The lowercase form of the error string above is intentional; it matches what the Go helper writes to the log.

# Security considerations

- __TLS is always verified.__ There is no insecure or `skip-verify` mode. Connections to Authelia use TLS 1.2 or later, and verification uses the system trust store or the [`ca-cert`](#ca-cert) you provide.
- __Credentials never reach logs.__ Passwords and 2FA tokens are never written to debug output. The debug log records HTTP status codes and the `status` JSON field from Authelia's responses, but never request or response bodies.
- __Credentials are zeroed from memory__ after use via `explicit_bzero(3)`.
- __Device Authorization verification URLs are validated.__ The URL returned by `/api/oidc/device-authorization` must use `https://`, point to the same host as [`url`](#url), and be under 2 KiB. This defends against a compromised or man-in-the-middled Authelia response phishing the user via an attacker-controlled URL rendered as a QR code.
- __Device Authorization tokens are bound to the requesting Linux username.__ After the token endpoint returns, the Go helper runs OIDC discovery, verifies the ID token, calls `/userinfo`, asserts that `userinfo.sub == id_token.sub`, and case-sensitively compares the custom `authelia.pam.username` claim against the Linux username the shim passed to it. Without this check, any Authelia account holder could approve another user's QR code and end up logged in as them. See [Device Authorization identity binding](#device-authorization-identity-binding) for the full check list and the required server-side `claims_policies` plus custom-scope configuration.
- __Client secrets live in `/etc/pam.d/*`.__ When using [`oauth2-client-secret`](#oauth2-client-secret) the value appears in plaintext in PAM configuration files. Verify that those files are not world-readable (`0644` owned by `root` is the default on most distributions), or use a public client to avoid the issue.
- __Client disconnect detection.__ If the SSH client disconnects mid-authentication, the C shim notices via `POLLRDHUP` on the client socket, kills the Go helper with `SIGTERM`, and returns `PAM_AUTH_ERR`. Without this, Device Authorization polling could outlive the SSH session and keep hitting Authelia's token endpoint until the device code expired.
- __Authelia's regulation still applies.__ Rate limiting, IP-based throttling, and any other [regulation](../../../configuration/security/regulation.md) rules enforced by your Authelia deployment apply to every login attempt through [pam_authelia]. Operators who previously relied on SSH's own per-source-IP penalties should verify that Authelia's regulation config is tuned appropriately.

# Limitations

- __No WebAuthn / FIDO2 over SSH.__ FIDO2 authenticators require direct USB or NFC access to the client device, which cannot be tunneled through `sshd`'s keyboard-interactive channel.
- __Device-flow QR codes need a Unicode-capable terminal.__ The QR is rendered with Unicode half-block characters (U+2580, U+2584, U+2588). Terminals that do not render these characters fall back to showing only the verification URL and user code, which still work but defeat the "point your phone at the screen" convenience.
- __WebAuthn users without a fallback will fail.__ If a user's only enrolled 2FA method is WebAuthn and [`method-priority`](#method-priority) is set to an explicit list that excludes both `user` and `device_authorization`, authentication fails. Enroll an additional TOTP or Duo credential, widen the priority list, or route the user through the Device Authorization flow (see the [WebAuthn over SSH](#authentication-flows) callout in the Authentication flows section for details).

# See also

- [github.com/authelia/pam](https://github.com/authelia/pam): source code, releases, and packaging metadata.
- [Authelia issue #497](https://github.com/authelia/authelia/issues/497): the original feature request that led to [pam_authelia].
- [Generating secure values](../../../reference/guides/generating-secure-values.md): how to generate the hashed client secret for the Device Authorization OIDC client.
- [OpenID Connect 1.0 Clients](../../../configuration/identity-providers/openid-connect/clients.md): full schema reference for OIDC client configuration on the Authelia side.
- [OpenID Connect 1.0 Provider: `claims_policies`](../../../configuration/identity-providers/openid-connect/provider.md#claims_policies): how to define the claims policy that emits `authelia.pam.username`.
- [OpenID Connect 1.0 Provider: `scopes`](../../../configuration/identity-providers/openid-connect/provider.md#scopes): how to declare the custom `authelia.pam` scope.
- [User attributes](../../../configuration/definitions/user-attributes.md): expression engine for deriving normalized values to use as the `authelia.pam.username` source.
- [RFC 8628]: OAuth 2.0 Device Authorization Grant specification.
- [RFC 6749]: OAuth 2.0 core specification (relevant for the scope format).

[Authelia]: https://www.authelia.com/
[Authelia APT repository]: ../../deployment/bare-metal.md#apt-repository
[pam_authelia]: https://github.com/authelia/pam
[OpenSSH]: https://www.openssh.com/
[PAM]: https://www.kernel.org/pub/linux/libs/pam/
[Go]: https://go.dev/dl/
[RFC 6749]: https://datatracker.ietf.org/doc/html/rfc6749
[RFC 8628]: https://datatracker.ietf.org/doc/html/rfc8628
