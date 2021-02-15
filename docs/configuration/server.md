---
layout: default
title: Server
parent: Configuration
nav_order: 7
---

# Server

The server section configures and tunes the http server module Authelia uses.

## Configuration

```yaml
server:
  read_buffer_size: 4096
  write_buffer_size: 4096
  path: ""
```

## Options

### read_buffer_size

Configures the maximum request size. The default of 4096 is generally sufficient for most use cases.

### write_buffer_size

Configures the maximum response size. The default of 4096 is generally sufficient for most use cases.

### path

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

## Additional Notes on Buffer Sizes

The read and write buffer sizes generally should be the same. This is because when Authelia verifies 
if the user is authorized to visit a URL, it also sends back nearly the same size response as the request. However
you're able to tune these individually depending on your needs.