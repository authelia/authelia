---
title: "Accessibility"
description: "Authelia Development Accessibility Guidelines"
summary: "This section covers the accessibility guidelines we aim to respect during development."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 350
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
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

Recommendations on resolutions which are common:

- Desktop/Laptop:
  1. 1920x1080
  2. 1366x768
  3. 2560x1440
  4. 1280x720
- Tablet Devices (With Touch and Landscape):
  1. 768x1024
  2. 810x1080
  3. 800x1280
- Mobile Devices (With Touch and Landscape):
  1. 360x800
  2. 390x844
  3. 414x896
  4. 412x915
