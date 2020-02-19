---
layout: default
title: Traefik 2.x
parent: Supported Proxies
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
    image: traefik:v2.1.2
    container_name: traefik
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    networks:
      - net
    labels:
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
      - '--entrypoints.http=true'
      - '--entrypoints.http.address=:80'
      - '--entrypoints.https=true'
      - '--entrypoints.https.address=:443'
      - '--log=true'
      - '--log.level=DEBUG'
      - '--log.filepath=/var/log/traefik.log'

  authelia:
    image: authelia/authelia
    container_name: authelia
    volumes:
      - /path/to/authelia:/var/lib/authelia
      - /path/to/authelia/config.yml:/etc/authelia/configuration.yml:ro
    networks:
      - net
    labels:
      - 'traefik.http.routers.authelia.rule=Host(`login.example.com`)'
      - 'traefik.http.routers.authelia.entrypoints=https'
      - 'traefik.http.routers.authelia.tls=true'
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
      - 'traefik.http.routers.nextcloud.rule=Host(`nextcloud.example.com`)'
      - 'traefik.http.routers.nextcloud.entrypoints=https'
      - 'traefik.http.routers.nextcloud.tls=true'
      - 'traefik.http.routers.nextcloud.middlewares=authelia'
      - 'traefik.http.middlewares.authelia.forwardAuth.address=http://authelia:9091/api/verify?rd=https://login.example.com/'
    expose:
      - 443
    restart: unless-stopped
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=Australia/Melbourne
```

[Traefik 2.x]: https://docs.traefik.io/
