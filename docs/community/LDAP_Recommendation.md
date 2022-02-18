---
layout: default
title: LDAP Recommendation
parent: Community
nav_order: 7
has_children: true
has_toc: false
---
# LDAP Recommendation
## Not Maintained By Authelia
[https://github.com/nitnelave/lldap](https://github.com/nitnelave/lldap)
Nitnelave create a simple and elegant way to create light weight ldap

# Prerequisite
* Linux Machine
* Docker
* Docker-compose
* Authelia installed on the same machine

# Setup
1. Create docker file in the same project
```yaml
volumes:
  lldap_data:
    driver: local

services:
  lldap:
    image: nitnelave/lldap:stable
    # Change this to the user:group you want.
    user: "33:33"
    ports:
      # For LDAP
      - "3890:3890"
      # For the web front-end
      - "17170:17170"
    volumes:
      - "lldap_data:/data"
      # Alternatively, you can mount a local folder
      # - "./lldap_data:/data"
    environment:
      - LLDAP_JWT_SECRET=REPLACE_WITH_RANDOM
      - LLDAP_LDAP_USER_PASS=REPLACE_WITH_PASSWORD
      - LLDAP_LDAP_BASE_DN=dc=example,dc=com
```
2. Replace `REPLACE_WITH_RANDOM` with a random 64 letter/numbers
3. Replace `REPLACE_WITH_PASSWORD` with a password (please make it strong)
4. Replace `dc=example` with anything (you dont need to own the domain)
5. If you need to mount it anywhere replace `lldap_data:/data` with `$location/lldap_data:/data`
6. Add this to the data directory

```toml
## Default configuration for Docker.
## All the values can be overridden through environment variables, prefixed
## with "LLDAP_". For instance, "ldap_port" can be overridden with the
## "LLDAP_LDAP_PORT" variable.

## The port on which to have the LDAP server.
#ldap_port = 3890

## The port on which to have the HTTP server, for user login and
## administration.
#http_port = 17170

## The public URL of the server, for password reset links.
#http_url = "http://localhost"

## Random secret for JWT signature.
## This secret should be random, and should be shared with application
## servers that need to consume the JWTs.
## Changing this secret will invalidate all user sessions and require
## them to re-login.
## You should probably set it through the LLDAP_JWT_SECRET environment
## variable from a secret ".env" file.
## This can also be set from a file's contents by specifying the file path
## in the LLDAP_JWT_SECRET_FILE environment variable
## You can generate it with (on linux):
## LC_ALL=C tr -dc 'A-Za-z0-9!"#%&'\''()*+,-./:;<=>?@[\]^_{|}~' </dev/urandom | head -c 32; echo ''
#jwt_secret = "REPLACE_WITH_RANDOM"

## Base DN for LDAP.
## This is usually your domain name, and is used as a
## namespace for your users. The choice is arbitrary, but will be needed
## to configure the LDAP integration with other services.
## The sample value is for "example.com", but you can extend it with as
## many "dc" as you want, and you don't actually need to own the domain
## name.
#ldap_base_dn = "dc=example,dc=com"

## Admin username.
## For the LDAP interface, a value of "admin" here will create the LDAP
## user "cn=admin,ou=people,dc=example,dc=com" (with the base DN above).
## For the administration interface, this is the username.
#ldap_user_dn = "admin"

## Admin password.
## Password for the admin account, both for the LDAP bind and for the
## administration interface. It is only used when initially creating
## the admin user.
## It should be minimum 8 characters long.
## You can set it with the LLDAP_LDAP_USER_PASS environment variable.
## This can also be set from a file's contents by specifying the file path
## in the LLDAP_LDAP_USER_PASS_FILE environment variable
## Note: you can create another admin user for user administration, this
## is just the default one.
#ldap_user_pass = "REPLACE_WITH_PASSWORD"

## Database URL.
## This encodes the type of database (SQlite, Mysql and so
## on), the path, the user, password, and sometimes the mode (when
## relevant).
## Note: Currently, only SQlite is supported. SQlite should come with
## "?mode=rwc" to create the DB if not present.
## Example URLs:
##  - "postgres://postgres-user:password@postgres-server/my-database"
##  - "mysql://mysql-user:password@mysql-server/my-database"
##
## This can be overridden with the DATABASE_URL env variable.
database_url = "sqlite:///data/users.db?mode=rwc"

## Private key file.
## Contains the secret private key used to store the passwords safely.
## Note that even with a database dump and the private key, an attacker
## would still have to perform an (expensive) brute force attack to find
## each password.
## Randomly generated on first run if it doesn't exist.
key_file = "/data/private_key"

## Options to configure SMTP parameters, to send password reset emails.
## To set these options from environment variables, use the following format
## (example with "password"): LLDAP_SMTP_OPTIONS__PASSWORD
#[smtp_options]
## Whether to enabled password reset via email, from LLDAP.
#enable_password_reset=true
## The SMTP server.
#server="smtp.gmail.com"
## The SMTP port.
#port=587
## Whether to connect with TLS.
#tls_required=true
## The SMTP user, usually your email address.
#user="sender@gmail.com"
## The SMTP password.
#password="password"
## The header field, optional: how the sender appears in the email. The first
## is a free-form name, followed by an email between <>.
#from="LLDAP Admin <sender@gmail.com>"
## Same for reply-to, optional.
#reply_to="Do not reply <noreply@localhost>"

## Tune the logging to be more verbose by setting this to be true.
## You can set it with the LLDAP_VERBOSE environment variable.
# verbose=false
```

7. Run the file with `docker-compose up -d`
8. Add this to authelia `configuration.yml`

```yml
authentication_backend:
  # Password reset through authelia works normally.
  disable_reset_password: false
  # How often authelia should check if there is an user update in LDAP
  refresh_interval: 1m
  ldap:
    implementation: custom
    # Pattern is ldap://HOSTNAME-OR-IP:PORT
    # Normal ldap port is 389, standard in LLDAP is 3890
    url: ldap://lldap:3890
    # The dial timeout for LDAP.
    timeout: 5s
    # Use StartTLS with the LDAP connection, TLS not supported right now
    start_tls: false
    #tls:
    #  skip_verify: false
    #  minimum_version: TLS1.2
    # Set base dn, like dc=google,dc.com
    base_dn: dc=example,dc=com
    username_attribute: uid
    # You need to set this to ou=people, because all users are stored in this ou!
    additional_users_dn: ou=people
    # To allow sign in both with username and email, one can use a filter like
    # (&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))
    users_filter: (&({username_attribute}={input})(objectClass=person))
    # Set this to ou=groups, because all groups are stored in this ou
    additional_groups_dn: ou=groups
    # Only this filter is supported right now
    groups_filter: (member={dn})
    # The attribute holding the name of the group.
    group_name_attribute: cn
    # Email attribute
    mail_attribute: mail
    # The attribute holding the display name of the user. This will be used to greet an authenticated user.
    display_name_attribute: displayName
    # The username and password of the admin user.
    # "admin" should be the admin username you set in the LLDAP configuration
    user: cn=admin,ou=people,dc=example,dc=com
    # Password can also be set using a secret: https://www.authelia.com/docs/configuration/secrets.html
    password: 'REPLACE_ME'
```

9. Replace example with what ever domain you set for `- LLDAP_LDAP_BASE_DN=dc=example,dc=com`
10. Restart authelia
11. Access LLDAP user interface with `IP_Address:17170` in the web browser
12. Login with `admin` and password setup before

# Reminders
* You can set secrets for `- LLDAP_LDAP_USER_PASS=REPLACE_WITH_PASSWORD` in LLDAP by changing it to `- LLDAP_LDAP_USER_PASS_FILE=/run/secrets/lldap_admin_password`
* You can set secrets for `password: 'REPLACE_ME'` in authelia by specifing `AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE=/run/secrets/lldap_admin_password` in docker
