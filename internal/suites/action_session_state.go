// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
)

const sessionStateScript = `() => fetch('/api/state', {credentials: 'same-origin'}).then((r) => r.json()).then((b) => ({username: (b.data && b.data.username) || '', level: (b.data && b.data.authentication_level) || 0}))`

func (rs *RodSession) doGetSessionState(page *rod.Page) (username string, level int, err error) {
	result, err := page.Timeout(pageActionTimeout).Eval(sessionStateScript)
	if err != nil {
		return "", 0, err
	}

	return result.Value.Get("username").Str(), result.Value.Get("level").Int(), nil
}

func (rs *RodSession) doAwaitSignedIn(ctx context.Context, page *rod.Page, username string) error {
	deadline := time.Now().Add(pageSettleTimeout)

	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if current, level, err := rs.doGetSessionState(page); err == nil && current == username && level >= 1 {
			return nil
		}

		time.Sleep(pagePollInterval)
	}

	return fmt.Errorf("the account '%s' was not signed in after %s", username, pageSettleTimeout)
}
