---
title: "Jira"
description: "Trusted Header SSO Integration for Jira"
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
aliases:
  - /docs/community/using-remote-user-header-for-sso-with-jira.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Introduction

This is a guide on integration of __Authelia__ and [Jira] via the trusted header SSO authentication.

As with all guides in this section it's important you read the [introduction](../introduction.md) first.

## Tested Versions

* Authelia: v4.35.5
* Jira: Unknown
* EasySSO: Unknown

## Before You Begin

This example makes the following assumptions:

* The user accounts with the same names already exist in [Jira].
* You have purchased the third-party plugin from the [Atlassian marketplace](https://marketplace.atlassian.com/apps/1212581/easy-sso-jira-kerberos-ntlm-saml?hosting=server&tab=overview)

## Configuration

To configure [Jira] to trust the `Remote-User` and `Remote-Email` header do the following:

1. Visit the Easy SSO plugin settings
2. Under HTTP configure the `Remote-User` header
3. Check the `Username` checkbox

## See Also

* [EasySSO Documentation](https://techtime.co.nz/display/TECHTIME/EasySSO#documentation-area)

[Jira]: https://www.atlassian.com/software/jira
