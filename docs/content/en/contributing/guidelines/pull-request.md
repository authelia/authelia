---
title: "Pull Request"
description: "Authelia Development Pull Request Guidelines"
lead: "This section covers the pull request guidelines."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  contributing:
    parent: "guidelines"
weight: 320
toc: true
aliases:
  - /contributing/development/guidelines-pull-request/
---

[Pull Request] guidelines are in place in order to maintain consistency and clearly communicate our process for
processing merges into the [master] branch.

## Overview

* Ensure the `Allow edits by maintainers` checkbox is checked due to our [Squash Merge](#squash-merge) policy
* Ensure you avoid a [force push](#force-push) excluding the specific exceptions listed in the
  [force push section](#force-push)

[Developer Certificate of Origin]: commit-message.md#developer-certificate-of-origin

## Squash Merge

Every [Pull Request] will be squash merged into [master]. This requires the [Pull Request] branch to be up-to-date with
the [master] branch.

## Force Push

Please do not force push to your PR's branch after you have created your PR especially when a maintainer has either
performed a review or has indicated they are performing a review, as doing so makes it harder to review your commits
accurately. PRs will always be squashed by us when we merge your work. Commit as many times as you need in your
pull request branch.

A few exceptions exist to this rule and are as follows:

- Making adjustments to the commit message i.e. for the following reasons:
  - To comply with the [Developer Certificate of Origin] sign off requirement
  - To comply with the other [Commit Message] guidelines
- To rebase your changes off of master or another branch

## Review

Every [Pull Request] will undergo a formal review process. This process is heavily complicated if you rewrite history
and/or perform a force push, especially after a maintainer has started a review. As such we request that any action that
you merge `origin/master` into your branch to synchronize your commit after the initial review and any other action that
rewrites history.

### Requirements

The following requirements must be met for a pull request to be accepted. This list also acts as a checklist for
maintainers in their review process.

- The changes must be [documented](../prologue/documentation-contributions.md) if they add or change behaviour
- The changes must meet the following guidelines:
  - [General](introduction.md#general-guidelines)
  - [Commit Message]
  - [Database Schema](database-schema.md)
  - [Documentation](documentation.md)
  - [Testing](testing.md)
  - [Accessibility](accessibiliy.md)
  - [Style](style.md)
- The changes adhere to all of the relevant linting and quality testing automations
- The pull request closes related issues by mentioning them appropriately
- The contribution adhere to the security by design principles by:
  - Setting secure defaults
  - Disallows critically insecure settings
  - Requires explicit awareness by users that specific settings may reduce security
- Potential future items:
  - Contribution includes DCO
  - Contribution includes REUSE-compliance requirements

[Commit Message]: commit-message.md
[Pull Request]: https://github.com/authelia/authelia/pulls
[master]: https://github.com/authelia/authelia/tree/master/
