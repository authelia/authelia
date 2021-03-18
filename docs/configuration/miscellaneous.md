---
layout: default
title: Miscellaneous
parent: Configuration
nav_order: 3
---

# Miscellaneous

Here are the main customizable options in Authelia.

## Host & Port

```yaml
host: 0.0.0.0
port: 9091
```

### host

Defines the address to listen on. See also [port](#port). Should typically be `0.0.0.0` or `127.0.0.1`, the former for 
containerized environments and the later for daemonized environments like init.d and systemd.

Note: If utilising an IPv6 literal address it must be enclosed by square brackets and quoted:

```yaml
host: "[fd00:1111:2222:3333::1]"
```

### port

`optional: true`

Defines the port to listen on. See also [host](#host).

## TLS

Authelia's port typically listens for plain unencrypted connections. This is by design as most environments allow to 
security on lower areas of the OSI model. However it required, if you specify both of the tls options the port will 
listen for TLS connections.

```yaml
tls_key: /config/ssl/key.pem
tls_cert: /config/ssl/cert.pem
```

### tls_key

The path to the private key for TLS connections. Must be in DER base64/PEM format. 

### tls_cert

The path to the public certificate for TLS connections. Must be in DER base64/PEM format.

## certificates_directory

This option defines the location of additional certificates to load into the trust chain specifically for Authelia.
This currently affects both the SMTP notifier and the LDAP authentication backend. The certificates should all be in the
PEM format and end with the extension `.pem`, `.crt`, or `.cer`. You can either add the individual certificates public 
key or the CA public key which signed them (don't add the private key).

```yaml
certificates_directory: /config/certs/
```

## Logging

### log_level

Defines the level of logs used by Authelia. This level can be set to `trace`, `debug` or `info`. When setting log_level 
to `trace`, you will generate a large amount of log entries and expose the `/debug/vars` and `/debug/pprof/` endpoints 
which should not be enabled in production.

```yaml
log_level: debug
```

### log_format

Defines the format of the logs written by Authelia. This format can be set to `json` or `text`.

```yaml
log_format: json
```

#### JSON format
```
{"level":"info","msg":"Logging severity set to info","time":"2020-01-01T00:00:00+11:00"}
{"level":"info","msg":"Authelia is listening for non-TLS connections on 0.0.0.0:9091","time":"2020-01-01T00:00:00+11:00"}
```
#### Text format
```
time="2020-01-01T00:00:00+11:00" level=info msg="Logging severity set to info"
time="2020-01-01T00:00:00+11:00" level=info msg="Authelia is listening for non-TLS connections on 0.0.0.0:9091"
```

### log_file_path

Logs can be stored in a file when file path is provided. Otherwise logs are written to standard output. When setting the 
log_level to `debug` or `trace` this will generate large amount of log entries. Administrators will need to ensure that 
they rotate and/or truncate the logs over time to prevent significant long-term disk usage.

```yaml
log_file_path: /config/authelia.log
```

## jwt_secret

`optional: false`

Defines the secret used to craft JWT tokens leveraged by the identity
verification process. This can also be defined using a [secret](./secrets.md).

```yaml
jwt_secret: v3ry_important_s3cr3t
```

## default_redirection_url

The default redirection URL is the URL where users are redirected when Authelia cannot detect the target URL where the 
user was heading.

In a normal authentication workflow, a user tries to access a website and she gets redirected to the sign-in portal in 
order to authenticate. Since the user initially targeted a website, the portal knows where the user was heading and
can redirect her after the authentication process. However, when a user visits the sign in portal directly, the portal 
considers the targeted website is the portal. In that case and if the default redirection URL is configured, the user is 
redirected to that URL. If not defined, the user is not redirected after authentication.

```yaml
default_redirection_url: https://home.example.com:8080/
```