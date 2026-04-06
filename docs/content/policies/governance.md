---
title: "Governance Policy"
description: "The Authelia Governance Policy which describes how the Authelia project is governed."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
toc: true
type: legal
aliases:
  - /governance-policy
  - /governance
  - /governance.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia is free from any outside parties directly influencing its decision and architecture process and is entirely
governed as outlined on this page.

To date no party that has contributed financially or otherwise to the project has been directly involved in the
design or implementation of the project or has attempted to influence the project in any way that we're aware of. Our
promise is that if this changes we will publish the name of the party and the details of attempt transparently on this
page.

Our affiliations with external companies will be transparently communicated on this page and the
[sponsors](../information/about.md#sponsors) section.

This policy outlines how the Authelia project is governed and the various processes that are in place to ensure
that the project is run in a safe and sustainable manner.

## Roles and Responsibilities

The following describes the roles within the Authelia project and their associated responsibilities.

### Maintainers

{{% profile-team name="maintainers" %}}

Maintainers are responsible for the day-to-day maintenance of the project. Their responsibilities include:

- Reviewing and merging pull requests.
- Triaging issues and discussions.
- Ensuring contributions meet the project [guidelines](../contributing/guidelines/introduction.md).
- Maintaining code quality and test coverage.
- Participating in project discussions and decision-making.

### Core Team

{{% profile-team name="core" %}}

All core team members hold the same access and responsibilities as [maintainers](#maintainers). In addition, the
core team is responsible for:

- Strategic direction and long-term planning for the project.
- Governance, policy decisions, and enforcement of the [code of conduct](code-of-conduct.md).
- Security response and coordinated vulnerability disclosure as described in the [security policy](security.md).
- Release management and versioning decisions.
- Infrastructure administration and management of sensitive resources.
- Reviewing and approving escalated permissions for contributors.

## Sensitive Resource Access

The following table summarizes which sensitive resources each role has access to. For the list of
current members in each role see the [Maintainers](#maintainers) and [Core Team](#core-team) sections above.

The table only describes the default sensitive resources the role has access to, while there is no divergence at this
stage there may be in the future.

| Sensitive Resource                         | Maintainers | Core Team |
|:-------------------------------------------|:-----------:|:---------:|
| Repository write access (commit and merge) |      Y      |     Y     |
| CI/CD pipeline unblock/approval            |      Y      |     Y     |
| CI/CD pipeline secrets                     |             |     Y     |
| CI/CD pipeline configuration               |             |     Y     |
| Package registry publishing credentials    |             |     Y     |
| Infrastructure access                      |             |     Y     |
| Organization-level administrative access   |             |     Y     |

## Access Control Enforcement

The following technical controls are enforced at the platform level to protect the project's version control system
and sensitive resources.

### Multi-Factor Authentication

The GitHub organization requires multi-factor authentication (MFA) for all members. Any user who attempts to access
sensitive resources in the version control system must have completed MFA enrollment. Members who disable or fail to
configure MFA are automatically removed from the organization by GitHub.

Members are also not permitted to have less secure multi-factor authentication (MFA) methods such as SMS. See the
[GitHub documentation](https://docs.github.com/en/organizations/keeping-your-organization-secure/managing-two-factor-authentication-for-your-organization/requiring-two-factor-authentication-in-your-organization#requiring-secure-methods-of-two-factor-authentication-in-your-organization)
for more details.

### Primary Branch Protection

The project's primary branch (`master`) is protected by a
[GitHub repository ruleset](https://github.com/authelia/authelia/rules). This ruleset enforces the following:

- Direct commits to the primary branch are blocked. All changes must be submitted via a pull request and pass the
  required [review process](../contributing/guidelines/pull-request.md#review).
- Deletion of the primary branch is prevented. Any attempt to delete the primary branch is rejected by the ruleset.

## Public Discussion and Communication

The project maintains several mechanisms for public discussion about proposed changes, usage obstacles, and general
community interaction:

- [GitHub Issues](https://github.com/authelia/authelia/issues): for bug reports and feature requests.
- [GitHub Discussions](https://github.com/authelia/authelia/discussions): for ideas, questions, support requests,
  and sharing configuration or setups.
- [Matrix](https://matrix.to/#/#community:authelia.com) and [Discord](https://discord.authelia.com): real-time
  community chat with dedicated rooms for support, contributing/development, and off-topic discussion.

For full details see the [contact](../information/contact.md) page.

## Contributing

The project welcomes contributions from anyone. The contribution process and requirements for acceptable
contributions are documented in the [contributing](../contributing/prologue/introduction.md) section. In summary:

1. Contributors should discuss their intended changes in advance to ensure it aligns with the governed direction via
   [GitHub Issues](https://github.com/authelia/authelia/issues),
   [GitHub Discussions](https://github.com/authelia/authelia/discussions), or
   [chat](../information/contact.md#chat).
2. Fork the repository and create a pull request following the
   [pull request guidelines](../contributing/guidelines/pull-request.md).
3. All contributions must adhere to the project [guidelines](../contributing/guidelines/introduction.md), which
   include requirements for [commit messages](../contributing/guidelines/commit-message.md),
   [code style](../contributing/guidelines/style.md),
   [testing](../contributing/guidelines/testing.md), and
   [documentation](../contributing/guidelines/documentation.md).
4. All status checks must pass and test coverage must not regress.
5. At least one maintainer must review and approve the pull request before it is merged.
6. Pull requests are squash-merged by maintainers.

## Contributor Access and Escalated Permissions

Contributors are reviewed prior to being granted escalated permissions to sensitive resources either directly or via
the [Roles and Responsibilities](#roles-and-responsibilities). Access to sensitive resources are reviewed on a
case-by-case basis, and each contributor granted access to one sensitive resource is re-reviewed should they need or
want access to any additional sensitive resource. Sensitive resources include but are not limited to:

- Repository write access (commit and merge permissions)
- CI/CD pipeline:
  - Secrets
  - Configuration
  - Unblock/Approval Permission
- Package registry publishing credentials
- Infrastructure access
- Organization-level administrative permissions

### Review Requirements

Before a contributor may be granted escalated permissions to any sensitive resource, the following requirements must
be satisfied:

1. The contributor must have a minimum of six months of active, consistent contributions to the project.
2. At least one existing member of the core team who holds the relevant access level must review and approve the
   escalation in discussion with one other member of the core team.
3. The review must evaluate:
   - The quality and consistency of the contributor's contribution history.
   - Adherence to project coding standards, guidelines, and review processes.
   - Alignment with the goals and direction of the project.
   - Demonstrated security awareness, including no history of introducing known vulnerabilities and an understanding
     of responsible disclosure practices.

### Default Permissions

The GitHub organization base permission is set to read-only. When a new collaborator is added to the organization
they receive no write access by default. All escalated permissions including repository write access are granted
exclusively through manual team assignment by a core team member after the [review requirements](#review-requirements)
have been satisfied.

### Granting Process

Escalated permissions are granted only after the review requirements above are met. The core team member approving
the escalation is responsible for ensuring the review has been conducted thoroughly. Permissions are scoped to the
minimum level necessary for the contributor's role and responsibilities.

### Revocation

Escalated permissions may be revoked at any time by the core team if a contributor is found to have violated project
policies, acted in bad faith, or is no longer actively contributing to the project.
