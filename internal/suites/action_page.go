// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/stretchr/testify/require"
)

const (
	pageActionTimeout = time.Second * 30
	pageHasTimeout    = time.Second * 2
	pagePollInterval  = time.Millisecond * 250
	pageSettleTimeout = time.Second * 30
)

func (rs *RodSession) hasElement(page *rod.Page, selector string) bool {
	has, _, err := page.Timeout(pageHasTimeout).Has(selector)

	return err == nil && has
}

func (rs *RodSession) pageURL(page *rod.Page) string {
	info, err := page.Timeout(pageActionTimeout).Info()
	if err != nil {
		return ""
	}

	return info.URL
}

func (rs *RodSession) pageTitle(page *rod.Page) string {
	info, err := page.Timeout(pageActionTimeout).Info()
	if err != nil {
		return ""
	}

	return info.Title
}

func (rs *RodSession) doOpen(ctx context.Context, page *rod.Page, uri string) (err error) {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	loaded := page.Timeout(pageActionTimeout).WaitNavigation(proto.PageLifecycleEventNameDOMContentLoaded)

	if err = rs.doNavigate(page.Timeout(pageActionTimeout), uri); err != nil {
		return fmt.Errorf("error navigating to '%s': %w", uri, err)
	}

	loaded()

	return nil
}

func (rs *RodSession) doInput(page *rod.Page, selector, value string) (err error) {
	element, err := page.Timeout(elementLocateTimeout).Element(selector)
	if err != nil {
		return fmt.Errorf("error locating '%s': %w", selector, err)
	}

	if err = element.SelectAllText(); err != nil {
		return fmt.Errorf("error selecting the text of '%s': %w", selector, err)
	}

	if err = element.Input(value); err != nil {
		return fmt.Errorf("error typing into '%s': %w", selector, err)
	}

	return nil
}

func (rs *RodSession) doClick(page *rod.Page, selector string) (err error) {
	element, err := page.Timeout(elementLocateTimeout).Element(selector)
	if err != nil {
		return fmt.Errorf("error locating '%s': %w", selector, err)
	}

	if err = element.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("error clicking '%s': %w", selector, err)
	}

	return nil
}

func (rs *RodSession) doAwaitEnabled(page *rod.Page, selector string) (err error) {
	element, err := page.Timeout(elementLocateTimeout).Element(selector)
	if err != nil {
		return fmt.Errorf("error locating '%s': %w", selector, err)
	}

	if err = element.Timeout(pageActionTimeout).WaitEnabled(); err != nil {
		return fmt.Errorf("error waiting for '%s' to be enabled: %w", selector, err)
	}

	return nil
}

func (rs *RodSession) doAwaitElement(ctx context.Context, page *rod.Page, selector string) error {
	deadline := time.Now().Add(pageSettleTimeout)

	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if rs.hasElement(page, selector) {
			return nil
		}

		time.Sleep(pagePollInterval)
	}

	return fmt.Errorf("'%s' was not on the page '%s' after %s", selector, rs.pageURL(page), pageSettleTimeout)
}

func (rs *RodSession) doAwaitSubmitted(ctx context.Context, page *rod.Page, selector, from string) error {
	deadline := time.Now().Add(pageSettleTimeout)

	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !rs.hasElement(page, selector) || rs.pageURL(page) != from {
			return nil
		}

		time.Sleep(pagePollInterval)
	}

	return fmt.Errorf("the form at '%s' was submitted but '%s' was still on screen %s later", from, selector, pageSettleTimeout)
}

func (rs *RodSession) doAwaitURLPrefix(ctx context.Context, page *rod.Page, prefix string, has bool) (current string, err error) {
	for {
		if current = rs.pageURL(page); current != "" && strings.HasPrefix(current, prefix) == has {
			return current, nil
		}

		select {
		case <-ctx.Done():
			return current, ctx.Err()
		case <-time.After(elementRetryInterval):
		}
	}
}

func (rs *RodSession) waitURLHasPrefix(t *testing.T, page *rod.Page, prefix string) {
	t.Helper()

	rs.waitURL(t, page, prefix, true)
}

func (rs *RodSession) waitURLLacksPrefix(t *testing.T, page *rod.Page, prefix string) {
	t.Helper()

	rs.waitURL(t, page, prefix, false)
}

func (rs *RodSession) waitURL(t *testing.T, page *rod.Page, prefix string, has bool) {
	t.Helper()

	current, err := rs.doAwaitURLPrefix(page.GetContext(), page, prefix, has)

	require.NoErrorf(t, err, "the page did not reach a url which %s the prefix '%s', it is at '%s'", map[bool]string{true: "has", false: "lacks"}[has], prefix, current)
}

func (rs *RodSession) doClearCookies(page *rod.Page) error {
	return proto.NetworkClearBrowserCookies{}.Call(page.Timeout(pageActionTimeout))
}

func (rs *RodSession) doMarkDocument(page *rod.Page, mark string) (err error) {
	_, err = page.Timeout(pageActionTimeout).Eval(fmt.Sprintf(`() => { window.%s = true; }`, mark))

	return err
}

func (rs *RodSession) isMarkedDocument(page *rod.Page, mark string) (marked, ok bool) {
	result, err := page.Timeout(pageHasTimeout).Eval(fmt.Sprintf(`() => window.%s === true`, mark))
	if err != nil {
		return false, false
	}

	return result.Value.Bool(), true
}

func isURLPath(uri, expected string) bool {
	u, err := url.Parse(uri)
	if err != nil {
		return false
	}

	return u.Path == expected
}
