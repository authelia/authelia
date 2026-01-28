---
title: "Team & Governance"
description: "About the Authelia Team and Project Governance"
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
toc: true
aliases:
  - '/team'
  - '/governance'
  - '/team-governance'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Teams

The following section describes the various teams within the Authelia project.

### Core Team

{{% profile-team name="core" %}}

### Maintainers Team

{{% profile-team name="maintainers" %}}

## Access and Permissions

The following section describes how team membership and access permissions are managed within the Authelia project.


### Access Levels

**External Contributors**

- Submit pull requests via forked repositories
- No direct repository access
- CI/CD workflow runs require approval from team members

**Maintainers Team**

- Push access to feature branches (not main/master)
- Can approve CI/CD workflow runs for external contributions
- Can review and approve pull requests
- Subject to all branch protection and approval requirements
- Access to specific infrastructure as needed for their responsibilities

**Core Team**

- Full repository access subject to branch protection rules
- Access to infrastructure, hosting, and deployment systems
- Access to release signing credentials
- Handle security vulnerabilities and incidents
- All operations still subject to multi-person approval requirements

### Team Growth Process

New team members are identified through their contributions and community involvement:

1. **Contribution Period**: Potential members demonstrate quality work through external contributions
2. **Invitation**: Core Team members may invite consistent contributors to join as Maintainers
3. **Gradual Access**: Access is granted incrementally based on responsibilities and trust
4. **Consensus**: All team additions are discussed and approved by Core Team members

### Security Controls

To protect against account compromise and ensure code quality, we implement multiple safeguards:

- **Multi-Person Approval**: All pull requests require approval from 2 team members before merge (excluding automated dependency updates and minor maintenance)
- **Release Protection**: All releases require approval from 2 members with write access.
- **Branch Protection**: The main branch cannot be pushed to directly; all changes require pull requests
- **Two-Factor Authentication**: Required for all team members to perform sensitive operations
- **No Unilateral Access**: Even with write permissions, no single team member can make changes alone

These controls guard against compromised accounts, accidental errors, and ensure all changes receive peer review.

### Access Review

- Team access is reviewed as needed
- Inactive members may have access removed after Core Team discussion
- Access can be revoked immediately for security concerns or code of conduct violations
- All access changes are discussed among Core Team members

## Governance

Authelia is free from any outside governance and is entirely governed as outlined on this page. We do not have any affiliations which have ever asked us to modify our governance structure.

Our affiliations with external companies are transparently communicated on the [Sponsors](./sponsors.md) page.

### Decision-Making

- Core Team makes governance decisions unanimously
- Technical decisions follow consensus-based discussion
- Major changes are discussed openly in issues and discussions
- Community input is valued and considered in decision-making

## Compliance

The following section contains various compliance related information.

### Key Individuals

There is no key individual who if they were incapacitated or unavailable would prevent future operations of the project.

All of the following areas can be reset or are otherwise accessible to all of the members of the [Core Team](#core-team):

- Private Keys
- Access Rights
- Passwords

### Bus Factor

The Authelia team has a bus factor of 3. Meaning that the project would stall if 3 team members were suddenly hit by a
bus.
