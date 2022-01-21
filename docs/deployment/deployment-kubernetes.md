---
layout: default
title: Deployment - Kubernetes
parent: Deployment
nav_order: 3
---

# Deployment on Kubernetes

<p>
    <img src="../images/logos/kubernetes.png" width="100" style="padding-right: 10px">
</p>

## UNDER CONSTRUCTION

The following areas are actively being worked on for Kubernetes:
1. Detailed Documentaiton
2. [Helm Chart (v3)](https://github.com/authelia/chartrepo)
3. Kustomize Deployment
4. Manifest Examples

Users are welcome to reach out directly by using any of our various [contact options](../about-us.md#contact-options). 

###  NGINX Ingress Controller 
If you use NGINX Ingress Controller you can protect an ingress with the following annotations.
The assumptions are that your public domain where authelia is running would be https://auth.mypublicdomain.com
and there would be a service called authelia with port 80 in the default namespace.

```yaml
annotations:
  nginx.ingress.kubernetes.io/auth-response-headers: Remote-User,Remote-Name,Remote-Groups,Remote-Email
  nginx.ingress.kubernetes.io/auth-signin: https://auth.mypublicdomain.com
  nginx.ingress.kubernetes.io/auth-snippet: |
    proxy_set_header X-Forwarded-Method $request_method;
  nginx.ingress.kubernetes.io/auth-url: http://authelia.default.svc.cluster.local/api/verify
  nginx.ingress.kubernetes.io/configuration-snippet: |
    proxy_set_header X-Forwarded-Method $request_method;
```

## FAQ

### RAM usage

If using file-based authentication, the argon2id provider will by default use 1GB of RAM for password generation. This means you should allow for at least this amount in your deployment/daemonset spec and have this much available on your node, alternatively you can [tweak the providers settings](https://www.authelia.com/docs/configuration/authentication/file.html#memory). Otherwise, your Authelia may OOM during login. See [here](https://github.com/authelia/authelia/issues/1234#issuecomment-663910799) for more info.
