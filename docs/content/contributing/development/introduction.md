---
title: "Development"
description: "An introduction into contributing to the Authelia project via development."
summary: "An introduction into contributing to the Authelia project via development."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 210
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

We encourage anyone who wishes to contribute via development to utilize GitHub [Issues] or [Discussions] or one of the
[other contact methods](../../information/contact.md) to discuss their contribution in advance and come up with a design
plan.

It's also important that you read guidelines and try to follow them. The development section is arranged in the order
we recommend reading it, and you can utilize the pagination at the bottom to navigate to the next part of the
development guide.

## Licensing

As the main Authelia repository and all supporting Authelia repositories are hosted on [GitHub](https://github.com)
users are explicitly making all contributions under the license agreement included with the repository. This is a
commonly accepted practice in Open Source and it is also explicitly expressed here and in the
[GitHub Terms of Service](https://docs.github.com/en/site-policy/github-terms/github-terms-of-service#6-contributions-under-repository-license)
 which users must have agreed to in order to attempt a contribution.

## Dependencies

### Selection

Dependencies should be kept to a minimum. When a new dependency is required, the following criteria apply:

- Prefer well-maintained libraries with active communities and a track record of timely security fixes.
- The dependency must use a compatible open-source license (e.g., MIT, Apache 2.0, BSD).
- Avoid introducing a dependency when the required functionality is small enough to implement directly.
- Discuss the addition of any new dependency with the maintainers before submitting a pull request.

### Obtaining

Dependencies are obtained via the standard package managers for each language:

- **Backend (Go):** Dependencies are declared in `go.mod` with integrity checksums recorded in `go.sum`. Versions
  are explicitly pinned.
- **Frontend (Node.js):** Dependencies are declared in `web/package.json` with integrity checksums and pinned
  versions recorded in `web/pnpm-lock.yaml`.

All tools that support explicit version pinning use pinned versions.

### Tracking and Updates

Dependency updates are managed automatically by [Renovate](https://docs.renovatebot.com/), which monitors for new
versions and submits pull requests. These pull requests follow the standard
[review process](../guidelines/pull-request.md#review) and must pass all status checks before being merged.

Each release includes Software Bill of Materials (SBOM) artifacts in both
[CycloneDX](https://cyclonedx.org/) and [SPDX](https://spdx.dev/) formats for all release artifacts. Release
provenance is generated at [SLSA](https://slsa.dev/) Build Level 3 using the SLSA GitHub Generator. For more
details on verifying release artifacts see the
[artifact signing and provenance](../../overview/security/artifact-signing-and-provenance.md) documentation.

Automated vulnerability scanning via [Grype](https://github.com/anchore/grype) is integrated into the CI/CD
pipeline and runs against both container images and SBOM artifacts.

Additional scanning is performed routinely via 

[Issues]: https://github.com/authelia/authelia/issues
[Discussions]: https://github.com/authelia/authelia/discussions
