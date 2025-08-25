---
title: "Server Asset Overrides"
description: "A reference guide on overriding server assets"
summary: "This section contains reference documentation for Authelia's server asset override capabilities."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 220
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

This guide effectively documents the usage of the
[asset_path](../../configuration/miscellaneous/server.md#asset_path) server configuration option.

## Structure

```console
/config/assets/
├── favicon.ico
├── logo.png
└── locales/<lang>[-[variant]]/<namespace>.json
```

## Assets

|        Asset        |  File Name  | Directory |          Notes          |
|:-------------------:|:-----------:|:---------:|:-----------------------:|
|       Favicon       | favicon.ico |    No     |           N/A           |
|        Logo         |  logo.png   |    No     |           N/A           |
| Translation Locales |   locales   |    Yes    | see [locales](#locales) |

## locales

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Currently users can only override languages that already exist in this list either by overriding
the language itself, or adding a variant form of that language. If you'd like support for another language feel free
to make a PR. We also encourage people to make PR's for variants where the difference in the variants is significant.
{{< /callout >}}

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Users wishing to override the locales files should be aware that we do not provide any guarantee
that the file will not change in a breaking way between releases as per our [Versioning Policy](../../policies/versioning.md). Users who are planning to
utilize these overrides should either check for changes to the files in the
[en](https://github.com/authelia/authelia/tree/master/internal/server/locales/en) translation prior to upgrading or
[Contribute](../../contributing/prologue/translations.md) their translation to ensure it is maintained.
{{< /callout >}}

The locales directory holds folders of internationalization locales. This directory can be utilized to override these
locales. They are the names of locales that are returned by the `navigator.language` ECMAScript command. These are
generally those in the [RFC5646 / BCP47 Format](https://datatracker.ietf.org/doc/html/rfc5646) specifically the language
codes from [Crowdin](https://support.crowdin.com/api/language-codes/).

Each directory has JSON files which you can explore the format of in the
[internal/server/locales](https://github.com/authelia/authelia/tree/master/internal/server/locales) directory on
GitHub. The important part is the key names you wish to override.

A full example for the `en-US` locale for the portal namespace is `locales/en-US/portal.json`.

Languages in browsers are supported in two forms. In their language only form such as `en` for English, and in their
variant form such as `en-AU` for English (Australian). If a user has the browser language `en-AU` we automatically load
the `en` and `en-AU` languages, where any keys in the `en-AU` language take precedence over the `en` language, and the
translations for the `en` language only applying when a translation from `en-AU` is not available.

### Namespaces

Each file in a locale directory represents a translation namespace. The list of current namespaces are below:

| Namespace |       Purpose       |
|:---------:|:-------------------:|
|  portal   | Portal Translations |

### Supported Languages

List of supported languages and variants:

{{% table-i18n-overrides %}}

More information may be available from the [Internationalization Reference Guide](./internationalization.md).
