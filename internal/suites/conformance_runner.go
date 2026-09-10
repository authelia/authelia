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

	log "github.com/sirupsen/logrus"
)

const (
	conformanceStatusConfigured = "CONFIGURED"
	conformanceStatusWaiting    = "WAITING"
	conformanceStatusFinished   = "FINISHED"

	// conformanceStatusInterrupted is a module which stopped on a failing condition. Like FINISHED it is never left,
	// so waiting on FINISHED alone spends the whole module budget on a verdict the server has already reached.
	conformanceStatusInterrupted = "INTERRUPTED"

	// conformanceModuleTimeout is the budget for one module, browser interaction included.
	conformanceModuleTimeout = time.Minute * 5

	// conformancePlaceholderTimeout is how long a module which fills placeholders is given after its browser legs
	// complete, before its remaining placeholders are considered absent.
	conformancePlaceholderTimeout = time.Second * 30

	// conformanceUploadSettleInterval and conformanceUploadSettleTimeout govern the wait after a placeholder is
	// filled. The conformance server does not act on the upload immediately: the module stays in WAITING for several
	// seconds while it picks the image up, which the suite's own UI shows as a spinner. Polling the status directly
	// for that window, rather than falling back into the outer loop which re-reads the browser status and the whole
	// log on every pass, keeps the poll at the once-a-second it is meant to be.
	conformanceUploadSettleInterval = time.Second
	conformanceUploadSettleTimeout  = time.Second * 10

	// conformancePlaceholderInterval is how often the log is re-read while waiting for a placeholder to appear.
	conformancePlaceholderInterval = time.Second

	// conformanceOutcomeInfoTimeout bounds the status read which completes an outcome. It runs on a context detached
	// from the module's, because a module which ran out of time is the one whose status is most worth reporting.
	conformanceOutcomeInfoTimeout = time.Second * 10
)

// conformanceTerminalStates are the states a module never leaves.
var conformanceTerminalStates = []string{conformanceStatusFinished, conformanceStatusInterrupted}

// ConformanceOutcome is the result of one module, produced off the test goroutine and asserted on it.
type ConformanceOutcome struct {
	Name   string
	Module string

	// ID is the module instance's identifier on the conformance server. It is what fetches the module's log for a
	// failure message, and it names the module's file inside the exported plan archive.
	ID string

	Result  string
	Status  string
	LogURL  string
	Skipped bool
	Reason  string
	Err     error
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
	// it, so it goes on waiting for a callback which is never sent. Once Assert has confirmed the page, the stub image
	// is uploaded to the module's log, which is what marks it for REVIEW, and the module is stopped.
	UploadErrorPage bool
}

// provesErrorPage reports whether the last leg ended on an error page this module is released from by upload.
func (o ConformanceOverride) provesErrorPage(legs *conformanceLegs) bool {
	return o.UploadErrorPage && legs.errorURL != ""
}

// conformanceOverrides holds only what differs from the reactive driver. Anything absent runs on the default path, so
// a module added by a future conformance suite release runs rather than failing.
var conformanceOverrides = map[string]ConformanceOverride{
	// The second authorization carries prompt=login, so Authelia must ask for credentials again despite the session.
	"oidcc-prompt-login": {Assert: conformanceAssertReauthentication},

	// max_age=1 with a second of delay must force re-authentication and an auth_time claim.
	"oidcc-max-age-1": {Assert: conformanceAssertReauthentication},

	// max_age=10000 is longer than the session has existed, so the provider must not ask again.
	"oidcc-max-age-10000": {Assert: conformanceAssertNoReauthentication},

	// Both send a redirect_uri which is not registered, so Authelia must show its own error page rather than send the
	// browser to it. This is the page the module's screenshot placeholder asks for.
	"oidcc-ensure-registered-redirect-uri":          {Assert: conformanceAssertErrorPage},
	"oidcc-ensure-request-object-with-redirect-uri": {Assert: conformanceAssertErrorPage},

	// Authelia rejects the unsigned request object on its own error page, which is within the specification, but the
	// module has no placeholder for that page and only accepts the rejection by a screenshot uploaded to its log.
	"oidcc-unsigned-request-object-supported-correctly-or-rejected-as-unsupported": {Assert: conformanceAssertErrorPage, UploadErrorPage: true},

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

// conformanceUnattended holds the modules which cannot run without a person, mapped to why. They are reported as
// skipped rather than failed. Keep this list minimal and justified: a module which leaves it becomes a real failure,
// which is the point.
var conformanceUnattended = map[string]string{
	"oidcc-server-rotate-keys": "requires the provider's signing keys to be rotated by hand while the module waits",
}

// conformanceReauthenticationLeg is the index of the browser leg the re-authentication expectations apply to. These
// modules authorize twice: the first round trip only establishes the session, and by the time it runs the plan's
// browser context may already hold one from an earlier module, so whether the sign in form appears on it says nothing
// about the parameter under test. The second round trip is the one which carries prompt=login or max_age.
const conformanceReauthenticationLeg = 1

func conformanceAssertReauthentication(leg ConformanceLeg) error {
	if leg.Index != conformanceReauthenticationLeg {
		return nil
	}

	// Either route counts as being asked again: Authelia normally asks on the consent decision form, but a session
	// which has expired outright sends the browser back to the sign in page instead.
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

	if reason, ok := conformanceUnattended[module.TestModule]; ok {
		outcome.Skipped, outcome.Reason = true, reason

		return outcome
	}

	ctx, cancel := context.WithTimeout(ctx, conformanceModuleTimeout)
	defer cancel()

	override := conformanceOverrides[module.TestModule]

	id, state, err := r.createAndAdvance(ctx, module, override)

	log.Debugf("Conformance module '%s' created as '%s' in state '%s'", module.TestModule, id, state)

	if id != "" {
		outcome.ID, outcome.LogURL = id, r.client.LogDetailURL(id)
	}

	if err != nil {
		outcome.Err = err

		return outcome
	}

	if state == conformanceStatusWaiting {
		if err = r.interact(ctx, id, module.TestModule, override); err != nil {
			outcome.Err = err
		}
	}

	// Only wait for the module to finish when the interaction succeeded. A module whose browser leg failed will never
	// leave WAITING, so waiting on it would burn the whole module budget before reporting a failure already known.
	if outcome.Err == nil {
		if _, err = r.client.WaitState(ctx, id, conformanceTerminalStates...); err != nil {
			outcome.Err = err
		}
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

// createAndAdvance creates module's test instance and, if it starts out CONFIGURED, starts it. It returns the id as
// soon as CreateTest has succeeded - even alongside a non-nil err from a later step - so the caller can still surface
// the module's log URL, and the state the module reached: WAITING or FINISHED on success.
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

// interact drives every URL the module hands over and fills every placeholder it raises, until the module reaches a
// terminal state. A module such as oidcc-prompt-login performs two authorization round trips, so this loops rather
// than assuming a single URL.
//
// Leaving WAITING is not the end: between two legs the module is RUNNING while it processes the first callback, and
// only publishes the second URL once it is back in WAITING.
func (r *ConformanceRunner) interact(ctx context.Context, id, module string, override ConformanceOverride) (err error) {
	if r.browser == nil {
		return errors.New("the module needs a browser but the runner has none")
	}

	legs := &conformanceLegs{visited: map[string]bool{}}

	// The stall deadline is only meaningful once the module has been asked at least once what it is waiting for, so
	// it is armed after the first poll rather than before it. A module which publishes its first URL more than
	// conformancePlaceholderTimeout after entering WAITING would otherwise be reported as stalled.
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

		progressed, err := r.visitURLs(ctx, id, module, override, legs, status.URLs)
		if err != nil {
			return err
		}

		if override.provesErrorPage(legs) {
			return r.proveErrorPage(ctx, id, legs.errorURL)
		}

		filled, err := r.fillPlaceholders(ctx, id)
		if err != nil {
			return err
		}

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

// advance reads where the module has got to. A module in a terminal state is done. One which is RUNNING between browser
// legs is waited on until it is back in WAITING or done, and reported as settled, since it has made progress the
// caller could not see.
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

	log.Debugf("Conformance module '%s' is in state '%s' between browser legs", id, state)

	if state, err = r.client.WaitState(ctx, id, append([]string{conformanceStatusWaiting}, conformanceTerminalStates...)...); err != nil {
		return false, false, err
	}

	return state != conformanceStatusWaiting, true, nil
}

// conformanceLegs is one module's browser leg bookkeeping, carried across the iterations of interact's poll loop: the
// URLs already driven, and how many legs have been driven so far. The count is what lets an override distinguish the
// second authorization round trip of a module such as oidcc-prompt-login from the first.
type conformanceLegs struct {
	visited map[string]bool
	count   int

	// errorURL is the page the most recent leg ended on when that was Authelia reporting an error, and empty when it
	// reached the client.
	errorURL string
}

// stalled explains a module which is still waiting with nothing left to drive. When the last leg ended on Authelia's
// error page the module is waiting for a callback which will not come, and that is Authelia's doing rather than the
// driver's, so the error says what Authelia reported.
func (l *conformanceLegs) stalled(id string) error {
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

// conformanceDescribeErrorPage renders the error parameters Authelia's completion view carries in its query, falling
// back to the URL for an error page which carries none.
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

// visitURLs drives every URL in urls which has not already been driven, recording each as it goes. It reports whether
// it drove any URL at all, so the caller can tell a quiet pass (nothing new to visit) from one that made progress.
func (r *ConformanceRunner) visitURLs(ctx context.Context, id, module string, override ConformanceOverride, legs *conformanceLegs, urls []string) (progressed bool, err error) {
	for _, uri := range urls {
		if legs.visited[uri] {
			continue
		}

		legs.visited[uri] = true
		progressed = true

		index := legs.count
		legs.count++

		log.Debugf("Conformance module '%s' driving leg %d at '%s'", module, index, uri)

		leg, err := r.browser.Drive(ctx, index, uri)
		if err != nil {
			return progressed, err
		}

		log.Debugf("Conformance module '%s' finished leg %d: first factor %t, consent %t, reauthentication %t, error page %t",
			module, index, leg.FirstFactor, leg.Consent, leg.Reauthentication, leg.AutheliaError)

		legs.errorURL = leg.ErrorURL

		// The assertion is made against what the driver observed over the leg, before the placeholder upload releases
		// the module. It cannot be made against the live page: Drive only returns once the flow has left Authelia.
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

// proveErrorPage uploads the stub image to the log of a module which has no placeholder for the error page it was
// shown, then stops the module. The upload goes first: it is what records the REVIEW, and it is recorded in place only
// while the module is still running.
func (r *ConformanceRunner) proveErrorPage(ctx context.Context, id, errorURL string) (err error) {
	log.Debugf("Conformance module '%s' uploading the error page it has no placeholder for", id)

	description := fmt.Sprintf("Authelia's error page, confirmed by the suite: %s", conformanceDescribeErrorPage(errorURL))

	if err = r.client.UploadImage(ctx, id, description, conformancePlaceholderImage); err != nil {
		return fmt.Errorf("error uploading the error page for module '%s': %w", id, err)
	}

	if err = r.client.StopTest(ctx, id); err != nil {
		return fmt.Errorf("error stopping module '%s' after uploading its error page: %w", id, err)
	}

	return nil
}

// fillPlaceholders uploads the stub image to every unfilled placeholder, reporting whether it filled any. The image
// carries no information: the module's page was already checked by the override's assertion, and the upload exists
// only because waitForPlaceholders() will not release the module without one.
func (r *ConformanceRunner) fillPlaceholders(ctx context.Context, id string) (filled bool, err error) {
	entries, err := r.client.Log(ctx, id)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if entry.Upload == "" {
			continue
		}

		log.Debugf("Conformance module '%s' filling placeholder '%s'", id, entry.Upload)

		if err = r.client.UploadPlaceholder(ctx, id, entry.Upload, conformancePlaceholderImage); err != nil {
			return filled, err
		}

		filled = true
	}

	return filled, nil
}

// awaitPlaceholderSettled polls a module's status until it leaves WAITING, or until timeout elapses, and reports the
// status it last saw. The first read happens before any wait, so a module which has already moved on is not made to
// wait for it.
func (r *ConformanceRunner) awaitPlaceholderSettled(ctx context.Context, id string, interval, timeout time.Duration) (current string, err error) {
	deadline := time.Now().Add(timeout)

	for {
		var info *ConformanceTestInfo

		if info, err = r.client.Info(ctx, id); err != nil {
			return "", err
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

// moduleState reports the module's current status. A module whose placeholder was just filled gets its settle window
// first: a single check straight after an upload would almost always see the old status, because the conformance
// server takes several seconds to act on the image.
func (r *ConformanceRunner) moduleState(ctx context.Context, id string, filled bool) (state string, err error) {
	if filled {
		if state, err = r.awaitPlaceholderSettled(ctx, id, conformanceUploadSettleInterval, conformanceUploadSettleTimeout); err != nil {
			return "", err
		}

		log.Debugf("Conformance module '%s' is in state '%s' after its placeholder was filled", id, state)

		return state, nil
	}

	var info *ConformanceTestInfo

	if info, err = r.client.Info(ctx, id); err != nil {
		return "", err
	}

	return info.Status, nil
}
