---
title: "Security Sensitive Values"
description: "An introduction into configuring Authelia's security sensitive values."
summary: "An introduction into configuring Authelia's security sensitive values."
date: 2024-03-31T23:24:06+00:00
draft: false
images: []
weight: 100150
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia contains several security sensitive values which are documented as such and are also generally are named
`secret`, `key`, `password`, `token`, or `certificate_chain`; alternatively they may be suffixed with a `_` followed by one
of the previous values.

We generally recommend not leaving these values directly in the configuration itself, as this often leads to accidentally
leaking the values when getting support and is generally slightly less secure.

There are three special ways to achieve this goal:

1. Using the native [Secrets](../methods/secrets.md) system which:
   - Loads the value from a file provided an environment variable with the file's location.
   - Generally easy to set up.
   - Can't be used for keys located within lists.
   - Does not include the value in the environment which is slightly more secure.
2. Using the `template` [file filter](../methods/files.md#file-filters) system which:
   - Loads the value from a file provided a template within the configuration itself making it easy to share.
   - Generally easy to set up but has a small learning curve.
   - Can be used anywhere in the configuration generally for any purpose.
   - Does not include the value in the environment which is slightly more secure.
3. Using the native [Environment](../methods/environment.md) system which:
  - Loads the value from the environment variable itself
  - Generally easy to set up.
  - Can't be used keys located within lists.
  - Does include the value in the environment which is slightly less secure.


## Template Example

This explains option 2 in the context of using it specifically for secret values. For more information on templating
see the [Reference Guide](../../reference/guides/templating.md).

### Single-Line Value

This example shows how to do a single-line value. The single quotes are only relevant if the value is a string and can
be excluded for other value types.

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    hmac_secret: '{{ secret "/config/secrets/absolute/path/to/hmac_secret" }}'
```

Alternatively you can use the special `m` variants of the `indent` and `squote` functions to automatically adjust the
layout depending on if the file has multiple lines, [msquote] will automatically single quote the value if it's not
multiple lines, see [Multi-Line Value](#multi-line-value) for more information on [mindent].

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    hmac_secret: {{ secret "/config/secrets/absolute/path/to/hmac_secret" | mindent 10 "|" | msquote }}
```

### Multi-Line Value

This example shows how to do a multi-line value. QuotiThng is not possible in this scenario as such it's excluded.

It's important to note the use of [mindent]:

1. The value of `10` indicates the value should be indented with 10 spaces:
   - You will need to adjust the indent depending on the context. The indent should be an additional indentation level
     (in this case 2 spaces) than the start of the key name. In this case the `jwks` key name `key` is indented exactly
     8 characters, so the value `10` is correct.
2. The value of `|` indicates what multiline prefix to use.

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    jwks:
      - key: {{ secret "/config/secrets/absolute/path/to/jwks/rsa.2048.pem" | mindent 10 "|" | msquote }}
```

[mindent]: ../../reference/guides/templating.md#mindent
[msquote]: ../../reference/guides/templating.md#msquote
