---
title: "Paperless"
description: "Trusted Header SSO Integration for Paperless"
summary: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 420
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Introduction

This is a guide on integration of __Authelia__ and [Paperless] (specifically Paperless-ngx) via the trusted header SSO
authentication.

As with all guides in this section it's important you read the [introduction](../introduction.md) first.

## Tested Versions

* Authelia:
  * v4.38.7
* Paperless:
  * v2.7.2

## Before You Begin

This example makes the following assumptions:

* __Application Root URL:__ `https://paperless.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`

## Configuration

To configure [Organizr] to trust the `Remote-User` header do the following:

1. Configure the environment variables:

```env
PAPERLESS_ENABLE_HTTP_REMOTE_USER=true
PAPERLESS_HTTP_REMOTE_USER_HEADER_NAME=HTTP_REMOTE_USER
PAPERLESS_LOGOUT_REDIRECT_URL=https://auth.example.com/logout
```

## See Also

[Organizr] does not appear to have documentation around their `Auth Proxy` configuration.

[Organizr]: https://organizr.app/
