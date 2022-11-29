---
title: "NGINX"
description: "An integration guide for Authelia and the NGINX reverse proxy"
lead: "A guide on integrating Authelia with the nginx reverse proxy."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  integration:
    parent: "proxies"
weight: 350
toc: true
aliases:
  - /i/nginx
  - /docs/deployment/supported-proxies/nginx.html
---

[NGINX] is a reverse proxy supported by __Authelia__.

*__Important:__ When using these guides it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only and you need to understand the proxy
configuration and customize it to your needs. To-that-end we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Get Started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get Started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Requirements

You need the following to run __Authelia__ with [NGINX]:

* [NGINX] must be built with the `http_auth_request` module which is relatively common
* [NGINX] must be built with the `http_realip` module which is relatively common
* [NGINX] must be built with the `http_set_misc` module or the `nginx-mod-http-set-misc` package if you want to preserve
  more than one query parameter when redirected to the portal due to a limitation in [NGINX]

## Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

*__Important:__ The included example is __NOT__ meant for production use. It's used expressly as an example to showcase
how you can configure multiple IP ranges. You should customize this example to fit your specific architecture and needs.
You should only include the specific IP address ranges of the trusted proxies within your architecture and should not
trust entire subnets unless that subnet only has trusted proxies and no other services.*

[NGINX]'s `http_realip` module is used to configure the trusted proxies' configuration. In our examples this is
configured in the `proxy.conf` file. Each `set_realip_from` directive adds a trusted proxy address range to the trusted
proxies list. Any request that comes from a source IP not in one of the configured ranges results in the header being
replaced with the source IP of the client.

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
  nginx:
    container_name: nginx
    image: lscr.io/linuxserver/nginx
    restart: unless-stopped
    networks:
      net:
        aliases: []
    ports:
      - '80:80'
      - '443:443'
    volumes:
      - ${PWD}/data/nginx/snippets:/config/nginx/snippets:ro
      - ${PWD}/data/nginx/site-confs:/config/nginx/site-confs:ro
    environment:
      TZ: 'Australia/Melbourne'
      DOCKER_MODS: 'linuxserver/mods:nginx-proxy-confs'
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

Below you will find commented examples of the following configuration:

* [Authelia Portal](#standard-example)
  * Running in Docker
  * Has the container name `authelia`
* [Protected Endpoint (Nextcloud)](#standard-example)
  * Running in Docker
  * Has the container name `nextcloud`
* [Supporting Configuration Snippets](#supporting-configuration-snippets)
* Assumes the following since we cannot reasonably provide a configuration for every architecture:
  * [NGINX] is also running in Docker and uses Docker DNS as a
    [resolver](https://nginx.org/en/docs/http/ngx_http_core_module.html#resolver) which is standard
  * [NGINX] shares a network with the `authelia` and `nextcloud` containers

### Assumptions

* Authelia is accessible to [NGINX] process with the hostname `authelia` on port `9091` making the URL
  `http://authelia:9091`. If this is not the case adjust all instances of this as appropriate.
* The [NGINX] configuration is in the folder `/config/nginx`. If this is not the case adjust all instances of this as
  appropriate.
* The URL you wish Authelia to be accessible on is `https://auth.example.com`. If this is not the case adjust all
  instances of this as appropriate.

### Standard Example

This example is for using the __Authelia__ portal redirection flow on a specific endpoint. It requires you to have the
[authelia-location.conf](#authelia-locationconf),
[authelia-authrequest.conf](#authelia-authrequestconf), and [proxy.conf](#proxyconf) snippets. In the example these
files exist in the `/config/nginx/snippets/` directory. The `/config/nginx/snippets/ssl.conf` snippet is expected to have
the configuration for TLS or SSL but is not included as part of the examples.

The directive `include /config/nginx/snippets/authelia-authrequest.conf;` within the `location` block is what directs
[NGINX] to perform authorization with Authelia. Every `location` block you wish for Authelia to perform authorization for
should include this directive.

{{< details "/config/nginx/site-confs/auth.conf (Authelia Portal)" >}}
```nginx
server {
    listen 80;
    server_name auth.*;

    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name auth.*;

    include /config/nginx/snippets/ssl.conf;

    set $upstream http://authelia:9091;

    location / {
        include /config/nginx/snippets/proxy.conf;
        proxy_pass $upstream;
    }

    location /api/verify {
        proxy_pass $upstream;
    }
}
```
{{< /details >}}

{{< details "/config/nginx/site-confs/nextcloud.conf (Protected Application - Nextcloud)" >}}
```nginx
server {
    listen 80;
    server_name nextcloud.*;

    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name nextcloud.*;

    include /config/nginx/snippets/ssl.conf;
    include /config/nginx/snippets/authelia-location.conf;

    set $upstream http://nextcloud;

    location / {
        include /config/nginx/snippets/proxy.conf;
        include /config/nginx/snippets/authelia-authrequest.conf;
        proxy_pass $upstream;
    }
}
```
{{< /details >}}

{{< details "/config/nginx/site-confs/whoami.conf (Protected Application - whoami)" >}}
```nginx
server {
    listen 80;
    server_name whoami.*;

    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name whoami.*;

    include /config/nginx/snippets/ssl.conf;
    include /config/nginx/snippets/authelia-location.conf;

    set $upstream http://whoami;

    location / {
        include /config/nginx/snippets/proxy.conf;
        include /config/nginx/snippets/authelia-authrequest.conf;
        proxy_pass $upstream;
    }
}
```
{{< /details >}}

### HTTP Basic Authentication Example

This example is for using HTTP basic auth on a specific endpoint. It is based on the full example above. It requires you
to have the [authelia-location-basic.conf](#authelia-location-basicconf),
[authelia-authrequest-basic.conf](#authelia-authrequest-basicconf), and [proxy.conf](#proxyconf) snippets. In the
example these files exist in the `/config/nginx/snippets/` directory. The `/config/nginx/snippets/ssl.conf` snippet is expected to have
the configuration for TLS or SSL but is not included as part of the examples.

The Authelia Portal file from the [Standard Example](#standard-example) configuration can be reused for this example as
such it isn't repeated.

{{< details "/config/nginx/site-confs/nextcloud.conf (Protected Application - Nextcloud)" >}}
```nginx
server {
    listen 80;
    server_name nextcloud.*;

    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name nextcloud.*;

    include /config/nginx/snippets/ssl.conf;
    include /config/nginx/snippets/authelia-location-basic.conf; # Use the "basic" endpoint

    set $upstream https://nextcloud;

    location / {
        include /config/nginx/snippets/proxy.conf;
        include /config/nginx/snippets/authelia-authrequest-basic.conf;
        proxy_pass $upstream;
    }
}
```
{{< /details >}}

### Supporting Configuration Snippets

The following configuration files are snippets that are used as includes in other files. The includes in the other files
match the headings, so if you wish to put them in a specific location or rename them, then make sure to update the
includes appropriately. Only the [proxy.conf](#proxyconf), [authelia-location.conf](#authelia-locationconf), and
[authelia-authrequest.conf](#authelia-authrequestconf) are required; see the descriptions for the others as to their
use cases.

#### proxy.conf

The following is an example `proxy.conf`. The important directives include the `real_ip` directives which you should read
[Trusted Proxies](#trusted-proxies) section to understand, or set the `X-Forwarded-Proto`, `X-Forwarded-Host`,
`X-Forwarded-Uri`, and `X-Forwarded-For` headers.

##### Standard Variant

Generally this variant is the suggested variant.

{{< details "/config/nginx/snippets/proxy.conf" >}}
```nginx
## Headers
proxy_set_header Host $host;
proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header X-Forwarded-Host $http_host;
proxy_set_header X-Forwarded-Uri $request_uri;
proxy_set_header X-Forwarded-Ssl on;
proxy_set_header X-Forwarded-For $remote_addr;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header Connection "";

## Basic Proxy Configuration
client_body_buffer_size 128k;
proxy_next_upstream error timeout invalid_header http_500 http_502 http_503; ## Timeout if the real server is dead.
proxy_redirect  http://  $scheme://;
proxy_http_version 1.1;
proxy_cache_bypass $cookie_session;
proxy_no_cache $cookie_session;
proxy_buffers 64 256k;

## Trusted Proxies Configuration
## Please read the following documentation before configuring this:
##     https://www.authelia.com/integration/proxies/nginx/#trusted-proxies
# set_real_ip_from 10.0.0.0/8;
# set_real_ip_from 172.16.0.0/12;
# set_real_ip_from 192.168.0.0/16;
# set_real_ip_from fc00::/7;
real_ip_header X-Forwarded-For;
real_ip_recursive on;

## Advanced Proxy Configuration
send_timeout 5m;
proxy_read_timeout 360;
proxy_send_timeout 360;
proxy_connect_timeout 360;
```
{{< /details >}}

##### Headers Only Variant

Generally the [standard variant](#standard-variant) is the suggested variant. This variant only contains the required
headers for Authelia to operate.

{{< details "/config/nginx/snippets/proxy.conf" >}}
```nginx
## Headers
proxy_set_header Host $host;
proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header X-Forwarded-Host $http_host;
proxy_set_header X-Forwarded-Uri $request_uri;
proxy_set_header X-Forwarded-Ssl on;
proxy_set_header X-Forwarded-For $remote_addr;
```
{{< /details >}}

#### authelia-location.conf

*The following snippet is used within the `server` block of a virtual host as a supporting endpoint used by
`auth_request` and is paired with [authelia-authrequest.conf](#authelia-authrequestconf).*

{{< details "/config/nginx/snippets/authelia-location.conf" >}}
```nginx
set $upstream_authelia http://authelia:9091/api/verify;

## Virtual endpoint created by nginx to forward auth requests.
location /authelia {
    ## Essential Proxy Configuration
    internal;
    proxy_pass $upstream_authelia;

    ## Headers
    ## The headers starting with X-* are required.
    proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
    proxy_set_header X-Original-Method $request_method;
    proxy_set_header X-Forwarded-Method $request_method;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $http_host;
    proxy_set_header X-Forwarded-Uri $request_uri;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header Content-Length "";
    proxy_set_header Connection "";

    ## Basic Proxy Configuration
    proxy_pass_request_body off;
    proxy_next_upstream error timeout invalid_header http_500 http_502 http_503; # Timeout if the real server is dead
    proxy_redirect http:// $scheme://;
    proxy_http_version 1.1;
    proxy_cache_bypass $cookie_session;
    proxy_no_cache $cookie_session;
    proxy_buffers 4 32k;
    client_body_buffer_size 128k;

    ## Advanced Proxy Configuration
    send_timeout 5m;
    proxy_read_timeout 240;
    proxy_send_timeout 240;
    proxy_connect_timeout 240;
}
```
{{< /details >}}

#### authelia-authrequest.conf

*The following snippet is used within a `location` block of a virtual host which uses the appropriate location block
and is paired with [authelia-location.conf](#authelia-locationconf).*

{{< details "/config/nginx/snippets/authelia-authrequest.conf" >}}
```nginx
## Send a subrequest to Authelia to verify if the user is authenticated and has permission to access the resource.
auth_request /authelia;

## Set the $target_url variable based on the original request.

## Comment this line if you're using nginx without the http_set_misc module.
set_escape_uri $target_url $scheme://$http_host$request_uri;

## Uncomment this line if you're using NGINX without the http_set_misc module.
# set $target_url $scheme://$http_host$request_uri;

## Save the upstream response headers from Authelia to variables.
auth_request_set $user $upstream_http_remote_user;
auth_request_set $groups $upstream_http_remote_groups;
auth_request_set $name $upstream_http_remote_name;
auth_request_set $email $upstream_http_remote_email;

## Inject the response headers from the variables into the request made to the backend.
proxy_set_header Remote-User $user;
proxy_set_header Remote-Groups $groups;
proxy_set_header Remote-Name $name;
proxy_set_header Remote-Email $email;

## If the subreqest returns 200 pass to the backend, if the subrequest returns 401 redirect to the portal.
error_page 401 =302 https://auth.example.com/?rd=$target_url;
```
{{< /details >}}

#### authelia-location-basic.conf

*The following snippet is used within the `server` block of a virtual host as a supporting endpoint used by
`auth_request` and is paired with [authelia-authrequest-basic.conf](#authelia-authrequest-basicconf). This particular
snippet is rarely required. It's only used if you want to only allow
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-location.conf](#authelia-locationconf) instead.*

{{< details "/config/nginx/snippets/authelia-location-basic.conf" >}}
```nginx
set $upstream_authelia http://authelia:9091/api/verify?auth=basic;

# Virtual endpoint created by nginx to forward auth requests.
location /authelia-basic {
    ## Essential Proxy Configuration
    internal;
    proxy_pass $upstream_authelia;

    ## Headers
    ## The headers starting with X-* are required.
    proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
    proxy_set_header X-Original-Method $request_method;
    proxy_set_header X-Forwarded-Method $request_method;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $http_host;
    proxy_set_header X-Forwarded-Uri $request_uri;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header Content-Length "";
    proxy_set_header Connection "";

    ## Basic Proxy Configuration
    proxy_pass_request_body off;
    proxy_next_upstream error timeout invalid_header http_500 http_502 http_503; # Timeout if the real server is dead
    proxy_redirect http:// $scheme://;
    proxy_http_version 1.1;
    proxy_cache_bypass $cookie_session;
    proxy_no_cache $cookie_session;
    proxy_buffers 4 32k;
    client_body_buffer_size 128k;

    ## Advanced Proxy Configuration
    send_timeout 5m;
    proxy_read_timeout 240;
    proxy_send_timeout 240;
    proxy_connect_timeout 240;
}
```
{{< /details >}}

#### authelia-authrequest-basic.conf

*The following snippet is used within a `location` block of a virtual host which uses the appropriate location block
and is paired with [authelia-location-basic.conf](#authelia-location-basicconf). This particular snippet is rarely
required. It's only used if you want to only allow
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-authrequest.conf](#authelia-authrequestconf) instead.*

{{< details "/config/nginx/snippets/authelia-authrequest-basic.conf" >}}
```nginx
## Send a subrequest to Authelia to verify if the user is authenticated and has permission to access the resource.
auth_request /authelia-basic;

## Comment this line if you're using nginx without the http_set_misc module.
set_escape_uri $target_url $scheme://$http_host$request_uri;

## Uncomment this line if you're using NGINX without the http_set_misc module.
# set $target_url $scheme://$http_host$request_uri;

## Save the upstream response headers from Authelia to variables.
auth_request_set $user $upstream_http_remote_user;
auth_request_set $groups $upstream_http_remote_groups;
auth_request_set $name $upstream_http_remote_name;
auth_request_set $email $upstream_http_remote_email;

## Inject the response headers from the variables into the request made to the backend.
proxy_set_header Remote-User $user;
proxy_set_header Remote-Groups $groups;
proxy_set_header Remote-Name $name;
proxy_set_header Remote-Email $email;
```
{{< /details >}}

#### authelia-location-detect.conf

*The following snippet is used within the `server` block of a virtual host as a supporting endpoint used by
`auth_request` and is paired with [authelia-authrequest-detect.conf](#authelia-authrequest-detectconf). This particular
snippet is rarely required. It's only used if you want to conditionally require
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-location.conf](#authelia-locationconf) instead.*

{{< details "/config/nginx/snippets/authelia-location-detect.conf" >}}
```nginx
include /config/nginx/snippets/authelia-location.conf;

set $is_basic_auth ""; # false value

## Detect the client you want to force basic auth for here
## For the example we just match a path on the original request
if ($request_uri = "/force-basic") {
    set $is_basic_auth "true";
    set $upstream_authelia "$upstream_authelia?auth=basic";
}

## A new virtual endpoint to used if the auth_request failed
location  /authelia-detect {
    internal;

    if ($is_basic_auth) {
        ## This is a request where we decided to use basic auth, return a 401.
        ## Nginx will also proxy back the WWW-Authenticate header from Authelia's
        ## response. This is what informs the client we're expecting basic auth.
        return 401;
    }

    ## The original request didn't target /force-basic, redirect to the pretty login page
    ## This is what `error_page 401 =302 https://auth.example.com/?rd=$target_url;` did.
    return 302 https://auth.example.com/$is_args$args;
}
```
{{< /details >}}

#### authelia-authrequest-detect.conf

*The following snippet is used within a `location` block of a virtual host which uses the appropriate location block
and is paired with [authelia-location-detect.conf](#authelia-location-detectconf). This particular snippet is rarely
required. It's only used if you want to conditionally require
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-authrequest.conf](#authelia-authrequestconf) instead.*

{{< details "/config/nginx/snippets/authelia-authrequest-detect.conf" >}}
```nginx
## Send a subrequest to Authelia to verify if the user is authenticated and has permission to access the resource.
auth_request /authelia;

## Comment this line if you're using nginx without the http_set_misc module.
set_escape_uri $target_url $scheme://$http_host$request_uri;

## Uncomment this line if you're using NGINX without the http_set_misc module.
# set $target_url $scheme://$http_host$request_uri;

## Save the upstream response headers from Authelia to variables.
auth_request_set $user $upstream_http_remote_user;
auth_request_set $groups $upstream_http_remote_groups;
auth_request_set $name $upstream_http_remote_name;
auth_request_set $email $upstream_http_remote_email;

## Inject the response headers from the variables into the request made to the backend.
proxy_set_header Remote-User $user;
proxy_set_header Remote-Groups $groups;
proxy_set_header Remote-Name $name;
proxy_set_header Remote-Email $email;

## If the subreqest returns 200 pass to the backend, if the subrequest returns 401 redirect to the portal.
error_page 401 =302 /authelia-detect?rd=$target_url;
```
{{< /details >}}

## See Also

* [NGINX ngx_http_auth_request_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html)
* [NGINX ngx_http_realip_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_realip_module.html)
* [Forwarded Headers]

[NGINX]: https://www.nginx.com/
[Forwarded Headers]: fowarded-headers
[linuxserver.io]: https://www.linuxserver.io/
