// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

const (
	externalIdentityCallbackPathFmt      = "/api/firstfactor/external-identity/%s/callback"
	externalIdentitySignInSelectorFmt    = "#external-identity-sign-in-button-%s"
	externalIdentityLinkStartSelectorFmt = "#external-identity-link-start-%s"
	externalIdentityLinkPendingSelector  = "#external-identity-link-pending-panel"
	externalIdentityLinkAcceptSelector   = "#external-identity-link-accept"
	externalIdentityLinkListedSelector   = "[id^='external-identity-link-'][id$='-description']"
	externalIdentityLinkedAccountsPath   = "/settings/external-identity"
	identityVerificationDialogSelector   = "#dialog-verify-one-time-code"
	identityVerificationCodeSelector     = "#one-time-code"
	identityVerificationVerifySelector   = "#dialog-verify"
)

var errExternalIdentityCallbackNotRedirected = errors.New("the external identity callback was not redirected")

func (rs *RodSession) doAwaitExternalIdentityCallback(ctx context.Context, page *rod.Page, provider string, patience time.Duration, action func() error) (callback, location string, requested bool, err error) {
	path := fmt.Sprintf(externalIdentityCallbackPathFmt, provider)

	if err = (proto.NetworkEnable{}).Call(page.Timeout(pageActionTimeout)); err != nil {
		return "", "", false, fmt.Errorf("error enabling the network domain: %w", err)
	}

	listener, cancel := page.WithCancel()

	redirects := make(chan [2]string, 1)

	var seen atomic.Bool

	wait := listener.EachEvent(func(e *proto.NetworkRequestWillBeSent) (stop bool) {
		if isURLPath(e.Request.URL, path) {
			seen.Store(true)
		}

		if e.RedirectResponse == nil || !isURLPath(e.RedirectResponse.URL, path) {
			return false
		}

		redirects <- [2]string{e.RedirectResponse.URL, e.Request.URL}

		return true
	})

	done := make(chan struct{})

	go func() {
		wait()
		close(done)
	}()

	defer func() {
		cancel()
		<-done
	}()

	if err = action(); err != nil {
		return "", "", seen.Load(), fmt.Errorf("error sending the browser to the '%s' provider: %w", provider, err)
	}

	timer := time.NewTimer(patience)

	defer timer.Stop()

	select {
	case redirect := <-redirects:
		return redirect[0], redirect[1], true, nil
	case <-timer.C:
		return "", "", seen.Load(), fmt.Errorf("%w within %s", errExternalIdentityCallbackNotRedirected, patience)
	case <-ctx.Done():
		return "", "", seen.Load(), fmt.Errorf("%w: %w", errExternalIdentityCallbackNotRedirected, ctx.Err())
	}
}
