---
title: "Custom CSS"
description: "Custom CSS Configuration."
summary: "This describes the configuration for applying a custom CSS file to the Authelia portal."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 199110
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
custom_css: ''
```

## Options

This section describes the individual configuration options.

### custom_css

{{< confkey type="string" default="''" required="no" >}}

The `custom_css` option allows administrators to provide a URL to a custom CSS file that will be loaded by the Authelia
portal. This is useful for customizing the look and feel of the portal without needing to rebuild the frontend assets.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Users wishing to utilize this feature should be aware that we do not provide any guarantee that the CSS selectors or
the portal structure will not change in a breaking way between releases as per our
[Versioning Policy](../../policies/versioning.md).
{{< /callout >}}

This value must be an absolute path (starting with `/`) or an `https://` URL. Using an absolute
URL makes it easier to include images or other assets while complying with the
[Content Security Policy](#content-security-policy).

#### Content Security Policy

When using an external `https://` URL for `custom_css`, *Authelia* will automatically attempt to add the host of that URL
to the `style-src` directive of the default [Content Security Policy](server.md#content-security-policy) (CSP).

However, if you are using a custom `server.headers.csp_template`, or if your custom CSS references other assets such as
images or fonts from an external host, you must manually update your CSP template to allow these hosts in the relevant
directives (e.g., `img-src`, `font-src`).

