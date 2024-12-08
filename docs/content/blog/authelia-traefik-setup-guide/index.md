---
title: "Authelia + Traefik Setup Guide"
description: "There have been lots of issues with people using guides that are out of date. This guide will attempt to bridge the gap and give users a definitive best practice way to setup Authelia."
summary: "In this guide we will walk through setting up Authelia with Traefik as the reverse proxy. This guide aims to provide an opinionated way to setup Authelia that is fully supported by the Authelia team."
date: 2024-11-23T10:10:09+10:00
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
This is not a demo. If you would like an all-in-one demo, please take a look at our [local bundle](https://www.authelia.com/integration/deployment/docker/#local).
## Assumptions and Adaptation
This guide makes a few assumptions. These assumptions may require adaptation in more advanced and complex scenarios. We can not reasonablly have examples for every advanced configuration option that exists. Some of these values can be automatically replaced with documentation variables.

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
 ┃  ┃ ┣ 📄 configuration.yaml
 ┃  ┃ ┗ 📄 users.yaml
 ┃  ┗ 📁 secrets
 ┣ 📄 compose.yaml
 ┗ 📁 traefik
    ┣ 📁 config
    ┃ ┣ 📄 acme.json
    ┃ ┣ 📄 dynamic.yaml
    ┃ ┗ 📄 traefik.yaml
    ┣ 📁 logs
    ┗ 📁 secrets
```

## Traefik and Whoami
{{< callout context="note" title="Note" icon="outline/info-circle" >}}
We'll focus on the minimal configuration needed to work with Authelia. For advanced Traefik features and configurations, consult their documentation.
{{< /callout >}}

Next, we'll set up Traefik as our reverse proxy. For detailed Traefik documentation, refer to the [official Traefik docs](https://doc.traefik.io/traefik/).

#### Docker Compose


```yaml{title="compose.yaml"}
services:
  traefik:
    image: traefik:latest
    container_name: traefik
    restart: unless-stopped
    security_opt:
      - no-new-privileges=true
    networks:
      proxy: {}
      authelia:
        aliases:
          - '{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
    ports:
      - '80:80'
      - '443:443'
    environment:
      TZ: 'America/Los_Angeles' # see below
      CF_EMAIL_FILE: '/run/secrets/cloudflare_email'
      CF_DNS_API_TOKEN_FILE: '/run/secrets/cloudflare_api_key'
    volumes:
      - '/var/run/docker.sock:/var/run/docker.sock:ro'
      - './traefik/config/traefik.yaml:/traefik.yaml:ro'
      - './traefik/config/dynamic.yaml:/dynamic.yaml:ro'
      - './traefik/config/acme.json:/acme.json'
      - './traefik/logs:/logs'
    secrets:
      - cloudflare_email
      - cloudflare_api_key
    labels:
      traefik.enable: 'true'
      traefik.http.routers.dashboard.rule: 'Host(`traefik.{{< sitevar name="domain" nojs="example.com" >}})'
      traefik.http.routers.dashboard.entrypoints: 'https'
      traefik.http.routers.dashboard.middlewares: 'authelia@docker'
      traefik.http.routers.dashboard.service: 'api@internal'

  whoami:
    image: traefik/whoami
    restart: unless-stopped
    container_name: whoami
    labels:
      traefik.enable: 'true'
      traefik.http.routers.whoami.rule: Host(`whoami.{{< sitevar name="domain" nojs="example.com" >}}`)
      traefik.http.routers.whoami.entrypoints: 'https'
    networks:
      proxy: {}

secrets:
  cloudflare_email:
    file: ./traefik/secrets/cloudflare_email.txt
  cloudflare_api_key:
    file: ./traefik/secrets/cloudflare_api_key.txt

networks:
  proxy:
    external: true
  authelia: {}
```
Notes:
Timezone strings can be found [here](https://go.dev/src/time/zoneinfo_abbrs_windows.go)

#### Basic Traefik Configuration
Now we configure Traefik.
The following files contain the minimal Traefik configuration needed for Authelia integration:

```yaml{title="traefik/config/traefik.yaml"}
# Base Traefik configuration
api:
  dashboard: true
  debug: false
  insecure: false

log:
  level: INFO
  filePath: /logs/traefik.log
accessLog:
  filePath: /logs/access.log

entryPoints:
  http:
    address: ":80"
    http:
      redirections:
       entryPoint:
         to: https
         scheme: https
         permanent: true
  https:
    address: ":443"
    http:
      tls:
        certResolver: cloudflare@file
        domains:
          - main: '{{< sitevar name="domain" nojs="example.com" >}}.com'
            sans:
              - '*.{{< sitevar name="domain" nojs="example.com" >}}.com'

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
  file:
    filename: '/dynamic.yaml'

certificatesResolvers:
  cloudflare:
    acme:
      storage: acme.json
      dnsChallenge:
        provider: cloudflare
        resolvers:
          - "1.1.1.1:53"
          - "1.0.0.1:53"

tls:
  options:
    default:
      minVersion: "VersionTLS12"
      cipherSuites:
        - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
        - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
        - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
        - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
        - TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305
        - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
```
Note: Create acme.json with 600 permissions before starting Traefik.

### Domain Configuration
```yaml{title="traefik/config/dynamic.yaml"}
##This file can be used to define dynamic routers/services/middlewares.
```

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
These are minimal configurations focused on Authelia integration. Adjust them according to your needs using Traefik's documentation.
{{< /callout >}}

## Authelia Compose
This configuration sets up Authelia's core service and configures forward authentication with Traefik. The portal will be available at `{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`. It also defines a new whoami container that will be protected by authelia.

```yaml{title="compose.yaml"}
...
  authelia:
    image: authelia/authelia:4.38.17
    container_name: {{< sitevar name="host" nojs="authelia" >}}
    volumes:
      - ./authelia/secrets:/secrets:ro
      - ./authelia/config:/config
      - ./authelia/logs:/var/log/authelia/
    networks:
      authelia: {}
    labels:
      ## Expose Authelia through Traefik
      traefik.enable: 'true'
      traefik.docker.network: 'authelia'
      traefik.http.routers.authelia.rule: 'Host(`{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`)'
      traefik.http.routers.authelia.entrypoints: 'https'
      ## Setup Authelia Middlewares
      traefik.http.middlewares.authelia.forwardAuth.address: '{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}/api/authz/forward-auth'
      traefik.http.middlewares.authelia.forwardAuth.trustForwardHeader: 'true'
      traefik.http.middlewares.authelia.forwardAuth.authResponseHeaders: 'Remote-User,Remote-Groups,Remote-Name,Remote-Email'
    environment:
      TZ: 'America/Los_Angeles'
      X_AUTHELIA_CONFIG_FILTERS: template

  whoami-secure:
    image: traefik/whoami
    restart: unless-stopped
    container_name: whoami-secure
    labels:
      traefik.enable: 'true'
      traefik.http.routers.whoami-secure.rule: 'Host(`whoami-secure.{{< sitevar name="domain" nojs="example.com" >}}`)'
      traefik.http.routers.whoami-secure.entrypoints: 'https'
      traefik.http.routers.whoami-secure.middlewares: 'authelia@docker'
    networks:
      proxy: {}
...
```

#### Authelia Configuration
```yaml{title="authelia/config/configuration.yaml"}
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
    jwt_secret: {{ secret "/config_secrets/jwt_secret.txt" | mindent 0 "|" | msquote}}

totp:
  disable: false
  issuer: {{< sitevar name="domain" nojs="example.com" >}}
  period: 30
  skew: 1

password_policy:
  zxcvbn:
    enabled: true
    min_score: 4

authentication_backend:
  file:
    path: '/config/users.yaml'
    password:
      algorithm: argon2id
      iterations: 3
      salt_length: 16
      parallelism: 8
      memory: 32768
      key_length: 32

access_control:
  default_policy: deny
  rules:
    - domain: traefik.{{< sitevar name="domain" nojs="example.com" >}}
      policy: one_factor
    - domain: whoami-secure.{{< sitevar name="domain" nojs="example.com" >}}
      policy: two_factor

session:
  name: authelia_session
  secret: {{ secret "/secrets/session_secret.txt" | mindent 0 "|" | msquote}}
  cookies:
    - domain: '{{< sitevar name="domain" nojs="example.com" >}}'
      authelia_url: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'

regulation:
  max_retries: 4
  find_time: 120
  ban_time: 300

storage:
  encryption_key: {{ secret "/secrets/storage_encryption_key.txt" | mindent 0 "|" | msquote}}
  local:
    path: '/config/db.sqlite3'

notifier:
  disable_startup_check: false
  filesystem:
    filename: '/config/notification.txt'
```
Each section in the configuration file above has detailed documentation available. Below are direct links.
Note: There are options that have not been mentioned here.

###### Core Configuration
* [Server Configuration](https://www.authelia.com/configuration/miscellaneous/server/) - Configure the server address, ports, TLS settings, and other core server options
* [Logging](https://www.authelia.com/configuration/miscellaneous/logging/) - Configure log levels, output locations, and format options
* [Identity Validation](https://www.authelia.com/configuration/identity-validation/introduction/) - Configure settings for password reset, elevated sessions, and identity verification

###### Authentication & Security
* [TOTP Configuration](https://www.authelia.com/configuration/second-factor/time-based-one-time-password/) - Configure Time-based One-Time Password settings for two-factor authentication
* [Password Policy](https://www.authelia.com/configuration/security/password-policy/) - Configure password strength requirements and validation rules
* [Authentication Backend](https://www.authelia.com/configuration/first-factor/introduction/) - Configure the authentication provider and settings
* [Access Control](https://www.authelia.com/configuration/security/access-control/) - Configure access control rules and policies

###### Data & Sessions
* [Session Configuration](https://www.authelia.com/configuration/session/introduction/) - Configure session management, cookies, and timeouts
* [Storage Configuration](https://www.authelia.com/configuration/storage/introduction/) - Configure the storage backend for user data and sessions

###### Security & Notifications
* [Regulation](https://www.authelia.com/configuration/security/regulation/) - Configure brute-force protection and rate limiting
* [Notifier](https://www.authelia.com/configuration/notifications/introduction/) - Configure notification delivery methods and settings

These documentation pages provide comprehensive information about each configuration section, including all available options, examples, and best practices for setting up your Authelia instance.

#### Secrets
In the config there are [go templates](https://www.authelia.com/reference/guides/templating/) that can be identified by `{{ }}`. These are replaced with the contents of the files specified when Authelia is started.

There are 3 required secrets that we need to create and put in `authelia/secrets/` directory:
* jwt_secret.txt
* storage_encryption_key.txt
* session_secret.txt

It is *strongly recommended* that these 3 values are [Random Alphanumeric Strings](https://www.authelia.com/reference/guides/generating-secure-values/#generating-a-random-alphanumeric-string) with 64 or more characters.


#### Users Database

```yaml{title="authelia/config/users.yaml"}
users:
  authelia: # Username
    displayname: "Authelia User"
    # WARNING: This is a default password for testing only!
    # IMPORTANT: Change this password before deploying to production!
    # Generate a new hash using the instructions at:
    # https://www.authelia.com/reference/guides/passwords/#passwords
    # Password is authelia
    password: "$6$rounds=50000$BpLnfgDsc2WD8F2q$Zis.ixdg9s/UOJYrs56b5QEZFiZECu0qZVNsIYxBaNJ7ucIL.nlxVCT5tqh8KHG8X4tlwCFm5r6NTOZZ5qRFN/"
    email: authelia@authelia.com
    groups:
      - admin
      - dev
```
The current password listed is `authelia`. It is important you [Generate](https://www.authelia.com/reference/guides/passwords/#passwords) a new password hash.
