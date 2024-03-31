---
title: "Caddy"
description: "An integration guide for Authelia and the Caddy reverse proxy"
summary: "A guide on integrating Authelia with the Caddy reverse proxy."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 320
toc: true
aliases:
  - /i/caddy
  - /docs/deployment/supported-proxies/caddy.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[Caddy] is a reverse proxy supported by __Authelia__.

__Authelia__ offers integration support for the official forward auth integration method Caddy provides, we don't
officially support any plugin that supports this though we don't specifically prevent such plugins working and there may
be plugins that work fine provided they support the forward authentication specification correctly.

*__Important:__ When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only, and you need to understand the proxy
configuration and customize it to your needs. To-that-end, we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Requirements

You need the following to run __Authelia__ with [Caddy]:

* [Caddy] [v2.5.1](https://github.com/caddyserver/caddy/releases/tag/v2.5.1) or greater

## Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

*__Important:__ The included example is __NOT__ meant for production use. It's used expressly as an example to showcase
how you can configure multiple IP ranges. You should customize this example to fit your specific architecture and needs.
You should only include the specific IP address ranges of the trusted proxies within your architecture and should not
trust entire subnets unless that subnet only has trusted proxies and no other services.*

[Caddy] by default doesn't trust any other proxies and removes potentially fabricated headers that are likely to lead
to security issues, and it is difficult to configure this incorrectly. This is an important security feature that is
common with proxies with good security practices.

You should read the [Caddy Trusted Proxies Documentation] as part of configuring this. It's important to ensure you take
the time to configure this carefully and correctly.

In the example we have a commented `trusted_proxies` directive which shows an example on adding the following networks
to the trusted proxy list in [Caddy]:

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

[Caddy] utilizes the [ForwardAuth](../../reference/guides/proxy-authorization.md#forwardauth) Authz implementation. The
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

Below you will find commented examples of the following configuration:

* Authelia Portal
* Protected Endpoint (Nextcloud)

### Basic examples

This example is the preferred example for integration with [Caddy]. There is an [advanced example](#advanced-example)
but we *__strongly urge__* anyone who needs to use this for a particular reason to either reach out to us or Caddy for
support to ensure the basic example covers your use case in a secure way.

#### Subdomain

{{< details "Caddyfile" >}}
```caddyfile
## It is important to read the following document before enabling this section:
##     https://www.authelia.com/integration/proxies/caddy/#trusted-proxies
(trusted_proxy_list) {
       ## Uncomment & adjust the following line to configure specific ranges which should be considered as trustworthy.
       # trusted_proxies 10.0.0.0/8 172.16.0.0/12 192.168.0.0/16 fc00::/7
}

# Authelia Portal.
auth.example.com {
        reverse_proxy authelia:9091 {
                ## This import needs to be included if you're relying on a trusted proxies configuration.
                import trusted_proxy_list
        }
}

# Protected Endpoint.
nextcloud.example.com {
        forward_auth authelia:9091 {
                uri /api/authz/forward-auth
                ## The following commented line is for configuring the Authelia URL in the proxy. We strongly suggest
                ## this is configured in the Session Cookies section of the Authelia configuration.
                # uri /api/authz/forward-auth?authelia_url=https://auth.example.com/
                copy_headers Remote-User Remote-Groups Remote-Email Remote-Name

                ## This import needs to be included if you're relying on a trusted proxies configuration.
                import trusted_proxy_list
        }
        reverse_proxy nextcloud:80 {
                ## This import needs to be included if you're relying on a trusted proxies configuration.
                import trusted_proxy_list
        }
}
```
{{< /details >}}

#### Subpath

*__Important:__ In order to use a subpath, you must also update your Authelia
[server address configuration](../../configuration/miscellaneous/server.md#address) to listen on the new endpoint.*

{{< details "Caddyfile" >}}
```caddyfile
## It is important to read the following document before enabling this section:
##     https://www.authelia.com/integration/proxies/caddy/#trusted-proxies
(trusted_proxy_list) {
       ## Uncomment & adjust the following line to configure specific ranges which should be considered as trustworthy.
       # trusted_proxies 10.0.0.0/8 172.16.0.0/12 192.168.0.0/16 fc00::/7
}

example.com {
        # Authelia Portal.
        @authelia path /authelia /authelia/*
        handle @authelia {
                reverse_proxy authelia:9091 {
                        ## This import needs to be included if you're relying on a trusted proxies configuration.
                        import trusted_proxy_list
                }
        }

        # Protected Endpoint.
        @nextcloud path /nextcloud /nextcloud/*
        handle @nextcloud {
                forward_auth authelia:9091 {
                        uri /api/authz/forward-auth?authelia_url=https://example.com/authelia/
                        copy_headers Remote-User Remote-Groups Remote-Email Remote-Name

                        ## This import needs to be included if you're relying on a trusted proxies configuration.
                        import trusted_proxy_list
                }
                reverse_proxy nextcloud:80 {
                        ## This import needs to be included if you're relying on a trusted proxies configuration.
                        import trusted_proxy_list
                }
        }
}
```
{{< /details >}}

### Advanced example

The advanced example allows for more flexible customization, however the [basic example](#basic-examples) should be
preferred in *most* situations. If you are unsure of what you're doing please don't use this method.

*__Important:__ Making a mistake when configuring the advanced example could lead to authentication bypass or errors.*

{{< details "Caddyfile" >}}
```caddyfile
## It is important to read the following document before enabling this section:
##     https://www.authelia.com/integration/proxies/caddy/#trusted-proxies
(trusted_proxy_list) {
       ## Uncomment & adjust the following line to configure specific ranges which should be considered as trustworthy.
       # trusted_proxies 10.0.0.0/8 172.16.0.0/12 192.168.0.0/16 fc00::/7
}

# Authelia Portal.
auth.example.com {
        reverse_proxy authelia:9091 {
                ## This import needs to be included if you're relying on a trusted proxies configuration.
                import trusted_proxy_list
        }
}

# Protected Endpoint.
nextcloud.example.com {
        reverse_proxy authelia:9091 {
                ## This import needs to be included if you're relying on a trusted proxies configuration.
                import trusted_proxy_list

                method GET
                rewrite "/api/authz/forward-auth?authelia_url=https://auth.example.com/"

                header_up X-Forwarded-Method {method}
                header_up X-Forwarded-URI {uri}

                ## If the auth request:
                ##   1. Responds with a status code IN the 200-299 range.
                ## Then:
                ##   1. Proxy the request to the backend.
                ##   2. Copy the relevant headers from the auth request and provide them to the backend.
                @good status 2xx
                handle_response @good {
                        request_header Remote-User {http.reverse_proxy.header.Remote-User}
                        request_header Remote-Groups {http.reverse_proxy.header.Remote-Groups}
                        request_header Remote-Email {http.reverse_proxy.header.Remote-Email}
                        request_header Remote-Name {http.reverse_proxy.header.Remote-Name}
                }
        }

        reverse_proxy nextcloud:80 {
                ## This import needs to be included if you're relying on a trusted proxies configuration.
                import trusted_proxy_list
        }
}
```
{{< /details >}}

## See Also

* [Caddy General Documentation](https://caddyserver.com/docs/)
* [Caddy Forward Auth Documentation]
* [Caddy Trusted Proxies Documentation]
* [Caddy Snippet] Documentation
* [Forwarded Headers]

[Caddy]: https://caddyserver.com
[Caddy Snippet]: https://caddyserver.com/docs/caddyfile/concepts#snippets
[Caddy Forward Auth Documentation]: https://caddyserver.com/docs/caddyfile/directives/forward_auth
[Caddy Trusted Proxies Documentation]: https://caddyserver.com/docs/caddyfile/directives/reverse_proxy#trusted_proxies
[Forwarded Headers]: forwarded-headers
