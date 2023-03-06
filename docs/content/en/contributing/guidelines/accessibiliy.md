---
title: "Accessibility"
description: "Authelia Development Accessibility Guidelines"
lead: "This section covers the accessibility guidelines we aim to respect during development."
date: 2023-03-06T11:42:13+11:00
draft: false
images: []
menu:
  contributing:
    parent: "guidelines"
weight: 350
toc: true
---

## Backend

There are no specific guidelines for backend accessibility other than ensuring there are reasonable logging and this is
extremely subjective.


## Frontend

### Translations

We aim to ensure as much of the web frontend information displayed to users is translated by default. This allows for
both automatic and manual translations by the community to be contributed to the code base. In addition it allows for
admins to locally override these values.

### Responsive Design

We aim to make the web frontend responsive to multiple screen resolutions. There are a few guidelines which we aim to
abide by:

- The available space is utilized efficiently in order to avoid scrolling where possible.
- The user only has to scroll in one direction to view available information. This direction should always be
  vertically.
