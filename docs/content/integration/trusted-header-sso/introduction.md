---
title: "Trusted Header SSO"
description: "Trusted Header SSO Integration"
summary: "An introduction into integrating Authelia with an application which implements authentication via trusted headers."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 410
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia will respond to requests via the forward authentication flow with specific headers that can be utilized by some
applications to perform authentication. This section of the documentation discusses how to integrate these products with
this model.

Please see the [proxy integration](../proxies/introduction.md) for more information on how to return these headers to
the application.

## Terminology

This authentication method is referred to by many names; notably `trusted header authentication`,
`header authentication`, `header sso`, and probably many more.

## Specifics

The headers are not intended to be returned to a browser, instead these headers are meant for internal communications
only. These headers are returned to the reverse proxy and then injected into the request which the reverse proxy makes
to the application's http endpoints.

This allows these applications to decide if they wish to trust these headers, and if they do trust them, perform some
form of authentication flow. This flow usually takes the form of automatically logging users into the application,
however it can vary depending on how the application decides to do this.

## Response Headers

The following table represents the response headers that Authelia's `/api/authz/*` or `/api/verify` endpoints return
which can be forwarded over a trusted network via the reverse proxy when using the forward authentication flow. See
the [Server Authz Endpoints](../../configuration/miscellaneous/server-endpoints-authz.md) configuration documentation
and linked reference articles for more information on these endpoints.

|    Header     |      Description / Notes       |                         Example                         |
|:-------------:|:------------------------------:|:-------------------------------------------------------:|
|  Remote-User  |       The users username       |                          john                           |
| Remote-Groups | The groups the user belongs to |                        admin,dev                        |
|  Remote-Name  |     The users display name     |                       John Smith                        |
| Remote-Email  |    The users email address     | jsmith@{{< sitevar name="domain" nojs="example.com" >}} |

## Forwarding the Response Headers

It's essential if you wish to utilize the trusted header single sign-on flow that you forward the
[response headers](#response-headers) via the reverse proxy to the backend application, not the browser. Please refer to
the relevant [proxy documentation](../proxies/introduction.md) for more information.

## Trusted Remote Networks

Several applications which implement this authentication method allow or require you to configure a list of IP addresses
which are trusted to deliver these headers. It is our recommendation that you configure this even if it is optional.

The application itself will have a way to detect this IP address and most implementations utilize the TCP source address
as this is the most appropriate. This is the TCP source address of your *proxy*, it is __not__ the TCP source address of
Authelia. This is because headers may be returned by Authelia to the proxy, however the backend application is *not
able* to determine this reliably, instead the TCP source address of the request to the application is used, which is
made by the reverse proxy. This also means your proxy must ensure only Authelia is setting these headers, and any other
headers are never forwarded to the backend and are instead replaced by the Authelia headers.

In some environments the TCP source address of the proxy may be difficult to determine. For example in a docker
environment a container may be a member of multiple networks. This means the TCP source address that you must use is the
IP address of the proxy on the network that both the proxy and the application are members of. In this environment it
is also imperative that you utilize a static IP for the proxy container as configuring an entire docker network is not
considered secure as any compromised container may be able to be used to bypass authentication for any container
configured to use this authentication flow.

### Docker

In a [Docker] environment a [container] may be a member of multiple networks. This means the TCP source address that you
must use is the IP address of the proxy on the [Docker Network] that both the proxy and the application are members of.
In this environment it is also imperative that you utilize a static IP for the proxy [container] as configuring an
entire [Docker Network] is not considered secure as any compromised [container] may be able to be used to bypass
authentication for any [container] configured to use this authentication flow.

The following command will print out the IP for a container named `traefik` on the `authelia` network:

```bash
docker inspect -f '{{.NetworkSettings.Networks.authelia.IPAddress}}' traefik
```

The following command will print out all network names and the associated IP address for a container named `traefik`:

```bash
docker inspect -f '{{range $network, $config := .NetworkSettings.Networks}}{{ $network }}: {{ $config.IPAddress }} {{end}}' traefik
```

### Kubernetes

In a [Kubernetes] the TCP source address is likely the [Pod] IP. Generally these cannot be static and you should instead
ensure a [Pod] that is configured to utilize this method is secured via one of the various means available such as a
service mesh. Configuring any of these security methods is well beyond the scope of this document.

[Docker]: https://docker.com
[Kubernetes]: https://kubernetes.io/
[Pod]: https://kubernetes.io/docs/concepts/workloads/pods/
[container]: https://www.docker.com/resources/what-container/
[Docker Network]: https://docs.docker.com/network/
