---
layout: default
title: NGINX
parent: Proxy Integration
grand_parent: Deployment
nav_order: 2
---

# NGINX

[NGINX] is a reverse proxy supported by **Authelia**.

## Configuration

Below you will find commented examples of the following configuration:

* Authelia portal
* Protected endpoint (Nextcloud)
* Supplementary config

With the below configuration you can add `authelia.conf` to virtual hosts to support protection with Authelia.
`auth.conf` is utilised to enable the protection either at the root location or a more specific location/route.
`proxy.conf` is included just for completeness.

#### Supplementary config

##### authelia.conf

```nginx
set $upstream_authelia http://authelia:9091/api/verify;

# Virtual endpoint created by nginx to forward auth requests.
location /authelia {
    internal;
    proxy_pass_request_body off;
    proxy_pass $upstream_authelia;
    proxy_set_header Content-Length "";

    # Timeout if the real server is dead
    proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;

    # [REQUIRED] Needed by Authelia to check authorizations of the resource.
    # Provide either X-Original-URL and X-Forwarded-Proto or
    # X-Forwarded-Proto, X-Forwarded-Host and X-Forwarded-Uri or both.
    # Those headers will be used by Authelia to deduce the target url of the user.
    # Basic Proxy Config
    client_body_buffer_size 128k;
    proxy_set_header Host $host;
    proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-Method $request_method;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $http_host;
    proxy_set_header X-Forwarded-Uri $request_uri;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Ssl on;
    proxy_redirect  http://  $scheme://;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_cache_bypass $cookie_session;
    proxy_no_cache $cookie_session;
    proxy_buffers 4 32k;

    # Advanced Proxy Config
    send_timeout 5m;
    proxy_read_timeout 240;
    proxy_send_timeout 240;
    proxy_connect_timeout 240;
}
```

##### auth.conf

```nginx
# Basic Authelia Config
# Send a subsequent request to Authelia to verify if the user is authenticated
# and has the right permissions to access the resource.
auth_request /authelia;
# Set the `target_url` variable based on the request. It will be used to build the portal
# URL with the correct redirection parameter.
auth_request_set $target_url $scheme://$http_host$request_uri;
# Set the X-Forwarded-User and X-Forwarded-Groups with the headers
# returned by Authelia for the backends which can consume them.
# This is not safe, as the backend must make sure that they come from the
# proxy. In the future, it's gonna be safe to just use OAuth.
auth_request_set $user $upstream_http_remote_user;
auth_request_set $groups $upstream_http_remote_groups;
auth_request_set $name $upstream_http_remote_name;
auth_request_set $email $upstream_http_remote_email;
proxy_set_header Remote-User $user;
proxy_set_header Remote-Groups $groups;
proxy_set_header Remote-Name $name;
proxy_set_header Remote-Email $email;
# If Authelia returns 401, then nginx redirects the user to the login portal.
# If it returns 200, then the request pass through to the backend.
# For other type of errors, nginx will handle them as usual.
error_page 401 =302 https://auth.example.com/?rd=$target_url;
```

##### proxy.conf

```nginx
client_body_buffer_size 128k;

#Timeout if the real server is dead
proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;

# Advanced Proxy Config
send_timeout 5m;
proxy_read_timeout 360;
proxy_send_timeout 360;
proxy_connect_timeout 360;

# Basic Proxy Config
proxy_set_header Host $host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header X-Forwarded-Host $http_host;
proxy_set_header X-Forwarded-Uri $request_uri;
proxy_set_header X-Forwarded-Ssl on;
proxy_redirect  http://  $scheme://;
proxy_http_version 1.1;
proxy_set_header Connection "";
proxy_cache_bypass $cookie_session;
proxy_no_cache $cookie_session;
proxy_buffers 64 256k;

# If behind reverse proxy, forwards the correct IP
set_real_ip_from 10.0.0.0/8;
set_real_ip_from 172.16.0.0/12;
set_real_ip_from 192.168.0.0/16;
set_real_ip_from fc00::/7;
real_ip_header X-Forwarded-For;
real_ip_recursive on;
```

#### Authelia Portal

```nginx
server {
    server_name auth.example.com;
    listen 80;
    return 301 https://$server_name$request_uri;
}

server {
    server_name auth.example.com;
    listen 443 ssl http2;
    include /config/nginx/ssl.conf;

    location / {
        set $upstream_authelia http://authelia:9091; # This example assumes a Docker deployment
        proxy_pass $upstream_authelia;
        include /config/nginx/proxy.conf;
    }
}
```

#### Protected Endpoint

```nginx
server {
    server_name nextcloud.example.com;
    listen 80;
    return 301 https://$server_name$request_uri;
}

server {
    server_name nextcloud.example.com;
    listen 443 ssl http2;
    include /config/nginx/ssl.conf;
    include /config/nginx/authelia.conf; # Virtual endpoint to forward auth requests

    location / {
        set $upstream_nextcloud https://nextcloud;
        proxy_pass $upstream_nextcloud;
        include /config/nginx/auth.conf; # Activates Authelia for specified route/location, please ensure you have setup the domain in your configuration.yml
        include /config/nginx/proxy.conf; # Reverse proxy configuration
    }
}
```

### Basic Auth Example

Here's an example for using HTTP basic auth on a specific endpoint. It is based on the full example above.

##### authelia-basic.conf

```nginx
# Notice we added the auth=basic query arg here
set $upstream_authelia http://authelia:9091/api/verify?auth=basic;

location /authelia {
    internal;
    proxy_pass_request_body off;
    proxy_pass $upstream_authelia;
    proxy_set_header Content-Length "";

    # Timeout if the real server is dead
    proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;

    # [REQUIRED] Needed by Authelia to check authorizations of the resource.
    # Provide either X-Original-URL and X-Forwarded-Proto or
    # X-Forwarded-Proto, X-Forwarded-Host and X-Forwarded-Uri or both.
    # Those headers will be used by Authelia to deduce the target url of the user.
    # Basic Proxy Config
    client_body_buffer_size 128k;
    proxy_set_header Host $host;
    proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-Method $request_method;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Host $http_host;
    proxy_set_header X-Forwarded-Uri $request_uri;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Ssl on;
    proxy_redirect  http://  $scheme://;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_cache_bypass $cookie_session;
    proxy_no_cache $cookie_session;
    proxy_buffers 4 32k;

    # Advanced Proxy Config
    send_timeout 5m;
    proxy_read_timeout 240;
    proxy_send_timeout 240;
    proxy_connect_timeout 240;
}
```

##### auth-basic.conf

Same as `auth.conf` but without the `error_page` directive. We want nginx to proxy the 401 back to the client, not to return a 301.

```nginx
# Basic Authelia Config
# Send a subsequent request to Authelia to verify if the user is authenticated
# and has the right permissions to access the resource.
auth_request /authelia;
# Set the `target_url` variable based on the request. It will be used to build the portal
# URL with the correct redirection parameter.
auth_request_set $target_url $scheme://$http_host$request_uri;
# Set the X-Forwarded-User and X-Forwarded-Groups with the headers
# returned by Authelia for the backends which can consume them.
# This is not safe, as the backend must make sure that they come from the
# proxy. In the future, it's gonna be safe to just use OAuth.
auth_request_set $user $upstream_http_remote_user;
auth_request_set $groups $upstream_http_remote_groups;
auth_request_set $name $upstream_http_remote_name;
auth_request_set $email $upstream_http_remote_email;
proxy_set_header Remote-User $user;
proxy_set_header Remote-Groups $groups;
proxy_set_header Remote-Name $name;
proxy_set_header Remote-Email $email;
# If Authelia returns 401, then nginx passes it to the user.
# If it returns 200, then the request pass through to the backend.
```

#### Protected Endpoint

```nginx
server {
    server_name nextcloud.example.com;
    listen 80;
    return 301 https://$server_name$request_uri;
}

server {
    server_name nextcloud.example.com;
    listen 443 ssl http2;
    include /config/nginx/ssl.conf;
    include /config/nginx/authelia-basic.conf; # Use the "basic" endpoint

    location / {
        set $upstream_nextcloud https://nextcloud;
        proxy_pass $upstream_nextcloud;
        include /config/nginx/auth-basic.conf; # Activate authelia with basic auth
        include /config/nginx/proxy.conf; # this file is the exact same as above
    }
}
```


### Basic auth for specific client

If you'd like to force basic auth for some requests, you can use the following template:

##### authelia-detect.conf

```nginx
set $is_basic_auth ""; # false value
set $upstream_authelia http://authelia:9091/api/verify;

# Detect the client you want to force basic auth for here
# For the example we just match a path on the original request
if ($request_uri = "/force-basic") {
    set $is_basic_auth "true";
    set $upstream_authelia "$upstream_authelia?auth=basic";
}

location = /authelia {
    # Same as above
}

# A new virtual endpoint to used if the auth_request failed
location = /authelia-redirect {
    internal;

    if ($is_basic_auth) {
        # This is a request where we decided to use basic auth, return a 401.
        # Nginx will also proxy back the WWW-Authenticate header from Authelia's
        # response. This is what informs the client we're expecting basic auth.
        return 401;
    }

    # The original request didn't target /force-basic, redirect to the pretty login page
    # This is what `error_page 401 =302 https://auth.example.com/?rd=$target_url;` did.
    return 302 https://auth.example.com/$is_args$args;
}
```

##### auth.conf

Here we replace `error_page` directive to determine if basic auth should be utilised or not.

```nginx
# Basic Authelia Config
# Send a subsequent request to Authelia to verify if the user is authenticated
# and has the right permissions to access the resource.
auth_request /authelia;
# Set the `target_url` variable based on the request. It will be used to build the portal
# URL with the correct redirection parameter.
auth_request_set $target_url $scheme://$http_host$request_uri;
# Set the X-Forwarded-User and X-Forwarded-Groups with the headers
# returned by Authelia for the backends which can consume them.
# This is not safe, as the backend must make sure that they come from the
# proxy. In the future, it's gonna be safe to just use OAuth.
auth_request_set $user $upstream_http_remote_user;
auth_request_set $groups $upstream_http_remote_groups;
auth_request_set $name $upstream_http_remote_name;
auth_request_set $email $upstream_http_remote_email;
proxy_set_header Remote-User $user;
proxy_set_header Remote-Groups $groups;
proxy_set_header Remote-Name $name;
proxy_set_header Remote-Email $email;
# If Authelia returns 401, then nginx passes it to the user.
# If it returns 200, then the request pass through to the backend.
error_page 401 /authelia-redirect?rd=$target_url;
```

This tells nginx to use the virtual endpoint we defined above in case the auth_request failed.

[NGINX]: https://www.nginx.com/
