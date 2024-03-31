---
title: "Traefik v1"
description: "An integration guide for Authelia and the Traefik v1 reverse proxy"
summary: "A guide on integrating Authelia with the Traefik reverse proxy."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 371
toc: true
aliases:
  - /i/traefik/v1
  - /docs/deployment/supported-proxies/traefik1.x.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[Traefik] v1 is a reverse proxy supported by __Authelia__.

*__Important:__ When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only, and you need to understand the proxy
configuration and customize it to your needs. To-that-end, we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

*__Important:__ The included example is __NOT__ meant for production use. It's used expressly as an example to showcase
how you can configure multiple IP ranges. You should customize this example to fit your specific architecture and needs.
You should only include the specific IP address ranges of the trusted proxies within your architecture and should not
trust entire subnets unless that subnet only has trusted proxies and no other services.*

[Traefik] by default doesn't trust any other proxies requiring explicit configuration of which proxies are trusted
and removes potentially fabricated headers that are likely to lead to security issues, and it is difficult to configure
this incorrectly. This is an important security feature that is common with proxies with good security practices.

In the example we have four commented lines which configure `TrustedIPs` which show an example on adding the following
networks to the trusted proxy list in [Traefik]:

* 10.0.0.0/8
* 172.16.0.0/12
* 192.168.0.0/16
* fc00::/7

## Assumptions and Adaptation

This guide makes a few assumptions. These assumptions may require adaptation in more advanced and complex scenarios. We
can not reasonably have examples for every advanced configuration option that exists. The
following are the assumptions we make:

* Deployment Scenario:
  * Single Host
  * Authelia is deployed as a Container with the container name `authelia` on port `9091`
  * Proxy is deployed as a Container on a network shared with Authelia
* The above assumption means that Authelia should be accessible to the proxy on `http://authelia:9091` and as such:
  * You will have to adapt all instances of the above URL to be `https://` if Authelia configuration has a TLS key and
    certificate defined
  * You will have to adapt all instances of `authelia` in the URL if:
    * you're using a different container name
    * you deployed the proxy to a different location
  * You will have to adapt all instances of `9091` in the URL if:
    * you have adjusted the default port in the configuration
  * You will have to adapt the entire URL if:
    * Authelia is on a different host to the proxy
* All services are part of the `example.com` domain:
  * This domain and the subdomains will have to be adapted in all examples to match your specific domains unless you're
    just testing or you want to use that specific domain

## Implementation

[Traefik] utilizes the [ForwardAuth](../../reference/guides/proxy-authorization.md#forwardauth) Authz implementation. The
associated [Metadata](../../reference/guides/proxy-authorization.md#forwardauth-metadata) should be considered required.

The examples below assume you are using the default
[Authz Endpoints Configuration](../../configuration/miscellaneous/server-endpoints-authz.md) or one similar to the
following minimal configuration:

```yaml {title="configuration.yml"}
server:
  endpoints:
    authz:
      forward-auth:
        implementation: 'ForwardAuth'
```

## Configuration

Below you will find commented examples of the following docker deployment:

* [Traefik] 1.x
* Authelia portal
* Protected endpoint (Nextcloud)
* Protected endpoint with `Authorization` header for basic authentication (Heimdall)

The below configuration looks to provide examples of running [Traefik] v1 with labels to protect your endpoint
(Nextcloud in this case).

Please ensure that you also setup the respective [ACME](https://docs.traefik.io/v1.7/configuration/acme/) configuration
for your [Traefik] setup as this is not covered in the example below.

### Basic Authentication

Authelia provides the means to be able to authenticate your first factor via the `Proxy-Authorization` header.
Given that this is not compatible with [Traefik] 1.x you can call the __Authelia__ `/api/verify` endpoint with the
`auth=basic` query parameter to force a switch to the `Authentication` header.

##### docker-compose.yml

```yaml {title="docker-compose.yml"}
version: '3'

networks:
  net:
    driver: 'bridge'
services:
  traefik:
    image: 'traefik:v1.7.34-alpine'
    container_name: 'traefik'
    volumes:
      - '/var/run/docker.sock:/var/run/docker.sock'
    networks:
      - 'net'
    labels:
      - 'traefik.frontend.rule=Host:traefik.example.com'
      - 'traefik.port=8081'
    ports:
      - '80:80'
      - '443:443'
      - '8081:8081'
    restart: 'unless-stopped'
    command:
      - '--api'
      - '--api.entrypoint=api'
      - '--docker'
      - '--defaultEntryPoints=https'
      - '--logLevel=DEBUG'
      - '--traefikLog=true'
      - '--traefikLog.filepath=/var/log/traefik.log'
      - '--entryPoints=Name:http Address::80'
      - '--entryPoints=Name:https Address::443 TLS'
      ## See the Forwarded Header Trust section. Comment the above two lines, then uncomment and customize the next two lines to configure the TrustedIPs.
      # - '--entryPoints=Name:http Address::80 ForwardedHeaders.TrustedIPs:10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,fc00::/7 ProxyProtocol.TrustedIPs:10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,fc00::/7'
      # - '--entryPoints=Name:https Address::443 TLS ForwardedHeaders.TrustedIPs:10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,fc00::/7 ProxyProtocol.TrustedIPs:10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,fc00::/7'
      - '--entryPoints=Name:api Address::8081'
  authelia:
    image: 'authelia/authelia'
    container_name: 'authelia'
    volumes:
      - '/path/to/authelia:/config'
    networks:
      - 'net'
    labels:
      - 'traefik.frontend.rule=Host:auth.example.com'
    expose:
      - 9091
    restart: 'unless-stopped'
    environment:
      TZ: 'Australia/Melbourne'
  nextcloud:
    image: 'linuxserver/nextcloud'
    container_name: 'nextcloud'
    volumes:
      - '/path/to/nextcloud/config:/config'
      - '/path/to/nextcloud/data:/data'
    networks:
      - 'net'
    labels:
      - 'traefik.frontend.rule=Host:nextcloud.example.com'
      - 'traefik.frontend.auth.forward.address=http://authelia:9091/api/authz/forward-auth'
      ## The following commented line is for configuring the Authelia URL in the proxy. We strongly suggest this is
      ## configured in the Session Cookies section of the Authelia configuration.
      # - 'traefik.frontend.auth.forward.address=http://authelia:9091/api/authz/forward-auth?authelia_url=https%3A%2F%2Fauth.example.com%2F'
      - 'traefik.frontend.auth.forward.trustForwardHeader=true'
      - 'traefik.frontend.auth.forward.authResponseHeaders=Remote-User,Remote-Groups,Remote-Email,Remote-Name'
    expose:
      - 443
    restart: 'unless-stopped'
    environment:
      PUID: '1000'
      PGID: '1000'
      TZ: 'Australia/Melbourne'
  heimdall:
    image: 'linuxserver/heimdall'
    container_name: 'heimdall'
    volumes:
      - '/path/to/heimdall/config:/config'
    networks:
      - 'net'
    labels:
      - 'traefik.frontend.rule=Host:heimdall.example.com'
      - 'traefik.frontend.auth.forward.address=http://authelia:9091/api/authz/forward-auth/basic'
      - 'traefik.frontend.auth.forward.trustForwardHeader=true'
      - 'traefik.frontend.auth.forward.authResponseHeaders=Remote-User,Remote-Groups,Remote-Email,Remote-Name'
    expose:
      - 443
    restart: 'unless-stopped'
    environment:
      PUID: '1000'
      PGID: '1000'
      TZ: 'Australia/Melbourne'
```

## See Also

* [Traefik v1 Documentation](https://doc.traefik.io/traefik/v1.7/)
* [Traefik v1 All Available Options](https://doc.traefik.io/traefik/v1.7/configuration/entrypoints/#all-available-options)
* [Forwarded Headers]

[Traefik]: https://docs.traefik.io/v1.7/
[Forwarded Headers]: forwarded-headers
