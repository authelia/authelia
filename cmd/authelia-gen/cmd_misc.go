package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"

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
		newMiscLocaleMoveCmd(),
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
	cmd.Flags().String("token", "", "conformance api token")
	cmd.Flags().StringSlice("suites", nil, "names of the plans to generate")
	cmd.Flags().String("consent", "implicit", "name of the consent mode to use")
	cmd.Flags().String("policy", "one_factor", "name of the authorization policy to use")
	cmd.Flags().String("brand", "authelia", "brand name to use")

	return cmd
}

func miscOIDCConformanceRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		rawURL, version, token, consent, policy, brand string

		autheliaURL, suiteURL *url.URL

		suiteNames []string
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

	if token, err = cmd.Flags().GetString("token"); err != nil {
		return err
	}

	if consent, err = cmd.Flags().GetString("consent"); err != nil {
		return err
	}

	if policy, err = cmd.Flags().GetString("policy"); err != nil {
		return err
	}

	if brand, err = cmd.Flags().GetString("brand"); err != nil {
		return err
	}

	return miscOIDCConformance(version, token, consent, policy, brand, autheliaURL, suiteURL, suiteNames...)
}

func miscOIDCConformance(version, token, consent, policy, brand string, autheliaURL, suiteURL *url.URL, suiteNames ...string) (err error) {
	suites := miscOIDCConformanceBuildSuites(version, consent, policy, brand, suiteURL, autheliaURL, suiteNames...)

	clients := &OpenIDConnectClients{}

	clients.IdentityProviders.OIDC.Clients = []schema.IdentityProvidersOpenIDConnectClient{}

	var client *http.Client

	if suiteURL != nil && len(token) != 0 {
		client = &http.Client{
			Transport: &RequestHeaderTransport{
				RoundTripper: http.DefaultTransport,
				headers: map[string]string{
					"Content-Type":  "application/json",
					"Authorization": fmt.Sprintf("Bearer %s", token),
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

func miscOIDCConformanceBuildSuites(version, consent, policy, brand string, suiteURL, autheliaURL *url.URL, suiteNames ...string) (suites []OpenIDConnectConformanceSuite) {
	builders := []*OpenIDConnectConformanceSuiteBuilder{
		{brand, "config", "Config", true, version, consent, policy, nil, autheliaURL},
		{brand, "basic", "Basic", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, suiteNameBasicFormPost, "Basic (Form Post)", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, "hybrid", "Hybrid", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, suiteNameHybridFormPost, "Hybrid (Form Post)", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, "implicit", "Implicit", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, suiteNameImplicitFormPost, "Implicit (Form Post)", true, version, consent, policy, suiteURL, autheliaURL},
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

func newMiscLocaleMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locale-move [key]",
		Short: "Move locales between namespaces",
		Args:  cobra.ExactArgs(1),
		RunE:  miscLocaleMoveRunE,

		DisableAutoGenTag: true,
	}

	cmd.Flags().StringP("source", "s", "", "locale source namespace")
	cmd.Flags().StringP("destination", "D", "", "locale destination namespace")

	return cmd
}

func miscLocaleMoveRunE(cmd *cobra.Command, args []string) (err error) {
	var (
		source      string
		destination string
		pathLocales string
	)

	if source, err = cmd.Flags().GetString("source"); err != nil {
		return err
	}

	if destination, err = cmd.Flags().GetString("destination"); err != nil {
		return err
	}

	if pathLocales, err = cmd.Flags().GetString(cmdFlagDirLocales); err != nil {
		return err
	}

	if source == "" || destination == "" {
		return fmt.Errorf("--source and --destination are required")
	}

	locales, err := os.ReadDir(pathLocales)
	if err != nil {
		return err
	}

	for _, locale := range locales {
		if err = miscLocaleMoveSingle(args[0], pathLocales, source, destination, locale); err != nil {
			return err
		}
	}

	return nil
}

func miscLocaleMoveSingle(key, pathLocales, source, destination string, locale os.DirEntry) (err error) {
	var (
		src, dst *os.File
	)

	if src, err = os.OpenFile(filepath.Join(pathLocales, locale.Name(), fmt.Sprintf("%s.json", source)), os.O_RDWR, 0644); err != nil {
		return err
	}

	defer src.Close()

	srcDecoder := json.NewDecoder(src)

	srcValues := map[string]any{}

	if err = srcDecoder.Decode(&srcValues); err != nil {
		return err
	}

	if dst, err = os.OpenFile(filepath.Join(pathLocales, locale.Name(), fmt.Sprintf("%s.json", destination)), os.O_RDWR, 0644); err != nil {
		return err
	}

	defer dst.Close()

	dstDecoder := json.NewDecoder(dst)

	dstValues := map[string]any{}

	if err = dstDecoder.Decode(&dstValues); err != nil {
		return err
	}

	var (
		value any
		ok    bool
	)

	if value, ok = srcValues[key]; !ok {
		return fmt.Errorf("locale key '%s' not found in source namespace '%s'", key, source)
	}

	delete(srcValues, key)

	dstValues[key] = value

	if err = src.Truncate(0); err != nil {
		return err
	}

	if _, err = src.Seek(0, 0); err != nil {
		return err
	}

	srcEncoder := json.NewEncoder(src)
	srcEncoder.SetIndent("", "\t")
	srcEncoder.SetEscapeHTML(false)

	if err = srcEncoder.Encode(srcValues); err != nil {
		return err
	}

	if err = dst.Truncate(0); err != nil {
		return err
	}

	if _, err = dst.Seek(0, 0); err != nil {
		return err
	}

	dstEncoder := json.NewEncoder(dst)
	dstEncoder.SetIndent("", "\t")
	dstEncoder.SetEscapeHTML(false)

	if err = dstEncoder.Encode(dstValues); err != nil {
		return err
	}

	return nil
}
