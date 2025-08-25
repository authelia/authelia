---
title: "Privacy Policy"
description: "Privacy Policy Configuration."
summary: "This describes a section of the configuration for enabling a Privacy Policy link display."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 199100
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
privacy_policy:
  enabled: false
  require_user_acceptance: false
  policy_url: ''
```

## Options

This section describes the individual configuration options.

### enabled

{{< confkey type="boolean" default="false" required="no" >}}

Enables the display of the Privacy Policy link.

### require_user_acceptance

{{< confkey type="boolean" default="false" required="no" >}}

Requires users accept per-browser the Privacy Policy via a Dialog Drawer at the bottom of the page. The fact they have
accepted is recorded and checked in the browser
[localStorage](https://developer.mozilla.org/en-US/docs/Web/API/Window/localStorage).

If the user has not accepted the policy they should not be able to interact with the Authelia UI via normal means.

Administrators who are required to abide by the [GDPR] or other privacy laws should be advised that
[OpenID Connect 1.0](../identity-providers/openid-connect/provider.md) clients configured with the `implicit` consent
mode are unlikely to trigger the display of the Authelia UI if the user is already authenticated.

We won't be adding checks like this to the `implicit` consent mode when that mode in particular is unlikely to be
compliant with those laws, and that mode is not strictly compliant with the OpenID Connect 1.0 specifications. It is
therefore recommended if `require_user_acceptance` is enabled then administrators should avoid using the `implicit`
consent mode or do so at their own risk.

### policy_url

{{< confkey type="string" required="situational" >}}

The privacy policy URL is a URL which optionally is displayed in the frontend linking users to the administrators
privacy policy. This is useful for users who wish to abide by laws such as the [GDPR].
Administrators can view the particulars of what _Authelia_ collects out of the box with our
[Privacy Policy](https://www.authelia.com/privacy/#application).

This value must be an absolute URL, and must have the `https://` scheme.

This option is required if the [enabled](#enabled) option is true.

[GDPR]: https://gdpr-info.eu/

_**Example:**_

```yaml {title="configuration.yml"}
privacy_policy:
  enabled: true
  policy_url: 'https://www.{{< sitevar name="domain" nojs="example.com" >}}/privacy-policy'
```
