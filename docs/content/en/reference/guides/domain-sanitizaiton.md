---
title: "Domain Sanitization"
description: "This guide describes and helps users sanitize provided files to hide their domain"
lead: "This guide describes and helps users sanitize provided files to hide their domain."
date: 2022-08-26T13:50:51+10:00
draft: false
images: []
menu:
  reference:
    parent: "guides"
weight: 220
toc: true
---

Some users may wish to hide their domain in files provided during troubleshooting. While this is discouraged, if a user
decides to perform this action it's critical for these purposes that you hide your domain in a very specific
way. Most editors allow replacing all instances of a value, utilizing this is essential to making troubleshooting
possible.

## General Rule

Only replace the purchased portion of domains. For example if you have `auth.abc123.com` and `app.abc123.com` they
should become `auth.example.com` and `app.example.com`, i.e. replace all instances of `abc123.com` with `example.com`.

## Multiple Domains

*__Replacement Value:__* `example#.com` (where `#` is a unique number per domain)

In instances where there are multiple domains it's recommended these domains are replaced with `example1.com`,
`example2.com`, etc.
