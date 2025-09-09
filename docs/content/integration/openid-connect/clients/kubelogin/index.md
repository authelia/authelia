---
title: "Kube Login"
description: "Integrating Kube Login with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-07-11T12:25:57+00:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/kubelogin/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Kube Login | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Kube Login with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.9](https://github.com/authelia/authelia/releases/tag/v4.39.9)
- [Kube Login]
  - [v1.33.0](https://github.com/int128/kubelogin/releases/tag/v1.33.0)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `kube_login`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Kube Login] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'kube_login'
        client_name: 'Kubernetes Cluster Access'
        client_secret: 'insecure_secret'
        public: false
        authorization_policy: 'two_factor'
        redirect_uris:
          - 'http://localhost:8000'
          - 'http://localhost:18000'
        scopes:
          - 'openid'
          - 'groups'
          - 'email'
          - 'profile'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

{{< callout context="note" title="Token Authentication" icon="outline/info-circle" >}}
Kubernetes uses OIDC ID tokens (JWTs) for user authentication. While Kube Login supports access tokens (opaque) per the OAuth2 specification, Kubernetes has minimal support for this method.
{{< /callout >}}

### Kubernetes API Server Configuration

Configure your Kubernetes API server to trust Authelia as an OIDC provider by adding these arguments:

```bash
--oidc-issuer-url=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
--oidc-client-id=kube_login
--oidc-groups-claim=groups
```

See the [Kubernetes Flags Documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#using-flags) for more information on these options.

#### How to Apply These Arguments
The method for configuring API server arguments varies by Kubernetes distribution. Consult the [Kubernetes OIDC] Authentication documentation for detailed instructions on applying these arguments to your specific setup.

**Common distributions:**
- K3s: Add to `/etc/rancher/k3s/config.yaml` under `kube-apiserver-arg:`
- kubeadm: Edit `/etc/kubernetes/manifests/kube-apiserver.yaml`
- Managed services: Use provider-specific tools (AWS CLI, gcloud, az cli)

### RBAC Configuration

After configuring OIDC authentication, create RBAC rules to authorize your users. Choose the approach that fits your needs:

#### Group-Based Access

```yaml {title="group-rbac.yaml"}
# Admins group - full cluster access
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: 'authelia-admins'
subjects:
- kind: Group
  name: 'admins'
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: 'cluster-admin' # NOTE this role gives COMPLETE access to the kubernetes api
  apiGroup: rbac.authorization.k8s.io

---
# Developers group - namespace-specific access
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: 'authelia-developers'
  namespace: development
subjects:
- kind: Group
  name: 'developers'
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: 'edit'
  apiGroup: rbac.authorization.k8s.io
```

#### Per User Access

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: 'authelia-user-admin'
subjects:
- kind: User
  name: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}#your-user-sub-claim'
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: 'cluster-admin' # NOTE this role gives COMPLETE access to the kubernetes api
  apiGroup: rbac.authorization.k8s.io
```

**Note:** You can obtain all user `sub` identifiers using the following command: `authelia storage user identifiers export`

### Client Configuration (kubectl + kubelogin)

#### Install Required Tools

1. **Install kubectl** (if not already installed) - [Installation Guide](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
2. **Install krew** (kubectl plugin manager): - [Installation Guide](https://krew.sigs.k8s.io/docs/user-guide/setup/install/)
3. **Install kubelogin** - `kubectl krew install oidc-login`

#### Configure kubeconfig

Use kubectl commands to set up the OIDC user on your local machine:

```bash
kubectl config set-credentials authelia \
  --exec-api-version=client.authentication.k8s.io/v1beta1 \
  --exec-command=kubectl \
  --exec-arg=oidc-login \
  --exec-arg=get-token \
  --exec-arg=--oidc-issuer-url=https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}} \
  --exec-arg=--oidc-client-id=kube_login \
  --exec-arg=--oidc-client-secret=insecure_secret \
  --exec-arg=--oidc-extra-scope=groups
```

#### Setup Context
Create and use a context with the OIDC user:
```bash
# Create context (replace 'your-cluster' with your actual cluster name)
kubectl config set-context authelia \
  --cluster=your-cluster \
  --user=authelia

# Switch to the new context
kubectl config use-context authelia
```

#### Testing the Configuration
```bash
# This should start the OIDC authentication flow in your browser.
kubectl get nodes
```


## See Also

- [Kube Login Usage Documentation](https://github.com/int128/kubelogin/blob/master/docs/usage.md)

[Authelia]: https://www.authelia.com
[Kube Login]: https://github.com/int128/kubelogin
[Kubernetes]: https://kubernetes.io/
[Kubernetes OIDC]: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
