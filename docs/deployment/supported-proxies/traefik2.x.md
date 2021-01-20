---
layout: default
title: Traefik 2.x
parent: Proxy Integration
grand_parent: Deployment
nav_order: 3
---

# Traefik2

[Traefik 2.x] is a reverse proxy supported by **Authelia**.

## Configuration

Below you will find commented examples of the following configuration:

* Traefik 2.x
* Authelia portal
* Protected endpoint (Nextcloud)

The below configuration looks to provide examples of running Traefik 2.x with labels to protect your endpoint (Nextcloud in this case).

Please ensure that you also setup the respective [ACME configuration](https://docs.traefik.io/https/acme/) for your Traefik setup as this is not covered in the example below.

##### docker-compose.yml
```yml
version: '3'

networks:
  net:
    driver: bridge

services:

  traefik:
    image: traefik:v2.2
    container_name: traefik
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    networks:
      - net
    labels:
      - 'traefik.enable=true'
      - 'traefik.http.routers.api.rule=Host(`traefik.example.com`)'
      - 'traefik.http.routers.api.entrypoints=https'
      - 'traefik.http.routers.api.service=api@internal'
      - 'traefik.http.routers.api.tls=true'
    ports:
      - 80:80
      - 443:443
    command:
      - '--api'
      - '--providers.docker=true'
      - '--providers.docker.exposedByDefault=false'
      - '--entrypoints.http=true'
      - '--entrypoints.http.address=:80'
      - '--entrypoints.http.http.redirections.entrypoint.to=https'
      - '--entrypoints.http.http.redirections.entrypoint.scheme=https'
      - '--entrypoints.https=true'
      - '--entrypoints.https.address=:443'
      - '--log=true'
      - '--log.level=DEBUG'
      - '--log.filepath=/var/log/traefik.log'

  authelia:
    image: authelia/authelia
    container_name: authelia
    volumes:
      - /path/to/authelia:/config
    networks:
      - net
    labels:
      - 'traefik.enable=true'
      - 'traefik.http.routers.authelia.rule=Host(`login.example.com`)'
      - 'traefik.http.routers.authelia.entrypoints=https'
      - 'traefik.http.routers.authelia.tls=true'
      - 'traefik.http.middlewares.authelia.forwardauth.address=http://authelia:9091/api/verify?rd=https://login.example.com/'
      - 'traefik.http.middlewares.authelia.forwardauth.trustForwardHeader=true'
      - 'traefik.http.middlewares.authelia.forwardauth.authResponseHeaders=Remote-User, Remote-Groups, Remote-Name, Remote-Email'
    expose:
      - 9091
    restart: unless-stopped
    environment:
      - TZ=Australia/Melbourne

  nextcloud:
    image: linuxserver/nextcloud
    container_name: nextcloud
    volumes:
      - /path/to/nextcloud/config:/config
      - /path/to/nextcloud/data:/data
    networks:
      - net
    labels:
      - 'traefik.enable=true'
      - 'traefik.http.routers.nextcloud.rule=Host(`nextcloud.example.com`)'
      - 'traefik.http.routers.nextcloud.entrypoints=https'
      - 'traefik.http.routers.nextcloud.tls=true'
      - 'traefik.http.routers.nextcloud.middlewares=authelia@docker'
    expose:
      - 443
    restart: unless-stopped
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=Australia/Melbourne
```

## FAQ

### Middleware authelia@docker not found

If Traefik and Authelia are defined in different docker compose stacks you may experience
an issue where Traefik complains that: `middleware authelia@docker not found`.

This can be avoided a couple different ways:
1. Ensure Authelia container is up before Traefik is started:
    - Utilise the [`depends_on` option](https://docs.docker.com/compose/compose-file/#depends_on)
2. Define the Authelia middleware on your Traefik container
```yaml
- 'traefik.http.middlewares.authelia.forwardauth.address=http://authelia:9091/api/verify?rd=https://login.example.com/'
- 'traefik.http.middlewares.authelia.forwardauth.trustForwardHeader=true'
- 'traefik.http.middlewares.authelia.forwardauth.authResponseHeaders=Remote-User, Remote-Groups, Remote-Name, Remote-Email'
```
    
[Traefik 2.x]: https://docs.traefik.io/
