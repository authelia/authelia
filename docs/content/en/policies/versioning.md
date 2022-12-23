---
title: "Versioning Policy"
description: "The Authelia Versioning Policy which is important reading for administrators"
date: 2022-12-21T20:48:14+11:00
draft: false
images: []
aliases:
  - /versioning-policy
  - /versioning
---

The __Authelia__ team aims to abide by the [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) policy. This
means that we use the format `major.minor.patch` for our version numbers, where a change to `major` denotes a breaking
change which will likely require user interaction to upgrade, `minor` which denotes a new feature, and `patch` denotes a
fix.

It is therefore recommended users do not automatically upgrade the `minor` version without reading the patch notes, and
it's critically important users do not upgrade the `major` version without reading the patch notes. You should pin your
version to `4.37` for example to prevent automatic upgrades from negatively affecting you.

## Exceptions

There are exceptions to this versioning policy.

### Advanced Customizations

Some advanced customizations are not guaranteed by the versioning policy. These features require the administrator to
ensure they keep up to date with the changes relevant to their version. While the customizations exist as a feature we
cannot allow these customizations to hinder the development process.

Notable Advanced Customizations:

- Templates:
  - Email
  - Content Security Policy header
- Localization Assets

### Breaking Changes

All features which are marked as:

- beta
- experimental

Notable examples:

- OpenID Connect 1.0
- File Filters

The reasoning is as we develop these features there may be mistakes and we may need to make a change that should be
considered breaking. As these features graduate from their status to generally available they will move into our
standard versioning policy from this exception.
