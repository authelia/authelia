---
title: "SWAG"
description: "An integration guide for Authelia and the SWAG reverse proxy"
lead: "A guide on integrating Authelia with SWAG."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "proxies"
weight: 351
toc: true
aliases:
  - /i/swag
---

[SWAG] is a reverse proxy supported by __Authelia__. It's an [NGINX] proxy container with bundled configurations to make
your life easier.

*__Important:__ When using these guides it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only and you need to understand the proxy
configuration and customize it to your needs. To-that-end we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Introduction

As [SWAG] is a [NGINX] proxy with curated configurations, integration of __Authelia__ with [SWAG] is very easy and you
only need to enabled two includes.

*__Note:__ All paths in this guide are the locations inside the container. You will have to either edit the files within
the container or adapt the path to the path you have mounted the relevant container path to.*

## Get Started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get Started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Requirements

[SWAG] supports the required [NGINX](nginx.md#requirements) requirements for __Authelia__ out-of-the-box.

### SWAG Caveat

One current caveat of the [SWAG] implementation is that it serves Authelia as a subpath for each domain. We
*__strongly recommend__* instead of using the out of the box method and guide for [SWAG] that you follow the
[NGINX](nginx.md) guide (which *can be used* with [SWAG]) and run Authelia as it's own subdomain.

This is partly because Webauthn requires that the domain is an exact match when registering and authenticating and it is
possible that due to web standards this will never change.

In addition this represents a bad user experience in some instances such as:

  - Users sometimes visit the `https://app.example.com/authelia` URL which doesn't automatically redirect the user to
    `https://app.example.com` (if they visit `https://app.example.com` then they'll be redirected to authenticate then
    redirected back to their original URL).
  - Administrators may wish to setup OpenID Connect 1.0 in which case it also doesn't represent a good user experience.

Taking these factors into consideration we're adapting our [SWAG] guide to use what we consider best for the users and
most easily supported. Users who wish to use the [SWAG] guide are free to do so but may not receive the same support.

## Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

To configure trusted proxies for [SWAG] see the [NGINX] section on [Trusted Proxies](nginx.md#trusted-proxies).
Adapting this to [SWAG] is beyond the scope of this documentation.

## Docker Compose

The following docker compose example has various applications suitable for setting up an example environment.

It uses the [nginx image](https://github.com/linuxserver/docker-nginx) from [linuxserver.io] which includes all of the
required modules including the `http_set_misc` module.

It also includes the [nginx-proxy-confs](https://github.com/linuxserver/docker-mods/tree/nginx-proxy-confs) mod where
they have several configuration examples in the `/config/nginx/proxy-confs` directory. This can be omitted if desired.

If you're looking for a more complete solution [linuxserver.io] also have an nginx container called [SWAG](./swag.md)
which includes ACME and various other useful utilities.

{{< details "docker-compose.yaml" >}}
```yaml
---
version: "3.8"

networks:
  net:
    driver: bridge

services:
  swag:
    container_name: swag
    image: lscr.io/linuxserver/swag
    restart: unless-stopped
    networks:
      net:
        aliases: []
    ports:
      - '80:80'
      - '443:443'
    volumes:
      - ${PWD}/data/swag:/config
    environment:
      PUID: '1000'
      PGID: '1000'
      TZ: 'Australia/Melbourne'
      URL: 'example.com'
      SUBDOMAINS: 'www,whoami,auth,nextcloud,'
      VALIDATION: 'http'
      CERTPROVIDER: 'cloudflare'
      ONLY_SUBDOMAINS: 'false'
      STAGING: 'true'
    cap_add:
      - NET_ADMIN
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

## Prerequisite Steps

In the [SWAG] `/config` mount which is mounted to `${PWD}/data/swag` in our example:

1. Create a folder named `snippets/authelia`:
   - The `mkdir -p ${PWD}/data/swag/snippets/authelia` command should achieve this on Linux.
2. Create the `${PWD}/data/swag/nginxsnippets/authelia/location.conf` file which can be found [here](nginx.md#authelia-locationconf).
3. Create the `${PWD}/data/swag/nginxsnippets/authelia/authrequest.conf` file which can be found [here](nginx.md#authelia-authrequestconf).
   - Ensure you adjust the line `error_page 401 =302 https://auth.example.com/?rd=$target_url;` replacing `https://auth.example.com/` with your external Authelia URL.

## Protected Application

In the server configuration for the application you want to protect:

1. Edit the `/config/nginx/proxy-confs/` file for the application you wish to protect.
2. Under the `#include /config/nginx/authelia-server.conf;` line which should be within the `server` block
   but not inside any `location` blocks add the following line: ``.
3. Under the `#include /config/nginx/authelia-location.conf;` line which should be within the applications
   `location` block add the following line `include /config/nginx/snippets/authelia/authrequest.conf;`.

### Example

```nginx
server {
    listen 443 ssl;
    listen [::]:443 ssl;

    server_name whoami.*;

    include /config/nginx/ssl.conf;

    client_max_body_size 0;

    # Authelia: Step 1.
    #include /config/nginx/authelia-server.conf;
    include /config/nginx/snippets/authelia/location.conf;

    location / {
        # Authelia: Step 2.
        #include /config/nginx/authelia-location.conf;
        include /config/nginx/snippets/authelia/authrequest.conf;

        include /config/nginx/proxy.conf;
        resolver 127.0.0.11 valid=30s;
        set $upstream_app whoami;
        set $upstream_port 80;
        set $upstream_proto http;
        proxy_pass $upstream_proto://$upstream_app:$upstream_port;
    }
}
```

## See Also

* [Authelia NGINX Integration Documentation](nginx.md)
* [LinuxServer.io Setting Up Authelia With SWAG Documentation / Blog Post](https://www.linuxserver.io/blog/2020-08-26-setting-up-authelia)
* [NGINX ngx_http_auth_request_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html)
* [Forwarded Headers]

[SWAG]: https://docs.linuxserver.io/general/swag
[NGINX]: https://www.nginx.com/
[Forwarded Headers]: fowarded-headers
