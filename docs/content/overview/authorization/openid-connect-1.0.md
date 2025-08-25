---
title: "OpenID Connect 1.0"
description: "OpenID Connect 1.0 is a authorization identity framework supported by Authelia."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 330
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[OpenID Connect 1.0](https://openid.net/connect/) is a authorization identity framework supported by Authelia. You can
configure your applications to use Authelia as an [OpenID Connect 1.0 Provider](https://openid.net/connect/). We do not
currently operate as an [OpenID Connect 1.0 Relying Party](https://openid.net/connect/). This like all single-sign on
technologies requires support by the protected application.

See the [OpenID Connect 1.0 Provider Configuration Guide](../../configuration/identity-providers/openid-connect/provider.md),  and the
[OpenID Connect 1.0 Integration Guide](../../integration/openid-connect/introduction.md) for more information.

## Protocol Support

The [Support Chart](../../integration/openid-connect/introduction.md#support-chart) lists the OAuth 2.0 and OpenID
Connect 1.0 specifications and protocols we support.

## OpenID Certified™

Authelia is [OpenID Certified™] to conform to the [OpenID Connect™ protocol].

{{< figure src="/images/oid-certification.jpg" class="center" process="resize 200x" >}}

For more information please see the
[OpenID Connect 1.0 Integration Documentation](../../integration/openid-connect/introduction.md#openid-certified).

[OpenID Certified™]: https://openid.net/certification/
[OpenID Connect™ protocol]: https://openid.net/developers/how-connect-works/
