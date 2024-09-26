---
title: "Frequently Asked Questions"
description: "Frequently Asked Questions regarding integrating the Authelia Trusted Header SSO implementation with applications"
summary: "Frequently Asked Questions regarding integrating the Authelia Trusted Header SSO implementation with applications."
date: 2024-09-19T05:54:35+10:00
draft: false
images: []
weight: 615
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Questions

The following section lists individual questions.

### What are the particular security mechanisms involved in the Trusted Header SSO implementation?

The Trusted Header SSO implementation relies on fairly trivial to implement mechanisms where the headers are implicitly
trusted by backend applications. This simplicity is both a blessing and potential problem.

As the headers are implicitly trusted it is important to ensure that the application trusting the headers is only
accessible via the proxy so that the proxy can strip or override the original request headers so that they can only
come from the Authelia authorization server responses.

For example we suggest having the hosted applications inaccessible to the public in any way and only accessible to your
proxy. This can be done by not specifying the docker ports option, only listening on 127.0.0.1 (or another IP only
accessible to the proxy and other local applications) and either hosting the application on the same host as the proxy
or using a VPN to communicate with it, etc.

In instances where this can not be achieved we strongly urge users to use and implement a mechanism designed for such
complex architectures such as [OpenID Connect 1.0](../openid-connect/introduction.md).
