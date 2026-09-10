// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.yaml.in/yaml/v4"

	"github.com/authelia/authelia/v4/internal/oidc/conformance"
	"github.com/authelia/authelia/v4/internal/utils"
)

// conformanceAcceptedResults are the module results which pass. REVIEW is here because filling an image placeholder
// forces it, and this suite fills placeholders for every module which blocks on one, so REVIEW carries no signal. Add
// WARNING here to accept results which the conformance suite flags but does not fail.
var conformanceAcceptedResults = []string{"PASSED", "REVIEW", "SKIPPED", "WARNING"}

// conformancePlanTimeout is the budget for one whole plan.
// conformanceDiagnosticTimeout bounds reading one failing module's log for its failure report. It is deliberately
// independent of the plan context, so that a report is still produced for a plan whose context has expired.
const conformanceDiagnosticTimeout = time.Second * 30

const conformancePlanTimeout = time.Minute * 75

// conformanceDrainTimeout bounds how long release waits for one runner to notice its context was cancelled and
// return. elementLocateTimeout (webdriver.go) is the longest a single in-flight ConformanceBrowser call can block
// regardless of context cancellation, so this needs to clear it comfortably rather than merely exceed it.
const conformanceDrainTimeout = time.Second * 45

// OIDCConformanceSuite runs the OpenID Foundation conformance suite against Authelia, one test per certified profile.
type OIDCConformanceSuite struct {
	*RodSuite

	client   *ConformanceClient
	cancel   context.CancelFunc
	browsers []*ConformanceBrowser
	outcomes map[string]<-chan ConformanceOutcome
	planIDs  map[string]string
}

// NewOIDCConformanceSuite returns a new *OIDCConformanceSuite.
func NewOIDCConformanceSuite() *OIDCConformanceSuite {
	return &OIDCConformanceSuite{
		RodSuite: NewRodSuite(oidcConformanceSuiteName),
		outcomes: map[string]<-chan ConformanceOutcome{},
		planIDs:  map[string]string{},
	}
}

// SetupSuite creates every plan and starts a runner per plan. The runners are concurrent because seven plans in
// sequence is an unreasonable wall clock, and they are goroutines rather than parallel testify methods because
// testify's suite carries one *testing.T which suite.Run reassigns per method: calling t.Parallel() in a method races
// every other method against that field. Nothing a runner does touches testing.T; each outcome carries its own error
// and the profile method which drains it is what fails.
//
// The deferred release below exists because testify only registers its TearDownSuite call in a defer after
// SetupSuite returns normally (suite.go:226-241 in testify v1.12.1). s.Require().NoError failing here calls
// t.FailNow, which unwinds this goroutine with runtime.Goexit before that registration ever happens, so
// TearDownSuite is never invoked on this path. Goexit does still run this goroutine's own already-registered
// defers, which is what makes a defer placed here - before any Require call - the only thing that stops the
// runners and closes the browsers already started when a later plan fails to create.
func (s *OIDCConformanceSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()

	complete := false

	defer func() {
		if !complete {
			s.release()
		}
	}()

	plans, err := oidcConformanceReadPlans()
	s.Require().NoError(err)

	s.client, err = NewConformanceClient(oidcConformanceBaseURL)
	s.Require().NoError(err)

	session, err := NewRodSession()
	s.Require().NoError(err)

	s.RodSession = session

	ctx, cancel := context.WithTimeout(context.Background(), conformancePlanTimeout)

	s.cancel = cancel

	for _, plan := range plans {
		plan := plan

		created, err := s.client.CreatePlan(ctx, plan.PlanName, plan.Variant, &plan.Plan)
		s.Require().NoErrorf(err, "error creating the '%s' plan", plan.Name)

		s.planIDs[plan.Name] = created.ID

		browser, err := NewConformanceBrowser(s.RodSession)
		s.Require().NoError(err)

		s.browsers = append(s.browsers, browser)

		s.outcomes[plan.Name] = NewConformanceRunner(s.client, browser, created.ID, created.Modules).Run(ctx)
	}

	complete = true
}

// release cancels the plan context, waits for every started runner to finish, closes every browser context created
// so far, and stops the rod session - in that order. It is the single place both TearDownSuite and a SetupSuite
// which failed partway release resources from, so the two paths - one running after SetupSuite returns normally,
// the other running inside SetupSuite's own unwind - cannot drift apart.
//
// The drain has to happen before any browser is closed. On the success path every profile method already drained
// its plan's channel to completion in assertPlan, so by the time TearDownSuite calls release every runner has
// returned and stopped touching its browser. On the SetupSuite-failure path that invariant does not hold: a runner
// can still be mid module - for example blocked in ConformanceBrowser.input/click on an element locate bounded by
// elementLocateTimeout rather than the plan context - when release runs from SetupSuite's own defer. Closing that
// runner's browser out from under it while it is in flight would be unsynchronised concurrent access to the shared
// *rod.Page and browser context. Draining first, with cancellation already in effect so each runner errors out of
// its current step quickly, ensures no runner is still touching a browser when release closes it.
func (s *OIDCConformanceSuite) release() {
	if s.cancel != nil {
		s.cancel()
	}

	s.drainRunners()

	for _, browser := range s.browsers {
		browser.Close()
	}

	if s.RodSession != nil {
		// Stop does not reclaim the shared Chrome subprocess here: NewRodSession is called with no options in
		// SetupSuite, which takes the shared browser path (webdriver.go), and RodSession.Stop returns early for a
		// shared session before closing the browser or calling Launcher.Cleanup. All this does in our case is
		// dispose rs.contexts, which is empty - the ConformanceBrowser incognito contexts closed above are tracked
		// in s.browsers, not registered with the session. The shared Chrome subprocess is reclaimed by TestMain's
		// closeSharedBrowsers. This call is still correct and matches the pattern other suites use for their own
		// session bookkeeping and any non-shared resources RodSession.Stop does own.
		if err := s.Stop(); err != nil {
			s.T().Logf("error stopping rod session: %v", err)
		}
	}
}

// drainRunners waits for every started runner to finish walking its plan and close its outcome channel, discarding
// the outcomes - the assertions that matter for these outcomes only run when SetupSuite succeeded, which is the
// path assertPlan already drains. Each plan's wait is individually bounded by conformanceDrainTimeout: the caller
// has already cancelled the plan context, which makes a well-behaved runner return promptly, but a rod call that
// blocks past its own fixed timeout regardless of context cancellation must not be allowed to stall teardown
// indefinitely. A channel which is not closed within the bound is logged and abandoned rather than waited on
// further - leaving one goroutine running out its own timeout is preferable to hanging the whole teardown, and this
// path only runs when the suite has already failed.
func (s *OIDCConformanceSuite) drainRunners() {
	for name, outcomes := range s.outcomes {
		timer := time.NewTimer(conformanceDrainTimeout)

		for drained := false; !drained; {
			select {
			case _, ok := <-outcomes:
				if !ok {
					drained = true
				}
			case <-timer.C:
				s.T().Logf("timed out after %s waiting for the '%s' plan's runner to finish", conformanceDrainTimeout, name)

				drained = true
			}
		}

		timer.Stop()
	}
}

// TearDownSuite exports each plan's log for the CI artifacts, then releases the runners, browser contexts and rod
// session via release. This only runs when SetupSuite returned normally; a SetupSuite which failed partway releases
// through the defer it registers on itself instead, since testify never calls TearDownSuite on that path. The
// plan-log export stays here rather than in release because a failed setup has little worth exporting and s.client
// may not exist yet.
func (s *OIDCConformanceSuite) TearDownSuite() {
	if s.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
		defer cancel()

		for name, id := range s.planIDs {
			if err := s.client.ExportPlanHTML(ctx, id, conformanceArchivePath(name)); err != nil {
				s.T().Logf("Error exporting the '%s' plan log: %v", name, err)
			}
		}
	}

	s.release()
}

// assertPlan drains a plan's outcomes and asserts each as a subtest named after its module.
func (s *OIDCConformanceSuite) assertPlan(name string) {
	outcomes, ok := s.outcomes[name]
	s.Require().Truef(ok, "the '%s' plan was not created", name)

	count := 0

	for outcome := range outcomes {
		outcome := outcome
		count++

		s.Run(outcome.Name, func() {
			t := s.T()

			if outcome.Skipped {
				t.Skipf("%s cannot run unattended: %s", outcome.Module, outcome.Reason)

				return
			}

			// Reported before asserting, because a require aborts the subtest and anything logged after it is
			// never printed -- and this is the run where the detail is wanted most.
			if failed := outcome.Err != nil || !utils.IsStringInSlice(outcome.Result, conformanceAcceptedResults); failed {
				s.reportFailure(t, name, outcome)
			}

			require.NoError(t, outcome.Err)
			require.Containsf(t, conformanceAcceptedResults, outcome.Result,
				"module '%s' finished with the result '%s' and the status '%s'", outcome.Module, outcome.Result, outcome.Status)
		})
	}

	s.Require().NotZerof(count, "the '%s' plan produced no modules", name)
}

// reportFailure prints everything known about a module that did not pass, before the assertion that aborts the
// subtest.
//
// The conformance server's own log is the only place a failure explains itself -- the module result is a single word,
// and the message on a failing condition is usually incomplete without the arguments beside it. That log lives in a
// container which is torn down with the suite, so the log-detail URL below is useful while a run is in progress and
// useless by the time anyone reads CI output; the exported archive is the copy that survives, and the module id is
// what finds this module inside it.
func (s *OIDCConformanceSuite) reportFailure(t *testing.T, plan string, outcome ConformanceOutcome) {
	t.Logf("Module '%s' of the '%s' plan finished with the result '%s' and the status '%s'.", outcome.Module, plan, outcome.Result, outcome.Status)

	if outcome.Err != nil {
		t.Logf("Suite error: %v", outcome.Err)
	}

	if outcome.ID == "" {
		t.Log("The module was never created, so the conformance server has no log for it.")

		return
	}

	// A report must never be the thing that fails: without a client there is nothing to read the log with, and
	// panicking here would replace a described failure with an undescribed one.
	if s.client == nil {
		return
	}

	t.Logf("Module id %s, in %s of the exported plan archive. Live log while the suite is up: %s",
		outcome.ID, conformanceArchiveName(plan), outcome.LogURL)

	ctx, cancel := context.WithTimeout(context.Background(), conformanceDiagnosticTimeout)
	defer cancel()

	entries, err := s.client.Log(ctx, outcome.ID)
	if err != nil {
		t.Logf("The module's log could not be read for this report: %v", err)

		return
	}

	if detail := ConformanceDiagnostics(entries); detail != "" {
		t.Logf("Conformance log entries that did not pass:%s", detail)
	} else {
		t.Logf("The module logged no failing entries, which points at the suite driving it rather than at the provider.")
	}
}

// TestConfig runs the Config OP certification profile.
func (s *OIDCConformanceSuite) TestConfig() {
	s.assertPlan(conformance.NameConfig)
}

// TestBasic runs the Basic OP certification profile.
func (s *OIDCConformanceSuite) TestBasic() {
	s.assertPlan(conformance.NameBasic)
}

// TestBasicFormPost runs the Basic OP certification profile with the form post response mode.
func (s *OIDCConformanceSuite) TestBasicFormPost() {
	s.assertPlan(conformance.NameBasicFormPost)
}

// TestHybrid runs the Hybrid OP certification profile.
func (s *OIDCConformanceSuite) TestHybrid() {
	s.assertPlan(conformance.NameHybrid)
}

// TestHybridFormPost runs the Hybrid OP certification profile with the form post response mode.
func (s *OIDCConformanceSuite) TestHybridFormPost() {
	s.assertPlan(conformance.NameHybridFormPost)
}

// TestImplicit runs the Implicit OP certification profile.
func (s *OIDCConformanceSuite) TestImplicit() {
	s.assertPlan(conformance.NameImplicit)
}

// TestImplicitFormPost runs the Implicit OP certification profile with the form post response mode.
func (s *OIDCConformanceSuite) TestImplicitFormPost() {
	s.assertPlan(conformance.NameImplicitFormPost)
}

// TestOIDCConformanceSuite runs the OIDCConformance suite.
func TestOIDCConformanceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping suite test in short mode")
	}

	suite.Run(t, NewOIDCConformanceSuite())
}

// TestOIDCConformanceGenerateRoundTrip exercises the boundary between the two processes this suite spans: the
// authelia-scripts process writes the plans and the client configuration, and the go test process reads the plans
// back and creates them over the API. Everything which crosses that boundary as JSON has to survive it, which the
// generator alone cannot prove - conformance.Plan excludes the API plan name and the variant from its own JSON
// representation on purpose, so they only reach the test process because this file carries them itself.
func TestOIDCConformanceGenerateRoundTrip(t *testing.T) {
	t.Setenv("SUITE_TMP_PATH", t.TempDir())

	require.NoError(t, oidcConformanceGenerate())

	plans, err := oidcConformanceReadPlans()
	require.NoError(t, err)

	expected := map[string]struct {
		plan    string
		variant bool
	}{
		conformance.NameConfig:           {"oidcc-config-certification-test-plan", false},
		conformance.NameBasic:            {"oidcc-basic-certification-test-plan", true},
		conformance.NameBasicFormPost:    {"oidcc-formpost-basic-certification-test-plan", true},
		conformance.NameHybrid:           {"oidcc-hybrid-certification-test-plan", true},
		conformance.NameHybridFormPost:   {"oidcc-formpost-hybrid-certification-test-plan", true},
		conformance.NameImplicit:         {"oidcc-implicit-certification-test-plan", true},
		conformance.NameImplicitFormPost: {"oidcc-formpost-implicit-certification-test-plan", true},
	}

	require.Len(t, plans, len(expected))

	data, err := os.ReadFile(SuiteTmpPath(oidcConformanceClientsFile))
	require.NoError(t, err)

	clients := &oidcConformanceClients{}
	require.NoError(t, yaml.Unmarshal(data, clients))

	ids := map[string]bool{}

	for _, client := range clients.IdentityProviders.OIDC.Clients {
		ids[client.ID] = true
	}

	require.NotEmpty(t, ids)

	for _, plan := range plans {
		want, ok := expected[plan.Name]

		require.Truef(t, ok, "unexpected profile '%s'", plan.Name)

		assert.Equalf(t, want.plan, plan.PlanName, "the '%s' profile lost its conformance plan name", plan.Name)
		assert.NotEmptyf(t, plan.Plan.Alias, "the '%s' profile lost its alias", plan.Name)
		assert.NotEmptyf(t, plan.Plan.Server.DiscoveryURL, "the '%s' profile lost its discovery url", plan.Name)

		if want.variant {
			require.NotNilf(t, plan.Variant, "the '%s' profile lost its variant", plan.Name)
			assert.Equal(t, "discovery", plan.Variant.ServerMetadata)
			assert.Equal(t, "static_client", plan.Variant.ClientRegistration)
		} else {
			assert.Nilf(t, plan.Variant, "the '%s' profile gained a variant", plan.Name)
		}

		if plan.Name == conformance.NameConfig {
			assert.Nil(t, plan.Plan.Client)

			continue
		}

		require.NotNilf(t, plan.Plan.Client, "the '%s' profile lost its client", plan.Name)
		assert.NotEmpty(t, plan.Plan.Client.Secret)

		// The plan is what tells the conformance suite which client to authorize as, and the YAML is what tells
		// Authelia that client exists. If the two ever drift every module of that plan fails at the authorization
		// endpoint.
		assert.Truef(t, ids[plan.Plan.Client.ID], "the '%s' profile's client '%s' is not in the generated configuration", plan.Name, plan.Plan.Client.ID)
		assert.Truef(t, ids[plan.Plan.ClientAlternate.ID], "the '%s' profile's alternate client '%s' is not in the generated configuration", plan.Name, plan.Plan.ClientAlternate.ID)
		assert.Truef(t, ids[plan.Plan.ClientSecretPost.ID], "the '%s' profile's secret post client '%s' is not in the generated configuration", plan.Name, plan.Plan.ClientSecretPost.ID)
	}
}

// conformanceArchiveName returns the file name of a plan's exported log archive.
func conformanceArchiveName(plan string) string {
	return fmt.Sprintf("conformance-%s.zip", plan)
}

// conformanceArchivePath returns where a plan's exported log archive is written. The directory is the one the CI step
// collects as an artifact, and the test binary runs in its own package directory, so the path is relative to that.
func conformanceArchivePath(plan string) string {
	return fmt.Sprintf("../../screenshots/%s/%s", oidcConformanceSuiteName, conformanceArchiveName(plan))
}
