---
title: "Troubleshooting Sanitization"
description: "This guide describes and helps users sanitize provided files to hide privacy related values for troubleshooting"
lead: "This guide describes and helps users sanitize provided files to hide information for privacy."
date: 2022-08-26T13:50:51+10:00
draft: false
images: []
menu:
  reference:
    parent: "guides"
weight: 220
toc: true
aliases:
  - /r/sanitize
  - /reference/guides/domain-sanitizaiton
---

Some users may wish to hide their domain in files provided during troubleshooting. While this is discouraged, if a user
decides to perform this action it's critical for these purposes that you hide your domain in a very specific
way. Most editors allow replacing all instances of a value, utilizing this is essential to making troubleshooting
possible.

## General Rules

1. Only replace the purchased portion of domains:
   - For example if you have `auth.abc123.com` and `app.abc123.com` they
   should become `auth.example.com` and `app.example.com`, i.e. replace all instances of `abc123.com` with `example.com`.
2. Make sure value replaced is replaced with a unique value:
   - For example if you replace `abc123.com` with `example.com` DO NOT replace any other value other than `abc123.com` with
   `example.com`. The same rule applies to IP addresses, usernames, and groups.
3. Make sure the value replaced is replaced across logs, configuration, and any references:
   - For example if you replace `abc123.com` with `example.com` in your configuration, make exactly the same replacement
   for the log files.
4. Make sure this consistency is followed for all communication regarding a single issue.

## Multiple Domains

*__Replacement Value:__* `example#.com` (where `#` is a unique number per domain)

In instances where there are multiple domains it's recommended these domains are replaced with `example1.com`,
`example2.com`, etc.

## Specific Values
