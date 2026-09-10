// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"go.yaml.in/yaml/v4"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc/conformance"
	"github.com/authelia/authelia/v4/internal/random"
)

const (
	oidcConformanceSuiteName = "OIDCConformance"

	// oidcConformanceBaseURL is the conformance suite's base url. It is the value the server builds its redirect URIs
	// from, so it must match the redirect URIs registered with Authelia byte for byte.
	oidcConformanceBaseURL = "https://conformance.example.com:8443"

	// oidcConformanceAutheliaURL is the issuer the conformance suite discovers Authelia at.
	oidcConformanceAutheliaURL = "https://login.example.com:8080"

	// oidcConformancePlansFile carries the generated plans from the authelia-scripts process which sets the suite up
	// to the go test process which runs it.
	oidcConformancePlansFile = "conformance-plans.json"

	// oidcConformanceClientsFile is the configuration fragment Authelia loads the conformance clients from.
	oidcConformanceClientsFile = "conformance-clients.yml"
)

// OIDCConformancePlanFile is one entry of the plans file exchanged between the setup and test processes. The API plan
// name and the variant are carried as their own fields rather than being read back off the embedded
// [conformance.Plan]: that type tags both of them json:"-" because authelia-gen holds them in memory and sends them as
// query parameters while writing only the plan body to disk, so a round trip through this file would otherwise decode
// them as their zero values and the suite would create every plan nameless and unvarianted.
type OIDCConformancePlanFile struct {
	// Name is the builder name of the profile, which is what the suite's per profile test methods look plans up by.
	Name string `json:"name"`

	// PlanName is the conformance suite's own name for the plan, sent as the planName query parameter.
	PlanName string `json:"plan_name"`

	// Variant is the plan's variant, sent as the variant query parameter.
	Variant *conformance.PlanVariant `json:"variant,omitempty"`

	// Plan is the plan body, sent as the request body.
	Plan conformance.Plan `json:"plan"`
}

// oidcConformanceClients is the shape of the configuration fragment holding the generated clients.
type oidcConformanceClients struct {
	IdentityProviders struct {
		OIDC struct {
			Clients []schema.IdentityProvidersOpenIDConnectClient `yaml:"clients"`
		} `yaml:"oidc"`
	} `yaml:"identity_providers"`
}

// oidcConformanceGenerate builds every certified profile and writes the two files the rest of the suite reads: the
// client configuration Authelia loads, and the plans the test process creates over the API. Both come from one builder
// invocation so a plan's client identifiers, secrets and redirect URIs cannot drift from Authelia's configuration.
func oidcConformanceGenerate() (err error) {
	suiteURL, err := url.ParseRequestURI(oidcConformanceBaseURL)
	if err != nil {
		return err
	}

	autheliaURL, err := url.ParseRequestURI(oidcConformanceAutheliaURL)
	if err != nil {
		return err
	}

	// The alias a plan registers is derived from the version, and it is also the path segment of the redirect URIs. A
	// random value per run means a rerun against an environment which was not torn down cannot collide on an alias.
	version := random.New().StringCustom(8, random.CharSetAlphabeticLower+random.CharSetNumeric)

	clients := &oidcConformanceClients{}
	clients.IdentityProviders.OIDC.Clients = []schema.IdentityProvidersOpenIDConnectClient{}

	var plans []OIDCConformancePlanFile

	for _, builder := range conformance.Builders(version, "implicit", "one_factor", "authelia", suiteURL, autheliaURL) {
		suite := builder.Build()

		clients.IdentityProviders.OIDC.Clients = append(clients.IdentityProviders.OIDC.Clients, suite.Clients...)
		plans = append(plans, OIDCConformancePlanFile{Name: builder.Name, PlanName: suite.Plan.Name, Variant: suite.Plan.Variant, Plan: suite.Plan})
	}

	if err = oidcConformanceWriteYAML(SuiteTmpPath(oidcConformanceClientsFile), clients); err != nil {
		return err
	}

	return oidcConformanceWriteJSON(SuiteTmpPath(oidcConformancePlansFile), plans)
}

// oidcConformanceWriteYAML encodes value to path as YAML. Unlike the repo's general convention it reports the close
// errors: this writes the configuration fragment Authelia boots from, and a write which is only discovered to have
// failed at close time would otherwise be reported as a successful setup and then surface as every plan failing for
// reasons which look like the conformance server's fault.
func oidcConformanceWriteYAML(path string, value any) (err error) {
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("error closing '%s': %w", path, cerr)
		}
	}()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)

	defer func() {
		if cerr := encoder.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("error flushing '%s': %w", path, cerr)
		}
	}()

	return encoder.Encode(value)
}

// oidcConformanceWriteJSON encodes value to path as JSON. It reports the close error for the same reason
// oidcConformanceWriteYAML does: the entire test process reads this file back and can do nothing without it.
func oidcConformanceWriteJSON(path string, value any) (err error) {
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("error closing '%s': %w", path, cerr)
		}
	}()

	return json.NewEncoder(f).Encode(value)
}

// oidcConformanceReadPlans reads the plans the setup process generated.
func oidcConformanceReadPlans() (plans []OIDCConformancePlanFile, err error) {
	data, err := os.ReadFile(SuiteTmpPath(oidcConformancePlansFile))
	if err != nil {
		return nil, fmt.Errorf("error reading the generated conformance plans, was the suite set up?: %w", err)
	}

	if err = json.Unmarshal(data, &plans); err != nil {
		return nil, err
	}

	return plans, nil
}

func init() {
	dockerEnvironment := NewDockerEnvironment([]string{
		"internal/suites/compose.yml",
		"internal/suites/OIDCConformance/compose.yml",
		"internal/suites/example/compose/authelia/compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/compose.yml",
		"internal/suites/example/compose/nginx/portal/compose.yml",
		"internal/suites/example/compose/smtp/compose.yml",
		"internal/suites/example/compose/redis/compose.yml",
	})

	setup := func(suitePath string) (err error) {
		// Generated before the stack comes up because Authelia loads the clients file at startup.
		if err = oidcConformanceGenerate(); err != nil {
			return err
		}

		if err = dockerEnvironment.Up(); err != nil {
			return err
		}

		if err = waitUntilAutheliaIsReady(dockerEnvironment, oidcConformanceSuiteName); err != nil {
			return err
		}

		client, err := NewConformanceClient(oidcConformanceBaseURL)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()

		if err = client.WaitReady(ctx); err != nil {
			return err
		}

		return updateDevEnvFileForDomain(BaseDomain, dockerEnvironment)
	}

	displayLogs := func() error {
		// conformance-mongodb is included because the conformance server will not finish starting without it, so it is
		// one of the answers to a setup which timed out waiting for the JVM to become ready.
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend", "conformance-mongodb", "conformance-server", "conformance-nginx")
	}

	teardown := func(suitePath string) error {
		return dockerEnvironment.Down()
	}

	// The three timeouts below are wall clock budgets which Buildkite's own step timeout has to contain: setup and
	// teardown wrap the authelia-scripts commands, and TestTimeout becomes `go test -timeout`. Their sum is what
	// .buildkite/steps/e2etests.sh has to sit above, and it is kept there deliberately - a step killed by Buildkite
	// produces no test output, no plan export and no container logs, on precisely the run which needed them.
	//
	// SetUpTimeout has to cover pulling a roughly 1 GB image, building the keytool layer, `compose up --wait` and up
	// to five minutes of ConformanceClient.WaitReady on a cold agent, so it is 15 rather than 10 minutes.
	//
	// TestTimeout has to cover conformancePlanTimeout (75m, the budget the runners themselves work to) plus
	// TearDownSuite: the two minute plan export loop and, on a failed run, up to seven drains of
	// conformanceDrainTimeout. 85 minutes covers that with room, and leaves the plan context - not the Go test
	// timeout - as the thing which ends a run that overruns, which is the one that produces usable output.
	GlobalRegistry.Register(oidcConformanceSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    15 * time.Minute,
		OnSetupTimeout:  displayLogs,
		OnError:         displayLogs,
		TestTimeout:     85 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 5 * time.Minute,
		Description:     "This suite runs the OpenID Foundation conformance suite against Authelia for every profile Authelia is OpenID Certified for.",
	})
}
