---
title: "Ansible AWX and Ansible Tower"
description: "Integrating Ansible AWX and Ansible Tower with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2025-04-26T18:35:57+10:00
draft: false
images: []
weight: 620
toc: true
aliases: []
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Ansible AWX and Ansible Tower | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Ansible AWX and Ansible Tower with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.14](https://github.com/authelia/authelia/releases/tag/v4.39.14)
- [Ansible AWX]
  - [v24.6.1](https://github.com/ansible/awx/releases/tag/24.6.1)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://awx.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `awx`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Ansible AWX] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'awx'
        client_name: 'Ansible AWX'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://awx.{{< sitevar name="domain" nojs="example.com" >}}/sso/complete/oidc/'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Ansible AWX] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [Ansible AWX] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following instructions:

1. Login to [Ansible AWX] as an administrator.
2. Click `Settings` from the left navifation bar.
3. Click `Generic OIDC settings` on the left side of the settings window.
4. Click `Edit` and configure the following options:
   - OIDC Key: `awx`
   - OIDC Secret: `insecure_secret`
   - OIDC Provider URL: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Verify OIDC Provider Certificate: Enable
5. Click `Save`.

## See Also

- [Ansible AWX Setting up Enterprise Authentication Documentation](https://ansible.readthedocs.io/projects/awx/en/24.6.1/administration/ent_auth.html#generic-oidc-settings)

[Authelia]: https://www.authelia.com
[Ansible AWX]: https://github.com/ansible/awx
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
