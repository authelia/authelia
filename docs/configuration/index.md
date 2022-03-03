---
layout: default
title: Configuration
nav_order: 4
has_children: true
---

# Configuration
Authelia has several methods of configuration available to it. The order of precedence is as follows:

1. [Secrets](./secrets.md)
2. [Environment Variables](#environment)
3. [Files](#files) (in order of them being specified)

This order of precedence puts higher weight on things higher in the list. This means anything specified in the 
[files](#files) is overridden by [environment variables](#environment) if specified, and anything specified by 
[environment variables](#environment) is overridden by [secrets](./secrets.md) if specified.

## Files
When running **Authelia**, you can specify your configuration by passing the file path as shown below.

```console
$ authelia --config config.custom.yml
```

You can have multiple configuration files which will be merged in the order specified. If duplicate keys are specified 
the last one to be specified is the one that takes precedence. Example:

```console
$ authelia --config configuration.yml --config config-acl.yml --config config-other.yml
$ authelia --config configuration.yml,config-acl.yml,config-other.yml
```

Authelia's configuration files use the YAML format. A template with all possible options can be found at the root of the 
repository [here](https://github.com/authelia/authelia/blob/master/config.template.yml).

### Docker
By default, the container looks for a configuration file at `/config/configuration.yml`. This can be changed using the `command` setting.

## Environment
You may also provide the configuration by using environment variables. Environment variables are applied after the 
configuration file meaning anything specified as part of the environment overrides the configuration files. The 
environment variables must be prefixed with `AUTHELIA_`.

_**Please Note:** It is not possible to configure_ the _access control rules section or OpenID Connect identity provider
section using environment variables at this time._

_**Please Note:** There are compatability issues with Kubernetes and this particular configuration option. You must 
ensure you have the `enableServiceLinks: false` setting in your pod spec. You can read more about this in the 
[migration documentation](./migration.md#kubernetes-4300)._

Underscores replace indented configuration sections or subkeys. For example the following environment variables replace
the configuration snippet that follows it:

```
AUTHELIA_LOG_LEVEL=info
AUTHELIA_SERVER_READ_BUFFER_SIZE=4096
```

```yaml
log:
  level: info
server:
  read_buffer_size: 4096
```

# Documentation

We document the configuration in two ways:

1. The configuration yaml default has comments documenting it. All documentation lines start with `##`. Lines starting 
   with a single `#` are yaml configuration options which are commented to disable them or as examples.
    
2. This documentation site. Generally each section of the configuration is in its own section of the documentation 
   site. Each configuration option is listed in its relevant section as a heading, under that heading generally are two
   or three colored labels. 
   - The `type` label is purple and indicates the yaml value type of the variable. It optionally includes some 
     additional information in parentheses.
   - The `default` label is blue and indicates the default value if you don't define the option at all. This is not the 
     same value as you will see in the examples in all instances, it is the value set when blank or undefined.
   - The `required` label changes color. When required it will be red, when not required it will be green, when the 
     required state depends on another configuration value it is yellow.  

# Validation

Authelia validates the configuration when it starts. This process checks multiple factors including configuration keys
that don't exist, configuration keys that have changed, the values of the keys are valid, and that a configuration
key isn't supplied at the same time as a secret for the same configuration option.

You may also optionally validate your configuration against this validation process manually by using the validate-config
option with the Authelia binary as shown below. Keep in mind if you're using [secrets](./secrets.md) you will have to
manually provide these if you don't want to get certain validation errors (specifically requesting you provide one of
the secret values). You can choose to ignore them if you know what you're doing. This command is useful prior to
upgrading to prevent configuration changes from impacting downtime in an upgrade. This process does not validate
integrations, it only checks that your configuration syntax is valid.

```console
$ authelia validate-config --config configuration.yml
```

# Duration Notation Format

We have implemented a string/integer based notation for configuration options that take a duration of time. This section 
describes the implementation of this. You can use this implementation in various areas of configuration such as:

- session:
  - expiration
  - inactivity
  - remember_me_duration
- regulation:
  - ban_time
  - find_time
- ntp:
  - max_desync
- webauthn:
  - timeout

The way this format works is you can either configure an integer or a string in the specific configuration areas. If you
supply an integer, it is considered a representation of seconds. If you supply a string, it parses the string in blocks
of quantities and units (number followed by a unit letter).  For example `5h` indicates a quantity of 5 units of `h`.

While you can use multiple of these blocks in combination, ee suggest keeping it simple and use a single value.

## Duration Notation Format Unit Legend

|  Unit   | Associated Letter |
|:-------:|:-----------------:|
|  Years  |         y         |
| Months  |         M         |
|  Weeks  |         w         |
|  Days   |         d         |
|  Hours  |         h         |
| Minutes |         m         |
| Seconds |         s         |

## Duration Notation Format Examples

|     Desired Value     |        Configuration Examples         |
|:---------------------:|:-------------------------------------:|
| 1 hour and 30 minutes | `90m` or `1h30m` or `5400` or `5400s` |
|         1 day         | `1d` or `24h` or `86400` or `86400s`  |
|       10 hours        | `10h` or `600m` or `9h60m` or `36000` |

# TLS Configuration

Various sections of the configuration use a uniform configuration section called TLS. Notably LDAP and SMTP.
This section documents the usage.

## Server Name
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: ""
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The key `server_name` overrides the name checked against the certificate in the verification process. Useful if you
require to use a direct IP address for the address of the backend service but want to verify a specific SNI.

## Skip Verify
<div markdown="1">
type: boolean
{: .label .label-config .label-purple } 
default: false
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The key `skip_verify` completely negates validating the certificate of the backend service. This is not recommended,
instead you should tweak the `server_name` option, and the global option [certificates directory](./miscellaneous.md#certificates_directory).

## Minimum Version
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: TLS1.2
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>

The key `minimum_version` controls the minimum TLS version Authelia will use when opening TLS connections.
The possible values are `TLS1.3`, `TLS1.2`, `TLS1.1`, `TLS1.0`. Anything other than `TLS1.3` or `TLS1.2`
are very old and deprecated. You should avoid using these and upgrade your backend service instead of decreasing
this value.
