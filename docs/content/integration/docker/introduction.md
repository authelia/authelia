---
title: "Docker"
description: "A guide on installing Authelia in Docker."
summary: "This is one of the primary ways we deliver Authelia to users and the recommended path."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 410
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The [Docker] container is deployed with the following image names:

* [authelia/authelia](https://hub.docker.com/r/authelia/authelia)
* [docker.io/authelia/authelia](https://hub.docker.com/r/authelia/authelia)
* [ghcr.io/authelia/authelia](https://github.com/authelia/authelia/pkgs/container/authelia)

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Container

### Environment Variables

Several environment variables apply specifically to the official container. This table documents them. It is important
to note these environment variables are specific to the container and have no effect on the *Authelia* daemon itself and
this section is not meant to document the daemon environment variables.

| Name  | Default |                                             Usage                                             |
|:-----:|:-------:|:---------------------------------------------------------------------------------------------:|
| PUID  |    0    | If the container is running as UID 0, it will drop privileges to this UID via the entrypoint  |
| PGID  |    0    | If the container is running as UID 0, it will drop privileges to this GID via the entrypoint  |
| UMASK |   N/A   | If set the container will run with the provided UMASK by running the `umask ${UMASK}` command |

### Permission Context

By default the container runs as the configured [Docker] daemon user. Users can control this behavior in several ways.

The first and recommended way is instructing the [Docker] daemon to run the *Authelia* container as another user. See
the [docker run] or [Docker Compose file reference documentation](https://docs.docker.com/compose/compose-file/05-services/#user)
for more information. The best part of this method is the process will never have privileged access, and the only
negative is the user must manually configure the filesystem permissions correctly.

The second method is by using the environment variables listed above. The downside to this method is that the entrypoint
itself will run as UID 0 (root). The advantage is the container will automatically set owner and permissions on the
filesystem correctly.

The last method which is beyond our documentation or support is using the
[user namespace](https://docs.docker.com/engine/security/userns-remap/) facility [Docker] provides.

[docker run]: https://docs.docker.com/engine/reference/commandline/run/

## Docker Compose

We provide two main [Docker Compose] examples which can be utilized to help test *Authelia* or can be adapted into your
existing [Docker Compose].

The [Bundles](#bundles) offer a curated example of a full setup with a proxy and some applications. The
[Standalone Example](#standalone-example) offers a simple example just to get you on your way and is intended for more
experienced users who just want to see the recommended deployment options.

### Bundles

The [Docker Compose] bundles have their own [dedicated page](bundles.md).

### Standalone Example

The following examples are [Docker Compose] deployments with just *Authelia* and no bundled applications or
proxies.

It expects the following:

* The file `data/authelia/config/configuration.yml` is present and the configuration file.
* The directory `data/authelia/secrets/` exists and contain the relevant [secret](../../configuration/methods/secrets.md) files:
  * A file named `JWT_SECRET` for the [jwt_secret](../../configuration/identity-validation/reset-password.md#jwt_secret)
  * A file named `SESSION_SECRET` for the [session secret](../../configuration/session/introduction.md#secret)
  * A file named `STORAGE_PASSWORD` for the [PostgreSQL password secret](../../configuration/storage/postgres.md#password)
  * A file named `STORAGE_ENCRYPTION_KEY` for the [storage encryption_key secret](../../configuration/storage/introduction.md#encryption_key)
* You're using PostgreSQL.
* You have an external network named `net` which is in bridge mode.

#### Using Secrets

Use this [Standalone Example](#standalone-example) if you want to use
[docker secrets](https://docs.docker.com/engine/swarm/secrets/).

```yaml {title="compose.yml"}
---
secrets:
  JWT_SECRET:
    file: '${PWD}/data/authelia/secrets/JWT_SECRET'
  SESSION_SECRET:
    file: '${PWD}/data/authelia/secrets/SESSION_SECRET'
  STORAGE_PASSWORD:
    file: '${PWD}/data/authelia/secrets/STORAGE_PASSWORD'
  STORAGE_ENCRYPTION_KEY:
    file: '${PWD}/data/authelia/secrets/STORAGE_ENCRYPTION_KEY'
services:
  authelia:
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    image: 'docker.io/authelia/authelia:latest'
    restart: 'unless-stopped'
    networks:
      net:
        aliases: []
    secrets: ['JWT_SECRET', 'SESSION_SECRET', 'STORAGE_PASSWORD', 'STORAGE_ENCRYPTION_KEY']
    environment:
      AUTHELIA_IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET_FILE: '/run/secrets/JWT_SECRET'
      AUTHELIA_SESSION_SECRET_FILE: '/run/secrets/SESSION_SECRET'
      AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE: '/run/secrets/STORAGE_PASSWORD'
      AUTHELIA_STORAGE_ENCRYPTION_KEY_FILE: '/run/secrets/STORAGE_ENCRYPTION_KEY'
    volumes:
      - '${PWD}/data/authelia/config:/config'
networks:
  net:
    external: true
    name: 'net'
...
```

#### Using a Secrets Volume

Use this [Standalone Example](#standalone-example) if you want to use a standard
[docker volume](https://docs.docker.com/storage/volumes/) or bind mount for your secrets.

```yaml {title="compose.yml"}
---
services:
  authelia:
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    image: 'docker.io/authelia/authelia:latest'
    restart: 'unless-stopped'
    networks:
      net:
        aliases: []
    environment:
      AUTHELIA_IDENTITY_VALIDATION_RESET_PASSWORD_JWT_SECRET_FILE: '/secrets/JWT_SECRET'
      AUTHELIA_SESSION_SECRET_FILE: '/secrets/SESSION_SECRET'
      AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE: '/secrets/STORAGE_PASSWORD'
      AUTHELIA_STORAGE_ENCRYPTION_KEY_FILE: '/secrets/STORAGE_ENCRYPTION_KEY'
    volumes:
      - '${PWD}/data/authelia/config:/config'
      - '${PWD}/data/authelia/secrets:/secrets'
networks:
  net:
    external: true
    name: 'net'
...
```

## Frequently Asked Questions

#### Running the Proxy on the Host Instead of in a Container

If you wish to run the proxy as a systemd service or other daemon, you will need to adjust the configuration. While this
configuration is not specific to *Authelia* and is mostly a [Docker] concept we explain this here to help alleviate the
users asking how to accomplish this. It should be noted that we can't provide documentation or support for every
architectural choice our users make and you should expect to do your own research to figure this out where possible.

The example below includes the additional `ports` option which must be added in order to allow communication to
*Authelia* from daemons on the [Docker] host. The other values are used to show context within the
[Standalone Example](#standalone-example) above. The example allows *Authelia* to be communicated with over the
localhost IP address `127.0.0.1` on port `9091`. You need to adjust this to your specific needs.

```yaml {title="compose.yml"}
---
services:
  authelia:
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    image: 'docker.io/authelia/authelia:latest'
    restart: 'unless-stopped'
    networks:
      net:
        aliases: []
    ports:
      - '127.0.0.1:{{< sitevar name="port" nojs="9091" >}}:{{< sitevar name="port" nojs="9091" >}}'
...
```

#### How do I debug a container startup issue when the logs don't indicate anything helpful?

Generally the best way to debug this is to start Authelia interactively. While most of the time the logs will be helpful
there are some specific conditions that prevent this. This example assumes the following are true:

1. The container name for Authelia is `authelia`.
2. The file you're using for compose is `compose.yml`.

The following command will allow you to run your existing compose with the additional required composition of the
example `compose.debug.yml`:

```bash
docker compose up -f compose.yml -f compose.debug.yml -d
```

The following command sequence will allow you to run Authelia interactively within the container:

```bash
$ docker exec -it authelia sh
$ authelia
```

The following is the supporting `compose.debug.yml`:

```yaml {title="compose.debug.yml"}
---
services:
  authelia:
    healthcheck:
      disable: true
    environment:
      AUTHELIA_LOG_LEVEL: 'trace'
    command: 'sleep 3300'
...
```

[Docker]: https://docker.com
[Docker Compose]: https://docs.docker.com/compose/
