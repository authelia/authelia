---
title: "SWAG"
description: "An integration guide for Authelia and the SWAG reverse proxy"
summary: "A guide on integrating Authelia with SWAG."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 351
toc: true
aliases:
  - /i/swag
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[SWAG] is a reverse proxy supported by __Authelia__. It's an [NGINX] proxy container with bundled configurations to make
your life easier.

*__Important:__ When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only, and you need to understand the proxy
configuration and customize it to your needs. To-that-end, we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Introduction

As [SWAG] is a [NGINX] proxy with curated configurations, integration of __Authelia__ with [SWAG] is very easy and you
only need to enabled two includes.

*__Note:__ All paths in this guide are the locations inside the container. You will have to either edit the files within
the container or adapt the path to the path you have mounted the relevant container path to.*

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Requirements

[SWAG] supports the required [NGINX](nginx.md#requirements) requirements for __Authelia__ out-of-the-box.

### SWAG Caveat

One current caveat of the [SWAG] implementation is that it serves Authelia as a subpath for each domain by default. We
*__strongly recommend__* instead of using the defaults that you configure Authelia as a subdomain if possible.

There are two potential ways to achieve this:

1. Adjust the default `authelia-server.conf` as per the included directions.
2. Use the supplementary configuration snippets provided officially by Authelia.

This is partly because WebAuthn requires that the domain is an exact match when registering and authenticating and it is
possible that due to web standards this will never change.

In addition this represents a bad user experience in some instances such as:

* Users sometimes visit the `https://app.example.com/authelia` URL which doesn't automatically redirect the user to
  `https://app.example.com` (if they visit `https://app.example.com` then they'll be redirected to authenticate then
  redirected back to their original URL)
* Administrators may wish to setup [OpenID Connect 1.0](../../configuration/identity-providers/openid-connect/provider.md) in
  which case it also doesn't represent a good user experience as the `issuer` will be
  `https://app.example.com/authelia` for example
* Using the [SWAG] default configurations are more difficult to support as our specific familiarity is with our own
  example snippets

#### Option 1: Adjusting the Default Configuration

Open the generated `authelia-server.conf`. Adjust the following section. There are two snippets, one before and one
after. The only line that changes is the `set $signin_url` line, with `$http_host` replaced by `auth.example.com` and this configuration assumes you're
serving Authelia at `auth.example.com`.

```nginx
    if ($signin_url = '') {
        ## Set the $signin_url variable
        set $signin_url https://$http_host/authelia/?rd=$target_url;
    }
```

```nginx
    if ($signin_url = '') {
        ## Set the $signin_url variable
        set $signin_url https://auth.example.com/authelia/?rd=$target_url;
    }
```

#### Option 2: Using the Authelia Supplementary Configuration Snippets

See standard [NGINX](nginx.md) guide (which *can be used* with [SWAG]) and run Authelia as its own subdomain.

## Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

To configure trusted proxies for [SWAG] see the [NGINX] section on [Trusted Proxies](nginx.md#trusted-proxies).
Adapting this to [SWAG] is beyond the scope of this documentation.

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

## Docker Compose

The following docker compose example has various applications suitable for setting up an example environment.

It uses the [nginx image](https://github.com/linuxserver/docker-nginx) from [linuxserver.io] which includes all of the
required modules including the `http_set_misc` module.

It also includes the [nginx-proxy-confs](https://github.com/linuxserver/docker-mods/tree/nginx-proxy-confs) mod where
they have several configuration examples in the `/config/nginx/proxy-confs` directory. This can be omitted if desired.

If you're looking for a more complete solution [linuxserver.io] also have an nginx container called [SWAG](swag.md)
which includes ACME and various other useful utilities.

{{< details "docker-compose.yml" >}}
```yaml
---
version: "3.8"

networks:
  net:
    driver: 'bridge'

services:
  swag:
    container_name: 'swag'
    image: 'lscr.io/linuxserver/swag'
    restart: 'unless-stopped'
    networks:
      net:
        aliases: []
    ports:
      - '80:80'
      - '443:443'
    volumes:
      - '${PWD}/data/swag:/config'
      ## Uncomment the line below if you want to use the Authelia configuration snippets.
      #- '${PWD}/data/nginx/snippets:/snippets'
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
      - 'NET_ADMIN'
  authelia:
    container_name: 'authelia'
    image: 'authelia/authelia'
    restart: 'unless-stopped'
    networks:
      net:
        aliases: []
    expose:
      - 9091
    volumes:
      - '${PWD}/data/authelia/config:/config'
    environment:
      TZ: 'Australia/Melbourne'
  nextcloud:
    container_name: 'nextcloud'
    image: 'lscr.io/linuxserver/nextcloud'
    restart: 'unless-stopped'
    networks:
      net:
        aliases: []
    expose:
      - 443
    volumes:
      - '${PWD}/data/nextcloud/config:/config'
      - '${PWD}/data/nextcloud/data:/data'
    environment:
      PUID: '1000'
      PGID: '1000'
      TZ: 'Australia/Melbourne'
  whoami:
    container_name: 'whoami'
    image: 'docker.io/traefik/whoami'
    restart: 'unless-stopped'
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
   - The `mkdir -p ${PWD}/data/swag/nginx/snippets/authelia` command should achieve this on Linux.
2. Create the `${PWD}/data/swag/nginx/snippets/authelia/location.conf` file which can be found [here](nginx.md#authelia-locationconf).
3. Create the `${PWD}/data/swag/nginx/snippets/authelia/authrequest.conf` file which can be found [here](nginx.md#authelia-authrequestconf).
   - Ensure you adjust the line `error_page 401 =302 https://auth.example.com/?rd=$target_url;` replacing `https://auth.example.com/` with your external Authelia URL.

## Protected Application

In the server configuration for the application you want to protect:

1. Edit the `/config/nginx/proxy-confs/` file for the application you wish to protect.
2. Under the `#include /config/nginx/authelia-server.conf;` line which should be within the `server` block
   but not inside any `location` blocks add the following line: `include /config/nginx/snippets/authelia/location.conf;`.
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
[Forwarded Headers]: forwarded-headers
