// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
)

// Emitter accepts an event for asynchronous delivery. Implementations must never block and must never return an
// error, because an authentication or self-service flow must not fail or stall because a webhook receiver is
// unhealthy. Implementations must not retain the context beyond the call, because callers pass the request context and
// the request is recycled once the handler returns.
type Emitter interface {
	Emit(ctx context.Context, event *Event)
}

// NewNoOpEmitter returns an Emitter which discards every event. It is the provider default so that call sites never
// need a nil check.
func NewNoOpEmitter() (emitter Emitter) {
	return &NoOpEmitter{}
}

// NoOpEmitter is an Emitter which discards every event.
type NoOpEmitter struct{}

// Emit discards the event.
func (e *NoOpEmitter) Emit(_ context.Context, _ *Event) {}
