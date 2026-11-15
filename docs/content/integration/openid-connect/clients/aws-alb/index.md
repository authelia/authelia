---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "AWS Application Load Balancer"
description: "A guide on integrating AWS Application Load Balancer with the Authelia OpenID Connect 1.0 Provider for SSO, with configuration examples and available options."
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
  title: "AWS Application Load Balancer | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring AWS Application Load Balancer with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [v4.39.28](https://github.com/authelia/authelia/releases/tag/v4.39.28)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- **Application Root URL:** `https://app.{{< sitevar name="domain" nojs="example.com" >}}/`
  - This option determines the redirect URI in the format of
    `https://app.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/idpresponse`.
    This means if you change this value, you need to update the redirect URI.
- **Authelia Root URL:** `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- **Client ID:** `aws-alb`
- **Client Secret:** `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example **Authelia** [client configuration] for use with
[AWS Application Load Balancer] which will operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'aws-alb'
        client_name: 'AWS Application Load Balancer'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        application_type: 'web'
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://app.{{< sitevar name="domain" nojs="example.com" >}}/oauth2/idpresponse'
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
        token_endpoint_auth_method: 'client_secret_basic'
```

### Application

To configure [AWS Application Load Balancer] there are two methods, using the [Web GUI](#web-gui) or using the
[Configuration File](#configuration-file).

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The `authenticate-oidc` action is only supported on HTTPS listeners. The [AWS Application Load Balancer] must be able to
reach the Authelia token and user info endpoints over IPv4, the Authelia domain must be publicly resolvable, and the
certificate presented by Authelia must be issued by a publicly trusted certificate authority. The load balancer also
maintains its own session which is not ended when the user logs out of Authelia.
{{< /callout >}}

#### Web GUI

To configure [AWS Application Load Balancer] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the
following instructions:

1. Login to the [Amazon EC2 Console](https://console.aws.amazon.com/ec2/).
2. Navigate to `Load Balancers` and select the load balancer.
3. On the `Listeners and rules` tab select the HTTPS listener.
4. On the `Rules` tab click `Add rule`, or select and edit an existing rule.
5. Add the conditions that should match requests for the application, for example a `Host header` condition with the
   value `app.{{< sitevar name="domain" nojs="example.com" >}}`.
6. For the `Pre-routing action` select `Authenticate user` and select `OIDC` as the identity provider.
7. Configure the following options:
   - Issuer: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}`
   - Authorization endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization`
   - Token endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token`
   - User info endpoint: `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo`
   - Client ID: `aws-alb`
   - Client secret: `insecure_secret`
   - Scope: `openid email profile`
8. For the `Routing action` select `Forward to target groups` and select the target group of the application.
9. Click `Next`, enter a `Priority`, and click `Next` again.
10. Click `Create`.

#### Configuration File

To configure [AWS Application Load Balancer] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the
following [AWS CloudFormation](https://aws.amazon.com/cloudformation/) listener rule resource, where `HttpsListener`
and `AppTargetGroup` refer to an existing HTTPS listener and target group in the same template:

```yaml {title="template.yml"}
Resources:
  AutheliaListenerRule:
    Type: 'AWS::ElasticLoadBalancingV2::ListenerRule'
    Properties:
      ListenerArn: !Ref 'HttpsListener'
      Priority: 10
      Conditions:
        - Field: 'host-header'
          Values:
            - 'app.{{< sitevar name="domain" nojs="example.com" >}}'
      Actions:
        - Type: 'authenticate-oidc'
          Order: 1
          AuthenticateOidcConfig:
            Issuer: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}'
            AuthorizationEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization'
            TokenEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token'
            UserInfoEndpoint: 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo'
            ClientId: 'aws-alb'
            ClientSecret: 'insecure_secret'
            Scope: 'openid email profile'
        - Type: 'forward'
          Order: 2
          TargetGroupArn: !Ref 'AppTargetGroup'
```

After a user authenticates the [AWS Application Load Balancer] forwards the claims from the Authelia user info endpoint
to the application in the `x-amzn-oidc-data` header as a JWT signed by the load balancer. Applications should verify
this JWT as described in the AWS documentation before trusting any of its claims.

## See Also

- [AWS Authenticate Users Using an Application Load Balancer Documentation](https://docs.aws.amazon.com/elasticloadbalancing/latest/application/listener-authenticate-users.html)
- [AWS CloudFormation AuthenticateOidcConfig Documentation](https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-properties-elasticloadbalancingv2-listenerrule-authenticateoidcconfig.html)

[Authelia]: https://www.authelia.com
[AWS Application Load Balancer]: https://aws.amazon.com/elasticloadbalancing/application-load-balancer/
[OpenID Connect 1.0]: ../../introduction.md
[client configuration]: ../../../../configuration/identity-providers/openid-connect/clients.md
