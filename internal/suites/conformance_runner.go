// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	conformanceReauthenticationLeg  = 1
	conformanceStatusConfigured     = "CONFIGURED"
	conformanceStatusWaiting        = "WAITING"
	conformanceStatusFinished       = "FINISHED"
	conformanceStatusInterrupted    = "INTERRUPTED"
	conformanceModuleTimeout        = time.Minute * 5
	conformancePlaceholderTimeout   = time.Second * 30
	conformanceUploadSettleInterval = time.Second
	conformanceUploadSettleTimeout  = time.Second * 10
	conformancePlaceholderInterval  = time.Second
	conformanceOutcomeInfoTimeout   = time.Second * 10
	conformanceErrorSettleTimeout   = time.Second * 30
)

var (
	conformanceTerminalStates = []string{conformanceStatusFinished, conformanceStatusInterrupted}

	conformanceOverrides = map[string]ConformanceOverride{
		// The second authorization carries prompt=login, so Authelia must ask for credentials again despite the session.
		"oidcc-prompt-login": {Assert: conformanceAssertReauthentication, Screenshot: ConformanceScreenshotReauthentication},

		// The max age one module sets max_age=1 with a second of delay must force re-authentication and an auth_time claim.
		"oidcc-max-age-1": {Assert: conformanceAssertReauthentication, Screenshot: ConformanceScreenshotReauthentication},

		// The max age 1000 module sets max_age=10000 is longer than the session has existed, so the provider must not ask
		// again.
		"oidcc-max-age-10000": {Assert: conformanceAssertNoReauthentication},

		// Both send a redirect_uri which is not registered, so Authelia must show its own error page rather than send the
		// browser to it. This is the page the module's screenshot placeholder asks for.
		"oidcc-ensure-registered-redirect-uri":          {Assert: conformanceAssertErrorPage, Screenshot: ConformanceScreenshotErrorPage},
		"oidcc-ensure-request-object-with-redirect-uri": {Assert: conformanceAssertErrorPage, Screenshot: ConformanceScreenshotErrorPage},

		// Authelia rejects the unsigned request object on its own error page, which is within the specification, but the
		// module has no placeholder for that page and only accepts the rejection by a screenshot uploaded to its log.
		"oidcc-unsigned-request-object-supported-correctly-or-rejected-as-unsupported": {Assert: conformanceAssertErrorPage, UploadErrorPage: true, Screenshot: ConformanceScreenshotErrorPage},

		// Every module whose upstream summary says to remove any cookies received from the provider, and no others. The
		// plans also carry oidcc-registration-logo-uri, -policy-uri and -tos-uri under that instruction, but those are
		// Dynamic OP modules and none of the profiles Authelia is certified for includes them.
		//
		// oidcc-prompt-none-not-logged-in is the one where a stale session is not merely untidy: prompt=none against a
		// live session returns a code, and the module is asserting that the provider errors instead, so it fails.
		"oidcc-display-page":              {ClearCookies: true},
		"oidcc-display-popup":             {ClearCookies: true},
		"oidcc-login-hint":                {ClearCookies: true},
		"oidcc-prompt-none-not-logged-in": {ClearCookies: true},
		"oidcc-ui-locales":                {ClearCookies: true},
	}

	conformanceUnattended = map[string]string{
		"oidcc-server-rotate-keys": "requires the provider's signing keys to be rotated by hand while the module waits",
	}
)

// ConformanceOutcome is the result of one module, produced off the test goroutine and asserted on it.
type ConformanceOutcome struct {
	Name    string
	Module  string
	ID      string
	Result  string
	Status  string
	LogURL  string
	Skipped bool
	Reason  string
	Err     error
	Trace   []string
}

// ConformanceOverride is the per module behavior which differs from the reactive default.
type ConformanceOverride struct {
	// ClearCookies discards the context's cookies before the module's first browser leg, for the modules whose
	// instructions are to remove any cookies received from the provider before proceeding.
	ClearCookies bool

	// Assert stands in for the screenshot the conformance suite would otherwise want. It is handed what the driver
	// observed over one browser leg, once that leg has been driven to its end, and its error becomes the module's
	// verdict. It is given the record rather than the browser because by the time a leg ends the browser is on the
	// conformance suite's callback, so nothing about Authelia's pages can be established from the live DOM any more.
	// The record carries the leg's index, so an expectation which belongs to the second authorization round trip is
	// not applied to the first.
	Assert func(leg ConformanceLeg) error

	// UploadErrorPage is for a module which accepts Authelia's error page as an outcome but raises no placeholder for
	// it, so it goes on waiting for a callback which is never sent. Once Assert has confirmed the page, a screenshot of
	// it is uploaded to the module's log, which is what marks it for REVIEW, and the module is stopped.
	UploadErrorPage bool

	// Screenshot is the page this module's screenshot placeholder asks for. A placeholder is only filled with a capture
	// of that page, so one from a module without it is left unfilled and the module stalls.
	Screenshot ConformanceScreenshot
}

// ConformanceScreenshot is a page a module's screenshot placeholder asks for.
type ConformanceScreenshot int

const (
	// ConformanceScreenshotNone is a module which asks for no screenshot.
	ConformanceScreenshotNone ConformanceScreenshot = iota

	// ConformanceScreenshotReauthentication is Authelia asking for the password again on the second authorization.
	ConformanceScreenshotReauthentication

	// ConformanceScreenshotErrorPage is Authelia's error page.
	ConformanceScreenshotErrorPage
)

func (o ConformanceOverride) screenshotOf(leg ConformanceLeg) string {
	switch o.Screenshot {
	case ConformanceScreenshotReauthentication:
		if leg.Index == conformanceReauthenticationLeg {
			return leg.LoginScreenshot
		}
	case ConformanceScreenshotErrorPage:
		return leg.ErrorScreenshot
	}

	return ""
}

func (o ConformanceOverride) provesErrorPage(legs *conformanceLegs) bool {
	return o.UploadErrorPage && legs.errorURL != ""
}

func conformanceAssertReauthentication(leg ConformanceLeg) error {
	if leg.Index != conformanceReauthenticationLeg {
		return nil
	}

	if !leg.Reauthentication && !leg.FirstFactor {
		return errors.New("expected Authelia to ask for the password again on the second authorization but it did not")
	}

	return nil
}

func conformanceAssertNoReauthentication(leg ConformanceLeg) error {
	if leg.Index != conformanceReauthenticationLeg {
		return nil
	}

	if leg.Reauthentication || leg.FirstFactor {
		return errors.New("expected Authelia to reuse the existing session on the second authorization but it asked for the password again")
	}

	return nil
}

func conformanceAssertErrorPage(leg ConformanceLeg) error {
	if !leg.AutheliaError {
		return errors.New("expected Authelia to show an error page rather than redirect to the client but it did not")
	}

	return nil
}

// ConformanceRunner walks one plan's modules in order.
type ConformanceRunner struct {
	client  *ConformanceClient
	browser *ConformanceBrowser
	planID  string
	modules []ConformancePlanModule
	names   []string

	// debug is whether each module's trace is recorded.
	debug bool

	// trace is the trace of the module being run. Modules run one at a time, so it is replaced as each begins.
	trace *ConformanceTrace
}

// NewConformanceRunner returns a ConformanceRunner for one plan. browser may be nil only when no module in modules
// reaches the WAITING state, which in practice means only in tests.
func NewConformanceRunner(client *ConformanceClient, browser *ConformanceBrowser, planID string, modules []ConformancePlanModule) *ConformanceRunner {
	return &ConformanceRunner{
		client:  client,
		browser: browser,
		planID:  planID,
		modules: modules,
		names:   ConformanceSubtestNames(modules),
	}
}

// Run walks the plan on its own goroutine and returns the channel its outcomes arrive on, one per module in plan
// order. The channel closes when the plan is done. Nothing here touches [testing.T]: an outcome carries its error so the
// test goroutine can be the one that fails.
func (r *ConformanceRunner) Run(ctx context.Context) <-chan ConformanceOutcome {
	outcomes := make(chan ConformanceOutcome, len(r.modules))

	go func() {
		defer close(outcomes)

		for i, module := range r.modules {
			outcomes <- r.run(ctx, r.names[i], module)
		}
	}()

	return outcomes
}

func (r *ConformanceRunner) run(ctx context.Context, name string, module ConformancePlanModule) (outcome ConformanceOutcome) {
	outcome = ConformanceOutcome{Name: name, Module: module.TestModule}

	r.trace = NewConformanceTrace(r.debug)

	if r.browser != nil {
		r.browser.trace = r.trace
	}

	defer func() {
		outcome.Trace = r.trace.Lines()
	}()

	if reason, ok := conformanceUnattended[module.TestModule]; ok {
		outcome.Skipped, outcome.Reason = true, reason

		return outcome
	}

	parent := ctx

	ctx, cancel := context.WithTimeout(ctx, conformanceModuleTimeout)
	defer cancel()

	override := conformanceOverrides[module.TestModule]

	id, state, err := r.createAndAdvance(ctx, module, override)

	r.trace.Logf("Module '%s' created as '%s' in state '%s'", module.TestModule, id, state)

	if id != "" {
		outcome.ID, outcome.LogURL = id, r.client.LogDetailURL(id)
	}

	if err != nil {
		outcome.Err = err

		if id != "" {
			r.settleAfterError(parent, id)
		}

		return outcome
	}

	if state == conformanceStatusWaiting {
		if err = r.interact(ctx, id, override); err != nil {
			outcome.Err = err
		}
	}

	if outcome.Err == nil {
		if _, err = r.client.WaitState(ctx, id, conformanceTerminalStates...); err != nil {
			outcome.Err = err
		}
	}

	if outcome.Err != nil {
		r.settleAfterError(parent, id)
	}

	infoCtx, infoCancel := context.WithTimeout(context.WithoutCancel(ctx), conformanceOutcomeInfoTimeout)
	defer infoCancel()

	info, err := r.client.Info(infoCtx, id)
	if err != nil {
		if outcome.Err == nil {
			outcome.Err = err
		}

		return outcome
	}

	outcome.Status, outcome.Result = info.Status, info.Result

	return outcome
}

func (r *ConformanceRunner) settleAfterError(parent context.Context, id string) {
	ctx, cancel := context.WithTimeout(parent, conformanceErrorSettleTimeout)
	defer cancel()

	state, err := r.client.WaitState(ctx, id, append([]string{conformanceStatusConfigured, conformanceStatusWaiting}, conformanceTerminalStates...)...)

	r.trace.Logf("Module '%s' settled in state '%s' after the suite failed it (%v)", id, state, err)
}

func (r *ConformanceRunner) createAndAdvance(ctx context.Context, module ConformancePlanModule, override ConformanceOverride) (id, state string, err error) {
	if id, err = r.client.CreateTest(ctx, r.planID, module.TestModule, module.Variant); err != nil {
		return "", "", fmt.Errorf("error creating module '%s': %w", module.TestModule, err)
	}

	if override.ClearCookies && r.browser != nil {
		if err = r.browser.ClearCookies(); err != nil {
			return id, "", err
		}
	}

	if state, err = r.client.WaitState(ctx, id, append([]string{conformanceStatusConfigured, conformanceStatusWaiting}, conformanceTerminalStates...)...); err != nil {
		return id, "", err
	}

	if state != conformanceStatusConfigured {
		return id, state, nil
	}

	if err = r.client.StartTest(ctx, id); err != nil {
		return id, "", err
	}

	state, err = r.client.WaitState(ctx, id, append([]string{conformanceStatusWaiting}, conformanceTerminalStates...)...)

	return id, state, err
}

func (r *ConformanceRunner) interact(ctx context.Context, id string, override ConformanceOverride) (err error) {
	if r.browser == nil {
		return errors.New("the module needs a browser but the runner has none")
	}

	legs := &conformanceLegs{visited: map[string]bool{}}

	var deadline time.Time

	for {
		if ctx.Err() != nil {
			return fmt.Errorf("module '%s' did not leave the WAITING state: %w", id, ctx.Err())
		}

		status, err := r.client.BrowserStatus(ctx, id)
		if err != nil {
			return err
		}

		if deadline.IsZero() {
			deadline = time.Now().Add(conformancePlaceholderTimeout)
		}

		progressed, err := r.visitURLs(ctx, id, override, legs, status.URLs)
		if err != nil {
			return err
		}

		if override.provesErrorPage(legs) {
			return r.proveErrorPage(ctx, id, legs.errorURL, legs.screenshot)
		}

		filled, pending, err := r.fillPlaceholders(ctx, id, legs.screenshot)
		if err != nil {
			return err
		}

		legs.pendingPlaceholder = pending

		done, settled, err := r.advance(ctx, id, filled)
		if err != nil {
			return err
		}

		if done {
			return nil
		}

		if progressed || filled || settled {
			deadline = time.Now().Add(conformancePlaceholderTimeout)
		} else if time.Now().After(deadline) {
			return legs.stalled(id)
		}

		time.Sleep(conformancePlaceholderInterval)
	}
}

func (r *ConformanceRunner) advance(ctx context.Context, id string, filled bool) (done, settled bool, err error) {
	var state string

	if state, err = r.moduleState(ctx, id, filled); err != nil {
		return false, false, err
	}

	switch state {
	case conformanceStatusFinished, conformanceStatusInterrupted:
		return true, false, nil
	case conformanceStatusWaiting:
		return false, false, nil
	}

	r.trace.Logf("Module '%s' is in state '%s' between browser legs", id, state)

	if state, err = r.client.WaitState(ctx, id, append([]string{conformanceStatusWaiting}, conformanceTerminalStates...)...); err != nil {
		return false, false, err
	}

	return state != conformanceStatusWaiting, true, nil
}

type conformanceLegs struct {
	visited            map[string]bool
	count              int
	errorURL           string
	screenshot         string
	pendingPlaceholder bool
}

func (l *conformanceLegs) stalled(id string) error {
	if l.pendingPlaceholder {
		return fmt.Errorf("module '%s' is waiting on a screenshot placeholder, but no leg showed the page it asks for", id)
	}

	if l.errorURL != "" {
		return &ConformanceErrorPageStallError{ID: id, URL: l.errorURL}
	}

	return fmt.Errorf("module '%s' stayed in the WAITING state with nothing left to visit or fill", id)
}

// ConformanceErrorPageStallError is a module left waiting for the flow to return to the client after Authelia ended
// it on an error page, which the module had no placeholder for.
type ConformanceErrorPageStallError struct {
	ID  string
	URL string
}

// Error names the module and what Authelia's error page reported.
func (e *ConformanceErrorPageStallError) Error() string {
	return fmt.Sprintf("module '%s' is waiting for the flow to return to the client, but Authelia ended it on an error page and the module raised no placeholder for one: %s",
		e.ID, conformanceDescribeErrorPage(e.URL))
}

func conformanceDescribeErrorPage(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}

	query := u.Query()

	var parts []string

	for _, field := range []struct{ label, key string }{
		{"error", "error"},
		{"description", "error_description"},
		{"hint", "error_hint"},
		{"debug", "error_debug"},
	} {
		if value := query.Get(field.key); value != "" {
			parts = append(parts, fmt.Sprintf("%s '%s'", field.label, value))
		}
	}

	if len(parts) == 0 {
		return uri
	}

	return strings.Join(parts, ", ")
}

func (r *ConformanceRunner) visitURLs(ctx context.Context, id string, override ConformanceOverride, legs *conformanceLegs, urls []string) (progressed bool, err error) {
	for _, uri := range urls {
		if legs.visited[uri] {
			continue
		}

		legs.visited[uri] = true
		progressed = true

		index := legs.count
		legs.count++

		r.trace.Logf("Driving leg %d at '%s'", index, uri)

		leg, err := r.browser.Drive(ctx, index, uri)
		if err != nil {
			return progressed, err
		}

		r.trace.Logf("Finished leg %d: first factor %t, consent %t, reauthentication %t, error page %t",
			index, leg.FirstFactor, leg.Consent, leg.Reauthentication, leg.AutheliaError)

		legs.errorURL = leg.ErrorURL
		legs.screenshot = conformanceLatest(legs.screenshot, override.screenshotOf(leg))

		if override.Assert != nil {
			if err = override.Assert(leg); err != nil {
				return progressed, err
			}
		}

		if err = r.client.Visit(ctx, id, uri); err != nil {
			return progressed, err
		}
	}

	return progressed, nil
}

func (r *ConformanceRunner) proveErrorPage(ctx context.Context, id, errorURL, screenshot string) (err error) {
	if screenshot == "" {
		return fmt.Errorf("module '%s' ended on Authelia's error page but no screenshot of it could be taken", id)
	}

	r.trace.Logf("Module '%s' uploading the error page it has no placeholder for", id)

	description := fmt.Sprintf("Authelia's error page, confirmed by the suite: %s", conformanceDescribeErrorPage(errorURL))

	if err = r.client.UploadImage(ctx, id, description, screenshot); err != nil {
		return fmt.Errorf("error uploading the error page for module '%s': %w", id, err)
	}

	if err = r.client.StopTest(ctx, id); err != nil {
		return fmt.Errorf("error stopping module '%s' after uploading its error page: %w", id, err)
	}

	return nil
}

func (r *ConformanceRunner) fillPlaceholders(ctx context.Context, id, screenshot string) (filled, pending bool, err error) {
	entries, err := r.client.Log(ctx, id)
	if err != nil {
		return false, false, err
	}

	for _, entry := range entries {
		if entry.Upload == "" {
			continue
		}

		if screenshot == "" {
			pending = true

			continue
		}

		r.trace.Logf("Module '%s' filling placeholder '%s'", id, entry.Upload)

		if err = r.client.UploadPlaceholder(ctx, id, entry.Upload, screenshot); err != nil {
			return filled, pending, err
		}

		filled = true
	}

	return filled, pending, nil
}

func (r *ConformanceRunner) awaitPlaceholderSettled(ctx context.Context, id string, interval, timeout time.Duration) (current string, err error) {
	deadline := time.Now().Add(timeout)

	for {
		var info *ConformanceTestInfo

		if info, err = r.client.Info(ctx, id); err != nil {
			return "", fmt.Errorf("error reading the status of module '%s' while waiting for its placeholder to settle: %w", id, err)
		}

		if info.Status != conformanceStatusWaiting || time.Now().After(deadline) {
			return info.Status, nil
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("module '%s' did not leave the %s state after its placeholder was filled: %w", id, conformanceStatusWaiting, ctx.Err())
		case <-time.After(interval):
		}
	}
}

func (r *ConformanceRunner) moduleState(ctx context.Context, id string, filled bool) (state string, err error) {
	if filled {
		if state, err = r.awaitPlaceholderSettled(ctx, id, conformanceUploadSettleInterval, conformanceUploadSettleTimeout); err != nil {
			return "", err
		}

		r.trace.Logf("Module '%s' is in state '%s' after its placeholder was filled", id, state)

		return state, nil
	}

	var info *ConformanceTestInfo

	if info, err = r.client.Info(ctx, id); err != nil {
		return "", err
	}

	return info.Status, nil
}
