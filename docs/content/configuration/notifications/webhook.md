---
title: "Webhook"
description: "Configuring the Webhook Notifications Settings."
summary: "Authelia can send notifications to users through HTTP webhooks. This section describes how to configure this."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 108300
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
        Authorization: 'Bearer token123'

notifier:
  disable_startup_check: false
  webhook_ref: 'ntfy'
```

## Options

This section describes the individual configuration options.

### webhook_ref

{{< confkey type="string" required="yes" >}}

Reference to a webhook defined in `definitions.webhooks`. This allows webhooks to be reused across multiple features (notifications, audit logging, telemetry, etc.).

See the [Webhook Definitions](../definitions/webhooks.md) documentation for details on configuring webhook endpoints.

**Example:**

```yaml {title="configuration.yml"}
definitions:
  webhooks:
    my_webhook:
      url: 'https://webhook.example.com/notifications'
      method: 'POST'
      timeout: '5s'
      headers:
        Authorization: 'Bearer your-token-here'

notifier:
  webhook_ref: 'my_webhook'
```

## Webhook Payload

The webhook notifier sends a JSON payload with the following structure:

```json
{
  "$schema": "https://github.com/authelia/authelia/blob/master/docs/schemas/v1/webhook/notification.json",
  "recipient": "user@example.com",
  "subject": "[Authelia] Password Reset",
  "body": "Click here to reset your password: https://...",
  "timestamp": "2024-03-14T06:00:14Z"
}
```

### JSON Schema

Authelia publishes a formal [JSON Schema](https://json-schema.org/) for the webhook notification payload to facilitate integration with external systems. The schema is available in the repository at:

**https://github.com/authelia/authelia/blob/master/docs/schemas/v1/webhook/notification.json**

The `$schema` field in every webhook payload references this schema, allowing receiving systems to validate incoming webhooks and generate types or models automatically.

### Payload Fields

| Field | Type | Description |
|-------|------|-------------|
| `$schema` | string (URI) | JSON Schema reference URL for this webhook payload format |
| `recipient` | string (email) | The email address of the notification recipient |
| `subject` | string | The notification subject line |
| `body` | string | The plaintext body of the notification |
| `timestamp` | string (ISO 8601) | UTC timestamp when the notification was sent |

## Security Considerations

{{< callout context="caution" title="Important" icon="outline/alert-triangle" >}}
When using webhook notifications:
1. **Always use HTTPS** - HTTP URLs are rejected for security
2. **Secure your endpoint** - Implement authentication using the `headers` configuration
3. **Validate requests** - Verify incoming webhooks are from your Authelia instance
4. **Protect sensitive data** - Notification bodies may contain sensitive information like password reset links
{{< /callout >}}

## Example Webhook Receiver

Here's a simple example of a webhook receiver endpoint:

```python
from flask import Flask, request, jsonify
import hmac
import hashlib

app = Flask(__name__)

@app.route('/authelia/notifications', methods=['POST'])
def receive_notification():
    # Verify authentication header
    token = request.headers.get('Authorization')
    if token != 'Bearer your-secret-token':
        return jsonify({'error': 'Unauthorized'}), 401

    # Parse webhook payload
    data = request.json
    recipient = data.get('recipient')
    subject = data.get('subject')
    body = data.get('body')
    timestamp = data.get('timestamp')

    # Process notification (e.g., send to Slack, Discord, etc.)
    print(f"Notification for {recipient}: {subject}")

    return jsonify({'status': 'received'}), 200

if __name__ == '__main__':
    app.run(ssl_context='adhoc')
```

## Use Cases

The webhook notifier is useful for:
- **Integration with messaging platforms** - Forward notifications to Slack, Discord, Microsoft Teams, etc.
- **Custom notification systems** - Implement your own notification delivery mechanism
- **Logging and auditing** - Store notification events in external systems
- **Multi-channel delivery** - Send notifications through multiple channels simultaneously
- **Testing and development** - Use services like webhook.site for testing

## Template Support

{{< callout context="note" title="Note" icon="outline/info-circle" >}}
The webhook notifier uses the same email templates as other notification providers. Only the plaintext version of the template is included in the webhook payload's `body` field. HTML templates are not sent.
{{< /callout >}}

Custom templates can be configured using the `template_path` option. See the [Notification Templates Reference Guide](../../reference/guides/notification-templates.md) for details.
