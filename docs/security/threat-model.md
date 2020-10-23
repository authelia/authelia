---
layout: default
title: Threat Model
parent: Security
nav_order: 2
---

# Threat Model

The design goals for Authelia is to protect access to applications by collaborating with reverse proxies to prevent
attacks coming from the edge of the network. This document gives an overview of what Authelia is protecting against but some
of those points are also detailed in [Security Measures](./measures.md).

## General assumptions

Authelia is considered to be running within a trusted network and it heavily relies on the first level of security provided by reverse proxies. It's very important that you take time configuring your reverse proxy properly to get all the authentication benefits brought by Authelia.
Some general security tweaks are listed in [Security Measures](./measures.md) to give you some ideas.

## Guarantees

If properly configured, Authelia guarantees the following for security of your users and your apps:

* Applications cannot be accessed without proper authorization. The access control list is highly configurable allowing administrators to guarantee least privilege principle.
* Applications can be protected with two factors in order to fight against credentials theft and protect highly sensitive data or operations.
* Sessions are bound in time, limiting the impact of a cookie theft. Sessions can have both soft and hard limits. With soft limit, the user is logged out when inactive for a certain period. With hard limit, the user has to authenticate again after a certain period, whether they were active or not. 
* Brute force attacks against credentials are protected thanks to a regulation mechanism temporarily blocking the user account after too many attempts.
* Identity validation is required for performing administrative actions such as registering 2FA devices, preventing attackers to pass second factor by auto-registering their own 2FA device. An email with a link is sent to the user and a click is required to confirm the action.
* Prevention against session fixation by regenerating a new session after each privilege elevation.
* Prevention against LDAP injection by following OWASP recommendations regarding valid input characters (https://cheatsheetseries.owasp.org/cheatsheets/LDAP_Injection_Prevention_Cheat_Sheet.html).
* Connections between Authelia and thirdparty components like mail server, database, cache and LDAP server can be made over TLS to protect against man-in-the-middle attacks from within the infrastructure.
* Validation of user session group memberships gets refreshed regularly from the authentication backend (LDAP only).
 
## Potential future guarantees

* Define and enforce a password policy (to be designed since such a policy can clash with a policy set by the LDAP server).
* Detect credential theft and prevent malicious actions.
* Detect session cookie theft and prevent malicious actions.
* Authenticate communications between Authelia and reverse proxy.
* Securely transmit authentication data to backends (OAuth2 with bearer tokens).
* Protect secrets stored in DB with encryption to prevent secrets leak by DB exfiltration.
* Least privilege on LDAP binding operations (currently administrative user is used to bind while it could be anonymous).
* Extend the check of user group memberships to authentication backends other than LDAP (File currently).
* Invalidate user session after profile or membership has changed in order to drop remaining privileges on the fly.

## Trusted environment

It's important to note that Authelia is considered running in a trusted environment for two reasons

1. Requests coming to Authelia should be initiated by reverse proxies but CAN be initiated by any other server currently. There is no trusted relationship between Authelia and the reverse proxy so an attacker within the network could abuse Authelia and attack it.
2. Your environment should be considered trusted especially if you're using the `Remote-User`, `Remote-Name`, `Remote-Email` and `Remote-Groups` headers to forward authentication data to your backends. These headers are transmitted plain and unsigned to the backends, meaning a malicious user within the network could pretend to be Authelia and send those headers to bypass authentication and gain access to the service. A mitigation could be to transmit those headers with a digital signature which could be verified by the backend however, many backends just don't support it. It has therefore been decided to invest on OpenID Connect instead to solve that authentication delegation problem. Indeed, many backends
do support OAuth2 though since it has become a standard lately.
