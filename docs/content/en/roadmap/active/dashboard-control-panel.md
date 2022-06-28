---
title: "Dashboard / Control Panel"
description: "Authelia Dashboard Implementation"
lead: "A dashboard or control panel for users and administrators to adjust their settings or Authelia's settings is easily one of the most impactful features we can implment."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  roadmap:
    parent: "active"
weight: 240
toc: true
aliases:
  - /r/dashboard
---

This feature has several major impacts on other roadmap items. For example several OpenID Connect features would greatly
benefit from a dashboard. It would also be important when we implement WebAuthn features like passwordless
authentication allowing users to intentionally register a passwordless credential.

## Stages

This section represents the stages involved in implementation of this feature. The stages are either in order of
implementation due to there being an underlying requirement to implement them in this order, or in their likely order
due to how important or difficult to implement they are.

### Initial Implementation

{{< roadmap-status >}}

Add control panel with the ability to control all of the current settings, with the added benefit of being able to
register multiple WebAuthn keys.

Users should also be able to view all of their registered devices, and revoke them individually.

### Password Reset

{{< roadmap-status >}}

Add a method for users to reset their password given they know their current password.

### Language Option

{{< roadmap-status >}}

Allow users to override the detected language in their browser and choose from one of the available languages.

### Session Management

{{< roadmap-status >}}

Add ability for users to view their own sessions and end them, administrators the ability to view all sessions and end
them, and for administrators to be notified of and add/view/remove bans on users.

### Much More

The practical usage of this is endless.
