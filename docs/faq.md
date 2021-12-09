---
layout: default
title: FAQ
nav_order: 8
---

# Frequently asked questions

## How can the backend be aware of the authenticated users?

This question is solved
[here](https://www.authelia.com/docs/deployment/supported-proxies/#how-can-the-backend-be-aware-of-the-authenticated-users).

## Why only use a private issuer key and no public key with OIDC?

The reason for using only the private key here is that one is able to calculate the public key easily from the private
key (`openssl rsa -in rsa.key -pubout > rsa.pem`).

## My installation broke after updating. What do I need to fix?

Check the [migration log](https://www.authelia.com/docs/configuration/migration.html) for any steps you need to follow. It's a good idea to consult this prior to running an update.
