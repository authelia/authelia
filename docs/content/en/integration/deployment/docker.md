---
title: "Docker"
description: "A guide on installing Authelia in Docker."
lead: "This is one of the primary ways we deliver Authelia to users and the recommended path."
date: 2022-05-27T22:24:38+10:00
lastmod: 2022-06-03T10:43:55+10:00
draft: false
images: []
menu:
  integration:
    parent: "deployment"
weight: 230
toc: true
---

The [Docker] container is deployed with the following image names:

* [authelia/authelia](https://hub.docker.com/r/authelia/authelia)
* [docker.io/authelia/authelia](https://hub.docker.com/r/authelia/authelia)
* [ghcr.io/authelia/authelia](https://github.com/authelia/authelia/pkgs/container/authelia)

## Docker Compose

We provide two main [Docker Compose] examples which can be utilized to help test *Authelia* or can be adapted into your
existing [Docker Compose].

* [Unbundled Example](#standalone-example)
* [Bundle: lite](#lite)
* [Bundle: local](#local)

### Standalone Example

The following is an example [Docker Compose] deployment with just *Authelia* and no bundled applications or proxies.

It expects the following:

* The file `data/authelia/config/configuration.yml` is present and the configuration file.
* The files `data/authelia/secrets/*` exist and contain the relevant secrets.
* You're using PostgreSQL.
* You have an external network named `net` which is in bridge mode.

```yaml
version: "3.8"
secrets:
  JWT_SECRET:
    file: ${PWD}/data/authelia/secrets/JWT_SECRET
  SESSION_SECRET:
    file: ${PWD}/data/authelia/secrets/SESSION_SECRET
  STORAGE_PASSWORD:
    file: ${PWD}/data/authelia/secrets/STORAGE_PASSWORD
  STORAGE_ENCRYPTION_KEY:
    file: ${PWD}/data/authelia/secrets/STORAGE_ENCRYPTION_KEY
  OIDC_HMAC_KEY:
    file: ${PWD}/data/authelia/secrets/OIDC_HMAC_KEY
  OIDC_PRIVATE_KEY:
    file: ${PWD}/data/authelia/secrets/OIDC_PRIVATE_KEY
services:
  authelia:
    container_name: authelia
    image: docker.io/authelia/authelia:latest
    restart: unless-stopped
    networks:
      net:
        aliases: []
    expose:
      - 9091
    secrets: [JWT_SECRET, SESSION_SECRET, STORAGE_PASSWORD, STORAGE_ENCRYPTION_KEY, OIDC_HMAC_KEY, OIDC_PRIVATE_KEY]
    environment:
      AUTHELIA_JWT_SECRET_FILE: /run/secrets/JWT_SECRET
      AUTHELIA_SESSION_SECRET_FILE: /run/secrets/SESSION_SECRET
      AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE: /run/secrets/STORAGE_PASSWORD
      AUTHELIA_STORAGE_ENCRYPTION_KEY_FILE: /run/secrets/STORAGE_ENCRYPTION_KEY
      AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE: /run/secrets/OIDC_HMAC_KEY
      AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE: /run/secrets/OIDC_PRIVATE_KEY
    volumes:
      - ${PWD}/data/authelia/config:/config
networks:
  net:
    external: true
    name: net
```

### Bundles

To use the bundles we recommend first cloning the git repository and checking out the latest release on a Linux Desktop:

```bash
git clone https://github.com/authelia/authelia.git
cd authelia
git checkout $(git describe --tags `git rev-list --tags --max-count=1`)
```

#### lite

The [lite bundle](https://github.com/authelia/authelia/tree/master/examples/compose/lite) can be used by following this
process:

1. Perform the commands in [the bundles section](#bundles).
2. Run the `cd examples/compose/lite` command.
3. Edit `users_database.yml` and either change the username of the `authelia` user, or
   [generate a new password](../../configuration/first-factor/file.md#passwords), or both. The default password is
   `authelia`.
4. Edit the `configuration.yml` and `docker-compose.yml` with your respective domains and secrets.
5. Run `docker compose up -d` or `docker-compose up -d`.

#### local

The [local bundle](https://github.com/authelia/authelia/tree/master/examples/compose/local) can be setup after cloning
the repository as per the [bundles](#bundles) section then running the following commands on a Linux Desktop:

```bash
cd examples/compose/local
./setup.sh
```

The bundle setup modifies the `/etc/hosts` file which is performed with `sudo`. Once it is successfully setup you can
visit the following URL's to see Authelia in action (`example.com` will be replaced by the domain you specified):

* [https://public.example.com](https://public.example.com) - Bypasses Authelia
* [https://traefik.example.com](https://traefik.example.com) - Secured with Authelia one-factor authentication
* [https://secure.example.com](https://secure.example.com) - Secured with Authelia two-factor authentication (see note below)

You will need to authorize the self-signed certificate upon visiting each domain. To visit
[https://secure.example.com](https://secure.example.com) you will need to register a device for second factor
authentication and confirm by clicking on a link sent by email. Since this is a demo with a fake email address, the
content of the email will be stored in `./authelia/notification.txt`. Upon registering, you can grab this link easily by
running the following command:

```bash
grep -Eo '"https://.*" ' ./authelia/notification.txt.
```

[Docker]: https://docker.com
[Docker Compose]: https://docs.docker.com/compose/
