---
title: "NGINX"
description: "An integration guide for Authelia and the NGINX reverse proxy"
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

[NGINX] is a reverse proxy supported by __Authelia__.

*__Important:__ When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only, and you need to understand the proxy
configuration and customize it to your needs. To-that-end, we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Requirements

You need the following to run __Authelia__ with [NGINX]:

* [NGINX] must be built with the `http_auth_request` module which is relatively common
* [NGINX] must be built with the `http_realip` module which is relatively common
* [NGINX] must be built with the `http_set_misc` module or the `nginx-mod-http-set-misc` package if you want to use the
  legacy method and preserve more than one query parameter when redirected to the portal due to a limitation in [NGINX]

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

* Authelia is accessible to [NGINX] process with the hostname `{{< sitevar name="host" nojs="authelia" >}}` on port `{{< sitevar name="port" nojs="9091" >}}` making the URL
  `{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}`. If this is not the case adjust all instances of this as appropriate.
* The [NGINX] configuration is in the folder `/config/nginx`. If this is not the case adjust all instances of this as
  appropriate.
* The URL you wish Authelia to be accessible on is `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`. If this is not the case adjust all
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

```nginx {title="site-confs/auth.conf"}
server {
    listen 80;
    server_name auth.*;

    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name auth.*;

    include /config/nginx/snippets/ssl.conf;

    set $upstream {{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}};

    location / {
        include /config/nginx/snippets/proxy.conf;
        proxy_pass $upstream;
    }

    location = /api/verify {
        proxy_pass $upstream;
    }

    location /api/authz/ {
        proxy_pass $upstream;
    }
}
```

```nginx {title="site-confs/nextcloud.conf"}
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

```nginx {title="site-confs/portainer.conf"}
server {
    listen 80;
    server_name portainer.*;

    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
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

### Supporting Configuration Snippets

The following configuration files are snippets that are used as includes in other files. The includes in the other files
match the headings, so if you wish to put them in a specific location or rename them, then make sure to update the
includes appropriately. Only the [proxy.conf](#proxyconf), [authelia-location.conf](#authelia-locationconf), and
[authelia-authrequest.conf](#authelia-authrequestconf) are required; see the descriptions for the others as to their
use cases.

#### proxy.conf

The following is an example `proxy.conf`. The important directives include the `real_ip` directives which you should read
[Trusted Proxies](#trusted-proxies) section to understand, or set the `X-Forwarded-Proto`, `X-Forwarded-Host`,
`X-Forwarded-URI`, and `X-Forwarded-For` headers.

##### Standard Variant

Generally this variant is the suggested variant.

```nginx {title="proxy.conf"}
## Headers
proxy_set_header Host $host;
proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header X-Forwarded-Host $http_host;
proxy_set_header X-Forwarded-URI $request_uri;
proxy_set_header X-Forwarded-Ssl on;
proxy_set_header X-Forwarded-For $remote_addr;
proxy_set_header X-Real-IP $remote_addr;

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

##### Headers Only Variant

Generally the [standard variant](#standard-variant) is the suggested variant. This variant only contains the required
headers for Authelia to operate.

```nginx {title="proxy.conf"}
## Headers
proxy_set_header Host $host;
proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header X-Forwarded-Host $http_host;
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

*The following snippet is used within the `server` block of a virtual host as a supporting endpoint used by
`auth_request` and is paired with [authelia-authrequest.conf](#authelia-authrequestconf).*

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
    proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
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

#### authelia-authrequest.conf

*The following snippet is used within a `location` block of a virtual host which uses the appropriate location block
and is paired with [authelia-location.conf](#authelia-locationconf).*

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
## value to determine the redirection URL here. It's much simpler and compatible with the mutli-cookie domain easily.

## Modern Method: Set the $redirection_url to the Location header of the response to the Authz endpoint.
auth_request_set $redirection_url $upstream_http_location;

## Modern Method: When there is a 401 response code from the authz endpoint redirect to the $redirection_url.
error_page 401 =302 $redirection_url;

## Legacy Method: Set $target_url to the original requested URL.
## This requires http_set_misc module, replace 'set_escape_uri' with 'set' if you don't have this module.
# set_escape_uri $target_url $scheme://$http_host$request_uri;

## Legacy Method: When there is a 401 response code from the authz endpoint redirect to the portal with the 'rd'
## URL parameter set to $target_url. This requires users update '{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/' with their external authelia URL.
# error_page 401 =302 https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/?rd=$target_url;
```

#### authelia-location-basic.conf

*The following snippet is used within the `server` block of a virtual host as a supporting endpoint used by
`auth_request` and is paired with [authelia-authrequest-basic.conf](#authelia-authrequest-basicconf). This particular
snippet is rarely required. It's only used if you want to only allow
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-location.conf](#authelia-locationconf) instead.*

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
This example assumes you configured an authz endpoint with the name `auth-request/basic` and the
implementation `AuthRequest` which contains the `HeaderAuthorization` and `HeaderProxyAuthorization` strategies.
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
    proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
    proxy_set_header X-Original-Method $request_method;
    proxy_set_header X-Forwarded-Method $request_method;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $http_host;
    proxy_set_header X-Forwarded-URI $request_uri;
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

#### authelia-authrequest-basic.conf

*The following snippet is used within a `location` block of a virtual host which uses the appropriate location block
and is paired with [authelia-location-basic.conf](#authelia-location-basicconf). This particular snippet is rarely
required. It's only used if you want to only allow
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-authrequest.conf](#authelia-authrequestconf) instead.*

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

*The following snippet is used within the `server` block of a virtual host as a supporting endpoint used by
`auth_request` and is paired with [authelia-authrequest-detect.conf](#authelia-authrequest-detectconf). This particular
snippet is rarely required. It's only used if you want to conditionally require
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-location.conf](#authelia-locationconf) instead.*

```nginx {title="authelia-location-detect.conf"}
include /config/nginx/snippets/authelia-location.conf;

set $is_basic_auth ""; # false value

## Detect the client you want to force basic auth for here
## For the example we just match a path on the original request
if ($request_uri = "/force-basic") {
    set $is_basic_auth "true";
    set $upstream_authelia "$upstream_authelia?auth=basic";
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

*The following snippet is used within a `location` block of a virtual host which uses the appropriate location block
and is paired with [authelia-location-detect.conf](#authelia-location-detectconf). This particular snippet is rarely
required. It's only used if you want to conditionally require
[HTTP Basic Authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) for a particular
endpoint. It's recommended to use [authelia-authrequest.conf](#authelia-authrequestconf) instead.*

```nginx {title="authelia-authrequest-detect.conf"}
## Send a subrequest to Authelia to verify if the user is authenticated and has permission to access the resource.
auth_request /internal/authelia/authz;

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
error_page 401 =302 /internal/authelia/authz/detect?rd=$target_url;
```

## Kubernetes

Authelia supports some of the [NGINX] based Kubernetes Ingress. See the
[Kubernetes Integration Guide](../kubernetes/nginx-ingress.md) for more information.

## See Also

* [NGINX ngx_http_auth_request_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html)
* [NGINX ngx_http_realip_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_realip_module.html)
* [Forwarded Headers]

[NGINX]: https://www.nginx.com/
[Forwarded Headers]: forwarded-headers
[linuxserver.io]: https://www.linuxserver.io/
