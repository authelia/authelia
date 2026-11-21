// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Package events provides the typed event vocabulary Authelia emits to webhook destinations. It owns the payload
// types, the registry of valid type names, redaction of credential equivalent values, and the Emitter interface. It
// has no knowledge of HTTP or of how events are delivered.
package events
