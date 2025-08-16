---
title: "Traefik"
description: "An integration guide for Authelia and the Traefik reverse proxy"
summary: "A guide on integrating Authelia with the Traefik reverse proxy."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 370
toc: true
aliases:
  - /i/traefik
  - /docs/deployment/supported-proxies/traefik2.x.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[Traefik] is a reverse proxy supported by __Authelia__.

*__Important:__ When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only, and you need to understand the proxy
configuration and customize it to your needs. To-that-end, we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Requirements

Authelia by default only generally provides support for versions of products that are also supported by their respective
developer. As such we only support the versions [Traefik] officially provides support for. The versions and lifetime
of support for [Traefik] can be read about in the official
[Traefik Deprecation Notices](https://doc.traefik.io/traefik/deprecation/releases/) documentation.

It should be noted that while these are the listed versions that are supported you may have luck with older versions.

We can officially guarantee the following versions of [Traefik] as these are the versions we perform integration testing
with at the current time:

{{% supported-product product="traefik" format="* [Traefik $version](https://github.com/traefik/traefik/releases/tag/$version)" %}}

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

In the example we have four commented lines which configure `trustedIPs` which show an example on adding the following
networks to the trusted proxy list in [Traefik]:

* 10.0.0.0/8
* 172.16.0.0/12
* 192.168.0.0/16
* fc00::/7

See the [Entry Points](https://doc.traefik.io/traefik/routing/entrypoints) documentation for more information.

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

The examples below also assume you are using the modern
[Session Configuration](../../configuration/session/introduction.md) which includes the `domain`, `authelia_url`, and
`default_redirection_url` as a subkey of the `session.cookies` key as a list item. Below is an example of the modern
configuration as well as the legacy configuration for context.

{{< sessionTabs "Generate Random Password" >}}
{{< sessionTab "Modern" >}}
```yaml {title="configuration.yml"}
session:
  cookies:
    - domain: '{{</* sitevar name="domain" nojs="example.com" */>}}'
      authelia_url: 'https://{{</* sitevar name="subdomain-authelia" nojs="auth" */>}}.{{</* sitevar name="domain" nojs="example.com" */>}}'
      default_redirection_url: 'https://www.{{</* sitevar name="domain" nojs="example.com" */>}}'
```
{{< /sessionTab >}}
{{< sessionTab "Legacy" >}}
```yaml {title="configuration.yml"}
default_redirection_url: 'https://www.{{</* sitevar name="domain" nojs="example.com" */>}}'
session:
  domain: '{{</* sitevar name="domain" nojs="example.com" */>}}'
```
{{< /sessionTab >}}
{{< /sessionTabs >}}

## Configuration

Below you will find commented examples of the following docker deployment:

* [Traefik]
* Authelia portal
* Protected endpoint (Nextcloud)
* Protected endpoint with [Authorization] header for basic authentication (Heimdall)

The below configuration looks to provide examples of running [Traefik] 3.x with labels to protect your endpoint
(Nextcloud in this case).

Please ensure that you also setup the respective [ACME configuration](https://docs.traefik.io/https/acme/) for your
[Traefik] setup as this is not covered in the example below.

## Assumptions and Adaptation

This guide makes a few assumptions. These assumptions may require adaptation in more advanced and complex scenarios. We
can not reasonably have examples for every advanced configuration option that exists. Some of these values can
automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

The following are the assumptions we make:

* Deployment Scenario:
  * Single Host
  * Authelia is deployed as a Container with the container name `{{< sitevar name="host" nojs="authelia" >}}` on port `{{< sitevar name="port" nojs="9091" >}}`
  * Proxy is deployed as a Container on a network shared with Authelia
* The above assumption means that Authelia should be accessible to the proxy on `{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}` and as such:
  * You will have to adapt all instances of the above URL to be `https://` if Authelia configuration has a TLS key and
    certificate defined
  * You will have to adapt all instances of `{{< sitevar name="host" nojs="authelia" >}}` in the URL if:
    * you're using a different container name
    * you deployed the proxy to a different location
  * You will have to adapt all instances of `{{< sitevar name="port" nojs="9091" >}}` in the URL if:
    * you have adjusted the default port in the configuration
  * You will have to adapt the entire URL if:
    * Authelia is on a different host to the proxy
* All services are part of the `{{< sitevar name="domain" nojs="example.com" >}}` domain:
  * This domain and the subdomains will have to be adapted in all examples to match your specific domains unless you're
    just testing or you want to use that specific domain

### Docker Compose

This is an example configuration using [docker compose] labels:

```yaml {title="compose.yml"}
---
networks:
  net:
    driver: bridge
services:
  traefik:
    container_name: 'traefik'
    image: 'traefik:v3.5'
    restart: 'unless-stopped'
    command:
      - '--api=true'
      - '--api.dashboard=true'
      - '--api.insecure=false'
      - '--pilot.dashboard=false'
      - '--global.sendAnonymousUsage=false'
      - '--global.checkNewVersion=false'
      - '--log=true'
      - '--log.level=DEBUG'
      - '--log.filepath=/config/traefik.log'
      - '--providers.docker=true'
      - '--providers.docker.exposedByDefault=false'
      - '--entryPoints.http=true'
      - '--entryPoints.http.address=:8080/tcp'
      - '--entryPoints.http.http.redirections.entryPoint.to=https'
      - '--entryPoints.http.http.redirections.entryPoint.scheme=https'
      ## Please see the Forwarded Header Trust section of the Authelia Traefik Integration documentation.
      # - '--entryPoints.http.forwardedHeaders.trustedIPs=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,fc00::/7'
      # - '--entryPoints.http.proxyProtocol.trustedIPs=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,fc00::/7'
      - '--entryPoints.http.forwardedHeaders.insecure=false'
      - '--entryPoints.http.proxyProtocol.insecure=false'
      - '--entryPoints.https=true'
      - '--entryPoints.https.address=:8443/tcp'
      ## Please see the Forwarded Header Trust section of the Authelia Traefik Integration documentation.
      # - '--entryPoints.https.forwardedHeaders.trustedIPs=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,fc00::/7'
      # - '--entryPoints.https.proxyProtocol.trustedIPs=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,fc00::/7'
      - '--entryPoints.https.forwardedHeaders.insecure=false'
      - '--entryPoints.https.proxyProtocol.insecure=false'
    networks:
      net: {}
    ports:
      - '80:8080'
      - '443:8443'
    volumes:
      - '/var/run/docker.sock:/var/run/docker.sock'
      - '${PWD}/data/traefik:/config'
    labels:
      traefik.enable: 'true'
      traefik.http.routers.api.rule: 'Host(`traefik.{{< sitevar name="domain" nojs="example.com" >}}`)'
      traefik.http.routers.api.entryPoints: 'https'
      traefik.http.routers.api.tls: 'true'
      traefik.http.routers.api.service: 'api@internal'
      traefik.http.routers.api.middlewares: 'authelia@docker'
  authelia:
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    image: 'authelia/authelia'
    restart: 'unless-stopped'
    networks:
      net: {}
    volumes:
      - '${PWD}/data/authelia/config:/config'
    environment:
      TZ: "Australia/Melbourne"
    labels:
      traefik.enable: 'true'
      traefik.http.routers.authelia.rule: 'Host(`{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`)'
      traefik.http.routers.authelia.entryPoints: 'https'
      traefik.http.routers.authelia.tls: 'true'
      traefik.http.middlewares.authelia.forwardAuth.address: '{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/forward-auth'
      ## The following commented line is for configuring the Authelia URL in the proxy. We strongly suggest this is
      ## configured in the Session Cookies section of the Authelia configuration.
      # traefik.http.middlewares.authelia.forwardAuth.address: '{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/forward-auth?authelia_url=https%3A%2F%2F{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}%2F'
      traefik.http.middlewares.authelia.forwardAuth.trustForwardHeader: 'true'
      traefik.http.middlewares.authelia.forwardAuth.authResponseHeaders: 'Remote-User,Remote-Groups,Remote-Email,Remote-Name'
  nextcloud:
    container_name: 'nextcloud'
    image: 'linuxserver/nextcloud'
    restart: 'unless-stopped'
    networks:
      net: {}
    volumes:
      - '${PWD}/data/nextcloud/config:/config'
      - '${PWD}/data/nextcloud/data:/data'
    environment:
      PUID: '1000'
      PGID: '1000'
      TZ: 'Australia/Melbourne'
    labels:
      traefik.enable: 'true'
      traefik.http.routers.nextcloud.rule: 'Host(`nextcloud.{{< sitevar name="domain" nojs="example.com" >}}`)'
      traefik.http.routers.nextcloud.entryPoints: 'https'
      traefik.http.routers.nextcloud.tls: 'true'
      traefik.http.routers.nextcloud.middlewares: 'authelia@docker'
  heimdall:
    container_name: 'heimdall'
    image: 'linuxserver/heimdall'
    restart: 'unless-stopped'
    networks:
      net: {}
    volumes:
      - '${PWD}/data/heimdall/config:/config'
    environment:
      PUID: '1000'
      PGID: '1000'
      TZ: 'Australia/Melbourne'
    labels:
      traefik.enable: 'true'
      traefik.http.routers.heimdall.rule: 'Host(`heimdall.{{< sitevar name="domain" nojs="example.com" >}}`)'
      traefik.http.routers.heimdall.entryPoints: 'https'
      traefik.http.routers.heimdall.tls: 'true'
      traefik.http.routers.heimdall.middlewares: 'authelia-basic@docker'
...
```

### YAML

This example uses a `compose.yml` similar to the one above however it has two major differences:

1. A majority of the configuration is in YAML instead of the `labels` section of the `compose.yml` file.
2. It connects to __Authelia__ over TLS with client certificates which ensures that [Traefik] is a proxy
   authorized to communicate with __Authelia__. This expects that the
   [Server TLS](../../configuration/miscellaneous/server.md#tls) section is configured correctly.
   * The client certificates can easily be disabled by commenting the `cert` and `key` options in the `http.middlewares`
     section for the `forwardAuth` middlewares and the `certificates` in the `http.serversTransports` section.
   * The TLS communication can be disabled by commenting the entire `tls` section in the `http.middlewares` section for
     all `forwardAuth` middlewares, adjusting the `authelia` router in the `http.routers` section to use the
     `authelia-net@docker` service, and commenting the `authelia` service in the `http.service` section.

```yaml {title="compose.yml"}
---
networks:
  net:
    driver: 'bridge'
services:
  traefik:
    container_name: 'traefik'
    image: 'traefik:v3.5'
    restart: 'unless-stopped'
    command:
      - '--api=true'
      - '--api.dashboard=true'
      - '--api.insecure=false'
      - '--pilot.dashboard=false'
      - '--global.sendAnonymousUsage=false'
      - '--global.checkNewVersion=false'
      - '--log=true'
      - '--log.level=DEBUG'
      - '--log.filepath=/config/traefik.log'
      - '--providers.docker=true'
      - '--providers.docker.exposedByDefault=false'
      - '--providers.file=true'
      - '--providers.file.watch=true'
      - '--providers.file.directory=/config/dynamic'
      - '--entryPoints.http=true'
      - '--entryPoints.http.address=:8080/tcp'
      - '--entryPoints.http.http.redirections.entryPoint.to=https'
      - '--entryPoints.http.http.redirections.entryPoint.scheme=https'
      - '--entryPoints.https=true'
      - '--entryPoints.https.address=:8443/tcp'
    networks:
      net: {}
    ports:
      - '80:8080'
      - '443:8443'
    volumes:
      - '/var/run/docker.sock:/var/run/docker.sock'
      - '${PWD}/data/traefik/config:/config'
      - '${PWD}/data/traefik/certificates:/certificates'
    labels:
      traefik.enable: 'true'
  authelia:
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    image: 'authelia/authelia'
    restart: 'unless-stopped'
    networks:
      net: {}
    volumes:
      - '${PWD}/data/authelia/config:/config'
      - '${PWD}/data/authelia/certificates:/certificates'
    environment:
      TZ: 'Australia/Melbourne'
    labels:
      traefik.enable: 'true'
  nextcloud:
    container_name: 'nextcloud'
    image: 'linuxserver/nextcloud'
    restart: 'unless-stopped'
    networks:
      net: {}
    volumes:
      - '${PWD}/data/nextcloud/config:/config'
      - '${PWD}/data/nextcloud/data:/data'
    environment:
      PUID: '1000'
      PGID: '1000'
      TZ: 'Australia/Melbourne'
    labels:
      traefik.enable: 'true'
  heimdall:
    container_name: 'heimdall'
    image: 'linuxserver/heimdall'
    restart: 'unless-stopped'
    networks:
      net: {}
    volumes:
      - '${PWD}/data/heimdall/config:/config'
    environment:
      PUID: '1000'
      PGID: '1000'
      TZ: 'Australia/Melbourne'
    labels:
      traefik.enable: 'true'
  whoami:
    container_name: 'whoami'
    image: 'traefik/whoami:latest'
    restart: 'unless-stopped'
    networks:
      net: {}
    environment:
      TZ: 'Australia/Melbourne'
    labels:
      traefik.enable: 'true'
...
```

This file is part of the dynamic configuration and should have the path
`${PWD}/data/traefik/config/dynamic/traefik.yml`. Please see the [Traefik] service and the volume that mounts the
`${PWD}/data/traefik/config` in the docker compose above.

```yaml {title="traefik.yml"}
---
entryPoints:
  web:
    proxyProtocol:
      insecure: false
      trustedIPs: []
    forwardedHeaders:
      insecure: false
      trustedIPs: []
  websecure:
    proxyProtocol:
      insecure: false
      trustedIPs: []
    forwardedHeaders:
      insecure: false
      trustedIPs: []
tls:
  options:
    modern:
      minVersion: 'VersionTLS13'
    intermediate:
      minVersion: 'VersionTLS12'
      cipherSuites:
        - 'TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256'
        - 'TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256'
        - 'TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384'
        - 'TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384'
        - 'TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305'
        - 'TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305'
http:
  middlewares:
    authelia:
      forwardAuth:
        address: '{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/forward-auth'
        ## The following commented line is for configuring the Authelia URL in the proxy. We strongly suggest this is
        ## configured in the Session Cookies section of the Authelia configuration.
        # address: '{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/forward-auth?authelia_url=https%3A%2F%2F{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}%2F'
        trustForwardHeader: true
        authResponseHeaders:
          - 'Remote-User'
          - 'Remote-Groups'
          - 'Remote-Email'
          - 'Remote-Name'
        tls:
          ca: '/certificates/ca.public.crt'
          cert: '/certificates/traefik.public.crt'
          key: '/certificates/traefik.private.pem'
    authelia-basic:
      forwardAuth:
        address: '{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/verify?auth=basic'
        trustForwardHeader: true
        authResponseHeaders:
          - 'Remote-User'
          - 'Remote-Groups'
          - 'Remote-Email'
          - 'Remote-Name'
        tls:
          ca: '/certificates/ca.public.crt'
          cert: '/certificates/traefik.public.crt'
          key: '/certificates/traefik.private.pem'
  routers:
    traefik:
      rule: 'Host(`traefik.{{< sitevar name="domain" nojs="example.com" >}}`)'
      entryPoints: 'websecure'
      service: 'api@internal'
      middlewares:
        - 'authelia@file'
      tls:
        options: 'modern@file'
        certResolver: 'default'
        domains:
          - main: '{{< sitevar name="domain" nojs="example.com" >}}'
            sans:
              - '*.{{< sitevar name="domain" nojs="example.com" >}}'
    whoami:
      rule: 'Host(`whoami.{{< sitevar name="domain" nojs="example.com" >}}`)'
      entryPoints: 'websecure'
      service: 'whoami-net@docker'
      middlewares:
        - 'authelia@file'
      tls:
        options: 'modern@file'
        certResolver: 'default'
        domains:
          - main: '{{< sitevar name="domain" nojs="example.com" >}}'
            sans:
              - '*.{{< sitevar name="domain" nojs="example.com" >}}'
    nextcloud:
      rule: 'Host(`nextcloud.{{< sitevar name="domain" nojs="example.com" >}}`)'
      entryPoints: 'websecure'
      service: 'nextcloud-net@docker'
      middlewares:
        - 'authelia@file'
      tls:
        options: 'modern@file'
        certResolver: 'default'
        domains:
          - main: '{{< sitevar name="domain" nojs="example.com" >}}'
            sans:
              - '*.{{< sitevar name="domain" nojs="example.com" >}}'
    heimdall:
      rule: 'Host(`heimdall.{{< sitevar name="domain" nojs="example.com" >}}`)'
      entryPoints: 'websecure'
      service: 'heimdall-net@docker'
      middlewares:
        - 'authelia-basic@file'
      tls:
        options: 'modern@file'
        certResolver: 'default'
        domains:
          - main: '{{< sitevar name="domain" nojs="example.com" >}}'
            sans:
              - '*.{{< sitevar name="domain" nojs="example.com" >}}'
    authelia:
      rule: 'Host(`{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`)'
      entryPoints: 'websecure'
      service: 'authelia@file'
      tls:
        options: 'modern@file'
        certResolver: 'default'
        domains:
          - main: '{{< sitevar name="domain" nojs="example.com" >}}'
            sans:
              - '*.{{< sitevar name="domain" nojs="example.com" >}}'
  services:
    authelia:
      loadBalancer:
        servers:
          - url: '{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/'
        serversTransport: 'autheliaMutualTLS'
  serversTransports:
    autheliaMutualTLS:
      certificates:
        - certFile: '/certificates/traefik.public.crt'
          keyFile: '/certificates/traefik.private.pem'
      rootCAs:
        - '/certificates/ca.public.crt'
...
```

## Kubernetes

Authelia supports some of the [Traefik] based Kubernetes Ingress. See the
[Kubernetes Integration Guide](../kubernetes/traefik-ingress.md) for more information.

## Frequently Asked Questions

### Basic Authentication

Authelia provides the means to be able to authenticate your first factor via the [Proxy-Authorization] header, this
is compatible with [Traefik].

If you have a use-case which requires the use of the [Authorization] header/basic authentication login prompt you can
call Authelia's `/api/verify?auth=basic` endpoint to force a switch to the [Authorization] header.

### Middleware authelia@docker not found

If [Traefik] and __Authelia__ are defined in different docker compose stacks you may experience an issue where [Traefik]
complains that: `middleware authelia@docker not found`.

This can be avoided a couple different ways:

1. Ensure __Authelia__ container is up before [Traefik] is started:
   * Utilise the [depends_on](https://docs.docker.com/compose/compose-file/#depends_on) option
2. Define the __Authelia__ middleware on your [Traefik] container. See the below example.

```yaml {title="compose.yml"}
traefik.http.middlewares.authelia.forwardAuth.address: '{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/forward-auth'
## The following commented line is for configuring the Authelia URL in the proxy. We strongly suggest this is
## configured in the Session Cookies section of the Authelia configuration.
# traefik.http.middlewares.authelia.forwardAuth.address: '{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/forward-auth?authelia_url=https%3A%2F%2F{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}%2F'
traefik.http.middlewares.authelia.forwardAuth.trustForwardHeader: 'true'
traefik.http.middlewares.authelia.forwardAuth.authResponseHeaders: 'Remote-User,Remote-Groups,Remote-Email,Remote-Name'
```

## See Also

* [Traefik ForwardAuth Documentation](https://doc.traefik.io/traefik/middlewares/http/forwardauth/)
* [Traefik Forwarded Headers Documentation](https://doc.traefik.io/traefik/routing/entrypoints/#forwarded-headers)
* [Traefik Proxy Protocol Documentation](https://doc.traefik.io/traefik/routing/entrypoints/#proxyprotocol)
* [Forwarded Headers]

[docker compose]: https://docs.docker.com/compose/
[Traefik]: https://docs.traefik.io/
[Forwarded Headers]: forwarded-headers
[Authorization]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Authorization
[Proxy-Authorization]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Proxy-Authorization
