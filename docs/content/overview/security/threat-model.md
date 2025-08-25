---
title: "Threat Model"
description: "An overview of the Authelia threat model."
summary: "An overview of the Authelia threat model."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 430
toc: true
aliases:
  - /o/threatmodel
  - /docs/security/threat-model.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The design goals for Authelia is to protect access to applications by collaborating with reverse proxies to prevent
attacks coming from the edge of the network. This document gives an overview of what Authelia is protecting against.
Some of these ideas are expanded on or otherwise described in [Security Measures](measures.md).

## General assumptions

Authelia is considered to be running within a trusted network and it heavily relies on the first level of security
provided by reverse proxies. It's very important that you take time configuring your reverse proxy properly to get all
the authentication benefits brought by Authelia.

Some general security tweaks are listed in [Security Measures](measures.md) to give you some ideas.

## Guarantees

If properly configured, Authelia guarantees the following for security of your users and your apps:

* Applications cannot be accessed without proper authorization. The access control list is highly configurable allowing
  administrators to guarantee the least privilege principle.
* Applications can be protected with two-factor authentication in order to fight against credential theft and protect
  highly sensitive data or operations.
* Sessions are bound in time, limiting the impact of a cookie theft. Sessions can have both soft and hard expiration
  time limits. With the soft limit, the user is logged out when inactive for a certain period of time. With the hard
  limit, the user has to authenticate again after a certain period of time, whether they were active or not.
* Brute force attacks against credentials are protected thanks to a regulation mechanism temporarily blocking the user
  account after too many attempts and delays to the authentication process.
* Identity validation is required for performing administrative actions such as registering 2FA devices, preventing
  attackers to pass two-factor authentication by self-registering their own device. An email with a link is sent to the
  user with a link providing them access to the registration flow which can only be opened by a single session.
* Prevention against session fixation by regenerating a new session after each privilege elevation.
* Prevention against LDAP injection by following
  [OWASP recommendations](https://cheatsheetseries.owasp.org/cheatsheets/LDAP_Injection_Prevention_Cheat_Sheet.html)
  regarding valid input characters.
* Connections between Authelia and third-party components like the SMTP server, database, session cache, and LDAP server
  can be made over TLS to mitigate against man-in-the-middle attacks.
* Validation of user group memberships gets refreshed regularly from the authentication backend (LDAP only).

## Potential future guarantees

* Define and enforce a password policy (to be designed since such a policy can clash with a policy set by the LDAP
  server).
* Detect credential theft and prevent malicious actions.
* Detect session cookie theft and prevent malicious actions.
* Binding session cookies to single IP addresses.
* Authenticate communication between Authelia and reverse proxy.
* Securely transmit authentication data to backends (OAuth2 with bearer tokens).
* Least privilege on LDAP binding operations (currently administrative user is used to bind while it could be anonymous
  for most operations).
* Extend the check of user group memberships to authentication backends other than LDAP (File currently).
* Allow administrators to configure policies regarding password resets so a compromised email account does not leave
  accounts vulnerable.
* Allow users to view their active and past sessions to either destroy them, report malicious activity to the
  administrator, or both.
* Allow administrators to temporarily restrict users that are suspected of being compromised no matter which backend is
  being used.
* Log comprehensive information about user sessions so administrators can identify malicious activity and potential
  consequences or damage caused by identified malicious activity.
* Ensure the `X-Forwarded-*` and `X-Original-*` headers are able to be trusted by allowing configuration of trusted proxy
  servers.

## Trusted environment

It's important to note that Authelia is considered running in a trusted environment for two reasons:

1. Requests coming to Authelia should be initiated by reverse proxies but CAN be initiated by any other server
   currently. There is no trusted relationship between Authelia and the reverse proxy so an attacker within the network
   could abuse Authelia and attack it.
2. Your environment should be considered trusted especially if you're using the `Remote-User`, `Remote-Name`,
   `Remote-Email` and `Remote-Groups` headers to forward authentication data to your backends. These headers are
   transmitted unsigned to the backends, meaning a malicious user within the network could pretend to be
   Authelia and send those headers to bypass authentication and gain access to the service. This could be mitigated by
   transmitting those headers with a digital signature which could be verified by the backend however, many backends
   just won't support it. It has therefore been decided to invest in OpenID Connect 1.0 instead to solve that
   authentication delegation problem.
