---
layout: default
title: Server
parent: Configuration
nav_order: 9
---

# Server

The server section configures and tunes the http server module Authelia uses.

## Configuration

```yaml
server:
  read_buffer_size: 4096
  write_buffer_size: 4096
  path: ""
  enable_pprof: false
  enable_expvars: false
  cors:
    disable: false
```

## Options

### read_buffer_size
<div markdown="1">
type: integer 
{: .label .label-config .label-purple } 
default: 4096
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Configures the maximum request size. The default of 4096 is generally sufficient for most use cases.

### write_buffer_size
<div markdown="1">
type: integer 
{: .label .label-config .label-purple } 
default: 4096
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Configures the maximum response size. The default of 4096 is generally sufficient for most use cases.

### path
<div markdown="1">
type: string 
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Authelia by default is served from the root `/` location, either via its own domain or subdomain.

Modifying this setting will allow you to serve Authelia out from a specified base path. Please note
that currently only a single level path is supported meaning slashes are not allowed, and only
alphanumeric characters are supported.

Example: https://auth.example.com/, https://example.com/
```yaml
server:
  path: ""
```

Example: https://auth.example.com/authelia/, https://example.com/authelia/
```yaml
server:
  path: authelia
```

### enable_pprof
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Enables the go pprof endpoints.

### enable_expvars
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Enables the go expvars endpoints.

### cors

#### enable
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Before enabling [CORS] you should read the MDN [CORS] documentation or have some specific instructions from an app that
requires it. The general rule is you should configure this as accurately as possible to your specific needs.

This enables the automatic handling of [CORS] headers. If enabled we add CORS headers to GET/OPTIONS requests on the root
path when the [Origin] header has a valid value. It is determined as valid if all of the following conditions are true:

- the scheme of the origin is `https`
- the domain of the origin is a subdomain of the value set in session domain

The following headers will be set on these requests:

|Header                            |Value                                                   |
|:--------------------------------:|:------------------------------------------------------:|
|[Vary]                            |Accept-Encoding, [Origin]                               |
|[Access-Control-Allow-Origin]     |<value of [Origin] header>                              |
|[Access-Control-Allow-Credentials]|false                                                   |
|[Access-Control-Allow-Headers]    |<value of [Access-Control-Request-Headers] header>      |
|[Access-Control-Allow-Methods]    |<value of [Access-Control-Request-Method] header or GET>|
|[Access-Control-Max-Age]          |100                                                     |

#### include_protected
<div markdown="1">
type: boolean
{: .label .label-config .label-purple }
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

If you override [origins](#origins) and you want to also automatically include any domain that is a subdomain of the 
session domain, you can achieve this by enabling this option. This is particularly useful if you want to allow origins
that are not protected but also keep all protected domains allowed.

#### origins
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>

Overrides the default behaviour for validating the [Origin] header. Instead of checking the domain is protected by
Authelia it checks this list of URLs (which must all start with `https`).

You can alternatively set this to exactly `*` and it will allow all origins, however this not recommended for anyone
without a deep understanding of [CORS].

MDN Header Documentation: [Access-Control-Allow-Origin].

#### headers
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>

Overrides the default behaviour of automatically allowing the headers requested by the CORS preflight.

MDN Header Documentation: [Access-Control-Allow-Headers].

#### methods
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple } 
required: no
{: .label .label-config .label-green }
</div>

Overrides the default behaviour of automatically allowing the methods requested by the CORS preflight. Is a list of valid
HTTP request methods, or can be a single item `*` which allows all methods.

MDN Header Documentation: [Access-Control-Allow-Methods].

#### vary
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple }
default: Accept-Encoding, Origin
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

A list of headers to be included in the Vary header.

MDN Header Documentation: [Vary](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Vary).

#### max_age
<div markdown="1">
type: int
{: .label .label-config .label-purple }
default: 100
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The value of the [Access-Control-Max-Age](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age) 
header. Setting this to -1 disables it.

MDN Header Documentation: [Access-Control-Max-Age](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age).

## Additional Notes

### Buffer Sizes

The read and write buffer sizes generally should be the same. This is because when Authelia verifies
if the user is authorized to visit a URL, it also sends back nearly the same size response as the request. However
you're able to tune these individually depending on your needs.

[CORS]: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
[Vary]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Vary
[Origin]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Origin
[Access-Control-Allow-Origin]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
[Access-Control-Allow-Credentials]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
[Access-Control-Allow-Headers]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
[Access-Control-Allow-Methods]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
[Access-Control-Max-Age]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age
[Access-Control-Request-Headers]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Request-Headers
[Access-Control-Request-Method]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Request-Method
