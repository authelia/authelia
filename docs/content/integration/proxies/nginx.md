---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "NGINX"
description: "An integration guide for Authelia and the NGINX reverse proxy including the auth_request module configuration examples, deployment considerations, and caveats."
summary: "A guide on integrating Authelia with the nginx reverse proxy."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 350
toc: true
aliases:
  - '/i/nginx'
  - '/docs/deployment/supported-proxies/nginx.html'
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[NGINX] is a reverse proxy supported by **Authelia**.

_**Important:** When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only, and you need to understand the proxy
configuration and customize it to your needs. To-that-end, we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section._

## Get started

It's **_strongly recommended_** that users setting up _Authelia_ for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping _Authelia_.

## Requirements

You need the following to run **Authelia** with [NGINX]:

- [NGINX] must be version 1.29.8 or newer as the examples use the `$is_request_port` and `$request_port` variables which
  were introduced in 1.29.3 but were not available in subrequests such as those made by the `auth_request` directive
  until 1.29.8
- [NGINX] must be built with the `http_auth_request` module which is relatively common
- [NGINX] must have a [resolver](https://nginx.org/en/docs/http/ngx_http_core_module.html#resolver) configured as the
  examples use variables in the `proxy_pass` directive which are resolved at runtime rather than at startup. The
  [linuxserver.io] images, [SWAG](swag.md), and [NGINX Proxy Manager](nginx-proxy-manager/index.md) configure this for
  you, otherwise you will need to add a directive such as `resolver 127.0.0.11 valid=30s;` (the Docker DNS server) to
  the `http` block, or `no resolver defined to resolve` errors will occur
- [NGINX] must be built with the `http_realip` module which is relatively common
- [NGINX] must be built with the `http_set_misc` module or the `nginx-mod-http-set-misc` package if you want to use the
  legacy method and preserve more than one query parameter when redirected to the portal due to a limitation in [NGINX]

## Trusted Proxies and Integration Security

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
In addition to this section which is important to read, you should read the
[Validating Forwarded Authentication](../../reference/guides/validating-forwarded-authentication.md) reference guide
and perform the validation steps as part of your regular security validation routine when using this integration.
{{< /callout >}}

_**Important:** You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before._

_**Important:** The included example is **NOT** meant for production use. It's used expressly as an example to showcase
how you can configure multiple IP ranges. You should customize this example to fit your specific architecture and needs.
You should only include the specific IP address ranges of the trusted proxies within your architecture and should not
trust entire subnets unless that subnet only has trusted proxies and no other services._

[NGINX]'s `http_realip` module is used to configure the trusted proxies' configuration. In our examples this is
configured in the `proxy.conf` file. Each `set_realip_from` directive adds a trusted proxy address range to the trusted
proxies list. Any request that comes from a source IP not in one of the configured ranges results in the header being
replaced with the source IP of the client.

## Assumptions and Adaptation

This guide makes a few assumptions. These assumptions may require adaptation in more advanced and complex scenarios. We
can not reasonably have examples for every advanced configuration option that exists. Some of these values can
automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

The following are the assumptions we make:

- Deployment Scenario:
  - Single Host
  - Authelia is deployed as a Container with the container name `{{< sitevar name="host" nojs="authelia" >}}` on port `{{< sitevar name="port" nojs="9091" >}}`
  - Proxy is deployed as a Container on a network shared with Authelia
- The above assumption means that Authelia should be accessible to the proxy on `{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}` and as such:
  - You will have to adapt all instances of the above URL to be `https://` if Authelia configuration has a TLS key and
    certificate defined
  - You will have to adapt all instances of `{{< sitevar name="host" nojs="authelia" >}}` in the URL if:
    - you're using a different container name
    - you deployed the proxy to a different location
  - You will have to adapt all instances of `{{< sitevar name="port" nojs="9091" >}}` in the URL if:
    - you have adjusted the default port in the configuration
  - You will have to adapt the entire URL if:
    - Authelia is on a different host to the proxy
- All services are part of the `{{< sitevar name="domain" nojs="example.com" >}}` domain:
  - This domain and the subdomains will have to be adapted in all examples to match your specific domains unless you're
    just testing or you want to use that specific domain

## Implementation

[NGINX] utilizes the [AuthRequest](../../reference/guides/proxy-authorization.md#authrequest) Authz implementation. The
associated [Metadata](../../reference/guides/proxy-authorization.md#authrequest-metadata) should be considered required.

The examples below assume you are using the default
[Authz Endpoints Configuration](../../configuration/miscellaneous/server-endpoints-authz.md) or one similar to the
following minimal configuration:

```yaml {title="configuration.yml"}
server:
  endpoints:
    authz:
      auth-request:
        implementation: 'AuthRequest'
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

## Docker Compose

The following docker compose example has various applications suitable for setting up an example environment.

It uses the [nginx image](https://github.com/linuxserver/docker-nginx) from [linuxserver.io] which includes all of the
required modules including the `http_set_misc` module.

It also includes the [nginx-proxy-confs](https://github.com/linuxserver/docker-mods/tree/nginx-proxy-confs) mod where
they have several configuration examples in the `/config/nginx/proxy-confs` directory. This can be omitted if desired.

If you're looking for a more complete solution [linuxserver.io] also have an nginx container called [SWAG](swag.md)
which includes ACME and various other useful utilities.

```yaml {title="compose.yml"}
---
networks:
  net:
    driver: 'bridge'

services:
  nginx:
    container_name: 'nginx'
    image: 'lscr.io/linuxserver/nginx'
    restart: 'unless-stopped'
    networks:
      net:
        aliases:
          - '{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
    ports:
      - '80:80/tcp'
      - '443:443/tcp'
      - '443:443/udp'
    volumes:
      - '${PWD}/data/nginx/snippets:/config/nginx/snippets'
      - '${PWD}/data/nginx/site-confs:/config/nginx/site-confs'
    environment:
      TZ: 'Australia/Melbourne'
      DOCKER_MODS: 'linuxserver/mods:nginx-proxy-confs'
  authelia:
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    image: 'authelia/authelia'
    restart: 'unless-stopped'
    networks:
      net: {}
    volumes:
      - '${PWD}/data/authelia/config:/config'
    environment:
      TZ: 'Australia/Melbourne'
  nextcloud:
    container_name: 'nextcloud'
    image: 'lscr.io/linuxserver/nextcloud'
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
  whoami:
    container_name: 'whoami'
    image: 'docker.io/traefik/whoami'
    restart: 'unless-stopped'
    networks:
      net: {}
    environment:
      TZ: 'Australia/Melbourne'
...
```

## Configuration

Below you will find commented examples of the following configuration:

- [Authelia Portal](#standard-example)
  - Running in Docker
  - Has the container name `authelia`
- [Protected Endpoint (Nextcloud)](#standard-example)
  - Running in Docker
  - Has the container name `nextcloud`
- [Supporting Configuration Snippets](#supporting-configuration-snippets)
- Assumes the following since we cannot reasonably provide a configuration for every architecture:
  - [NGINX] is also running in Docker and uses Docker DNS as a
    [resolver](https://nginx.org/en/docs/http/ngx_http_core_module.html#resolver) (see the [requirements](#requirements))
  - [NGINX] shares a network with the `authelia` and `nextcloud` containers

### Assumptions

- Authelia is accessible to [NGINX] process with the hostname `{{< sitevar name="host" nojs="authelia" >}}` on port `{{< sitevar name="port" nojs="9091" >}}` making the URL
  `{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}`. If this is not the case adjust all instances of this as appropriate.
- The [NGINX] configuration is in the folder `/config/nginx`. If this is not the case adjust all instances of this as
  appropriate.
- The URL you wish Authelia to be accessible on is `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`. If this is not the case adjust all
  instances of this as appropriate.

### Standard Example

This example is for using the **Authelia** portal redirection flow on a specific endpoint. It requires you to have the
[authelia-location.conf](#authelia-locationconf),
[authelia-authrequest.conf](#authelia-authrequestconf), and [proxy.conf](#proxyconf) snippets. In the example these
files exist in the `/config/nginx/snippets/` directory. The `/config/nginx/snippets/ssl.conf` snippet is expected to have
the configuration for TLS or SSL but is not included as part of the examples.

The directive `include /config/nginx/snippets/authelia-authrequest.conf;` within the `location` block is what directs
[NGINX] to perform authorization with Authelia. Every `location` block you wish for Authelia to perform authorization for
should include this directive.

```nginx {title="site-confs/auth.conf"}
server {
    listen 80;
    server_name auth.*;

    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    http2 on;
    server_name auth.*;

    include /config/nginx/snippets/ssl.conf;

    set $upstream {{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}};

    location / {
        include /config/nginx/snippets/proxy.conf;
        proxy_pass $upstream;
    }
}
```

```nginx {title="site-confs/nextcloud.conf"}
server {
    listen 80;
    server_name nextcloud.*;

    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    http2 on;
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

```nginx {title="site-confs/portainer.conf"}
server {
    listen 80;
    server_name portainer.*;

    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    http2 on;
    server_name portainer.*;

    include /config/nginx/snippets/ssl.conf;
    include /config/nginx/snippets/authelia-location.conf;

    set $upstream http://portainer:9000;

    location / {
        include /config/nginx/snippets/proxy.conf;
        include /config/nginx/snippets/authelia-authrequest.conf;
        proxy_pass $upstream;
    }

    location /api/websocket/ {
        include /config/nginx/snippets/proxy.conf;
        include /config/nginx/snippets/websocket.conf;
        include /config/nginx/snippets/authelia-authrequest.conf;
        proxy_pass $upstream;
    }
}
```

```nginx {title="site-confs/whoami.conf"}
server {
    listen 80;
    server_name whoami.*;

    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    http2 on;
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

### HTTP Basic Authentication Example

This example is for using HTTP basic auth on a specific endpoint. It is based on the full example above. It requires you
to have the [authelia-location-basic.conf](#authelia-location-basicconf),
[authelia-authrequest-basic.conf](#authelia-authrequest-basicconf), and [proxy.conf](#proxyconf) snippets. In the
example these files exist in the `/config/nginx/snippets/` directory. The `/config/nginx/snippets/ssl.conf` snippet is expected to have
the configuration for TLS or SSL but is not included as part of the examples.

The Authelia Portal file from the [Standard Example](#standard-example) configuration can be reused for this example as
such it isn't repeated.

```nginx {title="site-confs/nextcloud.conf"}
server {
    listen 80;
    server_name nextcloud.*;

    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    http2 on;
    server_name nextcloud.*;

    include /config/nginx/snippets/ssl.conf;
    include /config/nginx/snippets/authelia-location-basic.conf; # Use the "basic" endpoint

    set $upstream http://nextcloud;

    location / {
        include /config/nginx/snippets/proxy.conf;
        include /config/nginx/snippets/authelia-authrequest-basic.conf;
        proxy_pass $upstream;
    }
}
```

### Supporting Configuration Snippets

The following configuration files are snippets that are used as includes in other files. The includes in the other files
match the headings, so if you wish to put them in a specific location or rename them, then make sure to update the
includes appropriately. Only the [proxy.conf](#proxyconf), [authelia-location.conf](#authelia-locationconf), and
[authelia-authrequest.conf](#authelia-authrequestconf) are required; see the descriptions for the others as to their
use cases.

#### proxy.conf

The following is an example `proxy.conf`. The important directives include the `real_ip` directives which you should read
[Trusted Proxies](#trusted-proxies-and-integration-security) section to understand, or set the `X-Forwarded-Proto`, `X-Forwarded-Host`,
`X-Forwarded-URI`, and `X-Forwarded-For` headers.

##### Standard Variant

Generally this variant is the suggested variant.

```nginx {title="proxy.conf"}
## Headers
proxy_set_header Host $host$is_request_port$request_port;
proxy_set_header X-Original-URL $scheme://$host$is_request_port$request_port$request_uri;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header X-Forwarded-Host $host$is_request_port$request_port;
proxy_set_header X-Forwarded-URI $request_uri;
proxy_set_header X-Forwarded-Ssl on;
proxy_set_header X-Forwarded-For $remote_addr;
proxy_set_header X-Real-IP $remote_addr;

## Basic Proxy Configuration
client_body_buffer_size 128k;
proxy_next_upstream error timeout invalid_header http_500 http_502 http_503; ## Timeout if the real server is dead.
proxy_http_version 1.1;
## Replace authelia_session with the session cookie name if it has been changed.
proxy_cache_bypass $cookie_authelia_session;
proxy_no_cache $cookie_authelia_session;
proxy_buffers 64 256k;

## Trusted Proxies Configuration
## Please read the following documentation before configuring this:
##     https://www.authelia.com/integration/proxies/nginx/#trusted-proxies-and-integration-security
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

##### Headers Only Variant

Generally the [standard variant](#standard-variant) is the suggested variant. This variant only contains the required
headers for Authelia to operate.

```nginx {title="proxy.conf"}
## Headers
proxy_set_header Host $host$is_request_port$request_port;
proxy_set_header X-Original-URL $scheme://$host$is_request_port$request_port$request_uri;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header X-Forwarded-Host $host$is_request_port$request_port;
proxy_set_header X-Forwarded-URI $request_uri;
proxy_set_header X-Forwarded-Ssl on;
proxy_set_header X-Forwarded-For $remote_addr;
```

#### websocket.conf

The following is an example `websocket.conf`. This can be utilized on locations that require websockets. The standard
example has an example usage of this file.

```nginx {title="websocket.conf"}
## WebSocket Example
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
```

#### authelia-location.conf

_The following snippet is used within the `server` block of a virtual host as a supporting endpoint used by
`auth_request` and is paired with [authelia-authrequest.conf](#authelia-authrequestconf)._

```nginx {title="authelia-location.conf"}
set $upstream_authelia {{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/auth-request;

## Virtual endpoint created by nginx to forward auth requests.
location /internal/authelia/authz {
    ## Essential Proxy Configuration
    internal;
    proxy_pass $upstream_authelia;

    ## Headers
    ## The headers starting with X-* are required.
    proxy_set_header X-Original-Method $request_method;
    proxy_set_header X-Original-URL $scheme://$host$is_request_port$request_port$request_uri;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header Content-Length "";
    proxy_set_header Connection "";

    ## Basic Proxy Configuration
    proxy_pass_request_body off;
    proxy_next_upstream error timeout invalid_header http_500 http_502 http_503; # Timeout if the real server is dead
    proxy_http_version 1.1;
    ## Replace authelia_session with the session cookie name if it has been changed.
    proxy_cache_bypass $cookie_authelia_session;
    proxy_no_cache $cookie_authelia_session;
    proxy_buffers 4 32k;
    client_body_buffer_size 128k;

    ## Advanced Proxy Configuration
    send_timeout 5m;
    proxy_read_timeout 240;
    proxy_send_timeout 240;
    proxy_connect_timeout 240;
}
```

#### authelia-authrequest.conf

_The following snippet is used within a `location` block of a virtual host which uses the appropriate location block
and is paired with [authelia-location.conf](#authelia-locationconf)._

```nginx {title="authelia-authrequest.conf"}
## Send a subrequest to Authelia to verify if the user is authenticated and has permission to access the resource.
auth_request /internal/authelia/authz;

## Save the upstream metadata response headers from Authelia to variables.
auth_request_set $user $upstream_http_remote_user;
auth_request_set $groups $upstream_http_remote_groups;
auth_request_set $name $upstream_http_remote_name;
auth_request_set $email $upstream_http_remote_email;

## Inject the metadata response headers from the variables into the request made to the backend.
proxy_set_header Remote-User $user;
proxy_set_header Remote-Groups $groups;
proxy_set_header Remote-Email $email;
proxy_set_header Remote-Name $name;

## Configure the redirection when the authz failure occurs. Lines starting with 'Modern Method' and 'Legacy Method'
## should be commented / uncommented as pairs. The modern method uses the session cookies configuration's authelia_url
## value to determine the redirection URL here. It's much simpler and compatible with the multi-cookie domain easily.

## Modern Method: Set the $redirection_url to the Location header of the response to the Authz endpoint.
auth_request_set $redirection_url $upstream_http_location;

## Modern Method: When there is a 401 response code from the authz endpoint redirect to the $redirection_url.
error_page 401 =302 $redirection_url;

## Legacy Method: Set $target_url to the original requested URL.
## This requires http_set_misc module, replace 'set_escape_uri' with 'set' if you don't have this module.
# set_escape_uri $target_url $scheme://$host$is_request_port$request_port$request_uri;

## Legacy Method: When there is a 401 response code from the authz endpoint redirect to the portal with the 'rd'
## URL parameter set to $target_url. This requires users update '{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/' with their external authelia URL.
# error_page 401 =302 https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/?rd=$target_url;
```

#### authelia-location-basic.conf

_The following snippet is used within the `server` block of a virtual host as a supporting endpoint used by
`auth_request` and is paired with [authelia-authrequest-basic.conf](#authelia-authrequest-basicconf). This particular
snippet is rarely required. It's only used if you want to only allow
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-location.conf](#authelia-locationconf) instead._

This snippet requires an [authz endpoint](../../configuration/miscellaneous/server-endpoints-authz.md) named
`auth-request/basic` with the `AuthRequest` implementation and only the `HeaderAuthorization` strategy, which responds
with a `401` status code and the `WWW-Authenticate` header when the `Authorization` header is absent or invalid.
Configuring any authz endpoint replaces the default endpoints, so the default `auth-request` endpoint must also be
configured if it's used by other snippets. For example:

```yaml {title="configuration.yml"}
server:
  endpoints:
    authz:
      auth-request:
        implementation: 'AuthRequest'
      auth-request/basic:
        implementation: 'AuthRequest'
        authn_strategies:
          - name: 'HeaderAuthorization'
            schemes:
              - 'Basic'
```

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Do not add the `HeaderProxyAuthorization` strategy to an endpoint with the `AuthRequest` implementation. It responds
with a `407` status code which the `auth_request` directive treats as an error. Use the
`HeaderAuthRequestProxyAuthorization` strategy instead if the `Proxy-Authorization` header is required.
{{< /callout >}}

```nginx {title="authelia-location-basic.conf"}
set $upstream_authelia {{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/auth-request/basic;

# Virtual endpoint created by nginx to forward auth requests.
location /internal/authelia/authz/basic {
    ## Essential Proxy Configuration
    internal;
    proxy_pass $upstream_authelia;

    ## Headers
    ## The headers starting with X-* are required.
    proxy_set_header X-Original-Method $request_method;
    proxy_set_header X-Original-URL $scheme://$host$is_request_port$request_port$request_uri;
    proxy_set_header X-Forwarded-Method $request_method;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $host$is_request_port$request_port;
    proxy_set_header X-Forwarded-URI $request_uri;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header Content-Length "";
    proxy_set_header Connection "";

    ## Basic Proxy Configuration
    proxy_pass_request_body off;
    proxy_next_upstream error timeout invalid_header http_500 http_502 http_503; # Timeout if the real server is dead
    proxy_http_version 1.1;
    ## Replace authelia_session with the session cookie name if it has been changed.
    proxy_cache_bypass $cookie_authelia_session;
    proxy_no_cache $cookie_authelia_session;
    proxy_buffers 4 32k;
    client_body_buffer_size 128k;

    ## Advanced Proxy Configuration
    send_timeout 5m;
    proxy_read_timeout 240;
    proxy_send_timeout 240;
    proxy_connect_timeout 240;
}
```

#### authelia-authrequest-basic.conf

_The following snippet is used within a `location` block of a virtual host which uses the appropriate location block
and is paired with [authelia-location-basic.conf](#authelia-location-basicconf). This particular snippet is rarely
required. It's only used if you want to only allow
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-authrequest.conf](#authelia-authrequestconf) instead._

```nginx {title="authelia-authrequest-basic.conf"}
## Send a subrequest to Authelia to verify if the user is authenticated and has permission to access the resource.
auth_request /internal/authelia/authz/basic;

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

#### authelia-location-detect.conf

_The following snippet is used within the `server` block of a virtual host as a supporting endpoint used by
`auth_request` and is paired with [authelia-authrequest-detect.conf](#authelia-authrequest-detectconf). This particular
snippet is rarely required. It's only used if you want to conditionally require
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-location.conf](#authelia-locationconf) instead._

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This snippet requires both the `auth-request` and `auth-request/basic` authz endpoints described in the
[authelia-location-basic.conf](#authelia-location-basicconf) section.
{{< /callout >}}

```nginx {title="authelia-location-detect.conf"}
include /config/nginx/snippets/authelia-location.conf;

set $is_basic_auth ""; # false value

## Detect the client you want to force basic auth for here
## For the example we just match a path on the original request
if ($request_uri = "/force-basic") {
    set $is_basic_auth "true";
    set $upstream_authelia "${upstream_authelia}/basic";
}

## A new virtual endpoint to used if the auth_request failed
location  /internal/authelia/authz/detect {
    internal;

    if ($is_basic_auth) {
        ## This is a request where we decided to use basic auth, return a 401.
        ## Nginx will also proxy back the WWW-Authenticate header from Authelia's
        ## response. This is what informs the client we're expecting basic auth.
        return 401;
    }

    ## IMPORTANT: The below URL `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/` MUST be replaced with the externally accessible URL of the
    ## Authelia Portal/Site.
    ##
    ## The original request didn't target /force-basic, redirect to the pretty login page
    ## This is what `error_page 401 =302 https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/?rd=$target_url;` did.
    return 302 https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/$is_args$args;
}
```

#### authelia-authrequest-detect.conf

_The following snippet is used within a `location` block of a virtual host which uses the appropriate location block
and is paired with [authelia-location-detect.conf](#authelia-location-detectconf). This particular snippet is rarely
required. It's only used if you want to conditionally require
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-authrequest.conf](#authelia-authrequestconf) instead._

```nginx {title="authelia-authrequest-detect.conf"}
## Send a subrequest to Authelia to verify if the user is authenticated and has permission to access the resource.
auth_request /internal/authelia/authz;

## Comment this line if you're using nginx without the http_set_misc module.
set_escape_uri $target_url $scheme://$host$is_request_port$request_port$request_uri;

## Uncomment this line if you're using NGINX without the http_set_misc module.
# set $target_url $scheme://$host$is_request_port$request_port$request_uri;

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
error_page 401 =302 /internal/authelia/authz/detect?rd=$target_url;
```

## FAQ

### Why are both the X-Forwarded-* and X-Original-URL headers required?

The `X-Forwarded-*` headers are required for Authelia to function properly itself when you are making requests to the
portal or API. This means that the `X-Forwarded-*` headers are not essential for anything other than Authelia itself. We
do however recommend you use them.

The `X-Original-URL` header is required for the `auth_request` directive to work properly. This header was chosen simply
so that it's natively compatible with the [Kubernetes](#kubernetes) integrations. This means most of the time you do not
need to include it, only for the `auth_request` directive to work.

### Why do the examples use $host$is_request_port$request_port instead of $http_host?

The `$http_host` variable is the raw value of the `Host` request header. Clients using HTTP/3 send the host in the
`:authority` pseudo-header instead of the `Host` header, which leaves `$http_host` empty. This causes the
`X-Forwarded-Host` and `X-Original-URL` headers to be sent without a host, i.e. `https:///`, and Authelia is unable to
determine the URL of the request.

The `$host` variable is populated regardless of the HTTP version, and the `$request_port` variable contains the port the
client included in the request if any, with `$is_request_port` containing the `:` separator when it does. Together they
produce the equivalent of `$http_host` without depending on the `Host` header. You should not replace these variables
with `$http_host` when adapting the examples.

### NGINX is behind another proxy which terminates TLS, how do I configure that?

If [NGINX] receives requests over plain HTTP from another proxy which terminates TLS, such as a Cloudflare Tunnel or a
load balancer, the `$scheme` variable is `http`. This causes the `X-Forwarded-Proto` and `X-Original-URL` headers to
report the `http` scheme, and Authelia rejects these requests as they use an insecure scheme.

To resolve this you can use the `X-Forwarded-Proto` header sent by that proxy, but only when the request comes from
that proxy. Add the following to the `http {}` block, adjusting the address ranges to only include the proxies which
terminate TLS in front of [NGINX]:

```nginx
geo $realip_remote_addr $is_trusted_proxy {
    default     0;
    10.0.0.0/8  1;
}

map "$is_trusted_proxy:$http_x_forwarded_proto" $forwarded_proto {
    default                  $scheme;
    "~^1:(?<proto>https?)$"  $proto;
}

map $forwarded_proto $forwarded_ssl {
    default  off;
    https    on;
}
```

Then replace `$scheme` with `$forwarded_proto` in every `X-Forwarded-Proto` and `X-Original-URL` header and every
`$target_url` variable in the [proxy.conf](#proxyconf), `authelia-location*.conf`, and `authelia-authrequest*.conf`
snippets, and replace `on` with `$forwarded_ssl` in the `X-Forwarded-Ssl` header, for example:

```nginx
proxy_set_header X-Original-URL $forwarded_proto://$host$is_request_port$request_port$request_uri;
proxy_set_header X-Forwarded-Proto $forwarded_proto;
proxy_set_header X-Forwarded-Ssl $forwarded_ssl;
```

The `server` blocks must also accept plain HTTP from that proxy on the port it forwards requests to, rather than
redirecting them to HTTPS like the `listen 80` blocks in the examples.

### Why does NGINX respond with a 500 status code and log auth request unexpected status?

The `auth_request` directive only accepts a `2xx` status code which allows the request, or a `401` or `403` status code
which denies the request. Any other status code returned by the authz endpoint causes [NGINX] to respond with a `500`
status code and log an error such as `auth request unexpected status: 400`. The status code in the log message, along
with the Authelia logs, indicates the cause:

- `302` or `303`: The authz endpoint is using the `ForwardAuth` implementation instead of the `AuthRequest`
  implementation. Ensure the [authz endpoint](../../configuration/miscellaneous/server-endpoints-authz.md) used in
  the `authelia-location.conf` snippet has the `AuthRequest` implementation, which is the case for the default
  `auth-request` endpoint.
- `400`: The request could not be processed. The Authelia logs contain the specific reason, which is usually one of:
  - The `X-Original-URL` header is missing or is not a valid URL, see [authelia-location.conf](#authelia-locationconf).
  - The `X-Original-URL` header has the `http` scheme, see
    [NGINX is behind another proxy which terminates TLS](#nginx-is-behind-another-proxy-which-terminates-tls-how-do-i-configure-that).
  - The domain of the `X-Original-URL` header does not match any of the
    [session cookies](../../configuration/session/introduction.md#cookies) configurations.
  - The legacy session configuration is used, which does not have the `authelia_url` option required by the
    `AuthRequest` implementation, see the [session cookies](../../configuration/session/introduction.md#cookies)
    configuration.
- `404`: The URL in the `$upstream_authelia` variable does not match an authz endpoint. The path must be
  `/api/authz/` followed by the name of an endpoint which exists in the configuration, for example the
  `authelia-location-basic.conf` snippet requires you to configure the `auth-request/basic` endpoint.
- `407`: The authz endpoint has the `HeaderProxyAuthorization` strategy, which is not compatible with the
  `AuthRequest` implementation. Use the `HeaderAuthRequestProxyAuthorization` strategy instead.
- `502` or `504`: [NGINX] could not get a timely response from Authelia. Check that Authelia is running, that the
  URL in the `$upstream_authelia` variable is reachable from [NGINX], and that a
  [resolver](#requirements) is configured.

## Kubernetes

Authelia supports some of the [NGINX] based Kubernetes Ingress. See the
[Kubernetes Integration Guide](../kubernetes/nginx-ingress.md) for more information.

## See Also

- [NGINX ngx_http_auth_request_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html)
- [NGINX ngx_http_realip_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_realip_module.html)
- [Forwarded Headers]

[NGINX]: https://www.nginx.com/
[Forwarded Headers]: forwarded-headers
[linuxserver.io]: https://www.linuxserver.io/
