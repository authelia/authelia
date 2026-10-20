---
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

title: "Envoy"
description: "An integration guide for Authelia and the Envoy reverse proxy including external authorization filter configuration examples, deployment tips, and key caveats."
summary: "A guide on integrating Authelia with the Envoy reverse proxy."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 330
toc: true
aliases:
  - /i/envoy
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

[Envoy] is supported by **Authelia**.

_**Important:** When using these guides, it's important to recognize that we cannot provide a guide for every possible
method of deploying a proxy. These guides show a suggested setup only, and you need to understand the proxy
configuration and customize it to your needs. To-that-end, we include links to the official proxy documentation
throughout this documentation and in the [See Also](#see-also) section._

## Get started

It's **_strongly recommended_** that users setting up _Authelia_ for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping _Authelia_.

## Trusted Proxies and Integration Security

{{< callout context="danger" title="Security Note" icon="outline/alert-octagon" >}}
In addition to this section which is important to read, you should read the
[Validating Forwarded Authentication](../../reference/guides/validating-forwarded-authentication.md) reference guide
and perform the validation steps as part of your regular security validation routine when using this integration.
{{< /callout >}}

_**Important:** You should read the [Forwarded Headers] section and this section as part of any proxy configuration.
Especially if you have never read it before._

[Envoy] determines the trusted client address and whether to trust the `X-Forwarded-Proto` header sent by the client
using the [use_remote_address] and [xff_num_trusted_hops] options of the HTTP connection manager. The included example
sets [use_remote_address] to `true` and does not set [xff_num_trusted_hops], which is appropriate when [Envoy] is the
first proxy the client connects to. In this configuration the trusted client address is the source address of the
connection, and any `X-Forwarded-Proto` header sent by the client is replaced. The example also replaces the
`X-Forwarded-For` header with the trusted client address for both the authorization request and the request to the
backend, as Authelia uses the first address in this header as the client address.

If [Envoy] is behind another proxy, such as a load balancer which terminates TLS, you will need to trust that proxy. It's
important to understand the following before doing so:

- The [xff_num_trusted_hops] option trusts the `X-Forwarded-For` and `X-Forwarded-Proto` headers based on a number of
  hops rather than the address of the proxy. When it's set on its own, these headers are trusted for every connection
  including clients which connect to [Envoy] directly.
- The `xff_trusted_cidrs` option of the [XFF original IP detection extension] only affects how the client address is
  determined, and it requires [use_remote_address] to be disabled. When [use_remote_address] is disabled [Envoy] does
  not replace the `X-Forwarded-Proto` header sent by any client.

To trust individual proxies you should instead add a filter chain to the listener which only matches connections from
the addresses of those proxies using the [direct_source_prefix_ranges] option, and set [xff_num_trusted_hops] on that
filter chain only. Connections from any other address use the filter chain without a [filter_chain_match], where the
headers sent by the client are not trusted. The following example trusts a single proxy with the address `192.168.1.10`:

```yaml {title="envoy.yml"}
static_resources:
  listeners:
    - name: 'listener_https'
      filter_chains:
        ## Connections from the trusted proxy.
        - filter_chain_match:
            direct_source_prefix_ranges:
              - address_prefix: '192.168.1.10'
                prefix_len: 32
          filters:
            - name: 'envoy.filters.network.http_connection_manager'
              typed_config:
                "@type": 'type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager'
                use_remote_address: true
                xff_num_trusted_hops: 1
                ## The remainder of the configuration is the same as the main example.
        ## Connections from all other addresses.
        - filters:
            - name: 'envoy.filters.network.http_connection_manager'
              typed_config:
                "@type": 'type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager'
                use_remote_address: true
                ## The remainder of the configuration is the same as the main example.
```

The value of [xff_num_trusted_hops] should be the number of trusted proxies in front of [Envoy] which append to the
`X-Forwarded-For` header. Each filter chain requires its own `transport_socket` if the proxy connects to [Envoy] using
TLS.

## Assumptions and Adaptation

This guide makes a few assumptions. These assumptions may require adaptation in more advanced and complex scenarios. We
can not reasonably have examples for every advanced configuration option that exists. Some of these values can
automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

The following are the assumptions we make:

- Deployment Scenario:
  - Single Host
  - Authelia is deployed as a Container with the container name `{{< sitevar name="host" nojs="authelia" >}}` on port `{{< sitevar name="port" nojs="9091" >}}`
  - Proxy is deployed as a Container on a network shared with Authelia
- The above assumption means that Authelia should be accessible to the proxy on `{{< sitevar name="tls" nojs="http" >}}://{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}` and as such:
  - You will have to adapt all instances of the above URL to be `https://` if Authelia configuration has a TLS key and
    certificate defined
  - You will have to adapt all instances of `{{< sitevar name="host" nojs="authelia" >}}` in the URL if:
    - you're using a different container name
    - you deployed the proxy to a different location
  - You will have to adapt all instances of `{{< sitevar name="port" nojs="9091" >}}` in the URL if:
    - you have adjusted the default port in the configuration
  - You will have to adapt the entire URL if:
    - Authelia is on a different host to the proxy
- All services are part of the `{{< sitevar name="domain" nojs="example.com" >}}` domain:
  - This domain and the subdomains will have to be adapted in all examples to match your specific domains unless you're
    just testing or you want to use that specific domain

## Implementation

[Envoy] utilizes the [ExtAuthz](../../reference/guides/proxy-authorization.md#extauthz) Authz implementation. The
associated [Metadata](../../reference/guides/proxy-authorization.md#extauthz-metadata) should be considered required.

The examples below assume you are using the default
[Authz Endpoints Configuration](../../configuration/miscellaneous/server-endpoints-authz.md) or one similar to the
following minimal configuration:

```yaml {title="configuration.yml"}
server:
  endpoints:
    authz:
      ext-authz:
        implementation: 'ExtAuthz'
```

The examples below also assume you are using the modern
[Session Configuration](../../configuration/session/introduction.md) which includes the `domain`, `authelia_url`, and
`default_redirection_url` as a subkey of the `session.cookies` key as a list item. Below is an example of the modern
configuration as well as the legacy configuration for context.

{{< sessionTabs "Generate Random Password" >}}
{{< sessionTab "Modern" >}}

```yaml {title="configuration.yml"}
session:
  cookies:
    - domain: '{{</* sitevar name="domain" nojs="example.com" */>}}'
      authelia_url: 'https://{{</* sitevar name="subdomain-authelia" nojs="auth" */>}}.{{</* sitevar name="domain" nojs="example.com" */>}}'
      default_redirection_url: 'https://www.{{</* sitevar name="domain" nojs="example.com" */>}}'
```

{{< /sessionTab >}}
{{< sessionTab "Legacy" >}}

```yaml {title="configuration.yml"}
default_redirection_url: 'https://www.{{</* sitevar name="domain" nojs="example.com" */>}}'
session:
  domain: '{{</* sitevar name="domain" nojs="example.com" */>}}'
```

{{< /sessionTab >}}
{{< /sessionTabs >}}

## Configuration

Below you will find commented examples of the following configuration:

- Authelia Portal
- Protected Endpoint (Nextcloud)

### Example

Support for [Envoy] is possible with Authelia v4.37.0 and higher via the [Envoy] proxy [external authorization] filter.
The example uses the [Authz Endpoints](../../configuration/miscellaneous/server-endpoints-authz.md) which require Authelia
v4.38.0 or higher.

The example expects the TLS certificate chain and private key to exist at `${PWD}/data/certificates/public.chain.pem`
and `${PWD}/data/certificates/private.pem` respectively, which are mounted in the `/certificates` directory of the
[Envoy] container.

The example expects clients to connect to [Envoy] using the standard ports `80` and `443`, which are mapped to the
container ports `8080` and `8443` respectively. If clients use a different port, the `domains` of each virtual host must
include that port, and the redirect on the HTTP listener must set the `port_redirect` option to the HTTPS port as the
`https_redirect` option keeps any port other than `80` which the client used.

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
[Envoy] 1.37.0 contains a bug where headers such as `Location` are not returned to the client when Authelia denies a
request, which results in the client seeing the body of the response rather than being redirected to the portal. This
was fixed in [Envoy] 1.37.1, and you should not use 1.37.0 with Authelia.
{{< /callout >}}

[external authorization]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto.html#extensions-filters-http-ext-authz-v3-extauthz

```yaml {title="compose.yml"}
---
networks:
  net:
    driver: 'bridge'
services:
  envoy:
    container_name: 'envoy'
    image: 'envoyproxy/envoy:v1.40-latest'
    restart: 'unless-stopped'
    networks:
      net: {}
    ports:
      - '80:8080'
      - '443:8443'
    volumes:
      - '${PWD}/data/envoy/envoy.yaml:/etc/envoy/envoy.yaml'
      - '${PWD}/data/certificates:/certificates'
  authelia:
    container_name: '{{< sitevar name="host" nojs="authelia" >}}'
    image: 'authelia/authelia'
    restart: 'unless-stopped'
    networks:
      net: {}
    volumes:
      - '${PWD}/data/authelia/config:/config'
    environment:
      TZ: 'Australia/Melbourne'
  nextcloud:
    container_name: 'nextcloud'
    image: 'linuxserver/nextcloud'
    restart: 'unless-stopped'
    networks:
      net: {}
    volumes:
      - '${PWD}/data/nextcloud/config:/config'
      - '${PWD}/data/nextcloud/data:/data'
    environment:
      PUID: '1000'
      PGID: '1000'
      TZ: 'Australia/Melbourne'
```

```yaml {title="envoy.yml"}
static_resources:
  listeners:
    - name: 'listener_http'
      address:
        socket_address:
          address: '0.0.0.0'
          port_value: 8080
      filter_chains:
        - filters:
            - name: 'envoy.filters.network.http_connection_manager'
              typed_config:
                "@type": 'type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager'
                codec_type: 'auto'
                stat_prefix: 'ingress_http'
                route_config:
                  name: 'local_route'
                  virtual_hosts:
                    - name: 'backend'
                      domains: ['*']
                      routes:
                        - match:
                            prefix: '/'
                          redirect:
                            https_redirect: true
                http_filters:
                  - name: 'envoy.filters.http.router'
                    typed_config:
                      "@type": 'type.googleapis.com/envoy.extensions.filters.http.router.v3.Router'
    - name: 'listener_https'
      address:
        socket_address:
          address: '0.0.0.0'
          port_value: 8443
      filter_chains:
        - filters:
            - name: 'envoy.filters.network.http_connection_manager'
              typed_config:
                "@type": 'type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager'
                stat_prefix: 'ingress_http'
                use_remote_address: true
                skip_xff_append: false
                route_config:
                  name: 'local_route'
                  ## Replace the X-Forwarded-For header with the trusted client address so that the value sent by the
                  ## client is not forwarded to the backends.
                  request_headers_to_add:
                    - header:
                        key: 'X-Forwarded-For'
                        value: '%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%'
                      append_action: 'OVERWRITE_IF_EXISTS_OR_ADD'
                  virtual_hosts:
                    - name: 'nextcloud_service'
                      domains: ['nextcloud.{{< sitevar name="domain" nojs="example.com" >}}']
                      routes:
                        - match:
                            prefix: '/'
                          route:
                            cluster: 'nextcloud'
                    - name: 'authelia_service'
                      domains: ['{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}']
                      typed_per_filter_config:
                        envoy.filters.http.ext_authz:
                          "@type": 'type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthzPerRoute'
                          disabled: true
                      routes:
                        - match:
                            prefix: '/'
                          route:
                            cluster: 'authelia'
                http_filters:
                  - name: 'envoy.filters.http.ext_authz'
                    typed_config:
                      "@type": 'type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz'
                      transport_api_version: 'v3'
                      allowed_headers:
                        patterns:
                          - exact: 'authorization'
                          - exact: 'proxy-authorization'
                          - exact: 'accept'
                          - exact: 'cookie'
                      http_service:
                        path_prefix: '/api/authz/ext-authz/'
                        server_uri:
                          uri: '{{< sitevar name="host" nojs="authelia" >}}:{{< sitevar name="port" nojs="9091" >}}'
                          cluster: 'authelia'
                          timeout: '2s'
                        authorization_request:
                          headers_to_add:
                            - key: 'X-Forwarded-Proto'
                              value: '%REQ(X-FORWARDED-PROTO)%'
                            - key: 'X-Forwarded-For'
                              value: '%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%'
                            ## The following commented lines are for configuring the Authelia URL in the proxy. We
                            ## strongly suggest this is configured in the Session Cookies section of the Authelia configuration.
                            # - key: X-Authelia-URL
                            #   value: https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}
                        authorization_response:
                          allowed_upstream_headers:
                            patterns:
                              - exact: 'location'
                              - exact: 'authorization'
                              - exact: 'proxy-authorization'
                              - prefix: 'remote-'
                              - prefix: 'authelia-'
                          allowed_client_headers:
                            patterns:
                              - exact: 'set-cookie'
                          allowed_client_headers_on_success:
                            patterns:
                              - exact: 'set-cookie'
                      failure_mode_allow: false
                      ## Respond with 503 rather than the default of 403 when Authelia returns an error or cannot be
                      ## reached so that these failures are not mistaken for an authorization decision.
                      status_on_error:
                        code: 'ServiceUnavailable'
                  - name: 'envoy.filters.http.router'
                    typed_config:
                      "@type": 'type.googleapis.com/envoy.extensions.filters.http.router.v3.Router'
          transport_socket:
            name: 'envoy.transport_sockets.tls'
            typed_config:
              "@type": 'type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext'
              common_tls_context:
                tls_certificates:
                  - certificate_chain:
                      filename: '/certificates/public.chain.pem'
                    private_key:
                      filename: '/certificates/private.pem'
  clusters:
    - name: 'nextcloud'
      connect_timeout: '0.25s'
      type: 'logical_dns'
      dns_lookup_family: 'v4_only'
      lb_policy: 'round_robin'
      load_assignment:
        cluster_name: 'nextcloud'
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: 'nextcloud'
                      port_value: 80
    - name: 'authelia'
      connect_timeout: '0.25s'
      type: 'logical_dns'
      dns_lookup_family: 'v4_only'
      lb_policy: 'round_robin'
      load_assignment:
        cluster_name: 'authelia'
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: 'authelia'
                      port_value: 9091
overload_manager:
  resource_monitors:
    - name: 'envoy.resource_monitors.global_downstream_max_connections'
      typed_config:
        "@type": 'type.googleapis.com/envoy.extensions.resource_monitors.downstream_connections.v3.DownstreamConnectionsConfig'
        max_active_downstream_connections: 50000
```

## FAQ

### Envoy is behind another proxy which terminates TLS, how do I configure that?

If [Envoy] receives requests over plain HTTP from another proxy which terminates TLS, such as a Cloudflare Tunnel or a
load balancer, [Envoy] replaces the `X-Forwarded-Proto` header with `http` unless it trusts that proxy. Authelia then
rejects the requests with a `400 Bad Request` status code and logs an error stating the Target URL has an insecure
scheme `http`.

To resolve this configure [Envoy] to trust that proxy as described in the
[Trusted Proxies and Integration Security](#trusted-proxies-and-integration-security) section. [Envoy] keeps the
`X-Forwarded-Proto` header sent by a trusted proxy, and the example sends this header to Authelia in the authorization
request.

## Kubernetes

Authelia supports some of the [Envoy] based Kubernetes Ingresses such as [Envoy Gateway](../kubernetes/envoy/gateway.md)
and [Istio](../kubernetes/envoy/istio.md). To see the full list see the
[Kubernetes Integration Guide](../kubernetes/envoy/introduction.md).

## See Also

- [Envoy External Authorization Documentation](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto.html#extensions-filters-http-ext-authz-v3-extauthz)
- [Forwarded Headers]

[Envoy]: https://www.envoyproxy.io/
[use_remote_address]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-field-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-use-remote-address
[xff_num_trusted_hops]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-field-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-xff-num-trusted-hops
[XFF original IP detection extension]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/http/original_ip_detection/xff/v3/xff.proto
[direct_source_prefix_ranges]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener_components.proto#envoy-v3-api-field-config-listener-v3-filterchainmatch-direct-source-prefix-ranges
[filter_chain_match]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener_components.proto#envoy-v3-api-field-config-listener-v3-filterchain-filter-chain-match
[Forwarded Headers]: forwarded-headers
