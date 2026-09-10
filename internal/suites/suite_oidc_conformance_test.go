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

var conformanceAcceptedResults = []string{"PASSED", "REVIEW", "SKIPPED", "WARNING"}

const (
	conformanceDiagnosticTimeout = time.Second * 30
	conformancePlanTimeout       = time.Minute * 75
	conformanceDrainTimeout      = time.Second * 45
)

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

// SetupSuite creates every plan and starts a runner per plan.
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

func (s *OIDCConformanceSuite) release() {
	if s.cancel != nil {
		s.cancel()
	}

	s.drainRunners()

	for _, browser := range s.browsers {
		browser.Close()
	}

	if s.RodSession != nil {
		if err := s.Stop(); err != nil {
			s.T().Logf("error stopping rod session: %v", err)
		}
	}
}

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
// session via release.
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

func (s *OIDCConformanceSuite) reportFailure(t *testing.T, plan string, outcome ConformanceOutcome) {
	t.Logf("Module '%s' of the '%s' plan finished with the result '%s' and the status '%s'.", outcome.Module, plan, outcome.Result, outcome.Status)

	if outcome.Err != nil {
		t.Logf("Suite error: %v", outcome.Err)
	}

	if outcome.ID == "" {
		t.Log("The module was never created, so the conformance server has no log for it.")

		return
	}

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

// TestOIDCConformanceGenerateRoundTrip exercises the boundary between the two processes this suite spans.
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
		assert.Truef(t, ids[plan.Plan.Client.ID], "the '%s' profile's client '%s' is not in the generated configuration", plan.Name, plan.Plan.Client.ID)
		assert.Truef(t, ids[plan.Plan.ClientAlternate.ID], "the '%s' profile's alternate client '%s' is not in the generated configuration", plan.Name, plan.Plan.ClientAlternate.ID)
		assert.Truef(t, ids[plan.Plan.ClientSecretPost.ID], "the '%s' profile's secret post client '%s' is not in the generated configuration", plan.Name, plan.Plan.ClientSecretPost.ID)
	}
}

func conformanceArchiveName(plan string) string {
	return fmt.Sprintf("conformance-%s.zip", plan)
}

func conformanceArchivePath(plan string) string {
	return fmt.Sprintf("../../screenshots/%s/%s", oidcConformanceSuiteName, conformanceArchiveName(plan))
}
