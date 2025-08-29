---
title: "Securing Applications with Basic Auth"
description: "Learn how to protect applications using Authelia's ForwardAuth while enabling API access through Basic Authentication and service accounts."
summary: "A comprehensive guide for adding authentication to unprotected applications using Authelia and Traefik, with support for both web users and programmatic API access via service accounts."
date: 2025-08-21T14:44:34-07:00
draft: false
images: []
weight: 550
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

# Introduction
We are going to discuss how to protect backend applications with [Authelia], while still allowing programmatic access to the backend APIs using basic authentication. This pattern is essential when you need both human users (who can use Authelia's web interface) and automated systems (like monitoring tools, CI/CD pipelines, or other services) to access the same protected resources.

# Assumptions
This guide makes the following assumptions:

- [Authelia] is already setup and running with [Traefik] as its proxy. It should be noted that while this guide explicitly uses [Traefik] as its proxy, you can achieve the same end state with other [supported proxies].
- The backend application you want to protect has no built-in authentication (or authentication is disabled).

## Core Concepts
There are some concepts that are central to this guide which we will explain here.

### ForwardAuth
{{< figure src="authforward.webp" caption="Traefik Forward Auth" alt="Flow chart illustrating how AuthForward handles HTTP request authentication" process="resize 650x" >}}

[ForwardAuth](https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/forwardauth/) is a way to allow a proxy ([Traefik]) to delegate authorization to an external service. When a client requests a resource protected by a forward auth middleware, Traefik forwards headers and connection information about the initial request to the auth server.

There are two possible responses for the auth server:
- OK: the initial request continues to the resource server (backend application).
- KO: the initial request is blocked, given a redirect, or a www-authenticate response.

This allows you to centralize authentication logic in Authelia rather than relying on the application's implementation or lack thereof.

### Basic Authentication

[Basic authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Authentication#basic_authentication_scheme) transmits user credentials as a base64 encoded string in the format `username:password` via the `Authorization: Basic <encoded-string>` HTTP header.

### Service Accounts
Service Accounts are non-human users designed to enable programmatic access to protected resources. Unlike regular user accounts, they typically use long-lived credentials and don't require interactive authentication flows (like [OpenID Connect]).

In our case, Authelia treats service accounts the same as regular users - it doesn't distinguish between them. However, since our default access control policy is `deny`, service accounts will only have access to applications that are explicitly granted through ACL rules. This is important from a security perspective because anyone with service account credentials could potentially log into the Authelia web portal and access any resources that the service account is authorized for.

**Best Practice:** Grant service accounts the minimum permissions necessary and consider using dedicated service accounts for each application or use case to limit potential exposure.

## Architecture

### Traefik Router, Service, Middleware

This configuration sets up Traefik to route requests to your application while protecting it with Authelia's ForwardAuth middleware. The router defines which domain should be protected (`myapp.example.com`), the service points to your backend application, and the middleware configuration tells Traefik to validate all requests through Authelia before allowing access to the application.

The `authResponseHeaders` are important but *optional* - they allow Authelia to pass user information (like username, groups, and email) to your backend application, which can be useful for logging or user-specific functionality. See [Trusted Header SSO](../../trusted-header-sso/introduction) for more info.

```yaml{title="/dynamic/myapp.yaml"}
http:
  routers:
    myapp-router:
      rule: 'Host(`myapp.example.com`)'
      entrypoints:
        - https
      middlewares:
      - authelia@file
      service: myapp-service
  services:
    myapp-service:
      loadBalancer:
        servers:
          - url: 'http://myapp:80/'
  middlewares:
    authelia:
      forwardAuth:
        address: 'https://authelia:9091/api/authz/forward-auth'
        trustForwardHeader: true
        authResponseHeaders:
          - 'Remote-User'
          - 'Remote-Groups'
          - 'Remote-Email'
          - 'Remote-Name'
```

### Service Account

The service account is configured just like a regular user in Authelia's user database, but with specific groups that identify it as a service account. The key differences are:

- Groups: Contains both `myapp` (for application access) and `service` (to identify it as a service account).
- Email: Should be a monitoring/catchall email rather than a personal one.
- Password: Use a [strong, randomly generated password](https://www.authelia.com/reference/guides/generating-secure-values/#generating-a-random-password-hash) (>64 characters) since this will be used programmatically and not be protected by multiple factor authentication.

Human users like `John` only have the `myapp` group (and not `service`), which means they'll be subject to different authentication requirements.

```yaml{title="users.yaml"}
users:
  my-service-account:
    displayname: 'My Service Account'
    password: '$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM' # digest for 'password'
    email: 'my-service-account@example.com'
    groups:
      - 'myapp'
      - 'service'
  john:
    displayname: 'John Doe'
    password: '$argon2id$v=19$m=65536,t=3,p=2$BpLnfgDsc2WD8F2q$o/vzA4myCqZZ36bUGsDY//8mKUYNZZaR0t4MFFSs+iM' # digest for 'password'
    email: 'john.doe@authelia.com'
    groups:
      - 'myapp'
```

### Authelia ACL

These access control rules create different authentication requirements based on group membership:

- First rule: Users with BOTH `myapp` AND `service` groups (service accounts) only need one-factor authentication, which allows basic auth to work.
- Second rule: Users with only the `myapp` group (human users) require two-factor authentication, forcing them to use the web interface to complete the second factor authentication.

The order of these rules matter. Authelia has two important concepts for rule matching:
- [Sequential Order](https://www.authelia.com/configuration/security/access-control/#rule-matching-concept-1-sequential-order): Rules are evaluated sequentially, top to bottom, and the first matching rule will be applied.
- [Subject Criteria Requires Authentication](../../../configuration/security/access-control.md/#rule-matching-concept-2-subject-criteria-requires-authentication): Authelia can't determine groups or username without the user being authenticated.

```yaml{title=configuration.yaml}
access_control:
  default_policy: deny
  rules:
    - domain:
        - myapp.example.com
    policy: one_factor
    subject:
      - ['group:myapp', 'group:service']

    - domain:
        - myapp.example.com
      policy: two_factor
      subject:
        - 'group:myapp'
```

### Protected Application

This represents your backend application that has no built-in authentication. It could be an API server, web application, or any service that you want to protect. The application doesn't need to be modified - Authelia handles all authentication logic, and the application receives requests only after they've been validated.

If your application needs to know who is accessing it, it can read the [headers that Authelia forwards](https://www.authelia.com/integration/trusted-header-sso/introduction/#response-headers) (like `Remote-User` and `Remote-Groups`) to implement user-specific behavior or logging.

## How It Works

1. **Human users** accessing `myapp.example.com` via a browser are redirected to the Authelia portal and must complete two-factor authentication to get access.
2. **Service accounts** can bypass the web interface by including the `Authorization: Basic <base64(username:password)>` header with their credentials.
3. Both access methods are validated by Authelia, but different ACL rules are applied based on group membership.
4. Once authenticated, the requests are forwarded to the backend application with additional information in the [headers](../../trusted-header-sso/introduction.md/#response-headers).


## Verification

### Web Browser Access
1. Navigate to `https://myapp.example.com` in your browser.
2. You should be redirected to Authelia.
3. Login as John (or your other user) and complete two-factor authentication.
4. You should be redirected back to your application.

### API Access

```bash
curl -u "my-service-account:password" https://myapp.example.com/
```

## Common Use Cases

### Monitoring and Health Checking
Applications such as [Uptime Kuma] can make use of push-based health checking, ie. an application sends an api request periodically with its current status. If [Uptime Kuma] is behind Authelia (with authentication disabled), you can allow those push api requests using a service account. See [Path Bypass](#path-bypass) for how to bypass specific paths only.

### Log Shipping
Applications such as [Loki](https://grafana.com/docs/loki/latest/), [Mimir](https://grafana.com/docs/mimir/latest/), [Tempo](https://grafana.com/docs/tempo/latest/), and other log/metric/trace aggregators may not include their own authentication which makes this solution very useful. By configuring collection agents to use basic auth, they can ship logs or metrics to a protected application.

## Advanced Configs

### Path Bypass
In the case where you want to bypass certain api paths for service accounts (rather than the entire api), you can achieve this with access_control [resources](../../../configuration/security/access-control.md#resources).

The following example allows the `myapp` service account access to `myapp.example.com/api/push/*` without allowing it to access the entirety of the myapp api.
```yaml{title=configuration.yaml}
access_control:
  default_policy: deny
  rules:
    - domain:
        - myapp.example.com
    policy: one_factor
    resources:
      - '^/api/push([/?].*)?$'
    subject:
      - ['group:myapp', 'group:service']
```

### Network-Based Access Control
You can also restrict source ip addresses that service accounts can be used from using the `networks` option in access control. [See Here](../../../configuration/security/access-control.md#networks)

## Security Considerations
When implementing service accounts for use with Authelia, there are several security practices that should be kept in mind:

- Use a default deny policy wherever possible, this ensures service accounts only access what is explicitly granted.
- Grant only the minimum access to service accounts.
- Since service accounts bypass multi-factor authentication, password strength is crucial. Service account passwords should be at least 64 characters in length and randomly generated.
- Avoid using the same service account in multiple locations, use multiple service accounts, that way you can better determine which accounts/machines were compromised and avoid rotating credentials in many places at once.
- Restrict the IP addresses service accounts can be used from whenever possible.
- Periodically review service accounts in use and remove any unused ones.

[Authelia]:https://authelia.com/
[Traefik]: https://doc.traefik.io/traefik/
[supported proxies]: ../../proxies/support
[OpenID Connect]: https://openid.net/developers/how-connect-works/
[Uptime Kuma]: https://github.com/louislam/uptime-kuma
