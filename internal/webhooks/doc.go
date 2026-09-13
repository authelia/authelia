// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Package webhooks delivers events to configured HTTPS destinations. It owns the per destination worker, buffer,
// retry policy, HMAC signing, and hardened HTTP client. It implements events.Emitter.
package webhooks
