# Installing **Authelia** on Debian bare metal from scratch

This document describes how to build **Authelia** from scratch and how to
create a basic working configuration.

For the purpose of this document the following examples are used;

    This is the hostname people will access when they want to access the protected site;
    landing.example.com
    Hostname of the Authelia service;
    authelia.example.com
    Hostname of the protected site;
    protected.example.com
    $editor to be replaced by your favourite editor
    $user to be replaced by the username you want to create to login to landing.example.com

The following passwords are needed;

    slapd-admin-password=<slapd-admin-password>
    redis-password=<redis-password>
    mariadb-password=<mariadb-password>
    user-password=<user-password>
    jwtsecret=<jwt-secret>
    sessionsecret=<session-secret>
    
These packages are optional and only needed if you are planning to compile from source;

    * GO
    * Docker
    * Docker Compose
    * Node.JS
    and you can then skip the section 'Download, Build and Install Authelia'

**NOTE** [you can download the binary from https://github.com/clems4ever/authelia/releases/](https://github.com/clems4ever/authelia/releases/)

This document assumes you have a domain-wildcard SSL certificate in /etc/nginx/ssl/

    # cp -aL /etc/letsencrypt/live/example.com/fullchain.pem /etc/nginx/ssl/xxx.example.com.fullchain
    # cp -aL /etc/letsencrypt/live/example.com/privkey.pem /etc/nginx/ssl/xxx.example.com.key
    # cp -aL /etc/letsencrypt/live/example.com/cert.pem /etc/nginx/ssl/xxx.example.com.cert

**NOTE** First copy the /etc/letsencrypt/live/example.com directory from the location where you have letsencrypt running

**NOTE** All installation steps below are to be executed as root or as a normal user using 'sudo'

### Install Debian the latest version of Debian 10
    Debian Network install from a minimal CD: https://www.debian.org/CD/netinst/
    
### Uninstall apparmor and install your favourite editor if not already installed
    # apt-get remove apparmor
    # apt-get install $editor gcc g++ make slapd ldap-utils redis-server curl mariadb-server nginx git apt-transport-https ca-certificates gnupg2 software-properties-common ntpdate

**NOTE** You will be prompted to set the $slapd-admin-password

#### Install GO;
    # cd /usr/src/
    # curl -O https://dl.google.com/go/go1.13.5.linux-amd64.tar.gz
    # tar -C /usr/local -xzf go1.13.5.linux-amd64.tar.gz
    # mkdir ~/work

**NOTE** [You can check the latest version of GO on https://golang.org/dl/](https://golang.org/dl/)

### add GO variables to ~/.profile;
    # echo "export GOPATH=\$HOME/work" >> ~/.profile   
    # echo "export PATH=\$PATH:/usr/local/go/bin:\$GOPATH/bin" >> ~/.profile
    # source ~/.profile

### Configure the ldap server, create Organizational Units
    # cd ~
    # cat <<EOF > user_group_base.ldif
    dn: ou=people,dc=example,dc=com
    objectClass: organizationalUnit
    ou: people
    
    dn: ou=groups,dc=example,dc=com
    objectClass: organizationalUnit
    ou: groups
    EOF

    # ldapadd -x -D cn=admin,dc=example,dc=com -W -f user_group_base.ldif

**NOTE** You will be prompted to enter the $slapd-admin-password

### Create $user-password

    # user-password=$(slappasswd)

**NOTE** type password twice

### Create user in LDAP
    # user=<NewUserNameToCreate>
    # userfirst=<NewUserFirstName>
    # userlast=<NewUserLastName>
    # useremail=<NewUserEmailAddress>
    # useruid=<NewUserUID>
    # usergid=<NewUserGID>
    # cat <<EOF > new_user.ldif
    dn: uid=$user,ou=people,dc=example,dc=com
    objectClass: inetOrgPerson
    objectClass: posixAccount
    objectClass: shadowAccount
    uid: $user
    cn: $user
    givenName: $user
    sn: $userlast
    userPassword: $user-password
    loginShell: /bin/bash
    uidNumber: $useruid
    gidNumber: $usergid
    homeDirectory: /home/$user
    shadowMax: 60
    shadowMin: 1
    shadowWarning: 7
    shadowInactive: 7
    shadowLastChange: 0
    mail: $useremail
    
    dn: cn=$user,ou=groups,dc=example,dc=com
    objectClass: posixGroup
    cn: $user
    gidNumber: 0
    memberUid: $user
    EOF

    # ldapadd -x -D cn=admin,dc=example,dc=com -W -f new_user.ldif

### Example commands to check if LDAP is working and to modify/delete users
    # ldapsearch -x -LLL -b "dc=example,dc=com"
    # ldapdelete -x -W -D "cn=admin,dc=example,dc=com" "uid=$user,ou=people,dc=example,dc=com"
    # ldapdelete -x -W -D "cn=admin,dc=example,dc=com" "cn=$user,ou=groups,dc=example,dc=com"
    # ldappasswd -H ldap://127.0.0.1 -x -D "cn=admin,dc=example,dc=com" -W -S "uid=$user,ou=people,dc=example,dc=com"
    # ldapwhoami -vvv -h 127.0.0.1 -D "uid=$user,ou=people,dc=example,dc=com" -x -W

**NOTE** the last command can be used to test a $user password; if you see Result: Success (0) then the password matches.

### Configure redis
    # redis-password=$(openssl rand 60 | openssl base64 -A)
    # $editor /etc/redis/redis.conf
    	Change the following to lines;
    	supervised systemd
    	requirepass $redis-password
    # systemctl restart redis

### Testing redis
    # redis-cli
    # auth $redis-password
    # ping
    # quit

**NOTE** Redis should answer with PONG

### Configure mariadb
    # mariadb-password=<mariadb-password>
    # mysql
    > CREATE DATABASE `authelia`;
    > GRANT ALL privileges ON `authelia`.* TO 'authelia'@localhost IDENTIFIED BY '$mariadb-password';
    > exit

### Configure NGINX
    # unlink /etc/nginx/sites-enabled/default
    # cd /etc/nginx/sites-available
    # cat <<EOF > authelia.conf
    server {
            listen 80;
    
            location / {
                    include default_headers;
                    return 301 https://$server_name$request_uri;
            }
    }
    
    server {
            server_name authelia.example.com;
            listen 443 ssl http2;
            include ssl.conf;
    
            location / {
                    add_header X-Forwarded-Host authelia.example.com;
                    add_header X-Forwarded-Proto $scheme;
                    set $upstream_authelia http://authelia.example.com:9091;
                    proxy_pass $upstream_authelia;
                    include proxy.conf;
            }
    }
    
    server {
            server_name landing.example.com;
            listen 80;
            return 301 https://$server_name$request_uri;
    }
    
    server {
            server_name landing.example.com;
            listen 443 ssl http2;
            include ssl.conf;
            include authelia.conf;
    
            location / {
                    set $upstream_target https://protected.example.com;
                    proxy_pass $upstream_target;
                    include auth.conf;
                    include proxy.conf;
                    #proxy_ssl_certificate /etc/nginx/ssl/xxx.example.com.cert;
                    #proxy_ssl_certificate_key /etc/nginx/ssl/xxx.example.com.key;
            }
    }
    EOF

**NOTE** If protected.example.com needs a different SSL certificate than landing.example.com you can un-comment the two proxy_ssl_certificate lines and point them to the correct certificates. It is even possible to have protected.example.com on a completely different domain-name but that is not guaranteed to work without issues.

### Other NGINX files needed
    # cd /etc/nginx
    # cat <<EOF > auth.conf
    # Basic Authelia Config
    auth_request /authelia;
    auth_request_set $target_url $scheme://$http_host$request_uri;
    auth_request_set $user $upstream_http_remote_user;
    auth_request_set $groups $upstream_http_remote_groups;
    proxy_set_header X-Forwarded-User $user;
    proxy_set_header X-Forwarded-Groups $groups;
    error_page 401 =302 https://authelia.example.com/#/?rd=$target_url;
    
    EOF

    # cat <<EOF > authelia.conf
    location /authelia {
            internal;
            set $upstream_authelia http://authelia.example.com:9091/api/verify;
            proxy_pass_request_body off;
            proxy_pass $upstream_authelia;
            proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
            proxy_set_header Content-Length "";
    
            # Timeout if the real server is dead
            proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    
            # Basic Proxy Config
            client_body_buffer_size 128k;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $remote_addr;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header X-Forwarded-Host $http_host;
            proxy_set_header X-Forwarded-Uri $request_uri;
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
    EOF

    # cat <<EOF > default_headers
    # Make sure to understand the purpose of each of these HTTP headers.
    # Some may be not relevant for your own setup.
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-XSS-Protection "1; mode=block";
    add_header X-Robots-Tag "noindex, nofollow, nosnippet, noarchive";
    add_header X-Download-Options noopen;
    add_header X-Permitted-Cross-Domain-Policies none;
    add_header X-Forwarded-Proto https;
    add_header X-Forwarded-Host authelia.example.com;
    proxy_headers_hash_max_size 512;
    proxy_headers_hash_bucket_size 128;
    EOF

    # cat <<EOF > proxy.conf
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
    set_real_ip_from 172.0.0.0/8;
    set_real_ip_from 192.168.0.0/16;
    set_real_ip_from fc00::/7;
    real_ip_header X-Forwarded-For;
    real_ip_recursive on;
    EOF

    # cat <<EOF > proxy_params
    proxy_set_header Host $http_host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-host authelia.example.com;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_cache_bypass $http_upgrade;
    EOF

    # cat <<EOF > ssl.conf
    resolver 8.8.8.8 valid=300s;
    resolver_timeout 10s;
    ssl_certificate /etc/nginx/ssl/xxx.example.com.fullchain;
    ssl_certificate_key /etc/nginx/ssl/xxx.example.com.key;
    ssl_protocols TLSv1.2;
    ssl_ciphers 'EECDH+AESGCM:EDH+AESGCM:AES256+EECDH:AES256+EDH';
    ssl_ecdh_curve secp384r1;
    ssl_prefer_server_ciphers on;
    ssl_stapling on;
    ssl_stapling_verify on;
    ssl_session_timeout 24h;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;
    EOF

### (re)Start NGINX
    # ln sites-available/authelia.conf sites-enabled/
    # systemctl restart nginx

### Install Docker
    # curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add -
    # add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable"
    # apt-get update
    # apt-get install docker-ce docker-ce-cli containerd.io

### Install Docker Compose
    # curl -s https://api.github.com/repos/docker/compose/releases/latest | grep browser_download_url | grep docker-compose-Linux-x86_64 | cut -d '"' -f 4 | wget -qi -
    # chmod +x docker-compose-Linux-x86_64
    # mv docker-compose-Linux-x86_64 /usr/local/bin/docker-compose

### Install Node.JS
    # cd /usr/src
    # curl -sL https://deb.nodesource.com/setup_12.x | bash -
    # apt-get install nodejs

### Download, Build and Install Authelia
    # cd /opt
    # git clone https://github.com/clems4ever/authelia.git
    # cd authelia
    # source bootstrap.sh
    # authelia-scripts build

### Setup Authelia
    # mkdir /etc/authelia/
    # cat <<EOF > /etc/authelia/config.yml
    ###############################################################
    #                   Authelia configuration                    #
    ###############################################################
    
    port: 9091
    
    logs_level: info
    jwt_secret: $jwtsecret
    default_redirection_url: https://landing.example.com/
    
    totp:
      issuer: landing.example.com
    
    authentication_backend:
      ldap:
        url: ldap://127.0.0.1
        base_dn: dc=example,dc=com
        additional_users_dn: ou=people
        users_filter: uid={0}
        additional_groups_dn: ou=groups
        groups_filter: (&(member={dn})(objectclass=groupOfNames))
        group_name_attribute: cn
        mail_attribute: mail
        user: cn=admin,dc=example,dc=com
        password: $slapd-admin-password
    
    access_control:
      default_policy: two_factor
    
    session:
      name: authelia_session
      secret: $sessionsecret
      expiration: 3600 # 1 hour
      inactivity: 300 # 5 minutes
      domain: example.com
      redis:
        host: 127.0.0.1
        port: 6379
        password: $redis-password
    
    regulation:
      max_retries: 3
      find_time: 120
      ban_time: 300
    
    storage:
      mysql:
        host: 127.0.0.1
        port: 3306
        database: authelia
        username: authelia
        password: $mysql-password
    
    notifier:
      smtp:
        host: mail.example.com
        port: 25
        sender: authelia@example.com
    EOF

    # cat <<EOF > /etc/systemd/system/authelia.service
    [Unit]
    Description=2FA Single Sign-On Authentication Server
    Requires=mariadb.service redis-server.service
    After=network.target
    
    [Service]
    User=authelia
    Group=authelia
    ExecStart=/usr/local/bin/startauthelia.sh
    Restart=always
    
    [Install]
    WantedBy=multi-user.target
    EOF

    # cat <<EOF > /usr/local/bin/startauthelia.sh
    #!/bin/sh
    
    cd /opt/authelia/dist
    ./authelia -config /etc/authelia/config.yml
    EOF

    # adduser authelia
    # chown root.authelia /usr/local/bin/startauthelia.sh
    # chmod 750 /usr/local/bin/startauthelia.sh
    # chown -R authelia.authelia /etc/authelia/
    # chown -R authelia.authelia /opt/authelia
    # systemctl start authelia
    # systemctl status authelia
    # systemctl enable authelia
