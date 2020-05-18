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
# Configuration options specific to the internal http server
server:
  # Buffers usually should be configured to be the same value.
  # Explanation at https://docs.authelia.com/configuration/server.html
  # Read buffer size configures the http server's maximum incoming request size in bytes.
  read_buffer_size: 4096
  # Write buffer size configures the http server's maximum outgoing response size in bytes.
  write_buffer_size: 4096
```

### Buffer Sizes

The read and write buffer sizes generally should be the same. This is because when Authelia verifies 
if the user is authorized to visit a URL, it also sends back nearly the same size response 
(write_buffer_size) as the request (read_buffer_size).
