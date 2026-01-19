// Package notification provides notification delivery implementations for Authelia.
// This file implements the webhook notifier for sending notifications via HTTP/HTTPS.
package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/webhooks"
)

// WebhookNotifier is a notifier that sends notifications via HTTP webhooks.
// It implements the Notifier interface by converting email notifications into
// JSON webhook payloads and sending them to a configured webhook endpoint.
type WebhookNotifier struct {
	// client is the webhook HTTP client used to send requests
	client *webhooks.Client
}

// WebhookPayload represents the JSON payload sent to the webhook endpoint.
// It contains all notification data in a structured format with a JSON schema reference.
type WebhookPayload struct {
	// Schema is the JSON schema URL that defines the payload structure
	Schema string `json:"$schema"`

	// Recipient is the email address of the notification recipient
	Recipient string `json:"recipient"`

	// Subject is the notification subject line
	Subject string `json:"subject"`

	// Body is the plaintext notification message body
	Body string `json:"body"`

	// Timestamp is the UTC time when the notification was sent
	Timestamp time.Time `json:"timestamp"`
}

// NewWebhookNotifier creates a WebhookNotifier using the provided webhook client.
// The client parameter must not be nil, or this function will panic.
// Returns a notifier that implements the Notifier interface for sending notifications via webhooks.
func NewWebhookNotifier(client *webhooks.Client) *WebhookNotifier {
	if client == nil {
		panic("webhook client cannot be nil")
	}

	return &WebhookNotifier{
		client: client,
	}
}

// StartupCheck implements the startup check provider interface.
// It verifies the webhook notifier is ready to send notifications.
// Returns nil as webhook configuration is validated during application initialization.
func (n *WebhookNotifier) StartupCheck() (err error) {
	// Webhook client is already validated during configuration loading.
	// This is just a placeholder for the interface.
	return nil
}

// Send sends a notification via webhook.
// It executes the email template to generate the notification body, creates a JSON payload,
// and sends it to the configured webhook endpoint using the HTTP client.
// The recipient parameter specifies who the notification is for, subject is the notification title,
// et is the email template to render, and data contains the template variables.
// Returns nil on success, or an error if template execution or webhook delivery fails.
func (n *WebhookNotifier) Send(ctx context.Context, recipient mail.Address, subject string, et *templates.EmailTemplate, data any) (err error) {
	// Execute template to get plaintext body.
	var bodyBuf bytes.Buffer

	if err = et.Text.Execute(&bodyBuf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Create webhook payload.
	payload := WebhookPayload{
		Schema:    "https://github.com/authelia/authelia/blob/master/docs/schemas/v1/webhook/notification.json",
		Recipient: recipient.Address,
		Subject:   subject,
		Body:      bodyBuf.String(),
		Timestamp: time.Now().UTC(),
	}

	// Marshal to JSON.
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Send via webhook client.
	if err = n.client.Send(ctx, jsonData); err != nil {
		return fmt.Errorf("failed to send webhook notification: %w", err)
	}

	return nil
}
