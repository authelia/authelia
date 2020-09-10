---
layout: default
title: HAProxy
parent: Proxy Integration
grand_parent: Deployment
nav_order: 1
---

# HAProxy

[HAProxy] is a reverse proxy supported by **Authelia**.

## Requirements

You need the following to run Authelia with HAProxy:

* HAProxy 1.8.4+ (2.2.0+ recommended)
  * `USE_LUA=1` set at compile time
  * [haproxy-lua-http](https://github.com/haproxytech/haproxy-lua-http) must be available within the Lua path
    * A `json` library within the Lua path (dependency of haproxy-lua-http, usually found as OS package `lua-json`)
    * With HAProxy 2.1.3+ you can use the [`lua-prepend-path`] configuration option to specify the search path.
  * [haproxy-auth-request](https://github.com/TimWolla/haproxy-auth-request/blob/master/auth-request.lua)


## Configuration

Below you will find commented examples of the following configuration:

* Authelia portal
* Protected endpoint (Nextcloud)
* [haproxy-auth-request](https://github.com/TimWolla/haproxy-auth-request/blob/master/auth-request.lua)

With this configuration you can protect your virtual hosts with Authelia, by following the steps below:
1. Add host(s) to the `protected-frontends` ACL to support protection with Authelia.
You can separate each subdomain with a `|` in the regex, for example:
    ```
    acl protected-frontends hdr(host) -m reg -i ^(jenkins|nextcloud|phpmyadmin)\.example\.com
    ```
2. Add host ACL(s) in the form of `host-service`, this will be utilised to route to the correct
backend upon successful authentication, for example:
    ```
    acl host-jenkins hdr(host) -i jenkins.example.com
    acl host-nextcloud hdr(host) -i nextcloud.example.com
    acl host-phpmyadmin hdr(host) -i phpmyadmin.example.com
    ```
3. Add backend route for your service(s), for example:
    ```
    use_backend be_jenkins if host-jenkins
    use_backend be_nextcloud if host-nextcloud
    use_backend be_phpmyadmin if host-phpmyadmin
    ```
4. Add backend definitions for your service(s), for example:
    ```
    backend be_jenkins
        server jenkins jenkins:8080
    backend be_nextcloud
        server nextcloud nextcloud:443 ssl verify none
    backend be_phpmyadmin
        server phpmyadmin phpmyadmin:80
    ```

### Secure Authelia with TLS
There is a [known limitation](https://github.com/TimWolla/haproxy-auth-request/issues/12) with haproxy-auth-request with regard to TLS-enabled backends.
If you want to run Authelia TLS enabled the recommended workaround utilises HAProxy itself to proxy the requests.
This comes at a cost of two additional TCP connections, but allows the full HAProxy configuration flexbility with regard
to TLS verification as well as header rewriting. An example of this configuration is also be provided below.

#### Configuration

##### haproxy.cfg
```
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
    acl protected-frontends hdr(host) -m reg -i ^(nextcloud)\.example\.com
    acl host-authelia hdr(host) -i auth.example.com
    acl host-nextcloud hdr(host) -i nextcloud.example.com

    http-request set-var(req.scheme) str(https) if { ssl_fc }
    http-request set-var(req.scheme) str(http) if !{ ssl_fc }
    http-request set-var(req.questionmark) str(?) if { query -m found }
   
    # Headers to construct redirection URL
    http-request set-header X-Real-IP %[src]
    http-request set-header X-Forwarded-Proto %[var(req.scheme)]
    http-request set-header X-Forwarded-Host %[req.hdr(Host)]
    http-request add-header X-Forwarded-Port %[dst_port]
    http-request set-header X-Forwarded-Uri %[path]%[var(req.questionmark)]%[query]

    # Protect endpoints with haproxy-auth-request and Authelia
    http-request lua.auth-request be_authelia /api/verify if protected-frontends
   
    # Authelia backend route
    use_backend be_authelia if host-authelia
    # Redirect protected-frontends to Authelia if not authenticated
    use_backend be_authelia if protected-frontends !{ var(txn.auth_response_successful) -m bool }
    # Service backend route(s)
    use_backend be_nextcloud if host-nextcloud

backend be_authelia
    server authelia authelia:9091

backend be_nextcloud
    # Pass Remote-User and Remote-Groups headers   
    acl remote_user_exist var(req.auth_response_header.remote_user) -m found
    acl remote_groups_exist var(req.auth_response_header.remote_groups) -m found
    http-request set-header Remote-User %[var(req.auth_response_header.remote_user)] if remote_user_exist
    http-request set-header Remote-Groups %[var(req.auth_response_header.remote_groups)] if remote_groups_exist

    server nextcloud nextcloud:443 ssl verify none
```

##### haproxy.cfg (TLS enabled Authelia)
```
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
    acl protected-frontends hdr(host) -m reg -i ^(nextcloud)\.example\.com
    acl host-authelia hdr(host) -i auth.example.com
    acl host-nextcloud hdr(host) -i nextcloud.example.com

    http-request set-var(req.scheme) str(https) if { ssl_fc }
    http-request set-var(req.scheme) str(http) if !{ ssl_fc }
    http-request set-var(req.questionmark) str(?) if { query -m found }
   
    # Headers to construct redirection URL
    http-request set-header X-Real-IP %[src]
    http-request set-header X-Forwarded-Proto %[var(req.scheme)]
    http-request set-header X-Forwarded-Host %[req.hdr(Host)]
    http-request add-header X-Forwarded-Port %[dst_port]
    http-request set-header X-Forwarded-Uri %[path]%[var(req.questionmark)]%[query]

    # Protect endpoints with haproxy-auth-request and Authelia
    http-request lua.auth-request be_authelia_proxy /api/verify if protected-frontends
   
    # Authelia backend route
    use_backend be_authelia if host-authelia
    # Redirect protected-frontends to Authelia if not authenticated
    use_backend be_authelia if protected-frontends !{ var(txn.auth_response_successful) -m bool }
    # Service backend route(s)
    use_backend be_nextcloud if host-nextcloud

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
    # Pass Remote-User and Remote-Groups headers   
    acl remote_user_exist var(req.auth_response_header.remote_user) -m found
    acl remote_groups_exist var(req.auth_response_header.remote_groups) -m found
    http-request set-header Remote-User %[var(req.auth_response_header.remote_user)] if remote_user_exist
    http-request set-header Remote-Groups %[var(req.auth_response_header.remote_groups)] if remote_groups_exist

    server nextcloud nextcloud:443 ssl verify none
```

[HAproxy]: https://www.haproxy.org/
