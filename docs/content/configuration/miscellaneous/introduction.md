---
title: "Miscellaneous"
description: "Miscellaneous Configuration."
summary: "Authelia has a few config items that don't fit into their own area. This describes these options."
date: 2020-02-29T01:43:59+01:00
draft: false
images: []
weight: 199100
toc: true
aliases:
  - /docs/configuration/miscellaneous.html
  - /docs/configuration/theme.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
certificates_directory: '/config/certs/'
default_redirection_url: 'https://home.example.com:8080/'
theme: 'light'
```

## Options

This section describes the individual configuration options.

### certificates_directory

This option defines the location of additional certificates to load into the trust chain specifically for Authelia.
This currently affects both the SMTP notifier and the LDAP authentication backend. The certificates should all be in the
PEM format and end with the extension `.pem`, `.crt`, or `.cer`. You can either add the individual certificates public
key or the CA public key which signed them (don't add the private key).

### default_redirection_url

{{< confkey type="string" required="no" >}}

*__Important Note:__ You should configure the domain specific redirection URL's in the
[session](../session/introduction.md#default_redirection_url) configuration instead of using this option.*

The default redirection URL is the URL where users are redirected when Authelia cannot detect the target URL where the
user was heading.

In a normal authentication workflow, a user tries to access a website and they get redirected to the sign-in portal in
order to authenticate. Since the user initially targeted a website, the portal knows where the user was heading and
can redirect them after the authentication process. However, when a user visits the sign in portal directly, the portal
considers the targeted website is the portal. In that case and if the default redirection URL is configured, the user is
redirected to that URL. If not defined, the user is not redirected after authentication.

### default_2fa_method

{{< confkey type="string" default="totp" required="no" >}}

Sets the default second factor method for users. This must be blank or one of the enabled methods. New users will by
default have this method selected for them. In addition if this was configured to `webauthn` and a user had the `totp`
method, and the `totp` method was disabled in the configuration, the users' method would automatically update to the
`webauthn` method.

Options are:

* totp
* webauthn
* mobile_push

```yaml {title="configuration.yml"}
default_2fa_method: totp
```

### theme

{{< confkey type="string " default="light" required="no" >}}

There are currently 3 available themes for Authelia:

* light (default)
* dark
* grey

To enable automatic switching between themes, you can set `theme` to `auto`. The theme will be set to either `dark` or
`light` depending on the user's system preference which is determined using media queries. To read more technical
details about the media queries used, read the
[MDN](https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme).
