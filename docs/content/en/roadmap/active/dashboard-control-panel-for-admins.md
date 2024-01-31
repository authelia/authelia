---
title: "Dashboard / Control Panel for Administrators"
description: "Authelia Administrator Dashboard."
lead: "A dashboard or control panel for administrators to adjust system settings is easily one of the most impactful features we can implement."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  roadmap:
    parent: "active"
weight: 245
toc: true
aliases:
  - /r/admin-dashboard
---

This feature will pave the way to adding lots of useful administrator facing features. It will require in database
settings storage as well as some minimal traditional settings via files or environment variables.

This feature should not be confused with the [Dashboard / Control Panel for Users](dashboard-control-panel-for-users.md)
which is the dashboard for the user self-managing their own settings, as this feature is for managing the system
settings instead of the user settings.

## Stages

This section represents the stages involved in implementation of this feature. The stages are either in order of
implementation due to there being an underlying requirement to implement them in this order, or in their likely order
due to how important or difficult to implement they are.

### Design Stage

{{< roadmap-status >}}

Decide on a design.

### Initial Implementation

{{< roadmap-status >}}

Implement the pivotal elements of the design.

### Design Element: Segregation

{{< roadmap-status >}}

Allow the admin UI to be run as a separate process, on a different port, and at a different URL to Authelia itself.
Alternatively allow it to run as part of the main process and port for minimal configurations.

### User Management

{{< roadmap-status >}}

Manage user accounts with either the internal or LDAP authentication backends. Allow for creation, modification, and
deletion.

### Session Management

{{< roadmap-status >}}

Manage user sessions for all users.

### OpenID Connect 1.0 Client Management

{{< roadmap-status >}}

Manage client registrations via a web frontend.
