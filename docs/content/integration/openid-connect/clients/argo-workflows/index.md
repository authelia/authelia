---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Argo Workflows"
description: "A guide on integrating Argo Workflows with the Authelia OpenID Connect 1.0 Provider for single sign-on (SSO) with configuration examples and available options."
summary: ""
date: 2026-11-12T00:00:00+00:00
draft: false
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Argo Workflows | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Argo Workflows with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [Argo Workflows]
  - [v4.1.6](https://github.com/argoproj/argo-workflows/releases/tag/v4.1.6)

{{% oidc-common bugs="claims-hydration" %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://argo-workflows.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://argo-workflows.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `argo-workflows`
- **Client Secret:** `insecure_secret`
- **Namespace:** `argo`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with [Argo Workflows]
which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'argo-workflows'
        client_name: 'Argo Workflows'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://argo-workflows.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

#### Configuration Escape Hatch

{{% oidc-escape-hatch-claims-hydration client_id="argo-workflows" claims="email,email_verified,preferred_username,name,groups" %}}

### Application

To configure [Argo Workflows] there is one method, using the [Configuration File](#configuration-file).

#### Configuration File

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `workflow-controller-configmap.yaml`.
{{< /callout >}}

To configure [Argo Workflows] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Create the [Kubernetes] secrets which contain the client credentials:

   ```shell
   kubectl create secret -n argo generic client-id-secret --from-literal=client-id-key=argo-workflows
   kubectl create secret -n argo generic client-secret-secret --from-literal=client-secret-key=insecure_secret
   ```

2. Add the following `sso` configuration to the workflow controller configmap:

   ```yaml {title="workflow-controller-configmap.yaml"}
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: workflow-controller-configmap
     namespace: argo
   data:
     sso: |
       issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
       clientId:
         name: 'client-id-secret'
         key: 'client-id-key'
       clientSecret:
         name: 'client-secret-secret'
         key: 'client-secret-key'
       redirectUrl: 'https://argo-workflows.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/callback'
       scopes:
         - 'profile'
         - 'email'
         - 'groups'
       rbac:
         enabled: false
   ```

3. Start the Argo Server with the `sso` auth mode, for example `argo server --auth-mode sso`.

## See Also

- [Argo Workflows Argo Server SSO Documentation](https://argo-workflows.readthedocs.io/en/latest/argo-server-sso/)
- [Argo Workflows Workflow Controller ConfigMap Documentation](https://argo-workflows.readthedocs.io/en/latest/workflow-controller-configmap/)

[Authelia]: https://www.authelia.com
[Argo Workflows]: https://argoproj.github.io/workflows/
[Kubernetes]: https://kubernetes.io/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
