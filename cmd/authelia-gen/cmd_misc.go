package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func newMiscCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdUseMisc,
		Short: "Generate miscellaneous things",

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newMiscOIDCCmd(),
	)

	return cmd
}

func newMiscOIDCCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oidc",
		Short: "Generate OpenID Connect 1.0 configurations",

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(
		newMiscOIDCConformanceCmd(),
	)

	return cmd
}

func newMiscOIDCConformanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conformance",
		Short: "Generate OpenID Connect 1.0 conformance configurations",
		RunE:  miscOIDCConformanceRunE,

		DisableAutoGenTag: true,
	}

	cmd.Flags().String("authelia-url", "https://auth.example.com", "authelia url for conformance plans")
	cmd.Flags().String("version", "", "version name")
	cmd.Flags().String("url", "https://conformance.example.com", "conformance suite url for conformance plans")
	cmd.Flags().String("api-key", "", "conformance api key")
	cmd.Flags().StringSlice("suites", nil, "names of the plans to generate")

	return cmd
}

func miscOIDCConformanceRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		rawURL, version, apikey string
		autheliaURL, suiteURL   *url.URL
		suiteNames              []string
	)

	if rawURL, err = cmd.Flags().GetString("authelia-url"); err != nil {
		return err
	} else if autheliaURL, err = url.ParseRequestURI(rawURL); err != nil {
		return err
	}

	if rawURL, err = cmd.Flags().GetString("url"); err != nil {
		return err
	} else if suiteURL, err = url.ParseRequestURI(rawURL); err != nil {
		return err
	}

	if version, err = cmd.Flags().GetString("version"); err != nil {
		return err
	}

	if suiteNames, err = cmd.Flags().GetStringSlice("suites"); err != nil {
		return err
	}

	if apikey, err = cmd.Flags().GetString("api-key"); err != nil {
		return err
	}

	return miscOIDCConformance(version, apikey, autheliaURL, suiteURL, suiteNames...)
}

func miscOIDCConformance(version, apikey string, autheliaURL, suiteURL *url.URL, suiteNames ...string) (err error) {
	suites := miscOIDCConformanceBuildSuites(version, suiteURL, autheliaURL, suiteNames...)

	clients := &OpenIDConnectClients{}

	clients.IdentityProviders.OIDC.Clients = []schema.IdentityProvidersOpenIDConnectClient{}

	var client *http.Client

	if suiteURL != nil && len(apikey) != 0 {
		client = &http.Client{
			Transport: &RequestHeaderTransport{
				RoundTripper: http.DefaultTransport,
				headers: map[string]string{
					"Content-Type":  "application/json",
					"Authorization": fmt.Sprintf("Bearer %s", apikey),
				},
			},
		}
	}

	var (
		f   *os.File
		buf *bytes.Buffer
	)

	for _, suite := range suites {
		if f, err = os.OpenFile(fmt.Sprintf("%s%s", suite.Name, extJSON), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			return err
		}

		defer f.Close()

		buf = bytes.NewBuffer(nil)

		encoder := json.NewEncoder(io.MultiWriter(f, buf))

		encoder.SetIndent("", "  ")

		if err = encoder.Encode(&suite.Plan); err != nil {
			return err
		}

		if err = doOIDCConformanceSuitePostPlan(client, suiteURL, suite.Plan.Name, suite.Plan.Variant, buf); err != nil {
			return err
		}

		clients.IdentityProviders.OIDC.Clients = append(clients.IdentityProviders.OIDC.Clients, suite.Clients...)
	}

	if f, err = os.OpenFile(fmt.Sprintf("%s%s", "conformance-clients", extYAML), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return err
	}

	defer f.Close()

	encoder := yaml.NewEncoder(f)

	encoder.SetIndent(2)

	if err = encoder.Encode(clients); err != nil {
		return err
	}

	return nil
}

func miscOIDCConformanceBuildSuites(version string, suiteURL, autheliaURL *url.URL, suiteNames ...string) (suites []OpenIDConnectConformanceSuite) {
	builders := []*OpenIDConnectConformanceSuiteBuilder{
		{"config", "Config", true, version, nil, autheliaURL},
		{"basic", "Basic", true, version, suiteURL, autheliaURL},
		{suiteNameBasicFormPost, "Basic (Form Post)", true, version, suiteURL, autheliaURL},
		{"hybrid", "Hybrid", true, version, suiteURL, autheliaURL},
		{suiteNameHybridFormPost, "Hybrid (Form Post)", true, version, suiteURL, autheliaURL},
		{"implicit", "Implicit", true, version, suiteURL, autheliaURL},
		{suiteNameImplicitFormPost, "Implicit (Form Post)", true, version, suiteURL, autheliaURL},
	}

	for _, builder := range builders {
		if len(suiteNames) != 0 && !utils.IsStringInSlice(builder.name, suiteNames) {
			continue
		}

		suites = append(suites, builder.Build())
	}

	return suites
}

func doOIDCConformanceSuitePostPlan(client *http.Client, base *url.URL, plan string, variant *OpenIDConnectConformanceSuitePlanVariant, body *bytes.Buffer) (err error) {
	if client == nil {
		return nil
	}

	uri := base.JoinPath("api", "plan")

	query := uri.Query()

	query.Set("planName", plan)

	if variant != nil {
		var dataVariant []byte

		if dataVariant, err = json.Marshal(variant); err != nil {
			return err
		}

		query.Set("variant", string(dataVariant))
	}

	uri.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodPost, uri.String(), body)
	if err != nil {
		return err
	}

	var resp *http.Response

	if resp, err = client.Do(req); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

type RequestHeaderTransport struct {
	http.RoundTripper

	headers map[string]string
}

func (t *RequestHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	return t.RoundTripper.RoundTrip(req)
}

type OpenIDConnectClients struct {
	IdentityProviders struct {
		OIDC struct {
			Clients []schema.IdentityProvidersOpenIDConnectClient `yaml:"clients"`
		} `yaml:"oidc"`
	} `yaml:"identity_providers"`
}
