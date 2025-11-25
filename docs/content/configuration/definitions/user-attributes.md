---
title: "User Attributes"
description: "User Attributes Definitions Configuration"
summary: "Authelia allows configuring reusable user attribute definitions."
date: 2025-02-22T06:40:08+00:00
draft: false
images: []
weight: 199100
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The user attributes section allows you to define custom attributes for your users using Common Expression Language (CEL).
These attributes can be used at the current time to:

- Enhance [OpenID Connect 1.0 claims](../../integration/openid-connect/openid-connect-1.0-claims.md) with dynamic values

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
definitions:
  user_attributes:
    # Boolean attribute example
    is_admin:
      expression: '"admin" in groups'

    # String attribute example
    department:
      expression: 'groups[0]'

    # Number attribute example
    access_level:
      expression: '"admin" in groups ? 10 : 5'
```

## Options

This section describes the individual configuration options. Currently, these attribute definitions are used in the
[OpenID Connect 1.0 Provider](../identity-providers/openid-connect/provider.md#pol).

The key name is the name of the resulting attribute. It is important to note that this attribute name must not conflict
with extra attributes defined within the authentication backend, or with the common attributes we have defined.

In the above example the following attributes are added:

- `is_admin`
- `department`
- `access_level`

### expression

The [Common Expression Language](https://github.com/google/cel-spec) expression for this attribute.

## Contextual Attributes

{{< callout context="danger" title="Security Notice" icon="outline/alert-octagon" >}}
The `openid_authreq_claim_value` and `openid_authreq_claim_values` attributes should not be used in a security sensitive
context unless they are used in conjunction with either
[OAuth 2.0 JWT-Secured Authorization Requests (JAR)](https://www.rfc-editor.org/rfc/rfc9101.html) (with the use of
[JSON Web Encryption (JWE)](https://datatracker.ietf.org/doc/html/rfc7516) in the instance that an attacker having
knowledge of the value would present a security risk) or
[OAuth 2.0 Pushed Authorization Requests (PAR)](https://datatracker.ietf.org/doc/html/rfc9126). Both of these mechanisms
prevent the claims values from being altered by an attacker (specifically in the case of man-in-the-middle attacks and
compromised clients).
{{< /callout >}}

The following attributes are available for use in expressions depending on the context:

|           Attribute           |                     Description                      |                  Context                  |
|:-----------------------------:|:----------------------------------------------------:|:-----------------------------------------:|
| `openid_authreq_claim_value`  | The `value` property of the relevant claims request  | OpenID Connect 1.0 Authorization Request  |
| `openid_authreq_claim_values` | The `values` property of the relevant claims request | OpenID Connect 1.0 Authorization Request  |
