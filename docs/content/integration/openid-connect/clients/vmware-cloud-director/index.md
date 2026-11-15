---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "VMware Cloud Director"
description: "A guide on integrating VMware Cloud Director with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "VMware Cloud Director | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring VMware Cloud Director with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)
- [VMware Cloud Director]
  - [v10.6.2](https://blogs.vmware.com/cloudprovider/2026/09/announcing-vmware-cloud-director-10-6-2-enabling-your-migration-to-vmware-cloud-foundation-9-1.html)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://vcd.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://vcd.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth?service=provider` for a provider-level
    integration, or `https://vcd.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth?service=tenant:<organization_name>`
    for a tenant-level integration. This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `vmware-cloud-director`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with
[VMware Cloud Director] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'vmware-cloud-director'
        client_name: 'VMware Cloud Director'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://vcd.{{< sitevar name="domain" nojs="example.com" >}}/login/oauth?service=provider'
        scopes:
          - 'openid'
          - 'email'
          - 'profile'
          - 'groups'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [VMware Cloud Director] there is one method, using the [Web GUI](#web-gui).

#### Web GUI

To configure [VMware Cloud Director] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following
instructions:

1. Login to [VMware Cloud Director] with a local administrator account.
2. Navigate to `Administration`, `Identity Providers`, `OIDC`.
3. Confirm the client configuration redirect URI matches the redirect URI configured in Authelia.
4. Click `Configure`.
5. Configure the following options:
   - Client ID: `vmware-cloud-director`
   - Client Secret: `insecure_secret`
   - Configuration Discovery: Enabled
   - IDP Well-known Configuration Endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/.well-known/openid-configuration`
6. Click `Next`.
7. On the `Endpoint` page confirm the values discovered from Authelia and click `Next`.
8. On the `Scopes` page configure the scopes `openid`, `profile`, `email`, and `groups` and click `Next`.
9. On the `Claim Mapping` page ensure the groups claim is mapped to `groups` and click `Next`.
10. On the `Key Configuration` page optionally enable `Automatic Key Refresh` and click `Next`.
11. On the `Button Label` page optionally customize the label shown on the login page, for example `Authelia`.
12. Click `Save`.

To authorize users based on their Authelia groups, navigate to `Administration`, `Provider Access Control`, `Groups`
for a provider-level integration or `Administration`, `Access Control`, `Groups` for a tenant-level integration, click
`Import Groups`, select `OIDC` as the source, enter the Authelia group names, and assign a role.

## See Also

- [VMware Cloud Director Configure Your System to Use an OpenID Connect Identity Provider Documentation](https://techdocs.broadcom.com/us/en/vmware-cis/cloud-director/vmware-cloud-director/10-6/vmware-cloud-director-service-provider-portal-guide-10-6.html)

[Authelia]: https://www.authelia.com
[VMware Cloud Director]: https://www.vmware.com/products/cloud-infrastructure/cloud-director
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
