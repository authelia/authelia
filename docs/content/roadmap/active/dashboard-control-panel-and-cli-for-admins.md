---
title: "Dashboard / Control Panel and CLI for Administrators"
description: "Authelia Administrator Dashboard."
summary: "A dashboard or control panel for administrators to adjust system settings is easily one of the most impactful features we can implement."
date: 2024-03-21T18:25:55+11:00
draft: false
images: []
weight: 245
toc: true
aliases:
  - /r/admin-dashboard
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Adding a way for administrators to dynamically interface with Authelia is one of the more anticipated features by users,
this article describes ideas about this feature some of which are certain to be implemented and some which may not end
up being implemented. Effectively we'd like to be able to optionally give administrators the ability to dynamically
control aspects of Authelia while it's running, either via a CLI or via a UI.

This feature will pave the way to adding lots of useful administrator facing features. It will require in database
settings storage as well as some minimal traditional settings via files or environment variables once the feature
reaches the theoretical ideal state.

This feature should not be confused with the [Dashboard / Control Panel for Users](dashboard-control-panel-for-users.md)
which is the dashboard for the user self-managing their own settings, as this feature is for managing the system
settings instead of the user settings.

## Broad Concepts: Optional, Explicit, and Modular

This general idea is about making it as secure and flexible as possible. Reasonably in any security solution like
Authelia administrators should have to explicitly and deliberately enable dynamic controls like this. Defaulting to a
static configuration allowing mitigation against potential dynamic configuration vulnerabilities. In addition a lot of
users love Authelia for the declarative configuration it already has so we don't want to take that away from them.

Another key aspect that we feel a feature like this must have is the ability to add additional mitigations to this like
an external firewall that only allows certain network addresses or ZTNA identities to access. To that end we ideally
will not only allow configuring Authelia without the dynamic configuration but also allow:

- Control via a CLI either additionally or exclusively
- Control via a UI either:
  - On the same port as the portal
  - On a different port from the portal
  - In a separate process and/or host entirely

In addition to this feature we want to ensure the admin UI is invisible to the normal user so there is no indication if
it is or is not enabled, and even make it possible that it can be configured in a way where a person with a compromised
admin account will not know they are an admin or that any admin UI exists. This can realistically be achieved by
configuration similar to the example below.

```yaml {title="configuration.yml"}
server:
  address: 'tcp://:9091'

administration:
  ## Explicitly enable Dynamic Configuration, either via CLI only or CLI and UI via enable_ui.
  enable: false

  ## Explicitly enable the Admin UI on this Authelia instance. If 'enable' is configured another process could also be
  ## separately configured with this enabled.
  enable_ui: false

  ## Configure the listener address for the UI. If it's exactly the same host and port component as above with a
  ## different path, listen on the same listener as above. If it's exactly the same, error.
  address: 'tcp://:9092'

  ## URL enabling the button/link in the portal for the UI.
  url: 'https://auth-admin.example.com'

  ## List of users who are allowed to view the admin UI. In addition to the groups.
  users:
    - 'john'

  ## List of groups who are allowed to view the admin UI. In addition to the users.
  groups:
    - 'admins'
```

## Broad Concepts: API

A separate API which is authorized via OAuth 2.0 or user sessions will be an absolute must for this feature at least
long term. For example if configured to listen on `https://auth.example.com/admin` the API should be serviced via
`https://auth.example.com/admin/api/v1` or similar.

The tokens should realistically be able to be generated with granular access, should require a special scope, and have
a particular audience to be considered valid.

## Stages

This section represents the stages involved in implementation of this feature. The stages are either in order of
implementation due to there being an underlying requirement to implement them in this order, or in their likely order
due to how important or difficult to implement they are.

### Design Stage

{{< roadmap-status stage="in-progress" >}}

Decide on a design.

### Initial Implementation

{{< roadmap-status version="v4.40.0" >}}

Implement the pivotal elements of the design.

### Design Element: Segregation

{{< roadmap-status version="v4.40.0" >}}

Allow the admin UI to be run as a separate process, on a different port, and at a different URL to Authelia itself.
Alternatively allow it to run as part of the main process and port for minimal configurations.

### Session Management

{{< roadmap-status version="v4.40.0" >}}

Manage user sessions for all users.

### OpenID Connect 1.0 Client Management

{{< roadmap-status version="v4.40.0" >}}

Manage client registrations via a web frontend.

### Access Control Management

{{< roadmap-status >}}

Manage Access Control rules.

### User Management

{{< roadmap-status >}}

Manage user accounts with either the internal or LDAP authentication backends. Allow for creation, modification, and
deletion.


