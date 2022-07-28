---
title: "HAProxy"
description: "An integration guide for Authelia and the HAProxy reverse proxy"
lead: "A guide on integrating Authelia with the HAProxy reverse proxy."
date: 2020-02-29T01:43:59+01:00
draft: false
images: []
menu:
  integration:
    parent: "proxies"
weight: 340
toc: true
aliases:
  - /i/haproxy
  - /docs/deployment/supported-proxies/haproxy.html
---

[HAProxy] is a reverse proxy supported by __Authelia__.

*__Important:__ When using these guides it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only and you need to understand the proxy
configuration and customize it to your needs. To-that-end we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section.*

## Get Started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get Started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
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
* 172.16.0.0/16
* 192.168.0.0/16
* fc00::/7

## Configuration

Below you will find commented examples of the following configuration:

* Authelia Portal
* Protected Endpoint (Nextcloud)
* Protected Endpoint with `Authorization` header for basic authentication (Heimdall)

With this configuration you can protect your virtual hosts with Authelia, by following the steps below:

1. Add host(s) to the `protected-frontends` or `protected-frontends-basic` ACLs to support protection with Authelia.
You can separate each subdomain with a `|` in the regex, for example:

    ```text
    acl protected-frontends hdr(host) -m reg -i ^(?i)(jenkins|nextcloud|phpmyadmin)\.example\.com
    acl protected-frontends-basic hdr(host) -m reg -i ^(?i)(heimdall)\.example\.com
    ```

2. Add host ACL(s) in the form of `host-service`, this will be utilised to route to the correct
backend upon successful authentication, for example:

    ```text
    acl host-jenkins hdr(host) -i jenkins.example.com
    acl host-nextcloud hdr(host) -i nextcloud.example.com
    acl host-phpmyadmin hdr(host) -i phpmyadmin.example.com
    acl host-heimdall hdr(host) -i heimdall.example.com
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

### Secure Authelia with TLS

There is a [known limitation](https://github.com/TimWolla/haproxy-auth-request/issues/12) with haproxy-auth-request with
regard to TLS-enabled backends. If you want to run Authelia TLS enabled the recommended workaround utilises [HAProxy]
itself to proxy the requests. This comes at a cost of two additional TCP connections, but allows the full [HAProxy]
configuration flexibility with regard to TLS verification as well as header rewriting. An example of this configuration
is also be provided below.

#### Configuration

##### trusted_proxies.src.acl

```text
10.0.0.0/8
172.16.0.0/12
192.168.0.0/16
fc00::/7
```

##### haproxy.cfg

```text
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
    bind *:443 ssl crt example.com.pem

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
    acl protected-frontends hdr(host) -m reg -i ^(?i)(nextcloud)\.example\.com
    acl protected-frontends-basic hdr(host) -m reg -i ^(?i)(heimdall)\.example\.com
    acl host-authelia hdr(host) -i auth.example.com
    acl host-nextcloud hdr(host) -i nextcloud.example.com
    acl host-heimdall hdr(host) -i heimdall.example.com

    # This is required if utilising basic auth with /api/verify?auth=basic
    http-request set-var(txn.host) hdr(Host)

    http-request set-var(req.scheme) str(https) if { ssl_fc }
    http-request set-var(req.scheme) str(http) if !{ ssl_fc }
    http-request set-var(req.questionmark) str(?) if { query -m found }

    # These are optional if you wish to use the Methods rule in the access_control section.
    #http-request set-var(req.method) str(CONNECT) if { method CONNECT }
    #http-request set-var(req.method) str(GET) if { method GET }
    #http-request set-var(req.method) str(HEAD) if { method HEAD }
    #http-request set-var(req.method) str(OPTIONS) if { method OPTIONS }
    #http-request set-var(req.method) str(POST) if { method POST }
    #http-request set-var(req.method) str(TRACE) if { method TRACE }
    #http-request set-var(req.method) str(PUT) if { method PUT }
    #http-request set-var(req.method) str(PATCH) if { method PATCH }
    #http-request set-var(req.method) str(DELETE) if { method DELETE }
    #http-request set-header X-Forwarded-Method %[var(req.method)]

    # Required headers
    http-request set-header X-Real-IP %[src]
    http-request set-header X-Forwarded-Method %[var(req.method)]
    http-request set-header X-Forwarded-Proto %[var(req.scheme)]
    http-request set-header X-Forwarded-Host %[req.hdr(Host)]
    http-request set-header X-Forwarded-Uri %[path]%[var(req.questionmark)]%[query]

    # Protect endpoints with haproxy-auth-request and Authelia
    http-request lua.auth-request be_authelia /api/verify if protected-frontends
    # Force `Authorization` header via query arg to /api/verify
    http-request lua.auth-request be_authelia /api/verify?auth=basic if protected-frontends-basic

    # Redirect protected-frontends to Authelia if not authenticated
    http-request redirect location https://auth.example.com/?rd=%[var(req.scheme)]://%[base]%[var(req.questionmark)]%[query] if protected-frontends !{ var(txn.auth_response_successful) -m bool }
    # Send 401 and pass `WWW-Authenticate` header on protected-frontend-basic if not pre-authenticated
    http-request set-var(txn.auth) var(req.auth_response_header.www_authenticate) if protected-frontends-basic !{ var(txn.auth_response_successful) -m bool }
    http-response deny deny_status 401 hdr WWW-Authenticate %[var(txn.auth)] if { var(txn.host) -m reg -i ^(?i)(heimdall)\.example\.com } !{ var(txn.auth_response_successful) -m bool }

    # Authelia backend route
    use_backend be_authelia if host-authelia

    # Service backend route(s)
    use_backend be_nextcloud if host-nextcloud
    use_backend be_heimdall if host-heimdall

backend be_authelia
    server authelia authelia:9091

backend be_nextcloud
    # Pass Remote-User, Remote-Name, Remote-Email and Remote-Groups headers
    acl remote_user_exist var(req.auth_response_header.remote_user) -m found
    acl remote_groups_exist var(req.auth_response_header.remote_groups) -m found
    acl remote_name_exist var(req.auth_response_header.remote_name) -m found
    acl remote_email_exist var(req.auth_response_header.remote_email) -m found
    http-request set-header Remote-User %[var(req.auth_response_header.remote_user)] if remote_user_exist
    http-request set-header Remote-Groups %[var(req.auth_response_header.remote_groups)] if remote_groups_exist
    http-request set-header Remote-Name %[var(req.auth_response_header.remote_name)] if remote_name_exist
    http-request set-header Remote-Email %[var(req.auth_response_header.remote_email)] if remote_email_exist

    server nextcloud nextcloud:443 ssl verify none

backend be_heimdall
    # Pass Remote-User, Remote-Name, Remote-Email and Remote-Groups headers
    acl remote_user_exist var(req.auth_response_header.remote_user) -m found
    acl remote_groups_exist var(req.auth_response_header.remote_groups) -m found
    acl remote_name_exist var(req.auth_response_header.remote_name) -m found
    acl remote_email_exist var(req.auth_response_header.remote_email) -m found
    http-request set-header Remote-User %[var(req.auth_response_header.remote_user)] if remote_user_exist
    http-request set-header Remote-Groups %[var(req.auth_response_header.remote_groups)] if remote_groups_exist
    http-request set-header Remote-Name %[var(req.auth_response_header.remote_name)] if remote_name_exist
    http-request set-header Remote-Email %[var(req.auth_response_header.remote_email)] if remote_email_exist

    server heimdall heimdall:443 ssl verify none
```

##### haproxy.cfg (TLS enabled Authelia)

```text
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

    # Host ACLs
    acl protected-frontends hdr(host) -m reg -i ^(?i)(nextcloud)\.example\.com
    acl protected-frontends-basic hdr(host) -m reg -i ^(?i)(heimdall)\.example\.com
    acl host-authelia hdr(host) -i auth.example.com
    acl host-nextcloud hdr(host) -i nextcloud.example.com
    acl host-heimdall hdr(host) -i heimdall.example.com

    # This is required if utilising basic auth with /api/verify?auth=basic
    http-request set-var(txn.host) hdr(Host)

    http-request set-var(req.scheme) str(https) if { ssl_fc }
    http-request set-var(req.scheme) str(http) if !{ ssl_fc }
    http-request set-var(req.questionmark) str(?) if { query -m found }

    # These are optional if you wish to use the Methods rule in the access_control section.
    #http-request set-var(req.method) str(CONNECT) if { method CONNECT }
    #http-request set-var(req.method) str(GET) if { method GET }
    #http-request set-var(req.method) str(HEAD) if { method HEAD }
    #http-request set-var(req.method) str(OPTIONS) if { method OPTIONS }
    #http-request set-var(req.method) str(POST) if { method POST }
    #http-request set-var(req.method) str(TRACE) if { method TRACE }
    #http-request set-var(req.method) str(PUT) if { method PUT }
    #http-request set-var(req.method) str(PATCH) if { method PATCH }
    #http-request set-var(req.method) str(DELETE) if { method DELETE }
    #http-request set-header X-Forwarded-Method %[var(req.method)]

    # Required headers
    http-request set-header X-Real-IP %[src]
    http-request set-header X-Forwarded-Proto %[var(req.scheme)]
    http-request set-header X-Forwarded-Host %[req.hdr(Host)]
    http-request set-header X-Forwarded-Uri %[path]%[var(req.questionmark)]%[query]

    # Protect endpoints with haproxy-auth-request and Authelia
    http-request lua.auth-request be_authelia_proxy /api/verify if protected-frontends
    # Force `Authorization` header via query arg to /api/verify
    http-request lua.auth-request be_authelia_proxy /api/verify?auth=basic if protected-frontends-basic

    # Redirect protected-frontends to Authelia if not authenticated
    http-request redirect location https://auth.example.com/?rd=%[var(req.scheme)]://%[base]%[var(req.questionmark)]%[query] if protected-frontends !{ var(txn.auth_response_successful) -m bool }
    # Send 401 and pass `WWW-Authenticate` header on protected-frontend-basic if not pre-authenticated
    http-request set-var(txn.auth) var(req.auth_response_header.www_authenticate) if protected-frontends-basic !{ var(txn.auth_response_successful) -m bool }
    http-response deny deny_status 401 hdr WWW-Authenticate %[var(txn.auth)] if { var(txn.host) -m reg -i ^(?i)(heimdall)\.example\.com } !{ var(txn.auth_response_successful) -m bool }

    # Authelia backend route
    use_backend be_authelia if host-authelia

    # Service backend route(s)
    use_backend be_nextcloud if host-nextcloud
    use_backend be_heimdall if host-heimdall

backend be_authelia
    server authelia authelia:9091

backend be_authelia_proxy
    mode http
    server proxy 127.0.0.1:9092

listen authelia_proxy
    mode http
    bind 127.0.0.1:9092
    server authelia authelia:9091 ssl verify none

backend be_nextcloud
    # Pass Remote-User, Remote-Name, Remote-Email and Remote-Groups headers
    acl remote_user_exist var(req.auth_response_header.remote_user) -m found
    acl remote_groups_exist var(req.auth_response_header.remote_groups) -m found
    acl remote_name_exist var(req.auth_response_header.remote_name) -m found
    acl remote_email_exist var(req.auth_response_header.remote_email) -m found
    http-request set-header Remote-User %[var(req.auth_response_header.remote_user)] if remote_user_exist
    http-request set-header Remote-Groups %[var(req.auth_response_header.remote_groups)] if remote_groups_exist
    http-request set-header Remote-Name %[var(req.auth_response_header.remote_name)] if remote_name_exist
    http-request set-header Remote-Email %[var(req.auth_response_header.remote_email)] if remote_email_exist

    server nextcloud nextcloud:443 ssl verify none

backend be_heimdall
    # Pass Remote-User, Remote-Name, Remote-Email and Remote-Groups headers
    acl remote_user_exist var(req.auth_response_header.remote_user) -m found
    acl remote_groups_exist var(req.auth_response_header.remote_groups) -m found
    acl remote_name_exist var(req.auth_response_header.remote_name) -m found
    acl remote_email_exist var(req.auth_response_header.remote_email) -m found
    http-request set-header Remote-User %[var(req.auth_response_header.remote_user)] if remote_user_exist
    http-request set-header Remote-Groups %[var(req.auth_response_header.remote_groups)] if remote_groups_exist
    http-request set-header Remote-Name %[var(req.auth_response_header.remote_name)] if remote_name_exist
    http-request set-header Remote-Email %[var(req.auth_response_header.remote_email)] if remote_email_exist

    server heimdall heimdall:443 ssl verify none
```

## See Also

* [HAProxy Auth Request lua plugin Documentation](https://github.com/TimWolla/haproxy-auth-request)
* [Forwarded Headers]

[HAproxy]: https://www.haproxy.org/
[Forwarded Headers]: fowarded-headers
