---
title: "Webhook Definitions"
description: "Configuring reusable webhook definitions."
summary: "Authelia allows you to define reusable webhook endpoints that can be referenced by multiple components."
date: 2025-11-10T00:00:00+00:00
draft: false
images: []
weight: 115200
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Configuration

{{< config-alert-example >}}

```yaml {title="configuration.yml"}
definitions:
  webhooks:
    ntfy:
      url: 'https://ntfy.example.com/authelia'
      method: 'POST'
      timeout: '5s'
      headers:
        Authorization: 'Bearer your-token-here'

    slack:
      url: 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX'
      method: 'POST'
      timeout: '10s'
```

## Overview

Webhook definitions provide a centralized way to configure HTTP webhook endpoints that can be referenced by multiple Authelia components. This promotes configuration reusability and consistency across different features.

## Use Cases

Webhook definitions can be referenced by:
- **Notifications**: Send user notifications (password resets, 2FA registration) via webhooks
- **Future Features**: Audit logging, telemetry, and other event notifications

## Options

### url

{{< confkey type="string" required="yes" >}}

The webhook endpoint URL where requests will be sent.

{{< callout context="caution" title="Important" icon="outline/alert-triangle" >}}
The URL **must** use HTTPS for security. HTTP URLs will be rejected during configuration validation.
{{< /callout >}}

**Example:**

```yaml {title="configuration.yml"}
definitions:
  webhooks:
    my_webhook:
      url: 'https://webhook.example.com/authelia/events'
```

### method

{{< confkey type="string" default="POST" required="no" >}}

The HTTP method to use when sending webhook requests. Supported methods are:
- `POST` (default)
- `PUT`
- `PATCH`

**Example:**

```yaml {title="configuration.yml"}
definitions:
  webhooks:
    my_webhook:
      url: 'https://webhook.example.com/events'
      method: 'PUT'
```

### timeout

{{< confkey type="string,integer" syntax="duration" default="5 seconds" required="no" >}}

The timeout for the webhook HTTP request.

**Example:**

```yaml {title="configuration.yml"}
definitions:
  webhooks:
    my_webhook:
      url: 'https://webhook.example.com/events'
      timeout: '10s'
```

### headers

{{< confkey type="map" required="no" >}}

Custom HTTP headers to include in webhook requests. This can be used for authentication, content negotiation, or custom metadata.

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The following headers are automatically set by the webhook client and should not be overridden:
- `Content-Type: application/json`
- `User-Agent: Authelia-Webhook-Client`
{{< /callout >}}

**Example:**

```yaml {title="configuration.yml"}
definitions:
  webhooks:
    my_webhook:
      url: 'https://webhook.example.com/events'
      headers:
        Authorization: 'Bearer your-secret-token'
        X-API-Key: 'your-api-key'
        X-Custom-Header: 'custom-value'
```

### tls

{{< confkey type="object" required="no" >}}

TLS configuration for the webhook connection. See [TLS Configuration](../miscellaneous/server.md#tls) for available options.

**Example:**

```yaml {title="configuration.yml"}
definitions:
  webhooks:
    my_webhook:
      url: 'https://webhook.example.com/events'
      tls:
        server_name: 'webhook.example.com'
        skip_verify: false
        minimum_version: 'TLS1.2'
```

## Security Considerations

{{< callout context="caution" title="Important" icon="outline/alert-triangle" >}}
When configuring webhooks:
1. **Always use HTTPS** - HTTP URLs are rejected for security
2. **Secure your endpoints** - Implement authentication using the `headers` configuration
3. **Validate incoming requests** - Verify webhooks originate from your Authelia instance
4. **Protect sensitive data** - Webhook payloads may contain sensitive information
{{< /callout >}}

## Examples

### Ntfy Integration

```yaml {title="configuration.yml"}
definitions:
  webhooks:
    ntfy:
      url: 'https://ntfy.sh/my-authelia-topic'
      method: 'POST'
      timeout: '5s'
      headers:
        X-Priority: '3'

notifier:
  webhook_ref: 'ntfy'
```

### Slack Integration

```yaml {title="configuration.yml"}
definitions:
  webhooks:
    slack:
      url: 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX'
      method: 'POST'
      timeout: '10s'

notifier:
  webhook_ref: 'slack'
```

### Discord Integration

```yaml {title="configuration.yml"}
definitions:
  webhooks:
    discord:
      url: 'https://discord.com/api/webhooks/000000000000000000/XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX'
      method: 'POST'
      timeout: '5s'

notifier:
  webhook_ref: 'discord'
```
