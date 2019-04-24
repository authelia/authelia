# Nginx

[nginx] is the only official reverse proxy supported by **Authelia** for now.

## Configuration

Here is a commented example of configuration

    server {
        listen 443 ssl;
        server_name     myapp.example.com;

        resolver 127.0.0.11 ipv6=off;
        set $upstream_verify https://authelia.example.com/api/verify;
        set $upstream_endpoint http://nginx-backend;

        ssl_certificate     /etc/ssl/server.cert;
        ssl_certificate_key /etc/ssl/server.key;

        # Use HSTS, please beware of what you're doing if you set it.
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
        add_header X-Frame-Options "SAMEORIGIN";

        location / {
            # Send a subsequent request to Authelia to verify if the user is authenticated
            # and has the right permissions to access the resource.
            auth_request /auth_verify;

            # Set the X-Forwarded-User and X-Forwarded-Groups with the headers
            # returned by Authelia for the backends which can consume them.
            # This is not safe, as the backend must make sure that they come from the
            # proxy. In the future, it's gonna be safe to just use OAuth.
            auth_request_set            $user $upstream_http_remote_user;
            proxy_set_header            X-Forwarded-User $user;

            auth_request_set            $groups $upstream_http_remote_groups;
            proxy_set_header            X-Forwarded-Groups $groups;

            # Set the `target_url` variable based on the request. It will be used to build the portal
            # URL with the correct redirection parameter.
            auth_request_set            $target_url $scheme://$http_host$request_uri;
                        
            # If Authelia returns 401, then nginx redirects the user to the login portal.
            # If it returns 200, then the request pass through to the backend.
            # For other type of errors, nginx will handle them as usual.
            # NOTE: do not forget to include /#/ representing the hash router of the web application.
            error_page                  401 =302 https://login.example.com:8080/#/?rd=$target_url;

            proxy_pass                  $upstream_endpoint;
        }

        # Virtual endpoint created by nginx to forward auth requests.
        location /auth_verify {
            internal;

            # [OPTIONAL] The IP of the client shown in Authelia logs.
            proxy_set_header            X-Real-IP $remote_addr;

            # [REQUIRED] Needed by Authelia to check authorizations of the resource.
            # Provide either X-Original-URL and X-Forwarded-Proto or
            # X-Forwarded-Proto, X-Forwarded-Host and X-Forwarded-Uri or both.
            # Those headers will be used by Authelia to deduce the target url of the user.
            #
            # X-Forwarded-Proto is mandatory since Authelia uses the "trust proxy" option.
            # See https://expressjs.com/en/guide/behind-proxies.html
            proxy_set_header            X-Original-URL $scheme://$http_host$request_uri;
            
            proxy_set_header            X-Forwarded-Proto $scheme;
            proxy_set_header            X-Forwarded-Host $http_host;
            proxy_set_header            X-Forwarded-Uri $request_uri;
                        
            # [OPTIONAL] The list of IPs of client and proxies in the chain.
            proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;

            proxy_pass_request_body     off;
            proxy_set_header            Content-Length "";

            proxy_pass                  $upstream_verify;
        }
    }


[nginx]: https://www.nginx.com/