---
title: "Organizr"
description: "Trusted Header SSO Integration for Organizr"
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
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

This is a guide on integration of __Authelia__ and [Organizr] via the trusted header SSO authentication.

As with all guides in this section it's important you read the [introduction](../introduction.md) first.

## Tested Versions

* Authelia:
  * v4.35.5
* Organizr:
  * 2.1.1890

## Before You Begin

This example makes the following assumptions:

* __Application Root URL:__ `https://organizr.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
* __Reverse Proxy IP:__ `172.16.0.1`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

To configure [Organizr] to trust the `Remote-User` and `Remote-Email` header do the following:

1. Visit System Settings
2. Visit Main
3. Visit Auth Proxy
4. Fill in the following information:
   1. Auth Proxy: `Enabled`
   2. Auth Proxy Whitelist: `172.16.0.1`
   3. Auth Proxy Header Name: `Remote-User`
   4. Auth Proxy Header Name for Email: `Remote-Email`
   5. Override Logout: `Enabled`
   6. Logout URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/logout`

{{< picture src="organizr.png" alt="Organizr" width="736" style="padding-right: 10px" >}}

## See Also

[Organizr] does not appear to have documentation around their `Auth Proxy` configuration.

[Organizr]: https://organizr.app/
