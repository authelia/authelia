---
layout: default
title: Deployment - Lite
parent: Deployment
nav_order: 1
---

# Lite Deployment

**Authelia** can be deployed as a lite setup with minimal external dependencies.
The setup is called lite because it reduces the number of components in the architecture
to a reverse proxy such as Nginx, Traefik or HAProxy, Authelia and Redis.

This setup assumes you have basic knowledge and understanding of IP addresses, DNS and port
forwarding. You should setup the domain you intend to protect with Authelia to point to your
external IP address and port forward ports `80` and `443` to the host you plan to host the
`docker-compose.yml` bundle.

Port 80 is utilised by LetsEncrypt for certificate challenges, this will [automatically
provision](https://docs.traefik.io/https/acme/) up-to-date certificates for your domain(s).

Traefik publishes the respective services with LetsEncrypt provided certificates on port `443`.
The provided examples protect the Traefik dashboard with Authelia's one-factor auth
(traefik.example.com) and two instances of the
[whoami container](https://hub.docker.com/r/containous/whoami) with Authelia being
bypassed (public.example.com) and another with it's two-factor auth (secure.example.com). 

If you happen to already have an external SQL instance (MariaDB, MySQL or Postgres) this 
setup can easily be adapted to utilise said [service](../configuration/storage/index.md).

## Steps

- `git clone https://github.com/authelia/authelia.git`
- `cd authelia/compose/lite`
- Modify the `users_database.yml` the default username and password is `authelia`
- Modify the `configuration.yml` and `docker-compose.yml` with your respective domains and secrets
- `docker-compose up -d`

## Reverse Proxy

The [Lite bundle](https://github.com/authelia/authelia/blob/master/compose/lite/docker-compose.yml)
provides pre-made examples with [Traefik2.x](./supported-proxies/traefik2.x.md), you can swap this
out for any of the [supported proxies](./supported-proxies/index.md).

## FAQ

### Can you give more details on why this is not suitable for production environments?

This documentation gives instructions that will make **Authelia** non
resilient to failures and non scalable by preventing you from running multiple
instances of the application. This means that **Authelia** won't be able to distribute
the load across multiple servers and it will prevent failover in case of a
crash or an hardware issue.