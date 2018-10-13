The following deployment process has been used by a member of the community but it is not tested in this repository. Please ask your questions to the community using Slack or Gitter before creating an issue.

## Requirements
- Your system runs a recent Debian-based GNU/Linux distribution – *i.e.* it makes use of *systemd* (System V is getting old).
- Basics are installed – Wget, NodeJS, NPM, Fail2Ban, Nginx, etc.
- You have an LDAP and Redis server.
- Authelia will be configured to use MongoDB for storage.

## Deployment

The following commands are to be typed as `root` user.

1. Install Authelia via NPM.
```bash
npm i -g authelia
```

2. Add a user `authelia` (service account) to be used by Authelia.
```bash
useradd -r -s /bin/false authelia
```

3. Configure Authelia.
```bash
mkdir -p /etc/authelia
chown root:authelia /etc/authelia
chmod 2750 /etc/authelia
wget -O /etc/authelia/config.yml 'https://raw.githubusercontent.com/clems4ever/authelia/master/config.template.yml'
$EDITOR /etc/authelia/config.yml
```

Then, make sure to adjust this configuration file to fit your setup.

4. Create a *systemd* unit file to properly start Authelia.
```bash
$EDITOR /etc/systemd/system/authelia.service
```

Then, type the following conent.
```systemd
[Unit]
Description=2FA Single Sign-On Authentication Server
Requires=mongod.service redis.service
After=network.target

[Service]
User=authelia
Group=authelia
ExecStart=/usr/bin/authelia /etc/authelia/config.yml
Restart=always

[Install]
WantedBy=multi-user.target
```
**Note**: Redis and MongoDB instances that are used by Authelia are assumed to be on the same server as the latter. Remove or adjust the line starting with `Requires=`, otherwise.

5. Start Authelia and enable it on system startup.
```bash
systemctl daemon-reload
systemctl start authelia
systemctl status authelia
systemctl enable authelia
```

6. Protect Authelia from brute-force attempts with Fail2Ban.
```bash
$EDITOR /etc/fail2ban/filter.d/authelia.conf
```

Then, you can type the following filter, for instance. **Make sure to adapt this filter with future releases of Authelia, in case logging messages eventually change.**

```fail2ban
# Fail2Ban filter for Authelia

# Make sure that the HTTP header "X-Forwarded-For" received by Authelia's backend
# only contains a single IP address (the one from the end-user), and not the proxy chain
# (it is misleading: usually, this is the purpose of this header).

[INCLUDES]

before = common.conf

[Definition]

_daemon = authelia

failregex = ^%(__prefix_line)serror: date='.*?' method='POST', path='.*?' requestId='.*?' sessionId='.*?' ip='<HOST>' message='Reply with error 200: Special character used in LDAP query\.'$
            ^%(__prefix_line)serror: date='.*?' method='(GET|POST)', path='.*?' requestId='.*?' sessionId='.*?' ip='<HOST>' message='Reply with error 200: No user DN found for user '.*''$
            ^%(__prefix_line)serror: date='.*?' method='POST', path='.*?' requestId='.*?' sessionId='.*?' ip='<HOST>' message='Reply with error 200: Invalid Credentials'$
            ^%(__prefix_line)serror: date='.*?' method='POST', path='.*?' requestId='.*?' sessionId='.*?' ip='<HOST>' message='Reply with error 200: Wrong TOTP token\.'$

ignoreregex =

[Init]

# "maxlines" is number of log lines to buffer for multi-line regex searches
maxlines = 1

journalmatch = _SYSTEMD_UNIT=authelia.service + _COMM=node
```

Now, enable Authelia's Fail2Ban jail.
```bash
$EDITOR /etc/fail2ban/jail.local
```

You may append the following content, for instance, or customize it – refer to [Fail2Ban manual](https://www.fail2ban.org/wiki/index.php/MANUAL_0_8#Jails).

```fail2ban
[authelia]
enabled = yes
backend = systemd
port = 80,443
findtime = 3600
maxretry = 3
bantime = 3600
```

7. Restart Fail2Ban and make sure everything is working as expected.
```bash
systemctl restart fail2ban
systemctl status fail2ban
fail2ban-client status authelia
```

8. Configure Nginx.

Below are a sample configurations of Nginx, suitable with Authelia, under the following assumptions:
- Authelia web page will be accessible from `https://login.example.com` and its backend is running at `http://127.0.0.1:4221` – Authelia's frontend and backend are served from the same machine.
- `private.example.com` is a domain meant to be protected by Authelia. In this example, `https://private.example.com` is proxified to `http://127.0.0.1:8000` – not necessarily served by the same machine than Authelia itself.
- You do have SSL (TLS) certificates for both of these (sub)domains.

First, let's configure the Nginx server serving Authelia.
```nginx
server {
    listen 80;
    server_name login.example.com;

    location / {
        include default_headers;
        return 301 https://$server_name$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name login.example.com;
    include ssl_login.example.com_params;

    access_log /var/log/nginx/login.example.com/access.log;
    error_log /var/log/nginx/login.example.com/error.log;

    # Authelia's API for requests coming from external Nginx servers

    location = /api/verify {
        include default_headers;
        # Set original requested Host from X-Forwared-Host.
        # Of course, the external Nginx server protecting his resources
        # must have properly set this header, as configured in this how-to.
        proxy_set_header Host $http_x_forwarded_host;
        proxy_pass http://127.0.0.1:4221;
    }

    # Authelia's frontend

    location /secondfactor/totp/identity/finish {
        # We don't want the user web browser to cache
        # TOTP secrets / QR codes, for security purposes.
        add_header Cache-Control "no-store";
        add_header Pragma "no-cache";
        include default_headers;
        include proxy_params;
        # Pass to Authelia's backend
        proxy_pass http://127.0.0.1:4221;
        proxy_intercept_errors on;
        if ($request_method !~ ^(POST)$){
            error_page 401 = /error/401;
            error_page 403 = /error/403;
            error_page 404 = /error/404;
        }
    }

    location / {
        include default_headers;
        include proxy_params;
        # Pass to Authelia's backend
        proxy_pass http://127.0.0.1:4221;
        proxy_intercept_errors on;
        if ($request_method !~ ^(POST)$){
            error_page 401 = /error/401;
            error_page 403 = /error/403;
            error_page 404 = /error/404;
        }
    }
}
```

Then, here is how to configure an Nginx server so that its resources are protected by Authelia.
Two variants are possible, depending on whether this server runs on the same machine as Authelia or not.

Below is the case of an external server. See inline comments to know how to deal with a server running on the same machine as Authelia.
```nginx
server {
    listen 80;
    server_name private.example.com;

    location / {
        include default_headers;
        return 301 https://$server_name$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name private.example.com;
    include ssl_private.example.com_params;

    access_log /var/log/nginx/private.example.com/access.log;
    error_log /var/log/nginx/private.example.com/error.log;

    # Authelia's API
    include authelia_check-auth_block_external_api;
    # Replace "external" with "internal" in the above line in case this server is
    # on the same machine as Authelia itself.

    location / {
        include default_headers;
        auth_request /.check-auth;
        include authelia_sso_params;
        include proxy_params;
        # Pass to the protected backend
        proxy_pass http://127.0.0.1:8000;
        proxy_redirect off;
    }
}
```

**Note:** For better legibility, I include snippets of codes – Nginx allows this. Therefore `include filename;` refers to the content of the file `filename` located under Nginx configuration directory, *i.e.* `/etc/nginx`, by default. I give below the content of the files used in previous Nginx configurations.

Content of `ssl_login.example.com_params` or `ssl_private.example.com_params`: adapt paths pointed by `ssl_certificate`, `ssl_certificate_key` and `ssl_trusted_certificate` directives:
```nginx
resolver 8.8.8.8 8.8.4.4 valid=300s;
resolver_timeout 10s;
ssl_certificate /path/to/fullchain.pem;
ssl_certificate_key /path/to/privkey.pem;
ssl_protocols TLSv1.2;
ssl_ciphers 'EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH';
ssl_dhparam /path/to/dhparams.pem;
ssl_ecdh_curve secp384r1;
ssl_prefer_server_ciphers on;
ssl_stapling on;
ssl_stapling_verify on;
ssl_trusted_certificate /path/to/chain.pem;
ssl_session_timeout 24h;
ssl_session_cache shared:SSL:50m;
ssl_session_tickets off;
```

Content of `default_headers`:
```nginx
# Make sure to understand the purpose of each of these HTTP headers.
# Some may be not relevant for your own setup.
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header X-Content-Type-Options nosniff;
add_header X-Frame-Options "SAMEORIGIN";
add_header X-XSS-Protection "1; mode=block";
add_header X-Robots-Tag "noindex, nofollow, nosnippet, noarchive";
add_header X-Download-Options noopen;
add_header X-Permitted-Cross-Domain-Policies none;
```

Content of `proxy_params`:
```nginx
proxy_set_header Host $host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_http_version 1.1;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
proxy_cache_bypass $http_upgrade;
```

Content of `authelia_check-auth_block_external_api`:
```nginx
    location = /.check-auth {
        # We want this location to be used only for internal Nginx requests.
        internal;

        # Authelia verifies ACLs with the two following headers:
        # Host and X-Original-URI.
        # We need to provide them.
        # First, give the original requested host name in X-Forwarded-Host.
        # The API endoint will set the Host header for Authelia's backend
        # based on the value of this header.
        # But... why not directly set the Host header?
        # Well, to make sure we the proxy_pass will actually reach the the API.
        # Indeed, the API endpoint is served by an Nginx server, which you will reach only
        # if the Host matches the server_name directive.
        proxy_set_header X-Forwarded-Host $host;
        # Then, give Authelia the original requested URI.
        proxy_set_header X-Original-URI $request_uri;

        # Authelia trust proxies in the way explained here:
        # http://expressjs.com/en/guide/behind-proxies.html
        # Therefore, we need to ensure that some headers are genuine.
        # This is important, to know whether session cookies shall or not be passed
        # (HTTPS is required to pass them) and so that Authelia logs genuine client IPs.
        # First, set the appropriate header (X-Forwarded-For) with the actual end-user IP
        # (not spoofable: this server set it to the IP of its client, i.e. the end-user).
        # Make sure that the HTTP header X-Forwarded-For that is received by Authelia's backend
        # only contains a single IP address (the one from the end-user), and not the proxy chain
        # (it is misleading: usually, this is the purpose of this header). This is why we set is to the value
        # of the Nginx $remote_addr variable.
        proxy_set_header X-Forwarded-For $remote_addr;
        # Then, let Authelia know whether the connection was made over HTTPS
        # (again, not spoofable).
        proxy_set_header X-Forwarded-Proto $scheme;

        # For Nginx auth_subrequest module
        # cf. https://www.nginx.com/resources/admin-guide/restricting-access-auth-request/
        proxy_set_header Content-Length "";
        proxy_pass_request_body off;

        # Make sure we connect to the backend server with strong SSL/TLS parameters.
        proxy_ssl_protocols TLSv1.2;
        proxy_ssl_ciphers 'EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH';

        # By default, Nginx does not check whether the backend server presents a trustworthy
        # certificate or not. To avoid a MitM (Man-in-the-Middle) attack, we want to make sure
        # the certificate is signed by a CA trusted by this machine.
        proxy_ssl_verify on;
        proxy_ssl_trusted_certificate "/etc/ssl/certs/DST_Root_CA_X3.pem";
        proxy_ssl_verify_depth 2;
        # WARNING: Adjust "proxy_ssl_trusted_certificate" and "proxy_ssl_verify_depth" to fit your
        # own PKI. This example should be valid for a certificate signed by Let's Encrypt.
        # This is not that bad from a security point of view, but it can definitely be improved: you should
        # use your own PKI.
        # To go further, here are a couple useful resources:
        # - Enforce mutual SSL/TLS authentication: https://www.nginx.com/resources/admin-guide/nginx-https-upstreams/
        # - Check the backend certificate against a CRL (Certificate Revocation List): http://nginx.org/en/docs/stream/ngx_stream_proxy_module.html#proxy_ssl_crl

        # Again, to make sure we actually reach the appropriate API,
        # we make sure to send the SNI (Server Name Indication) in the
        # SSL/TLS client Hello sent from this server.
        proxy_ssl_server_name on;

        # Make sur to adjust the line below so that Nginx queries appropriate DNS servers
        # to resolve "login.example.com".
        resolver 8.8.8.8 8.8.4.4 valid=300s ipv6=off;
        resolver_timeout 10s;
        proxy_pass https://login.example.com/api/verify;
    }
```

Content of `authelia_check-auth_block_internal_api` (slightly different from the previous file):
```nginx
location = /.check-auth {
    internal;
    proxy_set_header Host $host;
    proxy_set_header X-Original-URI $request_uri;
    proxy_set_header X-Forwarded-For $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header Content-Length "";
    proxy_pass_request_body off;
    proxy_pass http://127.0.0.1:4221/api/verify;
}
```

Content of `authelia_sso_params`:
```nginx
auth_request /.check-auth;

# Uncomment if needed: it gives the backend server protected by Authelia
# HTTP headers containing the user name and groups of the authenticated user.
# Useful for SSO, for instance.
#
# auth_request_set $user $upstream_http_remote_user;
# auth_request_set $groups $upstream_http_remote_groups;
# proxy_set_header X-Forwarded-User $user;
# proxy_set_header X-Forwarded-Groups $groups;

auth_request_set $redirect $upstream_http_redirect;
error_page 401 =302 https://login.example.com?redirect=$redirect;
error_page 403 = https://login.example.com/error/403;
```

9. Reload Nginx.
```bash
systemctl reload nginx
systemctl status nginx
```