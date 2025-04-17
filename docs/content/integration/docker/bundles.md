---
title: "Bundles"
description: "A guide on installing Authelia in Docker using curated docker compose bundles."
summary: "This helps get started with docker compose "
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 420
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

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
   [generate a new password](../../reference/guides/passwords.md#passwords), or both. The default password is
   `authelia`.
4. Edit the `configuration.yml` and `compose.yml` with your respective domains and secrets.
5. Edit the `configuration.yml` to configure the [SMTP Server](../../configuration/notifications/smtp.md).
6. Run `docker compose up -d`.

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
