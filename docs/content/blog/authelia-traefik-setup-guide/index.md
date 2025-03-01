---
title: "Authelia + Traefik Setup Guide"
description: "A temporary guide for setting up Authelia with Traefik while the official Getting Started documentation is being improved."
summary: "In this guide we will walk through setting up Authelia with Traefik as the reverse proxy. This guide aims to provide an opinionated way to setup Authelia that is fully supported by the Authelia team."
date: 2025-01-03T13:31:09+11:00
draft: false
weight: 50
categories: ["Guides"]
tags: ["guides"]
contributors: ["Brynn Crowley"]
pinned: false
homepage: false
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---
{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
This guide is a temporary solution while we work to improve our "Getting Started" section of the website. It is likely this guide will **not** be updated for future versions. At such a time, a deprecation notice will be posted.

This is not a demo. If you would like an all-in-one demo, please take a look at our [local bundle](https://www.authelia.com/integration/deployment/docker/#local).
{{< /callout >}}

## Assumptions and Adaptation

This guide makes a few assumptions. These assumptions may require adaptation in more advanced and complex scenarios. We can not reasonably have examples for every advanced configuration option that exists. Some of these values can be automatically replaced with documentation variables.

{{< sitevar-preferences >}}

We make the following assumptions:
- [Docker](https://docs.docker.com/engine/install/) is configured correctly.
- Single Host
- You will have to adapt all instances of `{{< sitevar name="host" nojs="authelia" >}}` in the URL if:
  - you're using a different container name
  - you deployed the proxy to a different location
- You will have to adapt all instances of `{{< sitevar name="port" nojs="9091" >}}` in the URL if:
  - you have adjusted the default port in the configuration
- You will have to adapt the entire URL if:
  - Authelia is on a different host to the proxy
- All services are part of the `{{< sitevar name="domain" nojs="example.com" >}}` domain:
- This domain and the subdomains will have to be adapted in all examples to match your specific domains unless you're
  just testing, or you want to use that specific domain

## File Structure

The first thing we want to do is set up the file structure. Which should look something like this:
```text
📁 project
 ┣ 📁 authelia
 ┃  ┣ 📁 config
 ┃  ┃ ┣ 📄 configuration.yml
 ┃  ┃ ┗ 📄 users.yml
 ┃  ┗ 📁 secrets
 ┣ 📄 compose.yml
 ┗ 📁 traefik
    ┣ 📁 config
    ┃ ┣ 📄 dynamic.yml
    ┃ ┗ 📄 traefik.yml
    ┣ 📁 data
    ┃ ┗ 📄 acme.json
    ┣ 📁 logs
    ┗ 📁 secrets
```

## Traefik and Whoami

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
We'll focus on the minimal configuration needed to work with Authelia. For advanced Traefik features and configurations, consult their documentation.
{{< /callout >}}

Next, we'll set up Traefik as our reverse proxy. For detailed Traefik documentation, refer to the [official Traefik docs](https://doc.traefik.io/traefik/).

#### Docker Compose

```yaml{title="compose.yml"}
services:
  traefik:
    image: 'traefik:latest'
    container_name: 'traefik'
    restart: 'unless-stopped'
    security_opt:
      - 'no-new-privileges=true'
    networks:
      proxy:
        aliases:
          - '{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
      authelia: {}
    ports:
      - '80:80'
      - '443:443'
    environment:
      TZ: 'America/Los_Angeles' ## see below
    volumes:
      - '/var/run/docker.sock:/var/run/docker.sock:ro'
      - './traefik/config/traefik.yml:/traefik.yml:ro'
      - './traefik/config/dynamic.yml:/dynamic.yml:ro'
      - './traefik/data/:/data'
      - './traefik/logs:/logs'
    labels:
      traefik.enable: 'true'
      traefik.http.routers.dashboard.rule: 'Host(`traefik.{{< sitevar name="domain" nojs="example.com" >}}`)'
      traefik.http.routers.dashboard.entrypoints: 'https'
      traefik.http.routers.dashboard.middlewares: 'authelia@docker'
      traefik.http.routers.dashboard.service: 'api@internal'

  whoami:
    image: 'traefik/whoami'
    restart: 'unless-stopped'
    container_name: 'whoami'
    labels:
      traefik.enable: 'true'
      traefik.http.routers.whoami.rule: 'Host(`whoami.{{< sitevar name="domain" nojs="example.com" >}}`)'
      traefik.http.routers.whoami.entrypoints: 'https'
    networks:
      proxy: {}

  ## Other Services Go Here

networks:
  proxy:
    external: true
    name: 'proxy'
  authelia:
    name: 'authelia'
```

Note: Timezone strings can be found [here](https://go.dev/src/time/zoneinfo_abbrs_windows.go).



#### Basic Traefik Configuration

Now we configure Traefik.
The following files contain the minimal Traefik configuration needed for Authelia integration:

```yaml{title="traefik/config/traefik.yml"}
## Base Traefik configuration
api:
  dashboard: true
  debug: false
  insecure: false

log:
  level: 'INFO'
accessLog:
  filePath: '/logs/access.log'

entryPoints:
  http:
    address: ':80'
    http:
      redirections:
       entryPoint:
         to: 'https'
         scheme: 'https'
         permanent: true
  https:
    address: ':443'
    http:
      tls:
        certResolver: 'myresolver'

providers:
  docker:
    endpoint: 'unix:///var/run/docker.sock'
    exposedByDefault: false
  file:
    filename: '/dynamic.yml'

certificatesResolvers:
  myresolver:
    acme:
      storage: '/data/acme.json'
      httpChallenge:
        entryPoint: 'http'

tls:
  options:
    default:
      minVersion: 'VersionTLS12'
      cipherSuites:
        - 'TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256'
        - 'TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256'
        - 'TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384'
        - 'TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384'
        - 'TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305'
        - 'TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305'
```

### Domain Configuration

```yaml{title="traefik/config/dynamic.yml"}
## This file can be used to define dynamic routers/services/middlewares.
```

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
These are minimal configurations focused on Authelia integration. Adjust them according to your needs using Traefik's documentation.
{{< /callout >}}

## Authelia Compose

This configuration sets up Authelia's core service and configures forward authentication with Traefik. The portal will be available at `{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`. It also defines a new whoami container that will be protected by authelia.
The docker compose services defined below should be added to the existing compose.yml created for traefik and whoami.

```yaml{title="compose.yml"}
  authelia:
    image: 'authelia/authelia:4.38'
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    volumes:
      - './authelia/secrets:/secrets:ro'
      - './authelia/config:/config'
      - './authelia/logs:/var/log/authelia/'
    networks:
      authelia: {}
    labels:
      ## Expose Authelia through Traefik
      traefik.enable: 'true'
      traefik.docker.network: 'authelia'
      traefik.http.routers.authelia.rule: 'Host(`{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`)'
      traefik.http.routers.authelia.entrypoints: 'https'
      ## Setup Authelia ForwardAuth Middlewares
      traefik.http.middlewares.authelia.forwardAuth.address: '{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/forward-auth'
      traefik.http.middlewares.authelia.forwardAuth.trustForwardHeader: 'true'
      traefik.http.middlewares.authelia.forwardAuth.authResponseHeaders: 'Remote-User,Remote-Groups,Remote-Name,Remote-Email'
    environment:
      TZ: 'America/Los_Angeles'
      X_AUTHELIA_CONFIG_FILTERS: 'template'

  whoami-secure:
    image: 'traefik/whoami'
    restart: 'unless-stopped'
    container_name: 'whoami-secure'
    labels:
      traefik.enable: 'true'
      traefik.http.routers.whoami-secure.rule: 'Host(`whoami-secure.{{< sitevar name="domain" nojs="example.com" >}}`)'
      traefik.http.routers.whoami-secure.entrypoints: 'https'
      traefik.http.routers.whoami-secure.middlewares: 'authelia@docker'
    networks:
      proxy: {}
```

### Docker Networks

There are a couple docker networks that need to be created.

#### proxy

The `proxy` network contains Traefik and can be used to connect any additional containers to the Traefik proxy.
It is created by running the following command:

```shell
docker network create proxy \
  --opt "com.docker.network.bridge.name"="br-docker-proxy"
```

#### authelia

The `authelia` network contains the containers required for Authelia to function and connects Authelia to Traefik over a separate network.

While not included in this guide, it would include the storage provider (PostgresSQL or MySQL), session provider (Redis), and LDAP authentication backend. This network does not need to be created since it will automatically be created when the containers are started.
**Note**: While the `whoami-secure` container is protected by the Authelia middleware, it is not in the `authelia` docker network. This is because we want to avoid any risk of http traffic being intercepted. Protected services should either be in the `proxy` network or a network shared with Traefik, while Authelia-specific services use the separate `authelia` network for enhanced security isolation.

#### Authelia Configuration

```yaml{title="authelia/config/configuration.yml"}
server:
  address: 'tcp4://:{{< sitevar name="port" nojs="9091" >}}'

log:
  level: debug
  file_path: '/var/log/authelia/authelia.log'
  keep_stdout: true

identity_validation:
  elevated_session:
    require_second_factor: true
  reset_password:
    jwt_lifespan: '5 minutes'
    jwt_secret: {{ secret "/secrets/jwt_secret.txt" | mindent 0 "|" | msquote }}

totp:
  disable: false
  issuer: '{{< sitevar name="domain" nojs="example.com" >}}'
  period: 30
  skew: 1

password_policy:
  zxcvbn:
    enabled: true
    min_score: 4

authentication_backend:
  file:
    path: '/config/users.yml'
    password:
      algorithm: 'argon2'
      argon2:
        variant: 'argon2id'
        iterations: 3
        memory: 65535
        parallelism: 4
        key_length: 32
        salt_length: 16

access_control:
  default_policy: 'deny'
  rules:
    - domain: 'traefik.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'one_factor'
    - domain: 'whoami-secure.{{< sitevar name="domain" nojs="example.com" >}}'
      policy: 'two_factor'

session:
  name: 'authelia_session'
  secret: {{ secret "/secrets/session_secret.txt" | mindent 0 "|" | msquote }}
  cookies:
    - domain: '{{< sitevar name="domain" nojs="example.com" >}}'
      authelia_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'

regulation:
  max_retries: 4
  find_time: 120
  ban_time: 300

storage:
  encryption_key: {{ secret "/secrets/storage_encryption_key.txt" | mindent 0 "|" | msquote }}
  local:
    path: '/config/db.sqlite3'

notifier:
  disable_startup_check: false
  filesystem:
    filename: '/config/notification.txt'
```

Each section in the configuration file above has detailed documentation available. Below are direct links.
**Note**: There are config options that are not a part of this guide.

###### Core Configuration

* [Server Configuration](https://www.authelia.com/configuration/miscellaneous/server/) - Configure the server address, ports, TLS settings, and other core server options
* [Logging](https://www.authelia.com/configuration/miscellaneous/logging/) - Configure log levels, output locations, and format options
* [Identity Validation](https://www.authelia.com/configuration/identity-validation/introduction/) - Configure settings for password reset and elevated sessions.

###### Authentication & Security

* [TOTP Configuration](https://www.authelia.com/configuration/second-factor/time-based-one-time-password/) - Configure Time-based One-Time Password (TOTP) settings for two-factor authentication.
* [Password Policy](https://www.authelia.com/configuration/security/password-policy/) - Configure password strength requirements and validation rules
* [Authentication Backend](https://www.authelia.com/configuration/first-factor/introduction/) - Configure the authentication provider and settings
* [Access Control](https://www.authelia.com/configuration/security/access-control/) - Configure access control rules and policies for protected domains

###### Data & Sessions

* [Session Configuration](https://www.authelia.com/configuration/session/introduction/) - Configure session management, cookies, and timeouts
* [Storage Configuration](https://www.authelia.com/configuration/storage/introduction/) - Configure the storage backend for user data and sessions

###### Security & Notifications

* [Regulation](https://www.authelia.com/configuration/security/regulation/) - Configure brute-force protection and rate limiting
* [Notifier](https://www.authelia.com/configuration/notifications/introduction/) - Configure notification delivery methods and settings

These documentation pages provide comprehensive information about each configuration section, including all available options, examples, and best practices for setting up your Authelia instance.

#### Secrets

In the config there are go templates that can be identified by `{{ }}`. These are replaced with the contents of the files specified when Authelia is started. More information on them and the directives involved can be found [here](https://www.authelia.com/reference/guides/templating/).

There are 3 required secrets that we need to create and put in `authelia/secrets/` directory:
* jwt_secret.txt
* storage_encryption_key.txt
* session_secret.txt

You can automatically generate these secrets by running the following commands in the project root directory `project/`.
```shell{{title="Ensure Correct Permissions"}}
chown 8000:8000 ./authelia/secrets && chmod 0700 ./authelia/secrets
```
```shell{{title="Generate Secrets"}
docker run --rm -u 8000:8000 -v ./authelia/secrets:/secrets docker.io/authelia/authelia sh -c "cd /secrets && authelia crypto rand --length 64 session_secret.txt storage_encryption_key.txt jwt_secret.txt"
```

**Note** If you elect to generate these secrets yourself, it is *Strongly Recommended* that these 3 values are [Random Alphanumeric Strings](https://www.authelia.com/reference/guides/generating-secure-values/#generating-a-random-alphanumeric-string) with 64 or more characters.

#### Users Database

```yaml{title="authelia/config/users.yml"}
users:
  authelia: ## Username
    displayname: 'Authelia User'
    ## WARNING: This is a default password for testing only!
    ## IMPORTANT: Change this password before deploying to production!
    ## Generate a new hash using the instructions at:
    ## https://www.authelia.com/reference/guides/passwords/#passwords
    ## Password is 'authelia'
    password: '$6$rounds=50000$BpLnfgDsc2WD8F2q$Zis.ixdg9s/UOJYrs56b5QEZFiZECu0qZVNsIYxBaNJ7ucIL.nlxVCT5tqh8KHG8X4tlwCFm5r6NTOZZ5qRFN/'
    email: 'authelia@authelia.com'
    groups:
      - 'admin'
      - 'dev'
```

The current password listed is `authelia`. It is important you [Generate](https://www.authelia.com/reference/guides/passwords/#passwords) a new password hash.

### Starting the Stack

Once all the configuration for [Traefik](https://doc.traefik.io/traefik/) and [Authelia](https://www.authelia.com/) are complete, from the `project/` directory run `docker compose up -d` to download and start the containers.

### Verifying the Setup

1. Check container status: `docker compose ps`
2. Access Traefik dashboard at `https://traefik.{{< sitevar name="domain" nojs="example.com" >}}`
3. Test authentication at `https://whoami-secure.{{< sitevar name="domain" nojs="example.com" >}}`

### Troubleshooting

- Check container logs: `docker logs authelia`
- Ensure all secrets files exist and have correct permissions.

### Next Steps

This guide is not intended to instruct users on how to set up every aspect of Authelia. There are other features that were not mentioned in this guide that provide additional functionality. Some of these include:
- [Open ID Connect 1.0](https://www.authelia.com/configuration/identity-providers/openid-connect/provider/) which allows Authelia to handle authentication for applications that support the [Open ID Connect](https://openid.net/developers/how-connect-works/) protocol.
- [External Databases](https://www.authelia.com/configuration/storage/introduction/). Authelia supports more database types than just [SQLite](https://www.sqlite.org/index.html), including [MySql](https://hub.docker.com/_/mysql/) and [Postgres](https://hub.docker.com/_/postgres).
- [Non-memory Session Storage](https://www.authelia.com/configuration/session/introduction/) using [Redis](https://hub.docker.com/_/redis/). The default session provider is memory-only, this means that when Authelia restarts, all user sessions are destroyed and users are required to reauthenticate. Redis allows sessions to persist across restarts and makes Authelia fully stateless.
- [Metrics](https://www.authelia.com/configuration/telemetry/metrics/) allows Authelia administrators to export [various statistics](https://www.authelia.com/reference/guides/metrics/) regarding their individual Authelia installation.
