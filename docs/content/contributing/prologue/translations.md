---
title: "Translations"
description: "Information on contributing translations to the Authelia project."
summary: "Authelia has translations for many using facing areas of the web portal. Contributing to these translations is a very easy process."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 130
toc: true
aliases:
  - /docs/contributing/translations.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The way the translation process is facilitated is via [Crowdin]. We encourage members of the community to
[join the Authelia Crowdin project](https://translate.authelia.com) and help translate their preferred language.

## New Languages

If the language you wish to translate is not on [Crowdin] then you have a few options:

1. Ask for the language to be added via the [Crowdin] interface.
2. Ask a maintainer to add it via one of the [contact options](../../overview/project/contact.md).
3. Make a pull request directly on GitHub modifying the translation files within
[this directory](https://github.com/authelia/authelia/tree/master/internal/server/locales).

## Overrides

Users can override translations easily locally using the
[assets](../../configuration/miscellaneous/server.md#asset_path) directory. This is useful if you wish to perform a
translation and see if it looks correct in the browser.

[Crowdin]: https://translate.authelia.com
