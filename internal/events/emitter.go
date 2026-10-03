// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
)

// NewNoOpEmitter returns an Emitter which discards every event. It is the provider default so that call sites never
// need a nil check.
func NewNoOpEmitter() (emitter Emitter) {
	return &NoOpEmitter{}
}

// NoOpEmitter is an Emitter which discards every event.
type NoOpEmitter struct{}

// Emit discards the event.
func (e *NoOpEmitter) Emit(_ context.Context, _ *Event) {}
