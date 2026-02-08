---
title: "Dashboard / Control Panel for Users"
description: "Authelia User Dashboard."
summary: "A dashboard or control panel for users to adjust their settings is easily one of the most impactful features we can implement."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 340
toc: true
aliases:
  - /r/dashboard
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

This feature will pave the way to adding lots of useful user facing features.

It will be important when we implement:
- WebAuthn features like passwordless authentication allowing users to intentionally register a passwordless credential.
- Session management features.
- Many other user self-service related features.

This feature should not be confused with the [Dashboard / Control Panel for Administrators](dashboard-control-panel-and-cli-for-admins.md)
which is the dashboard for managing the system settings, as this feature is for the user self-managing their own
settings instead of the system settings.

## Stages

This section represents the stages involved in implementation of this feature. The stages are either in order of
implementation due to there being an underlying requirement to implement them in this order, or in their likely order
due to how important or difficult to implement they are.

### Initial Implementation

{{< roadmap-status stage="complete" version="v4.38.0" >}}

Add control panel with the ability to control all of the current settings, with the added benefit of being able to
register multiple WebAuthn keys.

Users should also be able to view all of their registered devices, and revoke them individually.

### Password Reset

{{< roadmap-status stage="complete">}}

Add a method for users to reset their password given they know their current password.

### Language Option

{{< roadmap-status >}}

Allow users to override the detected language in their browser and choose from one of the available languages.

### Session Management

{{< roadmap-status >}}

Add ability for users to view their own sessions and end them.

### Much More

The practical usage of this is endless.
