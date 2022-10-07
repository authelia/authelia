---
title: "NGINX Proxy Manager"
description: "An integration guide for Authelia and the NGINX Proxy Manager reverse proxy"
lead: "A guide on integrating Authelia with NGINX Proxy Manager."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "proxies"
weight: 352
toc: true
aliases:
  - /i/npm
---

[NGINX Proxy Manager] is supported by __Authelia__. It's a [NGINX] proxy with a configuration UI.

*__Important:__ When using these guides it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only and you need to understand the proxy
configuration and customize it to your needs. To-that-end we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Get Started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get Started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Requirements

[NGINX Proxy Manager] supports the required [NGINX](nginx.md#requirements) requirements for __Authelia__ out-of-the-box.

## Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

To configure trusted proxies for [NGINX Proxy Manager] see the [NGINX] section on
[Trusted Proxies](nginx.md#trusted-proxies). Adapting this to [NGINX Proxy Manager] is beyond the scope of
this documentation.

## Docker Compose

The following docker compose example has various applications suitable for setting up an example environment.

{{< details "docker-compose.yaml" >}}
```yaml
---
version: "3.8"

networks:
  net:
    driver: bridge

services:
  nginx:
    container_name: nginx
    image: jc21/nginx-proxy-manager
    restart: unless-stopped
    networks:
      net:
        aliases: []
    ports:
      - '80:80'
      - '81:81'
      - '443:443'
    volumes:
      - ${PWD}/data/nginx-proxy-manager/data:/data
      - ${PWD}/data/nginx-proxy-manager/letsencrypt:/etc/letsencrypt
      - ${PWD}/data/nginx/snippets:/config/nginx/snippets:ro
    environment:
      TZ: 'Australia/Melbourne'
  authelia:
    container_name: authelia
    image: authelia/authelia
    restart: unless-stopped
    networks:
      net:
        aliases: []
    expose:
      - 9091
    volumes:
      - ${PWD}/data/authelia/config:/config
    environment:
      TZ: 'Australia/Melbourne'
  nextcloud:
    container_name: nextcloud
    image: lscr.io/linuxserver/nextcloud
    restart: unless-stopped
    networks:
      net:
        aliases: []
    expose:
      - 443
    volumes:
      - ${PWD}/data/nextcloud/config:/config
      - ${PWD}/data/nextcloud/data:/data
    environment:
      PUID: '1000'
      PGID: '1000'
      TZ: 'Australia/Melbourne'
  whoami:
    container_name: whoami
    image: docker.io/traefik/whoami
    restart: unless-stopped
    networks:
      net:
        aliases: []
    expose:
      - 80
    environment:
      TZ: 'Australia/Melbourne'
...
```
{{< /details >}}

## Configuration

### Assumptions

*__Important:__ Our examples make assumptions about your configuration. These assumptions represent sections that
either most likely require an adjustment, or may require an adjustment if you're not configuring it in the same way.*

* The domain for Authelia is `auth.example.com` which shoud be adjusted in all examples and snippets to your actual
  domain.
* The required configuration snippets are mounted in the container or otherwise available in the `/snippets/` directory.
  If you choose a different directory you're required to adjust every instance of `/snippets/` appropriately to your
  needs.
* You have not configured the Authelia configuration YAML with a server TLS certificate/key.
* You are running Authelia on the default port.
* You are running Authelia with the `container_name` of `authelia` or the Authelia process is otherwise resolvable by
  [NGINX Proxy Manager] as `authelia`.

### Snippets

The examples assume you've mounted a volume containing the relevant
[NGINX Snippets](nginx.md#supporting-configuration-snippets) from the [NGINX Integration Guide](nginx.md). The suggested
snippets are the `proxy.conf`, `authelia-location.conf`, and `authelia-authrequest.conf`. It may be fine to substitute
the standard variant of the `proxy.conf` for the headers only variant but this is untested.

These snippets make the addition of a protected proxy host substantially easier.

### Authelia Portal

The Authelia portal requires minimal configuration.

1. Create a new `Proxy Host`.
2. Set the following items in the `Details` tab:
   * Domain Names: `auth.example.com`
   * Scheme: `http`
   * Forward Hostname / IP: `authelia`
   * Forward Port: `9091`
3. Configure your `SSL` tab to:
   * Serve a valid certificate.
   * Force SSL: `true`
4. Configure your `Advanced` tab:
```nginx
location / {
    include /snippets/proxy.conf;
    proxy_pass $forward_scheme://$server:$port;
}
```

#### Authelia Portal Screenshots

Authelia Portal `Details` tab example:

{{< figure src="authelia.details.png" alt="Step 2" width="350" style="padding-right: 10px" >}}

Authelia Portal `Advanced` tab example:

{{< figure src="authelia.advanced.png" alt="Step 4" width="350" style="padding-right: 10px" >}}

### Protected Application

The following example shows how to configure a protected application. We often use Nextcloud for such examples.

1. Create a new `Proxy Host`.
2. Set the following items in the `Details` tab:
   * Domain Names: `nextcloud.example.com`
   * Scheme: `http`
   * Forward Hostname / IP: `nextcloud`
   * Forward Port: `80`
3. Configure your `SSL` tab to:
   * Serve a valid certificate.
   * Force SSL: `true`
4. Configure your `Advanced` tab:
```nginx
include /snippets/authelia-location.conf;

location / {
    include /snippets/proxy.conf;
    include /snippets/authelia-authrequest.conf;
    proxy_pass $forward_scheme://$server:$port;
}
```

#### Protected Application Screenshots

Protected Application (Nextcloud) `Details` tab example:

{{< figure src="nextcloud.details.png" alt="Step 2" width="350" style="padding-right: 10px" >}}

Protected Application (Nextcloud) `Advanced` tab example:

{{< figure src="protectedapp.advanced.png" alt="Step 4" width="350" style="padding-right: 10px" >}}

#### Proxy Hosts Screenshot

The following screenshot shows an example of following the directions for the Authelia Portal and two applications:

{{< figure src="proxyhosts.png" alt="Step 4" width="350" style="padding-right: 10px" >}}

## See Also

* [NGINX Proxy Manager Documentation](https://nginxproxymanager.com/setup/)
* [NGINX ngx_http_auth_request_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html)
* [Forwarded Headers]

[NGINX Proxy Manager]: https://nginxproxymanager.com/
[NGINX]: https://www.nginx.com/
[Forwarded Headers]: ../fowarded-headers
