---
title: "HAProxy"
description: "An integration guide for Authelia and the HAProxy reverse proxy"
summary: "A guide on integrating Authelia with the HAProxy reverse proxy."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 340
toc: true
aliases:
  - /i/haproxy
  - /docs/deployment/supported-proxies/haproxy.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[HAProxy] is a reverse proxy supported by __Authelia__.

*__Important:__ When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only, and you need to understand the proxy
configuration and customize it to your needs. To-that-end, we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Requirements

You need the following to run __Authelia__ with [HAProxy]:

* [HAProxy] 1.8.4+ (2.2.0+ recommended)
  -`USE_LUA=1` set at compile time
  * [haproxy-lua-http](https://github.com/haproxytech/haproxy-lua-http) must be available within the Lua path
    * A `json` library within the Lua path (dependency of haproxy-lua-http, usually found as OS package `lua-json`)
    * With [HAProxy] 2.1.3+ you can use the `lua-prepend-path` configuration option to specify the search path
  * [haproxy-auth-request](https://github.com/TimWolla/haproxy-auth-request/blob/master/auth-request.lua)

## Trusted Proxies

*__Important:__ You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before.*

*__Important:__ The included example is __NOT__ meant for production use. It's used expressly as an example to showcase
how you can configure multiple IP ranges. You should customize this example to fit your specific architecture and needs.
You should only include the specific IP address ranges of the trusted proxies within your architecture and should not
trust entire subnets unless that subnet only has trusted proxies and no other services.*

With [HAProxy] the most convenient method to configure trusted proxies is to create a src ACL from the contents of a
file. The example utilizes this method and trusted proxies can then easily be added or removed from the ACL file.

[HAProxy] implicitly trusts all external proxies by default so it's important you configure this for a trusted
environment.

[HAProxy] by default __does__ trust all other proxies. This means it's essential that you configure this correctly.

In the example we have a `trusted_proxies.src.acl` file which is used by one `http-request del-header X-Forwarded-For`
line in the main configuration which shows an example of not trusting any proxies or alternatively an example on adding
the following networks to the trusted proxy list in [HAProxy]:

* 10.0.0.0/8
* 172.16.0.0/12
* 192.168.0.0/16
* fc00::/7

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

[HAProxy] utilizes the [ForwardAuth](../../reference/guides/proxy-authorization.md#forwardauth) Authz implementation. The
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

Below you will find commented examples of the following configuration:

* Authelia Portal
* Protected Endpoints (Nextcloud)

With this configuration you can protect your virtual hosts with Authelia, by following the steps below:

1. Add host(s) to the `protected-frontends` ACLs to support protection with Authelia. You can separate each subdomain
   with a `|` in the regex, for example:

    ```text
    acl protected-frontends hdr(host) -m reg -i ^(?i)(jenkins|nextcloud|phpmyadmin)\.example\.com
    ```

2. Add host ACL(s) in the form of `host-service`, this will be utilised to route to the correct
backend upon successful authentication, for example:

    ```text
    acl host-jenkins hdr(host) -i jenkins.{{< sitevar name="domain" nojs="example.com" >}}
    acl host-nextcloud hdr(host) -i nextcloud.{{< sitevar name="domain" nojs="example.com" >}}
    acl host-phpmyadmin hdr(host) -i phpmyadmin.{{< sitevar name="domain" nojs="example.com" >}}
    acl host-heimdall hdr(host) -i heimdall.{{< sitevar name="domain" nojs="example.com" >}}
    ```

3. Add backend route for your service(s), for example:

    ```text
    use_backend be_jenkins if host-jenkins
    use_backend be_nextcloud if host-nextcloud
    use_backend be_phpmyadmin if host-phpmyadmin
    use_backend be_heimdall if host-heimdall
    ```

4. Add backend definitions for your service(s), for example:

    ```text
    backend be_jenkins
        server jenkins jenkins:8080
    backend be_nextcloud
        server nextcloud nextcloud:443 ssl verify none
    backend be_phpmyadmin
        server phpmyadmin phpmyadmin:80
    backend be_heimdall
        server heimdall heimdall:443 ssl verify none
    ```

### Common

```text {title="trusted_proxies.src.acl"}
10.0.0.0/8
172.16.0.0/12
192.168.0.0/16
fc00::/7
```

### Standard Example

```text {title="haproxy.cfg"}
global
    # Path to haproxy-lua-http, below example assumes /usr/local/etc/haproxy/haproxy-lua-http/http.lua
    lua-prepend-path /usr/local/etc/haproxy/?/http.lua
    # Path to haproxy-auth-request
    lua-load /usr/local/etc/haproxy/auth-request.lua
    log stdout format raw local0 debug

defaults
    mode http
    log global
    option httplog

frontend fe_http
    bind *:443 ssl crt {{< sitevar name="domain" nojs="example.com" >}}.pem

    ## Trusted Proxies.
    http-request del-header X-Forwarded-For

    ## Comment the above directive and the two directives below to enable the trusted proxies ACL.
    # acl src-trusted_proxies src -f trusted_proxies.src.acl
    # http-request del-header X-Forwarded-For if !src-trusted_proxies

    ## Ensure X-Forwarded-For is set for the auth request.
    acl hdr-xff_exists req.hdr(X-Forwarded-For) -m found
    http-request set-header X-Forwarded-For %[src] if !hdr-xff_exists
    option forwardfor

    # Host ACLs
    acl protected-frontends hdr(Host) -m reg -i ^(?i)(nextcloud|heimdall)\.example\.com
    acl host-authelia hdr(Host) -i {{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
    acl host-nextcloud hdr(Host) -i nextcloud.{{< sitevar name="domain" nojs="example.com" >}}
    acl host-heimdall hdr(Host) -i heimdall.{{< sitevar name="domain" nojs="example.com" >}}

    http-request set-var(req.scheme) str(https) if { ssl_fc }
    http-request set-var(req.scheme) str(http) if !{ ssl_fc }
    http-request set-var(req.questionmark) str(?) if { query -m found }

    # Required Headers
    http-request set-header X-Forwarded-Method %[method]
    http-request set-header X-Forwarded-Proto  %[var(req.scheme)]
    http-request set-header X-Forwarded-Host   %[req.hdr(Host)]
    http-request set-header X-Forwarded-URI    %[path]%[var(req.questionmark)]%[query]

    # Protect endpoints with haproxy-auth-request and Authelia
    http-request lua.auth-intercept be_authelia /api/authz/forward-auth HEAD * remote-user,remote-groups,remote-name,remote-email - if protected-frontends
    http-request deny if protected-frontends !{ var(txn.auth_response_successful) -m bool } { var(txn.auth_response_code) -m int 403 }
    http-request redirect location %[var(txn.auth_response_location)] if protected-frontends !{ var(txn.auth_response_successful) -m bool }

    # Authelia backend route
    use_backend be_authelia if host-authelia

    # Service backend route(s)
    use_backend be_nextcloud if host-nextcloud
    use_backend be_heimdall if host-heimdall

backend be_authelia
    server authelia {{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}

backend be_nextcloud
    ## Pass the Set-Cookie response headers to the user.
    acl set_cookie_exist var(req.auth_response_header.set_cookie) -m found
    http-response set-header Set-Cookie %[var(req.auth_response_header.set_cookie)] if set_cookie_exist

    server nextcloud nextcloud:443 ssl verify none

backend be_heimdall
    ## Pass the Set-Cookie response headers to the user.
    acl set_cookie_exist var(req.auth_response_header.set_cookie) -m found
    http-response set-header Set-Cookie %[var(req.auth_response_header.set_cookie)] if set_cookie_exist

    server heimdall heimdall:443 ssl verify none
```

### TLS Example

```text {title="haproxy.cfg"}
global
    # Path to haproxy-lua-http, below example assumes /usr/local/etc/haproxy/haproxy-lua-http/http.lua
    lua-prepend-path /usr/local/etc/haproxy/?/http.lua
    # Path to haproxy-auth-request
    lua-load /usr/local/etc/haproxy/auth-request.lua
    log stdout format raw local0 debug

defaults
    mode http
    log global
    option httplog
    option forwardfor

frontend fe_http
    bind *:443 ssl crt /usr/local/etc/haproxy/haproxy.pem

    ## Trusted Proxies.
    http-request del-header X-Forwarded-For

    ## Comment the above directive and the two directives below to enable the trusted proxies ACL.
    # acl src-trusted_proxies src -f trusted_proxies.src.acl
    # http-request del-header X-Forwarded-For if !src-trusted_proxies

    ## Ensure X-Forwarded-For is set for the auth request.
    acl hdr-xff_exists req.hdr(X-Forwarded-For) -m found
    http-request set-header X-Forwarded-For %[src] if !hdr-xff_exists
    option forwardfor

    # Host ACLs
    acl protected-frontends hdr(Host) -m reg -i ^(?i)(nextcloud|heimdall)\.example\.com
    acl host-authelia hdr(Host) -i auth.{{< sitevar name="domain" nojs="example.com" >}}
    acl host-nextcloud hdr(Host) -i nextcloud.{{< sitevar name="domain" nojs="example.com" >}}
    acl host-heimdall hdr(Host) -i heimdall.{{< sitevar name="domain" nojs="example.com" >}}

    http-request set-var(req.scheme) str(https) if { ssl_fc }
    http-request set-var(req.scheme) str(http) if !{ ssl_fc }
    http-request set-var(req.questionmark) str(?) if { query -m found }

    # Required Headers
    http-request set-header X-Forwarded-Method %[method]
    http-request set-header X-Forwarded-Proto  %[var(req.scheme)]
    http-request set-header X-Forwarded-Host   %[req.hdr(Host)]
    http-request set-header X-Forwarded-URI    %[path]%[var(req.questionmark)]%[query]

    # Protect endpoints with haproxy-auth-request and Authelia
    http-request lua.auth-intercept be_authelia_proxy /api/authz/forward-auth HEAD * remote-user,remote-groups,remote-name,remote-email - if protected-frontends
    http-request deny if protected-frontends !{ var(txn.auth_response_successful) -m bool } { var(txn.auth_response_code) -m int 403 }
    http-request redirect location %[var(txn.auth_response_location)] if protected-frontends !{ var(txn.auth_response_successful) -m bool }

    # Authelia backend route
    use_backend be_authelia if host-authelia

    # Service backend route(s)
    use_backend be_nextcloud if host-nextcloud
    use_backend be_heimdall if host-heimdall

backend be_authelia
    server authelia {{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}

backend be_authelia_proxy
    mode http
    server proxy 127.0.0.1:9092

listen authelia_proxy
    mode http
    bind 127.0.0.1:9092
    server authelia {{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}} ssl verify none

backend be_nextcloud
    ## Pass the Set-Cookie response headers to the user.
    acl set_cookie_exist var(req.auth_response_header.set_cookie) -m found
    http-response set-header Set-Cookie %[var(req.auth_response_header.set_cookie)] if set_cookie_exist

    server nextcloud nextcloud:443 ssl verify none

backend be_heimdall
    ## Pass the Set-Cookie response headers to the user.
    acl set_cookie_exist var(req.auth_response_header.set_cookie) -m found
    http-response set-header Set-Cookie %[var(req.auth_response_header.set_cookie)] if set_cookie_exist

    server heimdall heimdall:443 ssl verify none
```

## See Also

* [HAProxy Auth Request lua plugin Documentation](https://github.com/TimWolla/haproxy-auth-request)
* [Forwarded Headers]

[HAproxy]: https://www.haproxy.org/
[Forwarded Headers]: forwarded-headers
