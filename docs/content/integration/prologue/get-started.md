---
title: "Get started"
description: "A getting started guide for Authelia."
summary: "This document serves as a get started guide for Authelia. It contains links to various sections and has some key notes in questions frequently asked by people looking to perform setup for the first time."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 120
toc: true
aliases:
  - '/docs'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

It's important to note that this guide has a layout which we suggest as the best order in areas to tackle, but you may
obviously choose a different path if you are so inclined.

## Prerequisites

The most important prerequisite that users understand that there is no single way to deploy software similar to
Authelia. We provide as much information as possible for users to configure the critical parts usually in the most
common scenarios however those using more advanced architectures are likely going to have to adapt. We can generally
help with answering less specific questions about this and it may be possible if provided adequate information more
specific questions may be answered.

1. Authelia *__MUST__* be served via the `https` scheme. This is not optional even for testing. This is a deliberate
   design decision to improve security directly (by using encrypted communication) and indirectly by reducing complexity.
2. Your proxy configuration for Authelia *__MUST__* include all of the
   [Required Headers](../proxies/introduction.md#required-headers).
3. When using the [Proxy Authorization](../../reference/guides/proxy-authorization.md) the proxy must include all of the
   required headers for the specific implementation that has been configured, similar to the curated examples.

### Forwarded Authentication

Forwarded Authentication is a simple per-request authorization flow that checks the metadata of a request and a session
cookie to determine if a user must be forwarded to the authentication portal.

In addition to the `https` scheme requirement for Authelia itself:

1. Due to the fact a cookie is used, it's an intentional design decision that *__ALL__* applications/domains protected
   via this method *__MUST__* use secure schemes (`https` and `wss`) for all of their communication.

### OpenID Connect 1.0

No additional requirements other than the use of the `https` scheme for Authelia itself exist excluding those mandated
by the relevant specifications.

## Important Notes

{{< callout context="danger" title="Important Notes" icon="outline/alert-octagon" >}}
The following section has general important notes for users getting started.
{{< /callout >}}

- When integrating Authelia with a proxy users should read the specific
  [Proxy Integration Important Notes](../proxies/introduction.md#important-notes).

## Documentation Variables

Some of the values presented in the documentation can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

It's important to customize the configuration for *Authelia* in advance of deploying it. The configuration is static and
not configured via web GUI. You can find a configuration template named {{< github-link path="config.template.yml" >}}
on GitHub which can be used as a basis for configuration, alternatively *Authelia* will write this template relevant for
your version the first time it is started. Users should expect that they have to configure elements of this file as part
of initial setup.

The important sections to consider in initial configuration are as follows:

1. [jwt_secret](../../configuration/identity-validation/reset-password.md#jwt_secret) which is used to sign identity
   verification emails for the reset password flow if enabled.
2. [authentication_backend](../../configuration/first-factor/introduction.md) which you must pick between
   [LDAP](../../configuration/first-factor/ldap.md) and a [YAML File](../../configuration/first-factor/file.md) and is
   essential for users to authenticate.
3. [storage](../../configuration/storage/introduction.md) which you must pick between the SQL Storage Providers, the
   recommended one for testing and lite deployments is [SQLite3](../../configuration/storage/sqlite.md) and the
   recommended one for production deployments otherwise is [PostgreSQL](../../configuration/storage/postgres.md).
4. [session](../../configuration/session/introduction.md) which is used to configure:
   1. The [session cookies](../../configuration/session/introduction.md#cookies) section should be configured with every SSO domain (none of them can be a suffix of
      the others) you wish to protect, and the most important options in this section are
      [domain](../../configuration/session/introduction.md#domain) and
      [authelia_url](../../configuration/session/introduction.md#authelia_url).
   2. The [secret](../../configuration/session/introduction.md#secret) is the most important, and
       [redis](../../configuration/session/redis.md) is recommended for production environments.
5. [notifier](../../configuration/notifications/introduction.md) which is used to send 2FA registration emails etc,
   there is an option for local file delivery but the [SMTP](../../configuration/notifications/smtp.md) option is
   recommended for production and you must only configure one of these.
6. [access_control](../../configuration/security/access-control.md) is also important but should be configured with a
   very basic policy to begin with. Something like:

```yaml {title="configuration.yml"}
access_control:
  default_policy: deny
  rules:
    - domain: '*.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: one_factor
```

## Deployment

There are several methods of deploying *Authelia* and we recommend reading the
[Deployment Documentation](../deployment/introduction.md) in order to perform deployment.

## Proxy Integration

The default method of utilizing *Authelia* is via the [Proxy Integrations](../proxies/introduction.md). It's
recommended that you read the relevant [Proxy Integration Documentation](../proxies/introduction.md).

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
When your [Deployment](#deployment) is on [Kubernetes](../kubernetes/introduction.md) we
recommend viewing the dedicated [Kubernetes Documentation](../kubernetes/introduction.md) prior to viewing the
[Proxy Integration Documentation](../proxies/introduction.md).
{{< /callout >}}

## Additional Useful Links

See the [Frequently Asked Questions](../../reference/guides/frequently-asked-questions.md) for helpful sections of the
documentation which may answer specific questions.

## Moving to Production

We consider it important to do several things in moving to a production environment.

1. Move all [secret values](../../configuration/methods/secrets.md#environment-variables) out of the configuration and
   into [secrets](../../configuration/methods/secrets.md).
2. Spend time understanding [access control](../../configuration/security/access-control.md) and granularly configure it
   to your requirements.
3. Review the [Security Measures](../../overview/security/measures.md) and
   [Threat Model](../../overview/security/threat-model.md) documentation.
4. Ensure you have reviewed the [Forwarded Headers](../proxies/forwarded-headers/index.md) documentation to ensure your
   proxy is not allowing insecure headers to be passed to *Authelia*.
5. Review the other [Configuration Options](../../configuration/prologue/introduction.md).
