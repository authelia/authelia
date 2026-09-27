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
	"github.com/authelia/authelia/v4/internal/utils"
)

const (
	oidcConformanceSuiteName          = "OIDCConformance"
	oidcConformanceBaseURL            = "https://conformance.example.com:8443"
	oidcConformanceAutheliaURL        = "https://login.example.com:8080"
	oidcConformancePlansFile          = "conformance-plans.json"
	oidcConformanceClientsFile        = "conformance-clients.yml"
	oidcConformanceMongoDBHostEnv     = "SUITE_OIDC_CONFORMANCE_MONGODB_HOST"
	oidcConformanceMongoDBComposeFile = "OIDCConformance/compose.mongodb.yml"
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

type oidcConformanceClients struct {
	IdentityProviders struct {
		OIDC struct {
			Clients []schema.IdentityProvidersOpenIDConnectClient `yaml:"clients"`
		} `yaml:"oidc"`
	} `yaml:"identity_providers"`
}

func oidcConformanceGenerate() (err error) {
	suiteURL, err := url.ParseRequestURI(oidcConformanceBaseURL)
	if err != nil {
		return err
	}

	autheliaURL, err := url.ParseRequestURI(oidcConformanceAutheliaURL)
	if err != nil {
		return err
	}

	version := random.New().StringCustom(8, random.CharSetAlphabeticLower+random.CharSetNumeric)

	clients := &oidcConformanceClients{}
	clients.IdentityProviders.OIDC.Clients = []schema.IdentityProvidersOpenIDConnectClient{}

	var plans []OIDCConformancePlanFile

	release := oidcConformanceRelease(oidcConformanceRevision())

	for _, builder := range conformance.Builders(version, "implicit", "one_factor", "authelia", suiteURL, autheliaURL) {
		suite := builder.Build()

		if release != "" {
			suite.Plan.Description = builder.Description(release)
		}

		clients.IdentityProviders.OIDC.Clients = append(clients.IdentityProviders.OIDC.Clients, suite.Clients...)
		plans = append(plans, OIDCConformancePlanFile{Name: builder.Name, PlanName: suite.Plan.Name, Variant: suite.Plan.Variant, Plan: suite.Plan})
	}

	if err = oidcConformanceWriteYAML(SuiteTmpPath(oidcConformanceClientsFile), clients); err != nil {
		return err
	}

	return oidcConformanceWriteJSON(SuiteTmpPath(oidcConformancePlansFile), plans)
}

func oidcConformanceRevision() (commit, tag string) {
	commit, _, _ = utils.RunCommandAndReturnOutput("git rev-parse HEAD")
	tag, _, _ = utils.RunCommandAndReturnOutput("git describe --tags --exact-match HEAD")

	return commit, tag
}

func oidcConformanceRelease(commit, tag string) string {
	switch {
	case commit == "":
		return ""
	case tag == "":
		return "commit " + commit
	default:
		return fmt.Sprintf("%s (commit %s)", tag, commit)
	}
}

func oidcConformanceComposeFiles(mongodbHost string) (files []string) {
	files = []string{
		"internal/suites/compose.yml",
		"internal/suites/OIDCConformance/compose.yml",
		"internal/suites/example/compose/authelia/compose.backend.{}.yml",
		"internal/suites/example/compose/authelia/compose.frontend.{}.yml",
		"internal/suites/example/compose/nginx/backend/compose.yml",
		"internal/suites/example/compose/nginx/portal/compose.yml",
		"internal/suites/example/compose/smtp/compose.yml",
		"internal/suites/example/compose/redis/compose.yml",
		"internal/suites/example/compose/postgres/compose.yml",
	}

	if mongodbHost == "" {
		files = append(files, "internal/suites/"+oidcConformanceMongoDBComposeFile)
	}

	return files
}

func oidcConformanceLogServices(mongodbHost string) (services []string) {
	services = []string{"authelia-backend", "authelia-frontend", "postgres"}

	if mongodbHost == "" {
		services = append(services, "conformance-mongodb")
	}

	return append(services, "conformance-server")
}

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
	mongodbHost := os.Getenv(oidcConformanceMongoDBHostEnv)

	dockerEnvironment := NewDockerEnvironment(oidcConformanceComposeFiles(mongodbHost))

	setup := func(suitePath string) (err error) {
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
		return dockerEnvironment.PrintLogs(oidcConformanceLogServices(mongodbHost)...)
	}

	teardown := func(suitePath string) error {
		return dockerEnvironment.Down()
	}

	GlobalRegistry.Register(oidcConformanceSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    2 * time.Minute,
		OnSetupTimeout:  displayLogs,
		OnError:         displayLogs,
		TestTimeout:     8 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 1 * time.Minute,
		Description:     "This suite runs the OpenID Foundation conformance suite against Authelia for every profile Authelia is OpenID Certified for.",
	})
}
