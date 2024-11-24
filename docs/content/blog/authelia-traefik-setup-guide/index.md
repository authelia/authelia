---
title: "Authelia + Traefik Setup Guide"
description: "There have been lots of issues with people using guides that are out of date. This guide will attempt to bridge the gap and give users a definitive best practice way to setup Authelia."
summary: "In this guide we will walk through setting up Authelia with Traefik as the reverse proxy. This guide aims to provide an opinionated way to setup Authelia that is fully supported by the Authelia team." #TODO: change this description
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
## Assumptions
We assume these items have already been completed prior to starting this guide.
- [Docker](https://docs.docker.com/engine/install/) needs to be configured.
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
    ┃ ┣ 📄 acme.json
    ┃ ┣ 📄 static.json
    ┃ ┗ 📄 traefik.json
    ┣ 📁 logs
    ┗ 📁 secrets
```

## Traefik and Whoami
{{< callout context="note" title="Note" icon="outline/info-circle" >}}
We'll focus on the minimal configuration needed to work with Authelia. For advanced Traefik features and configurations, consult their documentation.
{{< /callout >}}

First, we'll set up Traefik as our reverse proxy. For detailed Traefik documentation, refer to the [official Traefik docs](https://doc.traefik.io/traefik/).

#### Docker Compose


```yaml{title="compose.yml"}
services:
  traefik:
    image: traefik:latest
    container_name: traefik
    restart: unless-stopped
    depends_on:
      - authelia
    security_opt:
      - no-new-privileges=true
    networks:
      proxy:
      authelia:
    ports:
      - '80:80'
      - '443:443'
    environment:
      TZ: America/Los_Angeles # see below
      CF_EMAIL_FILE: /run/secrets/cloudflare_email
      CF_DNS_API_TOKEN_FILE: /run/secrets/cloudflare_api_key
    volumes:
      - '/var/run/docker.sock:/var/run/docker.sock:ro'
      - './traefik/config/traefik.yml:/traefik.yml:ro'
      - './traefik/config/static.yml:/static.yml:ro'
      - './traefik/config/acme.json:/acme.json'
      - './traefik/logs:/logs'
    secrets:
      - cloudflare_email
      - cloudflare_api_key

  whoami:
    image: traefik/whoami
    restart: unless-stopped
    container_name: whoami
    labels:
      - traefik.enable=true
      - traefik.http.routers.whoami.rule=Host(`whoami.example.com`)
      - traefik.http.routers.whoami.entrypoints=https
      - traefik.http.routers.whoami.tls=true
      - traefik.http.services.whoami.loadbalancer.server.port=80
    networks:
      - proxy

secrets:
  cloudflare_email:
    file: ./traefik/secrets/cloudflare_email.txt
  cloudflare_api_key:
    file: ./traefik/secrets/cloudflare_api_key.txt

networks:
  proxy:
    external: true
  authelia:
```
Notes:
Timezone strings can be found [here](https://go.dev/src/time/zoneinfo_abbrs_windows.go)

#### Basic Traefik Configuration
The following files contain the minimal Traefik configuration needed for Authelia integration:

```yaml{title="traefik/config/traefik.yml"}
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

serversTransport:
  insecureSkipVerify: true

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
  file:
    filename: /static.yml

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
Create acme.json with 600 permissions before starting Traefik: `touch acme.json && chmod 600 acme.json`

### Domain Configuration
```yaml{title="traefik/config/static.yml"}
http:
  routers:
    traefik:
      rule: Host(`traefik.example.com`)
      entrypoints:
        https:
      service: api@internal
      middlewares:
        - authelia
      tls:
        certResolver: cloudflare
        domains:
          ### Domain configuration
          ### Add your domains here for TLS certificate generation
          - main "example.com"
            sans:
              - "*.example.com" # Wildcard certificates require the DNS challenge.
          #- main "example2.com" ## Add additional domains as needed.
          #  sans:
          #    - "*.example2.com"
```

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
These are minimal configurations focused on Authelia integration. Adjust them according to your needs using Traefik's documentation.
{{< /callout >}}

## Authelia Compose
This configuration sets up Authelia's core service and configures forward authentication with Traefik. The portal will be available at `auth.example.com`. It also defines a new whoami container that will be protected by authelia.

```yaml{title="compose.yml"}
...
  authelia:
    image: authelia/authelia:4.38.17
    container_name: authelia
    restart: unless-stopped
    volumes:
      - ./authelia/secrets:/secrets:ro
      - ./authelia/config:/config
      - ./authelia/logs:/var/log/authelia/
    networks:
      proxy:
      authelia:
    labels:
      # Expose Authelia through Traefik
      - traefik.enable=true
      - traefik.docker.network=authelia
      - traefik.http.routers.authelia.rule=Host(`auth.example.com`)
      - traefik.http.routers.authelia.entrypoints=https
      # Forward auth config
      - traefik.http.middlewares.authelia.forwardAuth.address=http://authelia:9091/api/authz/forward-auth
      - traefik.http.middlewares.authelia.forwardAuth.trustForwardHeader=true
      - traefik.http.middlewares.authelia.forwardAuth.authResponseHeaders=Remote-User,Remote-Groups,Remote-Name,Remote-Email
    environment:
      - TZ=America/Los_Angeles
      - X_AUTHELIA_CONFIG_FILTERS=template

  whoami-secure:
    image: traefik/whoami
    restart: unless-stopped
    container_name: whoami-secure
    labels:
      - traefik.enable=true
      - traefik.http.routers.whoami.rule=Host(`whoami-secure.example.com`)
      - traefik.http.routers.whoami.entrypoints=https
      - traefik.http.routers.whoami.tls=true
      - traefik.http.services.whoami.loadbalancer.server.port=80
      - traefik.http.routers.whoami.middlewares=authelia
    networks:
      - proxy
...
```

#### Authelia Configuration
```yaml{title="authelia/config/configuration.yml"}
server: ## https://www.authelia.com/configuration/miscellaneous/server/
  address: 'tcp4://:9091'

log: ## https://www.authelia.com/configuration/miscellaneous/logging/
  level: debug
  file_path: '/var/log/authelia/authelia.log'
  keep_stdout: true

identity_validation: ## https://www.authelia.com/configuration/identity-validation/introduction/
  elevated_session:
    require_second_factor: true
  reset_password:
    jwt_lifespan: '5 minutes'
    jwt_secret: {{ secret "/config_secrets/jwt_secret.txt" | mindent 0 "|" | msquote}}

totp: ## https://www.authelia.com/configuration/second-factor/time-based-one-time-password/
  disable: false
  issuer: example.com
  period: 30
  skew: 1

password_policy: ## https://www.authelia.com/configuration/security/password-policy/
  zxcvbn:
    enabled: true
    min_score: 4

authentication_backend: ## https://www.authelia.com/configuration/first-factor/introduction/
  file:
    path: '/config/users.yml'
    password:
      algorithm: argon2id
      iterations: 3
      salt_length: 16
      parallelism: 8
      memory: 32768
      key_length: 32

access_control: ## https://www.authelia.com/configuration/security/access-control/
  default_policy: deny
  rules:
    - domain: traefik.example.com
      policy: one_factor
    - domain: whoami-secure.example.com
      policy: two_factor

session: ## https://www.authelia.com/configuration/session/introduction/
  name: authelia_session
  secret: {{ secret "/secrets/session_secret.txt" | mindent 0 "|" | msquote}}
  cookies:
    - domain: 'example.com'
      authelia_url: 'https://auth.example.com'

regulation: ## https://www.authelia.com/configuration/security/regulation/
  max_retries: 4
  find_time: 120
  ban_time: 300

storage: ## https://www.authelia.com/configuration/storage/introduction/
  encryption_key: {{ secret "/config_secrets/storage_encryption_key.txt" | mindent 0 "|" | msquote}}
  local:
    path: '/config/db.sqlite3'

notifier: ## https://www.authelia.com/configuration/notifications/introduction/
  disable_startup_check: false
  filesystem:
    filename: '/config/notification.txt'
```
