---
layout: default
title: Deployment - Kubernetes
parent: Deployment
nav_order: 3
---

# Deployment on Kubernetes

<p>
    <img src="../images/logos/kubernetes.logo.png" width="100" style="padding-right: 10px">
</p>

## UNDER CONSTRUCTION

The following areas are actively being worked on for Kubernetes:
1. Detailed Documentaiton
2. [Helm Chart (v3)](https://github.com/authelia/chartrepo)
3. Kustomize Deployment
4. Manifest Examples

Users are welcome to reach out directly on our [Matrix Room](https://riot.im/app/#/room/#authelia:matrix.org) or 
[Discord Server](https://discord.authelia.com) if they are looking for help setting up on Kubernetes in the meantime. 

## FAQ

### RAM usage

If using file-based authentication, the argon2id provider will by default use 1GB of RAM for password generation. This means you should allow for at least this amount in your deployment/daemonset spec and have this much available on your node, alternatively you can [tweak the providers settings](https://www.authelia.com/docs/configuration/authentication/file.html#memory). Otherwise, your Authelia may OOM during login. See [here](https://github.com/authelia/authelia/issues/1234#issuecomment-663910799) for more info.
