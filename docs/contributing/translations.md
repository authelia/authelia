---
layout: default
title: Translations
parent: Contributing
nav_order: 6
---

# Translations

Authelia has translations for many using facing areas of the web portal. Contributing to these translations is a very
easy process.

The way the translation process is facilitated is via [Crowdin]. We encourage members of the community to
[join the Authelia Crowdin project](https://crwd.in/authelia) and help translate their preferred language.

## Adding a New Language

If the language you wish to translate is not on [Crowdin] then you can either reach out to one of the maintainers and
ask them to add it or you may make a pull request directly on GitHub. The translation files are stored within 
[this directory](https://github.com/authelia/authelia/tree/master/internal/server/locales).

## Overrides

Users can override translations easily locally using the [assets](../configuration/server.md#locales) directory. This is
useful if you wish to perform a translation and see if it looks correct in the browser.


[Crowdin]: https://crowdin.com/project/authelia