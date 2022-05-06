---
layout: default
title: Caddy
parent: Proxy Integration
grand_parent: Deployment
nav_order: 1
---

[Caddy] is a reverse proxy supported by **Authelia**. Caddy officially supports the forward auth flow in version 2.5.1
and greater. Authelia offers integration support for the official forward auth integration method Caddy provides, we
can't reasonably be expected to offer support for all of the different plugins that exist. As we have direct contact
with the Caddy developers and their integration being written so well solving issues should be relatively straightforward.

## Configuration

Below you will find commented examples of the following configuration:

* Authelia portal
* Protected endpoint (Nextcloud)

### Basic example

The basic example is coming soon!

## Advanced example

The advanced example allows for more flexible customization, however the [basic example](#basic-example) should be
preferred in _most_ situations. If you are unsure of what you're doing please don't use this method.

```Caddyfile
authelia.example.com {
        log
        reverse_proxy authelia:9091
}

nextcloud.example.com {
        log
        route {
                reverse_proxy authelia:9091 {
                        method GET
                        rewrite "/api/verify?rd=https://authelia.example.com"

                        header_up X-Forwarded-Method {method}
                        header_up X-Forwarded-Uri {uri}

                        ## If the auth request:
                        ##   1. Responds with a status code IN the 200-299 range.
                        ## Then:
                        ##   1. Proxy the request to the backend.
                        ##   2. Copy the relevant headers from the auth request and provide them to the backend.
                        @good status 2xx
                        handle_response @good {
                                request_header Remote-User {http.reverse_proxy.header.Remote-User}
                                request_header Remote-Groups {http.reverse_proxy.header.Remote-Groups}
                                request_header Remote-Name {http.reverse_proxy.header.Remote-Name}
                                request_header Remote-Email {http.reverse_proxy.header.Remote-Email}
                        }

                        ## If the auth request:
                        ##   1. Responds with a status code NOT IN the 200-299 range.
                        ## Then:
                        ##   1. Respond with the status code of the auth request.
                        ##   1. Copy the response except for several headers.
                        @denied {
                                status 1xx 3xx 4xx 5xx
                        }
                        handle_response @denied {
                                copy_response
                                copy_response_headers {
                                        exclude Connection Keep-Alive Te Trailers Transfer-Encoding Upgrade
                                }
                        }
                }

                reverse_proxy nextcloud:80
        }
}
```


[Caddy]: https://caddyserver.com
