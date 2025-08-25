---
title: "NGINX Proxy Manager"
description: "An integration guide for Authelia and the NGINX Proxy Manager reverse proxy"
summary: "A guide on integrating Authelia with NGINX Proxy Manager."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 352
toc: true
aliases:
  - /i/npm
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[NGINX Proxy Manager] is supported by __Authelia__. It's a [NGINX] proxy with a configuration UI.

*__Important:__ When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only, and you need to understand the proxy
configuration and customize it to your needs. To-that-end, we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Requirements

[NGINX Proxy Manager] supports the required [NGINX](../nginx.md#requirements) requirements for __Authelia__ out-of-the-box.

## Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

To configure trusted proxies for [NGINX Proxy Manager] see the [NGINX] section on
[Trusted Proxies](../nginx.md#trusted-proxies). Adapting this to [NGINX Proxy Manager] is beyond the scope of
this documentation.

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

## Docker Compose

The following docker compose example has various applications suitable for setting up an example environment.

```yaml {title="compose.yml"}
---
networks:
  net:
    driver: 'bridge'

services:
  nginx:
    container_name: 'nginx'
    image: 'jc21/nginx-proxy-manager'
    restart: 'unless-stopped'
    networks:
      net:
        aliases: []
    ports:
      - '80:80'
      - '81:81'
      - '443:443'
    volumes:
      - '${PWD}/data/nginx-proxy-manager/data:/data'
      - '${PWD}/data/nginx-proxy-manager/letsencrypt:/etc/letsencrypt'
      - '${PWD}/data/nginx/snippets:/snippets'
    environment:
      TZ: 'Australia/Melbourne'
  authelia:
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    image: 'authelia/authelia'
    restart: 'unless-stopped'
    networks:
      net:
        aliases: []
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
    environment:
      TZ: 'Australia/Melbourne'
...
```

## Configuration

### Assumptions

*__Important:__ Our examples make assumptions about your configuration. These assumptions represent sections that
either most likely require an adjustment, or may require an adjustment if you're not configuring it in the same way.*

* The domain for Authelia is `{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}` which should be adjusted in all examples and snippets to your actual
  domain.
* The required configuration snippets are mounted in the container or otherwise available in the `/snippets/` directory.
  If you choose a different directory, you're required to adjust every instance of `/snippets/` appropriately to your
  needs.
* You have not configured the Authelia configuration YAML with a server TLS certificate/key.
* You are running Authelia on the default port.
* You are running Authelia with the `container_name` of `authelia` or the Authelia process is otherwise resolvable by
  [NGINX Proxy Manager] as `authelia`.
* If you want to use a [Custom Location](#protected-application-custom-locations) and wish for it to be protected, you should
  follow the [Protected Application Custom Location](#protected-application-custom-locations) guide.

### Snippets

The examples assume you've mounted a volume containing the relevant
[NGINX Snippets](../nginx.md#supporting-configuration-snippets) from the [NGINX Integration Guide](../nginx.md). The
suggested snippets are the `proxy.conf`, `authelia-location.conf`, and `authelia-authrequest.conf`. It may be fine to
substitute the standard variant of the `proxy.conf` for the headers only variant but this is untested.  You will need `websocket.conf` if any protected applications require websockets.

These snippets make the addition of a protected proxy host substantially easier.

### Authelia Portal

The Authelia portal requires minimal configuration.

1. Create a new `Proxy Host`.
2. Set the following items in the `Details` tab:
   * Domain Names: `{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
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

{{< picture src="authelia.details.png" alt="Step 2" width="450" >}}

Authelia Portal `Advanced` tab example:

{{< picture src="authelia.advanced.png" alt="Step 4" width="450" >}}

### Protected Application

The following example shows how to configure a protected application. We often use Nextcloud for such examples.

1. Create a new `Proxy Host`.
2. Set the following items in the `Details` tab:
   * Domain Names: `nextcloud.{{< sitevar name="domain" nojs="example.com" >}}`
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
{{< callout context="note" title="Websockets" icon="outline/info-circle" >}}
Note that because we are using the advanced configuration tab, the switches on the `Details` tab will have no effect.  If websockets are required for a protected application, you must include the websocket.conf from the [NGINX Snippets](../nginx.md#supporting-configuration-snippets). 
{{< /callout >}}

#### Protected Application Screenshots

Protected Application (Nextcloud) `Details` tab example:

{{< picture src="nextcloud.details.png" alt="Step 2" width="450" >}}

Protected Application (Nextcloud) `Advanced` tab example:

{{< picture src="protectedapp.advanced.png" alt="Step 4" width="450" >}}

#### Protected Application Custom Locations

It's important to note if you define locations in the `Custom Locations` tab of a proxy host that they will not be
checked with Authelia for authorization effectively bypassing the authorization policies you implement. If you want a
custom location then you can also define this in the advanced tab.

To replicate the `Custom Location` tab below a location block can be *__ADDED__* to the
[Protected Application](#protected-application) `Advanced` tab:

```nginx
location /custom {
    include /snippets/proxy.conf;
    include /snippets/authelia-authrequest.conf;
    proxy_pass http://192.168.1.20:8080;
}
```

{{< picture src="protectedapp.customlocation.png" alt="Custom Location" width="450" >}}

#### Proxy Hosts Screenshot

The following screenshot shows an example of following the directions for the Authelia Portal and two applications:

{{< picture src="proxyhosts.png" alt="Step 4" width="450" >}}

## See Also

* [NGINX Proxy Manager Documentation](https://nginxproxymanager.com/setup/)
* [NGINX ngx_http_auth_request_module Module Documentation](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html)
* [Forwarded Headers]

[NGINX Proxy Manager]: https://nginxproxymanager.com/
[NGINX]: https://www.nginx.com/
[Forwarded Headers]: ../forwarded-headers
