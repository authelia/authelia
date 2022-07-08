---
title: "Pull Request Guidelines"
description: "Authelia Development Pull Request Guidelines"
lead: "This section covers the pull request guidelines."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  contributing:
    parent: "development"
weight: 232
toc: true
---

[Pull Request] guidelines are in place in order to maintain consistency and clearly communicate our process for
processing merges into the [master] branch.

## Overview

* Ensure the `Allow edits by maintainers` checkbox is checked due to our [Squash Merge](#squash-merge) policy
* Ensure you avoid a force push due to our [Squash Merge](#squash-merge) policy and [Review](#review) complications

## Squash Merge

Every [Pull Request] will be squash merged into [master]. This requires the [Pull Request] branch to be up-to-date with
the [master] branch.

## Review

Every [Pull Request] will undergo a formal review process. This process is heavily complicated if you rewrite history
and/or perform a force push, especially after a maintainer has started a review. As such we request that any action that
