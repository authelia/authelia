---
layout: default
title: Server
parent: Configuration
nav_order: 13
---

# Server

The server section configures and tunes the http server module Authelia uses.

## Configuration

```yaml
server:
  host: 0.0.0.0
  port: 9091
  path: ""
  read_buffer_size: 4096
  write_buffer_size: 4096
  enable_pprof: false
  enable_expvars: false
  disable_healthcheck: false
  tls:
    key: ""
    certificate: ""
    client_certificates: []
  headers:
    csp_template: ""
```

## Options

## host
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: 0.0.0.0
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Defines the address to listen on. See also [port](#port). Should typically be `0.0.0.0` or `127.0.0.1`, the former for
containerized environments and the later for daemonized environments like init.d and systemd.

Note: If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:

```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port
<div markdown="1">
type: integer
{: .label .label-config .label-purple } 
default: 9091
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Defines the port to listen on. See also [host](#host).

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

### asset_path
<div markdown="1">
type: string 
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

Authelia by default serves all static assets from an embedded filesystem in the Go binary.

Modifying this setting will allow you to override and serve specific assets for Authelia from a specified path.
All files that can be overridden are documented below and must be placed in the `asset_path` with a flat file structure.

Example:
```console
/config/assets/
├── favicon.ico
├── logo.png
└── locales/<lang>[-[variant]]/<namespace>.json
```

|  Asset  |   File name   |
|:-------:|:-------------:|
| Favicon |  favicon.ico  |
|  Logo   |   logo.png    |
| locales | see [locales] |

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

### disable_healthcheck
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

On startup Authelia checks for the existence of /app/healthcheck.sh and /app/.healthcheck.env and if both of these exist
it writes the configuration vars for the healthcheck to the /app/.healthcheck.env file. In instances where this is not
desirable it's possible to disable these interactions entirely.

An example situation where this is the case is in Kubernetes when set security policies that prevent writing to the
ephemeral storage of a container or just don't want to enable the internal health check.

### tls

Authelia typically listens for plain unencrypted connections. This is by design as most environments allow to
security on lower areas of the OSI model. However it required, if you specify both the [tls key](#key) and 
[tls certificate](#certificate) options, Authelia will listen for TLS connections.

#### key
<div markdown="1">
type: string (path)
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: situational
{: .label .label-config .label-yellow }
</div>

The path to the private key for TLS connections. Must be in DER base64/PEM format.

#### certificate
<div markdown="1">
type: string (path)
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: situational
{: .label .label-config .label-yellow }
</div>

The path to the public certificate for TLS connections. Must be in DER base64/PEM format.

#### client_certificates
<div markdown="1">
type: list(string)
{: .label .label-config .label-purple }
default: []
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-yellow }
</div>

The list of file paths to certificates used for authenticating clients. Those certificates can be root
or intermediate certificates. If no item is provided mutual TLS is disabled.


### headers

#### csp_template
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

This customizes the value of the Content-Security-Policy header. It will replace all instances of `${NONCE}` with the 
nonce value of the Authelia react bundle. This is an advanced option to customize and you should do sufficient research 
about how browsers utilize and understand this header before attempting to customize it.

For example, the default CSP template is `default-src 'self'; object-src 'none'; style-src 'self' 'nonce-${NONCE}'`.

## Additional Notes

### Buffer Sizes

The read and write buffer sizes generally should be the same. This is because when Authelia verifies
if the user is authorized to visit a URL, it also sends back nearly the same size response as the request. However
you're able to tune these individually depending on your needs.

### Asset Overrides

If replacing the Logo for your Authelia portal it is recommended to upload a transparent PNG of your desired logo.
Authelia will automatically resize the logo to an appropriate size to present in the frontend.

#### locales

The locales folder holds folders of internationalization locales. This folder can be utilized to override these locales.
They are the names of locales that are returned by the `navigator.langauge` ECMAScript command. These are generally
those in the [RFC5646 / BCP47 Format](https://datatracker.ietf.org/doc/html/rfc5646) specifically the language codes
from [Crowdin](https://support.crowdin.com/api/language-codes/).

Each directory has json files which you can explore the format of in the
[internal/server/locales](https://github.com/authelia/authelia/tree/master/internal/server/locales) directory on
GitHub. The important part is the key names you wish to override. Each file represents a translation namespace. The list
of current namespaces are below:

| Namespace |       Purpose       |
|:---------:|:-------------------:|
|  portal   | Portal translations |

A full example for the `en-US` locale for the portal namespace is `locales/en-US/portal.json`.

Languages in browsers are supported in two forms. In their language only form such as `en` for English, and in their
variant form such as `en-AU` for English (Australian). If a user has the browser language `en-AU` we automatically load
the `en` and `en-AU` languages, where any keys in the `en-AU` language take precedence over the `en` language, and the
translations for the `en` language only applying when a translation from `en-AU` is not available.

List of supported languages and variants:

| Description | Language | Additional Variants |        Location        |
|:-----------:|:--------:|:-------------------:|:----------------------:|
|   English   |    en    |         N/A         | locales/en/portal.json |
|   Spanish   |    es    |         N/A         | locales/es/portal.json |
|   German    |    de    |         N/A         | locales/de/portal.json |

_**Important Note** Currently users can only override languages that already exist in this list either by overriding
the language itself, or adding a variant form of that language. If you'd like support for another language feel free
to make a PR. We also encourage people to make PR's for variants where the difference in the variants is important._

_**Important Note** Overriding these files will not guarantee any form of stability. Users who planning to utilize these
overrides should either check for changes to the files in the 
[en](https://github.com/authelia/authelia/tree/master/internal/server/locales/en) translation prior to upgrading or PR 
their translation to ensure it is maintained._