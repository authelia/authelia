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

### Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

To configure trusted proxies for [SWAG] see the [NGINX] section on [Trusted Proxies](nginx.md#trusted-proxies).
Adapting this to [SWAG] is beyond the scope of this documentation.

### Assumptions and Adaptation

This guide makes a few assumptions. These assumptions may require adaptation in more advanced and complex scenarios. We
can not reasonably have examples for every advanced configuration option that exists. The
following are the assumptions we make:

* You have followed the [Get Started](../prologue/get-started.md) guide and configured
* Deployment Scenario:
  * Single Host
  * Authelia is deployed as a Container with the container name `authelia` on port `9091`
  * Proxy is deployed as a Container on a network shared with Authelia
* The above assumption means that Authelia should be accessible to the proxy on `http://authelia:9091` and as such:
  * You will have to adapt all instances of the above URL to be `https://` if Authelia configuration has a TLS key and
    certificate defined
  * You will have to adapt all instances of `authelia` in the URL if:
    * You're using a different container name
    * You deployed the proxy to a different location
  * You will have to adapt all instances of `9091` in the URL if:
    * You have adjusted the default port in the configuration
  * You will have to adapt the entire URL if:
    * Authelia is on a different host to the proxy
* All services are part of the `example.com` domain:
  * This domain and the subdomains will have to be adapted in all examples to match your specific domains unless you're
    just testing or you want to use that specific domain

### Docker Compose

The following docker compose example has various applications suitable for setting up an example environment.

It uses the [nginx image](https://github.com/linuxserver/docker-nginx) from [linuxserver.io] which includes all of the
required modules including the `http_set_misc` module.

It also includes the [nginx-proxy-confs](https://github.com/linuxserver/docker-mods/tree/nginx-proxy-confs) mod where
they have several configuration examples in the `/config/nginx/proxy-confs` directory. This can be omitted if desired.

If you're looking for a more complete solution [linuxserver.io] also have an nginx container called [SWAG](swag.md)
which includes ACME and various other useful utilities.

```yaml {title="docker-compose.yml"}
---
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
      SUBDOMAINS: 'www,whoami,auth,organizr'
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
  organizr:
    container_name: 'organizr'
    image: 'organizr/organizr'
    restart: 'unless-stopped'
    networks:
      net:
        aliases: []
    expose:
      - 80
    volumes:
      - '${PWD}/data/organizr/config:/config'
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

### Configuration Options

There are two configuration options for [SWAG]. The recommended option is
[Adjusting the Default Configuration](#option-1-adjusting-the-default-configuration)

### Option 1: Adjusting the Default Configuration

The first option requires some minor adjustments to be made and is the recommended option.

#### Adjust authelia-server.conf

The generated `authelia-server.conf` includes the `proxy_pass http://$upstream_authelia:9091;` directive in two location
blocks, we recommend adjusting these locations so they include the part of the location match after the `/authelia` part
for example in the `location = /authelia/api/verify` set the directive to
`proxy_pass http://$upstream_authelia:9091/api/verify;` and the `location = /authelia/api/authz/auth-request` set the
directive to `proxy_pass http://$upstream_authelia:9091/api/authz/auth-request;`. See the below example.

```nginx
# location for authelia 4.37 and below auth requests
location = /authelia/api/verify {
    internal;

    include /config/nginx/proxy.conf;
    include /config/nginx/resolver.conf;
    set $upstream_authelia authelia;
    proxy_pass http://$upstream_authelia:9091/api/verify;

    ## Include the Set-Cookie header if present
    auth_request_set $set_cookie $upstream_http_set_cookie;
    add_header Set-Cookie $set_cookie;

    proxy_pass_request_body off;
    proxy_set_header Content-Length "";
}

# location for authelia 4.38 and above auth requests
location = /authelia/api/authz/auth-request {
    internal;

    include /config/nginx/proxy.conf;
    include /config/nginx/resolver.conf;
    set $upstream_authelia authelia;
    proxy_pass http://$upstream_authelia:9091/api/authz/auth-request;

    ## Include the Set-Cookie header if present
    auth_request_set $set_cookie $upstream_http_set_cookie;
    add_header Set-Cookie $set_cookie;

    proxy_pass_request_body off;
    proxy_set_header Content-Length "";
}
```

#### Configure Authelia Site Configuration

1. In the `/config/nginx/site-confs/` directory copy `authelia.subdomain.conf.sample` to `authelia.subdomain.conf`.
2. Edit `authelia.subdomain.conf` and adjust `server_name authelia.*;` to be `server_name auth.*;`.

#### Configure Organizr Site Configuration

We're using Organizr as an example application for this example.

1. In the `/config/nginx/site-confs/` directory copy `organizr.subdomain.conf.sample` to `organizr.subdomain.conf`.
2. Edit `organizr.subdomain.conf` and remove the leading `#` (i.e. uncomment) the
3. `#include /config/nginx/authelia-server.conf;` line and the `#include /config/nginx/authelia-location.conf;` line.

#### Restart the Container

Once these changes have occurred you can restart [SWAG] and Organizr and Authelia should be configured correctly.

## Option 2: Using the Authelia Supplementary Configuration Snippets

See standard [NGINX](nginx.md) guide (which *can be used* with [SWAG]) and run Authelia as its own subdomain.

### Prerequisite Steps

In the [SWAG] `/config` mount which is mounted to `${PWD}/data/swag` in our example:

1. Create a folder named `snippets/authelia`:
   - The `mkdir -p ${PWD}/data/swag/nginx/snippets/authelia` command should achieve this on Linux.
2. Create the `${PWD}/data/swag/nginx/snippets/authelia/location.conf` file which can be found [here](nginx.md#authelia-locationconf).
3. Create the `${PWD}/data/swag/nginx/snippets/authelia/authrequest.conf` file which can be found [here](nginx.md#authelia-authrequestconf).
   - Ensure you adjust the line `error_page 401 =302 https://auth.example.com/?rd=$target_url;` replacing `https://auth.example.com/` with your external Authelia URL.

### Protected Application

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
