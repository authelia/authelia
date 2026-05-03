---
title: "Seerr"
description: "Trusted Header SSO Integration for Seerr"
summary: ""
date: 2026-04-11T23:39:14+05:30
draft: false
images: []
menu:
integration:
parent: "trusted-header-sso"
weight: 420
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Introduction

This is a guide on integration of __Authelia__ and [Seerr] via the trusted header SSO authentication.

As with all guides in this section it's important you read the [introduction](../introduction.md) first.

## Tested Versions

* Authelia:
  * v4.39.18
* [Seerr] Server

## Before You Begin

This example makes the following assumptions:

* __Application Root URL:__ `https://seerr.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __User Email Domain:__ `@{{< sitevar name="domain" nojs="example.com" >}}`
* Seerr has been initialized already with a user to configure this. Trusted header SSO **cannot** be used to auto-create users.
* Trusted header SSO can only be used when trust proxy is enabled.
* This feature will only work if a list of trusted proxies is provided.

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

To configure [Seerr] to trust the `Remote-User` and `Remote-Email` header do the following:

### With GUI

1. Login as an admin user.
2. Navigate to **Settings -> Network**.
3. Enable "Trust Proxy"
4. Provide a list of trusted proxies in **Advanced Network Settings**.
5. Select `remote-user` and `remote-email` from the list. Seerr will look for both fields in requests if they are configured here.

Alternatively, you can select just the username and look for only the username in the requests.

6. Save changes.

### Editing the configuration directly

Update the configuration file to this

```json
...
 "network": {
  ...
  "trustProxy": true,
  "trustedProxies": {
   "v4": [
    "10.0.50.3"
   ],
   "v6": [
    "fd00:dead::beef"
   ]
  },
  "forwardAuth": {
   "enabled": true,
   "userHeader": "remote-user",
   "emailHeader": "remote-email"
  },
  "proxy": {
    ...
```

The header names in the GUI and config file are case insensitive.

[Seerr]: https://github.com/seerr-team/seerr
