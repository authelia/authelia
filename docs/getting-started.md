---
layout: default
title: Getting Started
nav_order: 2
---

# Getting Started

## Docker Compose

### Steps

These commands are intended to be run sequentially:

- `git clone https://github.com/authelia/authelia.git`
- `cd authelia/compose/local`
- `sudo ./setup.sh` *sudo is required to modify the `/etc/hosts` file*

You can now visit the following locations; replace example.com with the domain you specified in the setup script:
- https://public.example.com - Bypasses Authelia
- https://traefik.example.com - Secured with Authelia one-factor authentication
- https://secure.example.com - Secured with Authelia two-factor authentication (see note below)

You will need to authorize the self-signed certificate upon visiting each domain.
To visit https://secure.example.com you will need to register a device for second factor authentication and confirm by clicking on a link sent by email.
Since this is a demo with a fake email address, the content of the email will be stored in `./authelia/notification.txt`.
Upon registering, you can grab this link easily by running the following command: `grep -Eo '"https://.*" ' ./authelia/notification.txt`.

## Deployment

So you're convinced that Authelia is what you need. You can head to the deployment documentation [here](./deployment/index.md).
Some recipes have been crafted for helping with the bootstrap of your environment.
You can choose between a [lite](./deployment/deployment-lite.md) deployment which is deployment advised for a single server setup.
However, this setup just does not scale. If you want a full environment that can scale out, use the [HA](./deployment/deployment-ha.md) or [Kubernetes](./deployment/deployment-kubernetes.md) deployment documentation.