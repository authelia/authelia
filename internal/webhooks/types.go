// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"time"

	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/model"
)

// Store is the storage a Dispatcher records callback confirmations in, so that a destination which confirmed using
// the callback stays confirmed when Authelia restarts. It also signs the callback keys, so that every Authelia instance
// which shares the storage gives a destination the same callback address.
type Store interface {
	LoadCachedData(ctx context.Context, name string) (data *model.CachedData, err error)
	SaveCachedData(ctx context.Context, data model.CachedData) (err error)
	DeleteCachedData(ctx context.Context, name string) (err error)
	WebhookCallbackSignature(values ...[]byte) (signature string)
}

type grant struct {
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Rate      int       `json:"rate"`
	Confirmed time.Time `json:"confirmed"`
}

type delivery struct {
	events  []*events.Event
	batched bool
}
