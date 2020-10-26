---
layout: default
title: Traefik 1.x
parent: Proxy Integration
grand_parent: Deployment
nav_order: 3
---

# Traefik

[Traefik 1.x] is a reverse proxy supported by **Authelia**.

## Configuration

Below you will find commented examples of the following configuration:

* Traefik 1.x
* Authelia portal
* Protected endpoint (Nextcloud)

The below configuration looks to provide examples of running Traefik 1.x with labels to protect your endpoint (Nextcloud in this case).

Please ensure that you also setup the respective [ACME configuration](https://docs.traefik.io/v1.7/configuration/acme/) for your Traefik setup as this is not covered in the example below.

##### docker-compose.yml
```yml
version: '3'

networks:
  net:
    driver: bridge

services:

  traefik:
    image: traefik:v1.7.20-alpine
    container_name: traefik
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    networks:
      - net
    labels:
      - 'traefik.frontend.rule=Host:traefik.example.com'
      - 'traefik.port=8081'
    ports:
      - 80:80
      - 443:443
      - 8081:8081
    restart: unless-stopped
    command:
      - '--api'
      - '--api.entrypoint=api'
      - '--docker'
      - '--defaultentrypoints=https'
      - '--logLevel=DEBUG'
      - '--traefiklog=true'
      - '--traefiklog.filepath=/var/log/traefik.log'
      - '--entryPoints=Name:http Address::80'
      - '--entryPoints=Name:https Address::443 TLS'
      - '--entryPoints=Name:api Address::8081'

  authelia:
    image: authelia/authelia
    container_name: authelia
    volumes:
      - /path/to/authelia:/config
    networks:
      - net
    labels:
      - 'traefik.frontend.rule=Host:login.example.com'
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
      - 'traefik.frontend.rule=Host:nextcloud.example.com'
      - 'traefik.frontend.auth.forward.address=http://authelia:9091/api/verify?rd=https://login.example.com/'
      - 'traefik.frontend.auth.forward.trustForwardHeader=true'
      - 'traefik.frontend.auth.forward.authResponseHeaders=Remote-User,Remote-Groups,Remote-Name,Remote-Email'
    expose:
      - 443
    restart: unless-stopped
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=Australia/Melbourne
```

[Traefik 1.x]: https://docs.traefik.io/v1.7/
