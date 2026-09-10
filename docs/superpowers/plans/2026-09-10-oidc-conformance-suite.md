<!--
SPDX-FileCopyrightText: 2026 Authelia

SPDX-License-Identifier: Apache-2.0
-->

# OIDC Conformance Suite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an Authelia integration suite named `OIDCConformance` that runs the OpenID Foundation conformance suite in Docker, creates the seven certification test plans Authelia is certified for over its HTTP API, drives every module with go-rod, and asserts each module's result as a Go subtest.

**Architecture:** The plan/client configuration is generated once from a builder shared with `authelia-gen` and crosses the `authelia-scripts` / `go test` process boundary as files in `SuiteTmpPath()`. In the test process, one goroutine per profile owns a conformance plan and an isolated rod browser context, walks its modules serially, and publishes outcomes on a channel that the matching testify method drains and asserts.

**Tech Stack:** Go 1.25, testify suites, go-rod v0.116.2, Docker Compose, the OpenID Foundation conformance suite (`registry.gitlab.com/openid/conformance-suite:release-v5.2.4`), MongoDB 6.0.13.

**Spec:** `docs/superpowers/specs/2026-09-10-oidc-conformance-suite-design.md`

## Global Constraints

- Every new file needs a REUSE header. Go and shell: `// SPDX-FileCopyrightText: 2026 Authelia`, blank comment line, `// SPDX-License-Identifier: Apache-2.0`. YAML: the same with `#`. Copy the header verbatim from an existing sibling file. `reuse lint --lines` runs in `lefthook` pre-commit and in `.buildkite/steps/lint.sh` and must pass.
- Conformance suite images are pinned to `release-v5.2.4` — `registry.gitlab.com/openid/conformance-suite:release-v5.2.4` and `registry.gitlab.com/openid/conformance-suite/nginx:release-v5.2.4`. Never use `latest`.
- MongoDB is pinned to `mongo:6.0.13`, matching the conformance suite's own compose file.
- The conformance base URL is exactly `https://conformance.example.com:8443`, everywhere, with no trailing slash. Authelia's URL is exactly `https://login.example.com:8080`. These strings appear in the plan JSON, the registered redirect URIs and the container environment, and must be byte-identical in all three.
- The conformance server container must take the compose network alias `server`, because the published nginx image's baked-in configuration proxies to `http://server:8080`.
- The conformance nginx container takes address `${SUITE_SUBNET:-192.168.240}.140` and the alias `conformance.example.com`.
- Accepted module results are `PASSED`, `REVIEW` and `SKIPPED`. `WARNING` and `FAILED` fail.
- Suite name is `OIDCConformance`, so the top-level test function must be named exactly `TestOIDCConformanceSuite` — `authelia-scripts` runs `go test -run '^(Test<SuiteName>Suite)$'`.
- Never call `require`/`assert` or any `RodSession` helper taking `*testing.T` from a goroutine that is not the test's own. All runner and browser-driver code returns errors instead.
- Run `go build ./...` and `golangci-lint run` before each commit. `golangci-lint` is configured at the repository root.

---

### Task 1: Extract the conformance builder to `internal/oidc/conformance`

The builder that produces both the conformance plan JSON and the matching Authelia client definitions currently lives in `package main` under `cmd/authelia-gen`, so the suite cannot import it. Move it verbatim, with its tests, and have `authelia-gen` import it. Behavior must not change.

**Files:**

- Create: `internal/oidc/conformance/const.go`
- Create: `internal/oidc/conformance/types.go`
- Create: `internal/oidc/conformance/builder.go`
- Create: `internal/oidc/conformance/builder_test.go`
- Modify: `cmd/authelia-gen/openid_conformance.go` (delete; contents move)
- Modify: `cmd/authelia-gen/openid_conformance_test.go` (delete; contents move)
- Modify: `cmd/authelia-gen/types.go` (remove the moved types)
- Modify: `cmd/authelia-gen/const.go` (remove the moved constants)
- Modify: `cmd/authelia-gen/cmd_misc.go` (use the imported package)

**Interfaces:**

- Consumes: nothing.
- Produces:
  - `conformance.SuiteBuilder` struct with exported fields `Brand`, `Name`, `Friendly`, `Certification`, `Version`, `Consent`, `Policy`, `SuiteURL`, `AutheliaURL` and method `Build() conformance.Suite`.
  - `conformance.Suite` with fields `Name string`, `Plan conformance.Plan`, `Clients []schema.IdentityProvidersOpenIDConnectClient`.
  - `conformance.Plan` with the JSON tags the API expects, including `Name string` (json:"-") and `Variant *conformance.PlanVariant` (json:"-").
  - `conformance.MustHash(value string) *schema.PasswordDigest`.
  - `conformance.Builders(version, consent, policy, brand string, suiteURL, autheliaURL *url.URL) []*conformance.SuiteBuilder` returning the seven builders in a fixed order.
  - Constants `conformance.NameConfig`, `NameBasic`, `NameBasicFormPost`, `NameHybrid`, `NameHybridFormPost`, `NameImplicit`, `NameImplicitFormPost`.

- [ ] **Step 1: Read what is being moved**

Read these first so the move is mechanical rather than a rewrite:

```bash
cat cmd/authelia-gen/openid_conformance.go
sed -n '233,315p' cmd/authelia-gen/types.go
sed -n '44,60p' cmd/authelia-gen/const.go
sed -n '124,250p' cmd/authelia-gen/cmd_misc.go
```

- [ ] **Step 2: Create the constants file**

Create `internal/oidc/conformance/const.go`:

```go
// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Package conformance provides the OpenID Connect 1.0 conformance suite plan and client definitions Authelia is
// certified against. It is shared by the authelia-gen conformance generator and the OIDCConformance integration suite
// so that a plan's client identifiers, secrets and redirect URIs cannot drift from the clients Authelia is configured
// with.
package conformance

const (
	// NameConfig is the builder name of the Config OP profile.
	NameConfig = "config"

	// NameBasic is the builder name of the Basic OP profile.
	NameBasic = "basic"

	// NameBasicFormPost is the builder name of the Basic OP profile using the form post response mode.
	NameBasicFormPost = "basic-form-post"

	// NameHybrid is the builder name of the Hybrid OP profile.
	NameHybrid = "hybrid"

	// NameHybridFormPost is the builder name of the Hybrid OP profile using the form post response mode.
	NameHybridFormPost = "hybrid-form-post"

	// NameImplicit is the builder name of the Implicit OP profile.
	NameImplicit = "implicit"

	// NameImplicitFormPost is the builder name of the Implicit OP profile using the form post response mode.
	NameImplicitFormPost = "implicit-form-post"
)

const (
	planConfig           = "conformance-config"
	planBasic            = "conformance-basic"
	planBasicFormPost    = "conformance-basic-form-post"
	planImplicit         = "conformance-implicit"
	planImplicitFormPost = "conformance-implicit-form-post"
	planHybrid           = "conformance-hybrid"
	planHybridFormPost   = "conformance-hybrid-form-post"
)
```

- [ ] **Step 3: Move the types**

Create `internal/oidc/conformance/types.go` with the REUSE header, `package conformance`, and the types currently at `cmd/authelia-gen/types.go:233-315`, renamed by dropping the `OpenIDConnectConformanceSuite` prefix:

| Old name                                     | New name        |
| -------------------------------------------- | --------------- |
| `OpenIDConnectConformanceSuite`              | `Suite`         |
| `OpenIDConnectConformanceSuitePlan`          | `Plan`          |
| `OpenIDConnectConformanceSuitePlanVariant`   | `PlanVariant`   |
| `OpenIDConnectConformanceSuitePlanServer`    | `PlanServer`    |
| `OpenIDConnectConformanceSuitePlanClient`    | `PlanClient`    |
| `OpenIDConnectConformanceSuitePlanMutualTLS` | `PlanMutualTLS` |
| `OpenIDConnectConformanceSuitePlanResource`  | `PlanResource`  |

Keep every struct tag exactly as it is. The JSON tags are the wire format the conformance API consumes, so a changed tag is a silent behavior change. Give each exported type a doc comment beginning with its name.

- [ ] **Step 4: Move the builder**

Create `internal/oidc/conformance/builder.go` with the REUSE header, `package conformance`, and the contents of `cmd/authelia-gen/openid_conformance.go`, with these changes and no others:

- `OpenIDConnectConformanceSuiteBuilder` becomes `SuiteBuilder` and its fields become exported: `Brand`, `Name`, `Friendly`, `Certification`, `Version`, `Consent`, `Policy`, `SuiteURL`, `AutheliaURL`.
- The `switch name` and `switch b.Name` statements use the new constants from `const.go`.
- `MustHash` keeps its name and doc comment.

Then append the builder list, moved out of `cmd_misc.go:191-209` so both callers share one ordering:

```go
// Builders returns the conformance suite builders for every profile Authelia is certified for, in a fixed order.
func Builders(version, consent, policy, brand string, suiteURL, autheliaURL *url.URL) []*SuiteBuilder {
	return []*SuiteBuilder{
		{brand, NameConfig, "Config", true, version, consent, policy, nil, autheliaURL},
		{brand, NameBasic, "Basic", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, NameBasicFormPost, "Basic (Form Post)", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, NameHybrid, "Hybrid", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, NameHybridFormPost, "Hybrid (Form Post)", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, NameImplicit, "Implicit", true, version, consent, policy, suiteURL, autheliaURL},
		{brand, NameImplicitFormPost, "Implicit (Form Post)", true, version, consent, policy, suiteURL, autheliaURL},
	}
}
```

- [ ] **Step 5: Move the tests**

`git mv cmd/authelia-gen/openid_conformance_test.go internal/oidc/conformance/builder_test.go`, change the package clause to `package conformance`, and apply the same renames. The test is 526 lines of table-driven expectations covering all seven profiles; it is the proof the move preserved behavior, so change names only — never an expected value.

- [ ] **Step 6: Run the moved tests to verify they pass unchanged**

Run: `go test ./internal/oidc/conformance/... -v`
Expected: PASS, including `TestMustHash` and every `TestOpenIDConnectConformanceSuiteBuilder_Build` case.

If any expected value needs editing to make this pass, the move was not mechanical — revert that edit and fix the code instead.

- [ ] **Step 7: Update authelia-gen to import the package**

Delete `cmd/authelia-gen/openid_conformance.go`, delete the moved types from `cmd/authelia-gen/types.go` and the moved constants from `cmd/authelia-gen/const.go`, then in `cmd/authelia-gen/cmd_misc.go`:

- Add the import `"github.com/authelia/authelia/v4/internal/oidc/conformance"`.
- Replace `miscOIDCConformanceBuildSuites`'s body with a call to `conformance.Builders(...)` followed by the existing `suiteNames` filter, returning `[]conformance.Suite`.
- Change `doOIDCConformanceSuitePostPlan`'s `variant` parameter type to `*conformance.PlanVariant`.
- Leave every other line of the file alone, including the flag definitions and the file output.

- [ ] **Step 8: Verify authelia-gen still builds and passes**

Run: `go build ./... && go test ./cmd/authelia-gen/...`
Expected: PASS.

- [ ] **Step 9: Verify the generator's output is unchanged**

```bash
cd "$(mktemp -d)" && go run github.com/authelia/authelia/v4/cmd/authelia-gen misc oidc conformance --version 4.40 && ls
```

Expected: seven `conformance-*.json` files plus `conformance-clients.yaml`. Spot-check that `conformance-basic.json` has an `alias` of `conformance-basic-authelia440` and a `discoveryUrl` under `auth.example.com`.

- [ ] **Step 10: Commit**

```bash
golangci-lint run
git add internal/oidc/conformance cmd/authelia-gen
git commit -m "refactor(oidc): extract conformance builder to internal package"
```

---

### Task 2: Conformance HTTP API client

A typed client for the conformance suite's HTTP API. Devmode disables the suite's own OAuth login, so no token is needed. Two server behaviors drive the design: `wait-state` clamps its timeout to 30 seconds server-side and returns 404 once a module has finished and left the running registry, and unfilled image placeholders appear as log entries carrying an `upload` field.

**Files:**

- Create: `internal/suites/conformance_api.go`
- Test: `internal/suites/conformance_api_test.go`

**Interfaces:**

- Consumes: `NewHTTPTransport()` from `internal/suites/http.go`; `conformance.Plan` from Task 1.
- Produces:
  - `type ConformancePlanModule struct { TestModule string; Variant map[string]string }`
  - `type ConformanceCreatedPlan struct { ID string; Modules []ConformancePlanModule }`
  - `type ConformanceTestInfo struct { ID, TestName, Status, Result string }`
  - `type ConformanceBrowserStatus struct { URLs []string; Visited []string }`
  - `type ConformanceLogEntry struct { Upload string; Msg string; Result string }`
  - `func NewConformanceClient(base string) (*ConformanceClient, error)`
  - `func (c *ConformanceClient) WaitReady(ctx context.Context) error`
  - `func (c *ConformanceClient) CreatePlan(ctx context.Context, plan *conformance.Plan) (*ConformanceCreatedPlan, error)`
  - `func (c *ConformanceClient) CreateTest(ctx context.Context, planID, module string, variant map[string]string) (string, error)`
  - `func (c *ConformanceClient) StartTest(ctx context.Context, id string) error`
  - `func (c *ConformanceClient) WaitState(ctx context.Context, id string, states ...string) (string, error)`
  - `func (c *ConformanceClient) Info(ctx context.Context, id string) (*ConformanceTestInfo, error)`
  - `func (c *ConformanceClient) BrowserStatus(ctx context.Context, id string) (*ConformanceBrowserStatus, error)`
  - `func (c *ConformanceClient) Visit(ctx context.Context, id, url string) error`
  - `func (c *ConformanceClient) Log(ctx context.Context, id string) ([]ConformanceLogEntry, error)`
  - `func (c *ConformanceClient) UploadPlaceholder(ctx context.Context, id, placeholder, dataURI string) error`
  - `func (c *ConformanceClient) LogDetailURL(id string) string`
  - `func (c *ConformanceClient) ExportPlanHTML(ctx context.Context, planID, path string) error`

- [ ] **Step 1: Write the failing tests**

Create `internal/suites/conformance_api_test.go`:

```go
// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/oidc/conformance"
)

func TestConformanceClient_CreatePlan(t *testing.T) {
	var query string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/plan", r.URL.Path)

		query = r.URL.RawQuery

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"_id":"plan1","modules":[{"testModule":"oidcc-server","variant":{"response_type":"code"}}]}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	plan := &conformance.Plan{Name: "oidcc-basic-certification-test-plan", Alias: "alias", Variant: &conformance.PlanVariant{ServerMetadata: "discovery"}}

	created, err := client.CreatePlan(context.Background(), plan)
	require.NoError(t, err)

	assert.Equal(t, "plan1", created.ID)
	require.Len(t, created.Modules, 1)
	assert.Equal(t, "oidcc-server", created.Modules[0].TestModule)
	assert.Equal(t, "code", created.Modules[0].Variant["response_type"])
	assert.Contains(t, query, "planName=oidcc-basic-certification-test-plan")
	assert.Contains(t, query, "variant=")
}

func TestConformanceClient_WaitStateReturnsPersistedStatusWhenTestNoLongerRunning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/runner/abc/wait-state":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"test not found"}`))
		case "/api/info/abc":
			_, _ = w.Write([]byte(`{"_id":"abc","testName":"oidcc-server","status":"FINISHED","result":"PASSED"}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	state, err := client.WaitState(ctx, "abc", "FINISHED")
	require.NoError(t, err)
	assert.Equal(t, "FINISHED", state)
}

func TestConformanceClient_WaitStateRetriesOnTimeout(t *testing.T) {
	var calls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/runner/abc/wait-state", r.URL.Path)

		calls++

		if calls == 1 {
			_, _ = w.Write([]byte(`{"timeout":true}`))

			return
		}

		_, _ = w.Write([]byte(`{"state":"WAITING"}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	state, err := client.WaitState(ctx, "abc", "WAITING")
	require.NoError(t, err)
	assert.Equal(t, "WAITING", state)
	assert.Equal(t, 2, calls)
}

func TestConformanceClient_CreateTestReportsAliasCollision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"alias already in use"}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	_, err = client.CreateTest(context.Background(), "plan1", "oidcc-server", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alias already in use")
}

func TestConformanceClient_LogExposesUnfilledPlaceholders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/log/abc", r.URL.Path)

		_, _ = w.Write([]byte(`[{"msg":"a"},{"msg":"upload me","upload":"ph1","result":"REVIEW"}]`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	entries, err := client.Log(context.Background(), "abc")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "", entries[0].Upload)
	assert.Equal(t, "ph1", entries[1].Upload)
}

func TestConformanceClient_UploadPlaceholder(t *testing.T) {
	var body []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/log/abc/images/ph1", r.URL.Path)

		body, _ = io.ReadAll(r.Body)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	require.NoError(t, client.UploadPlaceholder(context.Background(), "abc", "ph1", conformancePlaceholderImage))
	assert.Equal(t, conformancePlaceholderImage, string(body))
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/suites/ -run TestConformanceClient -v`
Expected: FAIL to compile — `undefined: NewConformanceClient`.

- [ ] **Step 3: Write the client**

Create `internal/suites/conformance_api.go`:

```go
// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/oidc/conformance"
	"github.com/authelia/authelia/v4/internal/utils"
)

const (
	// conformanceWaitStateTimeout is the per-request long poll budget. The server clamps this to 30 seconds, so a
	// longer value would silently become 30 and a shorter one only wastes round trips.
	conformanceWaitStateTimeout = time.Second * 30

	// conformanceReadyInterval is how often the JVM is asked whether it has finished starting.
	conformanceReadyInterval = time.Second * 2

	// conformancePlaceholderImage is a 1x1 transparent PNG. Modules which call waitForPlaceholders() hold in WAITING
	// until an image is uploaded, and the endpoint accepts only PNG or JPEG data URIs. The suite verifies those pages
	// with go-rod instead, so this image carries no information and exists only to release the wait.
	conformancePlaceholderImage = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
)

// ConformancePlanModule is one entry in a conformance test plan's module list.
type ConformancePlanModule struct {
	TestModule string            `json:"testModule"`
	Variant    map[string]string `json:"variant"`
}

// ConformanceCreatedPlan is the response to creating a conformance test plan.
type ConformanceCreatedPlan struct {
	ID      string                  `json:"_id"`
	Modules []ConformancePlanModule `json:"modules"`
}

// ConformanceTestInfo is the persisted state of a single conformance test module instance.
type ConformanceTestInfo struct {
	ID       string `json:"_id"`
	TestName string `json:"testName"`
	Status   string `json:"status"`
	Result   string `json:"result"`
}

// ConformanceBrowserStatus describes the URLs a module is waiting for a browser to visit.
type ConformanceBrowserStatus struct {
	URLs    []string `json:"urls"`
	Visited []string `json:"visited"`
}

// ConformanceLogEntry is one entry of a module's log. An entry with a non-empty Upload is an image placeholder which
// has not been filled, and which is holding the module in WAITING.
type ConformanceLogEntry struct {
	Msg    string `json:"msg"`
	Result string `json:"result"`
	Upload string `json:"upload"`
}

// ConformanceClient talks to the OpenID Foundation conformance suite's HTTP API. The suite runs with
// fintechlabs.devmode enabled, which disables its OAuth login, so no credentials are sent.
type ConformanceClient struct {
	base   *url.URL
	client *http.Client
}

// NewConformanceClient returns a ConformanceClient for the conformance suite at base.
func NewConformanceClient(base string) (client *ConformanceClient, err error) {
	var parsed *url.URL

	if parsed, err = url.ParseRequestURI(base); err != nil {
		return nil, fmt.Errorf("error parsing conformance base url: %w", err)
	}

	return &ConformanceClient{
		base: parsed,
		client: &http.Client{
			Transport: NewHTTPTransport(),
			Timeout:   conformanceWaitStateTimeout + time.Second*10,
		},
	}, nil
}

func (c *ConformanceClient) uri(query url.Values, elem ...string) string {
	uri := c.base.JoinPath(append([]string{"api"}, elem...)...)

	if query != nil {
		uri.RawQuery = query.Encode()
	}

	return uri.String()
}

// LogDetailURL returns the human readable log page for a module, for inclusion in failure output.
func (c *ConformanceClient) LogDetailURL(id string) string {
	uri := c.base.JoinPath("log-detail.html")

	query := uri.Query()
	query.Set("log", id)
	uri.RawQuery = query.Encode()

	return uri.String()
}

func (c *ConformanceClient) do(ctx context.Context, method, uri string, body io.Reader, contentType string, expected int, out any) (err error) {
	var req *http.Request

	if req, err = http.NewRequestWithContext(ctx, method, uri, body); err != nil {
		return err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	var resp *http.Response

	if resp, err = c.client.Do(req); err != nil {
		return err
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != expected {
		return fmt.Errorf("%s %s: expected status %d but got %d: %s", method, uri, expected, resp.StatusCode, data)
	}

	if out == nil {
		return nil
	}

	return json.Unmarshal(data, out)
}

// WaitReady blocks until the conformance server answers an API request or ctx expires. The JVM binds its port well
// before Spring has finished starting, so a connection check is not sufficient.
func (c *ConformanceClient) WaitReady(ctx context.Context) (err error) {
	query := url.Values{}
	query.Set("length", "1")

	for {
		if err = c.do(ctx, http.MethodGet, c.uri(query, "plan"), nil, "", http.StatusOK, nil); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("conformance suite was not ready: %w (last error: %v)", ctx.Err(), err)
		case <-time.After(conformanceReadyInterval):
		}
	}
}

// CreatePlan creates a conformance test plan and returns its id and module list.
func (c *ConformanceClient) CreatePlan(ctx context.Context, plan *conformance.Plan) (created *ConformanceCreatedPlan, err error) {
	body, err := json.Marshal(plan)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("planName", plan.Name)

	if plan.Variant != nil {
		var variant []byte

		if variant, err = json.Marshal(plan.Variant); err != nil {
			return nil, err
		}

		query.Set("variant", string(variant))
	}

	created = &ConformanceCreatedPlan{}

	if err = c.do(ctx, http.MethodPost, c.uri(query, "plan"), bytes.NewReader(body), "application/json", http.StatusCreated, created); err != nil {
		return nil, err
	}

	return created, nil
}

// CreateTest creates an instance of a plan's module and returns its id. The variant is sent alongside the plan even
// though the endpoint documents it as applying to standalone tests: the upstream scripts/conformance.py sends both
// together for plan modules, and it is what makes plans which repeat a module across variants work.
func (c *ConformanceClient) CreateTest(ctx context.Context, planID, module string, variant map[string]string) (id string, err error) {
	query := url.Values{}
	query.Set("test", module)
	query.Set("plan", planID)

	if len(variant) != 0 {
		var data []byte

		if data, err = json.Marshal(variant); err != nil {
			return "", err
		}

		query.Set("variant", string(data))
	}

	out := struct {
		ID string `json:"id"`
	}{}

	if err = c.do(ctx, http.MethodPost, c.uri(query, "runner"), nil, "application/json", http.StatusCreated, &out); err != nil {
		return "", err
	}

	return out.ID, nil
}

// StartTest starts a module which is waiting in the CONFIGURED state.
func (c *ConformanceClient) StartTest(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodPost, c.uri(nil, "runner", id), nil, "", http.StatusOK, nil)
}

// WaitState long polls until the module reaches one of states, or ctx expires. The server clamps each poll to 30
// seconds and answers {"timeout":true} when it expires, so this loops. Once a module finishes it leaves the running
// registry and the endpoint answers 404; the persisted status from Info is authoritative at that point.
func (c *ConformanceClient) WaitState(ctx context.Context, id string, states ...string) (state string, err error) {
	query := url.Values{}
	query.Set("states", strings.Join(states, ","))
	query.Set("timeoutMs", strconv.FormatInt(conformanceWaitStateTimeout.Milliseconds(), 10))

	for {
		out := struct {
			State   string `json:"state"`
			Timeout bool   `json:"timeout"`
		}{}

		err = c.do(ctx, http.MethodGet, c.uri(query, "runner", id, "wait-state"), nil, "", http.StatusOK, &out)

		switch {
		case err == nil && out.State != "":
			return out.State, nil
		case err == nil && out.Timeout:
			// The poll expired without a transition. Fall through to the context check and poll again.
		case err != nil:
			var info *ConformanceTestInfo

			if info, err = c.Info(ctx, id); err != nil {
				return "", err
			}

			if utils.IsStringInSlice(info.Status, states) {
				return info.Status, nil
			}

			return "", fmt.Errorf("module '%s' is no longer running and its persisted status is '%s', which is not one of %v", id, info.Status, states)
		}

		if ctx.Err() != nil {
			return "", fmt.Errorf("module '%s' did not reach one of %v: %w", id, states, ctx.Err())
		}
	}
}

// Info returns the persisted status and result of a module.
func (c *ConformanceClient) Info(ctx context.Context, id string) (info *ConformanceTestInfo, err error) {
	info = &ConformanceTestInfo{}

	if err = c.do(ctx, http.MethodGet, c.uri(nil, "info", id), nil, "", http.StatusOK, info); err != nil {
		return nil, err
	}

	return info, nil
}

// BrowserStatus returns the URLs a module is waiting to have visited.
func (c *ConformanceClient) BrowserStatus(ctx context.Context, id string) (status *ConformanceBrowserStatus, err error) {
	status = &ConformanceBrowserStatus{}

	if err = c.do(ctx, http.MethodGet, c.uri(nil, "runner", "browser", id), nil, "", http.StatusOK, status); err != nil {
		return nil, err
	}

	return status, nil
}

// Visit marks a URL as visited. The server matches by exact string equality, so uri must be the URL as it was given.
func (c *ConformanceClient) Visit(ctx context.Context, id, uri string) error {
	query := url.Values{}
	query.Set("url", uri)

	return c.do(ctx, http.MethodPost, c.uri(query, "runner", "browser", id, "visit"), nil, "", http.StatusNoContent, nil)
}

// Log returns a module's log entries.
func (c *ConformanceClient) Log(ctx context.Context, id string) (entries []ConformanceLogEntry, err error) {
	if err = c.do(ctx, http.MethodGet, c.uri(nil, "log", id), nil, "", http.StatusOK, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// UploadPlaceholder fills an image placeholder, releasing a module which is blocked in waitForPlaceholders().
func (c *ConformanceClient) UploadPlaceholder(ctx context.Context, id, placeholder, dataURI string) error {
	return c.do(ctx, http.MethodPost, c.uri(nil, "log", id, "images", placeholder), bytes.NewReader([]byte(dataURI)), "text/plain", http.StatusOK, nil)
}

// ExportPlanHTML writes a plan's HTML export to path, for collection as a CI artifact.
func (c *ConformanceClient) ExportPlanHTML(ctx context.Context, planID, path string) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.uri(nil, "plan", "exporthtml", planID), nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d exporting plan '%s'", resp.StatusCode, planID)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, resp.Body)

	return err
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/suites/ -run TestConformanceClient -v`
Expected: PASS, six tests.

- [ ] **Step 5: Commit**

```bash
golangci-lint run
git add internal/suites/conformance_api.go internal/suites/conformance_api_test.go
git commit -m "test(suites): add oidc conformance api client"
```

---

### Task 3: Subtest naming

Turn a plan's module list into PascalCase subtest names, stripping the `oidcc-` prefix and appending a variant suffix only where a plan lists the same module more than once.

**Files:**

- Create: `internal/suites/conformance_names.go`
- Test: `internal/suites/conformance_names_test.go`

**Interfaces:**

- Consumes: `ConformancePlanModule` from Task 2.
- Produces: `func ConformanceSubtestNames(modules []ConformancePlanModule) []string` returning one name per module, index-aligned with the input.

- [ ] **Step 1: Write the failing test**

Create `internal/suites/conformance_names_test.go`:

```go
// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConformanceSubtestNames(t *testing.T) {
	testCases := []struct {
		name     string
		have     []ConformancePlanModule
		expected []string
	}{
		{
			"ShouldStripPrefixAndPascalCase",
			[]ConformancePlanModule{{TestModule: "oidcc-server"}},
			[]string{"Server"},
		},
		{
			"ShouldPascalCaseEverySegment",
			[]ConformancePlanModule{{TestModule: "oidcc-ensure-request-without-nonce-succeeds-for-code-flow"}},
			[]string{"EnsureRequestWithoutNonceSucceedsForCodeFlow"},
		},
		{
			"ShouldKeepDigits",
			[]ConformancePlanModule{{TestModule: "oidcc-max-age-10000"}},
			[]string{"MaxAge10000"},
		},
		{
			"ShouldHandleModulesWithoutThePrefix",
			[]ConformancePlanModule{{TestModule: "oidcc-config-certification"}, {TestModule: "discovery-issuer-not-matching-config"}},
			[]string{"ConfigCertification", "DiscoveryIssuerNotMatchingConfig"},
		},
		{
			"ShouldNotSuffixUniqueModules",
			[]ConformancePlanModule{
				{TestModule: "oidcc-server", Variant: map[string]string{"response_type": "code"}},
				{TestModule: "oidcc-scope-address", Variant: map[string]string{"response_type": "code"}},
			},
			[]string{"Server", "ScopeAddress"},
		},
		{
			"ShouldSuffixRepeatedModulesWithTheirVariant",
			[]ConformancePlanModule{
				{TestModule: "oidcc-server", Variant: map[string]string{"response_type": "code id_token"}},
				{TestModule: "oidcc-server", Variant: map[string]string{"response_type": "code id_token token"}},
			},
			[]string{"ServerCodeIdToken", "ServerCodeIdTokenToken"},
		},
		{
			"ShouldOrderVariantKeysDeterministically",
			[]ConformancePlanModule{
				{TestModule: "oidcc-server", Variant: map[string]string{"response_type": "code", "response_mode": "form_post"}},
				{TestModule: "oidcc-server", Variant: map[string]string{"response_type": "code", "response_mode": "query"}},
			},
			[]string{"ServerFormPostCode", "ServerQueryCode"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConformanceSubtestNames(tc.have))
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/suites/ -run TestConformanceSubtestNames -v`
Expected: FAIL to compile — `undefined: ConformanceSubtestNames`.

- [ ] **Step 3: Write the implementation**

Create `internal/suites/conformance_names.go`:

```go
// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"sort"
	"strings"
	"unicode"
)

// conformanceModulePrefix is redundant in a subtest name: every module in these plans carries it.
const conformanceModulePrefix = "oidcc-"

// ConformanceSubtestNames returns the PascalCase subtest name of each module, index aligned with modules. A module
// which a plan lists more than once has its variant appended so the names within a plan stay unique, which is what
// makes go test able to address a single one with -run.
func ConformanceSubtestNames(modules []ConformancePlanModule) (names []string) {
	counts := make(map[string]int, len(modules))

	for _, module := range modules {
		counts[module.TestModule]++
	}

	names = make([]string, len(modules))

	for i, module := range modules {
		name := conformancePascalCase(strings.TrimPrefix(module.TestModule, conformanceModulePrefix))

		if counts[module.TestModule] > 1 {
			name += conformanceVariantSuffix(module.Variant)
		}

		names[i] = name
	}

	return names
}

// conformanceVariantSuffix renders a variant as a PascalCase suffix, in sorted key order so that two runs of the same
// plan produce the same names.
func conformanceVariantSuffix(variant map[string]string) (suffix string) {
	keys := make([]string, 0, len(variant))

	for key := range variant {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		suffix += conformancePascalCase(variant[key])
	}

	return suffix
}

// conformancePascalCase splits value on every character which is neither a letter nor a digit and title cases each
// part, so that 'code id_token' and 'max-age-1' become 'CodeIdToken' and 'MaxAge1'.
func conformancePascalCase(value string) string {
	builder := &strings.Builder{}

	for _, part := range strings.FieldsFunc(value, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])

		builder.WriteString(string(runes))
	}

	return builder.String()
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/suites/ -run TestConformanceSubtestNames -v`
Expected: PASS, seven subtests.

- [ ] **Step 5: Commit**

```bash
golangci-lint run
git add internal/suites/conformance_names.go internal/suites/conformance_names_test.go
git commit -m "test(suites): add oidc conformance subtest naming"
```

---

### Task 4: Suite environment

Bring up the conformance suite alongside Authelia and register the Authelia suite. At the end of this task `authelia-scripts suites setup OIDCConformance` produces a running stack whose API answers, even though no tests exist yet.

**Files:**

- Create: `internal/suites/OIDCConformance/compose.yml`
- Create: `internal/suites/OIDCConformance/Dockerfile.server`
- Create: `internal/suites/OIDCConformance/configuration.yml`
- Create: `internal/suites/OIDCConformance/users.yml`
- Create: `internal/suites/suite_oidc_conformance.go`
- Modify: `internal/suites/hosts.go` (add the conformance host entry)

**Interfaces:**

- Consumes: `conformance.Builders` from Task 1; `NewConformanceClient`/`WaitReady` from Task 2.
- Produces:
  - `const oidcConformanceSuiteName = "OIDCConformance"`
  - `const oidcConformanceBaseURL = "https://conformance.example.com:8443"`
  - `const oidcConformancePlansFile = "conformance-plans.json"`
  - `const oidcConformanceClientsFile = "conformance-clients.yml"`
  - `type OIDCConformancePlanFile struct { Name string; Plan conformance.Plan }` — the on-disk handoff between setup and the test process.

- [ ] **Step 1: Read the suite patterns being followed**

```bash
cat internal/suites/suite_oidc.go
cat internal/suites/OIDC/compose.yml
cat internal/suites/OIDC/configuration.yml
cat internal/suites/OIDC/users.yml
sed -n '20,80p' internal/suites/hosts.go
```

- [ ] **Step 2: Add the host entry**

In `internal/suites/hosts.go`, inside `HostEntries()`, after the `// OIDC tester app.` block add:

```go
		// OpenID Connect 1.0 conformance suite.
		{Domain: "conformance.example.com", IP: SuiteAddress(140)},
```

- [ ] **Step 3: Create the conformance server Dockerfile**

Create `internal/suites/OIDCConformance/Dockerfile.server`:

```dockerfile
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM registry.gitlab.com/openid/conformance-suite:release-v5.2.4

# The conformance server fetches Authelia's discovery document, JWKS, token endpoint and userinfo endpoint over TLS
# signed by the suite development CA, which no public root vouches for. Importing into the JDK's own cacerts adds that
# CA while leaving the public roots in place, which a -Djavax.net.ssl.trustStore override would not.
COPY common/pki/ca.public.crt /tmp/authelia-ca.crt

RUN keytool -importcert -noprompt -alias authelia-development-ca -file /tmp/authelia-ca.crt -cacerts -storepass changeit
```

- [ ] **Step 4: Create the compose file**

Create `internal/suites/OIDCConformance/compose.yml`. The build context is `internal/suites` so the Dockerfile can copy from `common/pki`.

```yaml
# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

---
services:
  authelia-backend:
    environment:
      X_AUTHELIA_CONFIG_FILTERS: 'template'
      X_AUTHELIA_CONFIG: '/config/configuration.yml,/tmp/conformance-clients.yml'
    volumes:
      - './OIDCConformance/configuration.yml:/config/configuration.yml'
      - './OIDCConformance/users.yml:/config/users.yml'
      - './common/pki:/pki'
      - './common/pki/public.crt:/certs/public.crt'
      - '${SUITE_TMP:-/tmp}:/tmp'

  conformance-mongodb:
    image: 'mongo:6.0.13'
    cpus: 2
    mem_limit: 2g
    networks:
      authelianet: {}

  conformance-server:
    build:
      context: '.'
      dockerfile: 'OIDCConformance/Dockerfile.server'
    cpus: 4
    mem_limit: 4g
    environment:
      BASE_URL: 'https://conformance.example.com:8443'
      MONGODB_HOST: 'conformance-mongodb'
      JAVA_EXTRA_ARGS: '-Dfintechlabs.devmode=true'
    depends_on:
      - 'conformance-mongodb'
      - 'authelia-backend'
    expose:
      - 8080
    networks:
      authelianet:
        aliases:
          # The published nginx image proxies to http://server:8080 and that configuration is baked in.
          - 'server'

  conformance-nginx:
    image: 'registry.gitlab.com/openid/conformance-suite/nginx:release-v5.2.4'
    cpus: 1
    mem_limit: 512m
    depends_on:
      - 'conformance-server'
    volumes:
      - './common/pki/public.chain.pem:/etc/ssl/certs/nginx-selfsigned.crt'
      - './common/pki/private.pem:/etc/ssl/private/nginx-selfsigned.key'
    networks:
      authelianet:
        aliases:
          - 'conformance.example.com'
        ipv4_address: ${SUITE_SUBNET:-192.168.240}.140
...
```

- [ ] **Step 5: Create the Authelia configuration**

Create `internal/suites/OIDCConformance/configuration.yml` by copying `internal/suites/OIDC/configuration.yml`:

```bash
cp internal/suites/OIDC/configuration.yml internal/suites/OIDCConformance/configuration.yml
cp internal/suites/OIDC/users.yml internal/suites/OIDCConformance/users.yml
```

Then delete everything from the `    clients:` line (`internal/suites/OIDC/configuration.yml:82`) to the end of the
`identity_providers` block, leaving the section as:

```yaml
identity_providers:
  oidc:
    enable_client_debug_messages: true
    hmac_secret: 'IVPWBkAdJHje3uz7LtFTDU2pFUfh39Xm'
    jwks:
      - key: {{ secret "/pki/private.oidc.pem" | mindent 10 "|" | msquote }}
        certificate_chain: {{ secret "/pki/public.oidc.chain.pem" | mindent 10 "|" | msquote }}
```

Every client now comes from the generated `/tmp/conformance-clients.yml`, which Authelia merges in because the compose
file sets `X_AUTHELIA_CONFIG` to both paths. Leave every other section — server, log, storage, notifier, session,
authentication backend, WebAuthn and access control — exactly as the OIDC suite has it.

Verify the templating still resolves, since `X_AUTHELIA_CONFIG_FILTERS: 'template'` means the `{{ secret ... }}`
expressions above are evaluated at startup:

```bash
grep -c 'secret "/pki' internal/suites/OIDCConformance/configuration.yml
```

Expected: `2`.

`users.yml` is copied unchanged: the conformance plans sign in as `john` / `password`, which that file already defines, and the `profile`, `email`, `phone` and `address` scopes the plans request are satisfied by its attributes.

- [ ] **Step 6: Write the suite registration**

Create `internal/suites/suite_oidc_conformance.go`:

```go
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

	"gopkg.in/yaml.v3"

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

// OIDCConformancePlanFile is one entry of the plans file exchanged between the setup and test processes.
type OIDCConformancePlanFile struct {
	Name string           `json:"name"`
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
	version := random.New().StringCustom(8, random.CharSetAlphaNumericLowercase)

	clients := &oidcConformanceClients{}
	clients.IdentityProviders.OIDC.Clients = []schema.IdentityProvidersOpenIDConnectClient{}

	var plans []OIDCConformancePlanFile

	for _, builder := range conformance.Builders(version, "implicit", "one_factor", "authelia", suiteURL, autheliaURL) {
		suite := builder.Build()

		clients.IdentityProviders.OIDC.Clients = append(clients.IdentityProviders.OIDC.Clients, suite.Clients...)
		plans = append(plans, OIDCConformancePlanFile{Name: builder.Name, Plan: suite.Plan})
	}

	if err = oidcConformanceWriteYAML(SuiteTmpPath(oidcConformanceClientsFile), clients); err != nil {
		return err
	}

	return oidcConformanceWriteJSON(SuiteTmpPath(oidcConformancePlansFile), plans)
}

func oidcConformanceWriteYAML(path string, value any) (err error) {
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)

	defer encoder.Close()

	return encoder.Encode(value)
}

func oidcConformanceWriteJSON(path string, value any) (err error) {
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

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
		return dockerEnvironment.PrintLogs("authelia-backend", "authelia-frontend", "conformance-server", "conformance-nginx")
	}

	teardown := func(suitePath string) error {
		return dockerEnvironment.Down()
	}

	GlobalRegistry.Register(oidcConformanceSuiteName, Suite{
		SetUp:           setup,
		SetUpTimeout:    10 * time.Minute,
		OnSetupTimeout:  displayLogs,
		OnError:         displayLogs,
		TestTimeout:     90 * time.Minute,
		TearDown:        teardown,
		TearDownTimeout: 5 * time.Minute,
		Description:     "This suite runs the OpenID Foundation conformance suite against Authelia for every profile Authelia is OpenID Certified for.",
	})
}
```

Check that `random.CharSetAlphaNumericLowercase` exists with `grep -n 'CharSetAlphaNumeric' internal/random/const.go`; if the lowercase variant is absent, use `random.CharSetAlphaNumeric` and lowercase the result with `strings.ToLower`.

- [ ] **Step 7: Verify the suite is registered and builds**

```bash
go build ./... && go run ./cmd/authelia-scripts suites list
```

Expected: `OIDCConformance` appears in the list.

- [ ] **Step 8: Bring the environment up for real**

```bash
go run ./cmd/authelia-scripts suites setup OIDCConformance
```

Expected: setup completes without error. Then verify each moving part:

```bash
ls -l "${SUITE_TMP_PATH:-/tmp}/conformance-plans.json" "${SUITE_TMP_PATH:-/tmp}/conformance-clients.yml"
curl -sk https://conformance.example.com:8443/api/plan?length=1 | head -c 200
docker compose -p authelia exec -T conformance-server sh -c 'keytool -list -cacerts -storepass changeit -alias authelia-development-ca'
curl -sk https://login.example.com:8080/.well-known/openid-configuration | head -c 200
```

Expected: both files exist; the API returns JSON; keytool prints the imported certificate; Authelia's discovery document returns.

If the conformance server cannot reach Authelia, the JVM logs say so — check with `docker compose -p authelia logs conformance-server | grep -i 'PKIX\|SSLHandshake'`. A PKIX error means the CA import did not take effect.

- [ ] **Step 9: Tear down and commit**

```bash
go run ./cmd/authelia-scripts suites teardown OIDCConformance
golangci-lint run
git add internal/suites/OIDCConformance internal/suites/suite_oidc_conformance.go internal/suites/hosts.go
git commit -m "test(suites): add oidc conformance suite environment"
```

---

### Task 5: Browser driver

Drive the pages the conformance suite hands over, and classify what is on screen. Every function here may run off the test goroutine, so nothing takes `*testing.T` and nothing calls `require`.

**Files:**

- Create: `internal/suites/conformance_browser.go`
- Test: `internal/suites/conformance_browser_test.go`

**Interfaces:**

- Consumes: `RodSession.doNavigate` from `internal/suites/action_visit.go`.
- Produces:
  - `type ConformancePageState int` with values `ConformancePageUnknown`, `ConformancePageFirstFactor`, `ConformancePageConsent`, `ConformancePageAutheliaError`, `ConformancePageCallback`
  - `func ConformanceClassifyPage(url string, hasFirstFactor, hasConsent, hasError bool) ConformancePageState`
  - `type ConformanceBrowser struct` with `func NewConformanceBrowser(session *RodSession) (*ConformanceBrowser, error)`, `func (b *ConformanceBrowser) Close()`, `func (b *ConformanceBrowser) ClearCookies() error`, `func (b *ConformanceBrowser) Page() *rod.Page`, `func (b *ConformanceBrowser) Drive(ctx context.Context, uri string) error`, `func (b *ConformanceBrowser) HasFirstFactor() bool`

- [ ] **Step 1: Write the failing test**

Create `internal/suites/conformance_browser_test.go`:

```go
// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConformanceClassifyPage(t *testing.T) {
	testCases := []struct {
		name                            string
		url                             string
		firstFactor, consent, autheliaError bool
		expected                        ConformancePageState
	}{
		{
			"ShouldDetectTheCallbackByPath",
			"https://conformance.example.com:8443/test/a/conformance-basic-authelia1a2b3c4d/callback?code=abc",
			false, false, false,
			ConformancePageCallback,
		},
		{
			"ShouldPreferTheCallbackOverAnyStageStillInTheDOM",
			"https://conformance.example.com:8443/test/a/alias/callback",
			true, false, false,
			ConformancePageCallback,
		},
		{
			"ShouldDetectTheFirstFactorPage",
			"https://login.example.com:8080/",
			true, false, false,
			ConformancePageFirstFactor,
		},
		{
			"ShouldDetectTheConsentPage",
			"https://login.example.com:8080/consent/openid/decision",
			false, true, false,
			ConformancePageConsent,
		},
		{
			"ShouldDetectAnAutheliaError",
			"https://login.example.com:8080/",
			false, false, true,
			ConformancePageAutheliaError,
		},
		{
			"ShouldReportUnknownWhenNothingMatches",
			"https://login.example.com:8080/",
			false, false, false,
			ConformancePageUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConformanceClassifyPage(tc.url, tc.firstFactor, tc.consent, tc.autheliaError))
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/suites/ -run TestConformanceClassifyPage -v`
Expected: FAIL to compile — `undefined: ConformanceClassifyPage`.

- [ ] **Step 3: Write the implementation**

Create `internal/suites/conformance_browser.go`:

```go
// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// ConformancePageState is what the browser is looking at part way through a conformance module's authorization flow.
type ConformancePageState int

const (
	// ConformancePageUnknown is a page the driver has no action for.
	ConformancePageUnknown ConformancePageState = iota

	// ConformancePageFirstFactor is Authelia's first factor sign in form.
	ConformancePageFirstFactor

	// ConformancePageConsent is Authelia's OpenID Connect 1.0 consent decision stage.
	ConformancePageConsent

	// ConformancePageAutheliaError is Authelia reporting an error rather than presenting a stage.
	ConformancePageAutheliaError

	// ConformancePageCallback is the conformance suite's own callback, which ends a leg of the flow.
	ConformancePageCallback
)

const (
	conformanceSelectorFirstFactor   = "#first-factor-stage"
	conformanceSelectorConsent       = "#openid-consent-decision-stage"
	conformanceSelectorConsentAccept = "#openid-consent-accept"
	conformanceSelectorAutheliaError = `.notification[data-type="error"]`
	conformanceCallbackPathFragment  = "/test/a/"

	// conformanceElementTimeout is how long a single probe for an element waits. Probes run in a loop, so this is the
	// resolution of that loop rather than the budget for a page appearing.
	conformanceElementTimeout = time.Second * 2

	// conformanceDriveInterval is how often the driver re-examines a page which it had no action for.
	conformanceDriveInterval = time.Millisecond * 250
)

// ConformanceClassifyPage decides what the browser is looking at. It takes the observations rather than the page so
// that the decision is testable without a browser. The callback wins over every stage selector, because a stage left
// in the DOM of the document being navigated away from would otherwise be read as the current page.
func ConformanceClassifyPage(uri string, hasFirstFactor, hasConsent, hasError bool) ConformancePageState {
	switch {
	case strings.Contains(uri, conformanceCallbackPathFragment):
		return ConformancePageCallback
	case hasError:
		return ConformancePageAutheliaError
	case hasConsent:
		return ConformancePageConsent
	case hasFirstFactor:
		return ConformancePageFirstFactor
	default:
		return ConformancePageUnknown
	}
}

// ConformanceBrowser is one conformance plan's isolated browser context. Every method returns an error rather than
// failing a test, because the driver runs on a goroutine which is not the test's.
type ConformanceBrowser struct {
	session   *RodSession
	incognito *rod.Browser
	page      *rod.Page
	username  string
	password  string
}

// NewConformanceBrowser returns a ConformanceBrowser with its own cookie jar.
func NewConformanceBrowser(session *RodSession) (browser *ConformanceBrowser, err error) {
	incognito, err := session.WebDriver.Incognito()
	if err != nil {
		return nil, fmt.Errorf("error creating an isolated browser context: %w", err)
	}

	page, err := incognito.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("error creating a tab: %w", err)
	}

	return &ConformanceBrowser{
		session:   session,
		incognito: incognito,
		page:      page,
		username:  testUsername,
		password:  testPassword,
	}, nil
}

// Page returns the browser's tab, for a module override which needs to assert against it.
func (b *ConformanceBrowser) Page() *rod.Page {
	return b.page
}

// Close disposes of the browser context.
func (b *ConformanceBrowser) Close() {
	if b.incognito == nil {
		return
	}

	_ = b.incognito.Close()

	b.incognito = nil
}

// ClearCookies discards every cookie in this context, for the modules whose instructions are to remove any cookies
// received from the provider before proceeding.
func (b *ConformanceBrowser) ClearCookies() error {
	return proto.NetworkClearBrowserCookies{}.Call(b.page)
}

// HasFirstFactor reports whether Authelia's sign in form is on screen right now.
func (b *ConformanceBrowser) HasFirstFactor() bool {
	return b.has(conformanceSelectorFirstFactor)
}

func (b *ConformanceBrowser) has(selector string) bool {
	has, _, err := b.page.Timeout(conformanceElementTimeout).Has(selector)

	return err == nil && has
}

func (b *ConformanceBrowser) url() string {
	info, err := b.page.Info()
	if err != nil {
		return ""
	}

	return info.URL
}

// Drive navigates to uri and works the flow through to the conformance suite's callback. It is deliberately reactive:
// it acts on whatever page appears rather than on a script per module, so that a module added by a future conformance
// suite release runs instead of failing.
func (b *ConformanceBrowser) Drive(ctx context.Context, uri string) (err error) {
	if err = b.session.doNavigate(b.page, uri); err != nil {
		return fmt.Errorf("error navigating to '%s': %w", uri, err)
	}

	for {
		if ctx.Err() != nil {
			return fmt.Errorf("the flow beginning at '%s' did not reach the conformance callback, the last page was '%s': %w", uri, b.url(), ctx.Err())
		}

		switch ConformanceClassifyPage(b.url(), b.has(conformanceSelectorFirstFactor), b.has(conformanceSelectorConsent), b.has(conformanceSelectorAutheliaError)) {
		case ConformancePageCallback:
			return nil
		case ConformancePageFirstFactor:
			if err = b.signIn(); err != nil {
				return err
			}
		case ConformancePageConsent:
			if err = b.consent(); err != nil {
				return err
			}
		case ConformancePageAutheliaError:
			// An error page can be the point of a module, so it ends the leg without failing. The module's own result
			// decides whether it was expected.
			return nil
		case ConformancePageUnknown:
			time.Sleep(conformanceDriveInterval)
		}
	}
}

func (b *ConformanceBrowser) signIn() (err error) {
	if err = b.input("#username-textfield", b.username); err != nil {
		return err
	}

	if err = b.input("#password-textfield", b.password); err != nil {
		return err
	}

	return b.click("#sign-in-button")
}

func (b *ConformanceBrowser) consent() error {
	return b.click(conformanceSelectorConsentAccept)
}

func (b *ConformanceBrowser) input(selector, value string) (err error) {
	element, err := b.page.Timeout(elementLocateTimeout).Element(selector)
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

func (b *ConformanceBrowser) click(selector string) (err error) {
	element, err := b.page.Timeout(elementLocateTimeout).Element(selector)
	if err != nil {
		return fmt.Errorf("error locating '%s': %w", selector, err)
	}

	if err = element.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("error clicking '%s': %w", selector, err)
	}

	return nil
}
```

The three selectors are the ones the existing suites already drive, so they are known good rather than guessed:
`first-factor-stage` from `verify_is_first_factor_page.go`, `openid-consent-decision-stage` from
`verify_is_consent_page.go`, `openid-consent-accept` from `scenario_oidc_test.go:103`, and `.notification` from
`verify_notification.go`. Confirm they have not moved:

```bash
grep -rn 'first-factor-stage\|openid-consent-decision-stage\|openid-consent-accept' internal/suites/*.go | head
```

Expected: each appears in the file named above.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/suites/ -run TestConformanceClassifyPage -v`
Expected: PASS, six subtests.

- [ ] **Step 5: Commit**

```bash
golangci-lint run
git add internal/suites/conformance_browser.go internal/suites/conformance_browser_test.go
git commit -m "test(suites): add oidc conformance browser driver"
```

---

### Task 6: Plan runner

Walk one plan's modules in order, producing an outcome per module on a channel. This is the code that runs off the test goroutine.

**Files:**

- Create: `internal/suites/conformance_runner.go`
- Test: `internal/suites/conformance_runner_test.go`

**Interfaces:**

- Consumes: everything from Tasks 2, 3 and 5.
- Produces:
  - `type ConformanceOutcome struct { Name, Module, Result, Status, LogURL string; Skipped bool; Reason string; Err error }`
  - `type ConformanceOverride struct { ClearCookies bool; Assert func(browser *ConformanceBrowser) error }`
  - `var conformanceOverrides map[string]ConformanceOverride`
  - `var conformanceUnattended map[string]string`
  - `func NewConformanceRunner(client *ConformanceClient, browser *ConformanceBrowser, planID string, modules []ConformancePlanModule) *ConformanceRunner`
  - `func (r *ConformanceRunner) Run(ctx context.Context) <-chan ConformanceOutcome`

- [ ] **Step 1: Write the failing test**

Create `internal/suites/conformance_runner_test.go`. It exercises the module loop against a scripted `httptest` server, with no browser: a module which is `FINISHED` immediately needs no browser at all.

```go
// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConformanceRunner_ReportsOutcomesInPlanOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/runner":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"` + r.URL.Query().Get("test") + `-id"}`))
		case r.URL.Path == "/api/runner/oidcc-server-id/wait-state",
			r.URL.Path == "/api/runner/oidcc-scope-address-id/wait-state":
			_, _ = w.Write([]byte(`{"state":"FINISHED"}`))
		case r.URL.Path == "/api/info/oidcc-server-id":
			_, _ = w.Write([]byte(`{"_id":"oidcc-server-id","status":"FINISHED","result":"PASSED"}`))
		case r.URL.Path == "/api/info/oidcc-scope-address-id":
			_, _ = w.Write([]byte(`{"_id":"oidcc-scope-address-id","status":"FINISHED","result":"WARNING"}`))
		case r.URL.Path == "/api/log/oidcc-server-id", r.URL.Path == "/api/log/oidcc-scope-address-id":
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	modules := []ConformancePlanModule{{TestModule: "oidcc-server"}, {TestModule: "oidcc-scope-address"}}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var outcomes []ConformanceOutcome

	for outcome := range NewConformanceRunner(client, nil, "plan1", modules).Run(ctx) {
		outcomes = append(outcomes, outcome)
	}

	require.Len(t, outcomes, 2)

	assert.Equal(t, "Server", outcomes[0].Name)
	assert.Equal(t, "PASSED", outcomes[0].Result)
	require.NoError(t, outcomes[0].Err)

	assert.Equal(t, "ScopeAddress", outcomes[1].Name)
	assert.Equal(t, "WARNING", outcomes[1].Result)
}

func TestConformanceRunner_SkipsUnattendedModules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("the runner must not contact the server for a skipped module, got %s", r.URL.Path)
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	modules := []ConformancePlanModule{{TestModule: "oidcc-server-rotate-keys"}}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var outcomes []ConformanceOutcome

	for outcome := range NewConformanceRunner(client, nil, "plan1", modules).Run(ctx) {
		outcomes = append(outcomes, outcome)
	}

	require.Len(t, outcomes, 1)
	assert.True(t, outcomes[0].Skipped)
	assert.Equal(t, "ServerRotateKeys", outcomes[0].Name)
	assert.NotEmpty(t, outcomes[0].Reason)
}

func TestConformanceRunner_FillsPlaceholders(t *testing.T) {
	var uploaded bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/runner":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"m1"}`))
		case "/api/runner/m1/wait-state":
			_, _ = w.Write([]byte(`{"state":"FINISHED"}`))
		case "/api/log/m1":
			if uploaded {
				_, _ = w.Write([]byte(`[]`))

				return
			}

			_, _ = w.Write([]byte(`[{"msg":"upload","upload":"ph1","result":"REVIEW"}]`))
		case "/api/log/m1/images/ph1":
			uploaded = true

			_, _ = w.Write([]byte(`{}`))
		case "/api/info/m1":
			_, _ = w.Write([]byte(`{"_id":"m1","status":"FINISHED","result":"REVIEW"}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))

	defer server.Close()

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var outcomes []ConformanceOutcome

	for outcome := range NewConformanceRunner(client, nil, "plan1", []ConformancePlanModule{{TestModule: "oidcc-display-page"}}).Run(ctx) {
		outcomes = append(outcomes, outcome)
	}

	require.Len(t, outcomes, 1)
	assert.True(t, uploaded)
	assert.Equal(t, "REVIEW", outcomes[0].Result)
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/suites/ -run TestConformanceRunner -v`
Expected: FAIL to compile — `undefined: NewConformanceRunner`.

- [ ] **Step 3: Write the implementation**

Create `internal/suites/conformance_runner.go`:

```go
// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	conformanceStatusConfigured = "CONFIGURED"
	conformanceStatusWaiting    = "WAITING"
	conformanceStatusFinished   = "FINISHED"

	// conformanceModuleTimeout is the budget for one module, browser interaction included.
	conformanceModuleTimeout = time.Minute * 5

	// conformancePlaceholderTimeout is how long a module which fills placeholders is given after its browser legs
	// complete, before its remaining placeholders are considered absent.
	conformancePlaceholderTimeout = time.Second * 30

	// conformancePlaceholderInterval is how often the log is re-read while waiting for a placeholder to appear.
	conformancePlaceholderInterval = time.Second
)

// ConformanceOutcome is the result of one module, produced off the test goroutine and asserted on it.
type ConformanceOutcome struct {
	Name    string
	Module  string
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

	// Assert is the go-rod check which stands in for the screenshot the conformance suite would otherwise want. It runs
	// while the page which prompted the placeholder is still on screen, and its error becomes the module's verdict.
	Assert func(browser *ConformanceBrowser) error
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

	// These carry instructions to remove any cookies received from the provider, so that a normal login page is shown.
	"oidcc-display-page":    {ClearCookies: true},
	"oidcc-display-popup":   {ClearCookies: true},
	"oidcc-ui-locales":      {ClearCookies: true},
	"oidcc-claims-locales":  {ClearCookies: true},
}

// conformanceUnattended holds the modules which cannot run without a person, mapped to why. They are reported as
// skipped rather than failed. Keep this list minimal and justified: a module which leaves it becomes a real failure,
// which is the point.
var conformanceUnattended = map[string]string{
	"oidcc-server-rotate-keys": "requires the provider's signing keys to be rotated by hand while the module waits",
}

func conformanceAssertReauthentication(browser *ConformanceBrowser) error {
	if !browser.HasFirstFactor() {
		return errors.New("expected Authelia to present the sign in form a second time but it did not")
	}

	return nil
}

func conformanceAssertNoReauthentication(browser *ConformanceBrowser) error {
	if browser.HasFirstFactor() {
		return errors.New("expected Authelia to reuse the existing session but it presented the sign in form again")
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
// order. The channel closes when the plan is done. Nothing here touches testing.T: an outcome carries its error so the
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

	id, err := r.client.CreateTest(ctx, r.planID, module.TestModule, module.Variant)
	if err != nil {
		outcome.Err = fmt.Errorf("error creating module '%s': %w", module.TestModule, err)

		return outcome
	}

	outcome.LogURL = r.client.LogDetailURL(id)

	override := conformanceOverrides[module.TestModule]

	if override.ClearCookies && r.browser != nil {
		if err = r.browser.ClearCookies(); err != nil {
			outcome.Err = err

			return outcome
		}
	}

	state, err := r.client.WaitState(ctx, id, conformanceStatusConfigured, conformanceStatusWaiting, conformanceStatusFinished)
	if err != nil {
		outcome.Err = err

		return outcome
	}

	if state == conformanceStatusConfigured {
		if err = r.client.StartTest(ctx, id); err != nil {
			outcome.Err = err

			return outcome
		}

		if state, err = r.client.WaitState(ctx, id, conformanceStatusWaiting, conformanceStatusFinished); err != nil {
			outcome.Err = err

			return outcome
		}
	}

	if state == conformanceStatusWaiting {
		if err = r.interact(ctx, id, override); err != nil {
			outcome.Err = err
		}
	}

	// Only wait for the module to finish when the interaction succeeded. A module whose browser leg failed will never
	// leave WAITING, so waiting on it would burn the whole module budget before reporting a failure already known.
	if outcome.Err == nil {
		if _, err = r.client.WaitState(ctx, id, conformanceStatusFinished); err != nil {
			outcome.Err = err
		}
	}

	info, err := r.client.Info(ctx, id)
	if err != nil {
		if outcome.Err == nil {
			outcome.Err = err
		}

		return outcome
	}

	outcome.Status, outcome.Result = info.Status, info.Result

	return outcome
}

// interact drives every URL the module hands over and fills every placeholder it raises, until the module leaves
// WAITING. A module such as oidcc-prompt-login performs two authorization round trips, so this loops rather than
// assuming a single URL.
func (r *ConformanceRunner) interact(ctx context.Context, id string, override ConformanceOverride) (err error) {
	if r.browser == nil {
		return errors.New("the module needs a browser but the runner has none")
	}

	visited := map[string]bool{}

	deadline := time.Now().Add(conformancePlaceholderTimeout)

	for {
		if ctx.Err() != nil {
			return fmt.Errorf("module '%s' did not leave the WAITING state: %w", id, ctx.Err())
		}

		status, err := r.client.BrowserStatus(ctx, id)
		if err != nil {
			return err
		}

		progressed := false

		for _, uri := range status.URLs {
			if visited[uri] {
				continue
			}

			visited[uri] = true
			progressed = true

			if err = r.browser.Drive(ctx, uri); err != nil {
				return err
			}

			// The assertion runs while the page which prompted it is still on screen, before the placeholder upload
			// releases the module.
			if override.Assert != nil {
				if err = override.Assert(r.browser); err != nil {
					return err
				}
			}

			if err = r.client.Visit(ctx, id, uri); err != nil {
				return err
			}
		}

		filled, err := r.fillPlaceholders(ctx, id)
		if err != nil {
			return err
		}

		if progressed || filled {
			deadline = time.Now().Add(conformancePlaceholderTimeout)
		}

		info, err := r.client.Info(ctx, id)
		if err != nil {
			return err
		}

		if info.Status != conformanceStatusWaiting {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("module '%s' stayed in the WAITING state with nothing left to visit or fill", id)
		}

		time.Sleep(conformancePlaceholderInterval)
	}
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

		if err = r.client.UploadPlaceholder(ctx, id, entry.Upload, conformancePlaceholderImage); err != nil {
			return filled, err
		}

		filled = true
	}

	return filled, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/suites/ -run TestConformanceRunner -v`
Expected: PASS, three tests.

- [ ] **Step 5: Commit**

```bash
golangci-lint run
git add internal/suites/conformance_runner.go internal/suites/conformance_runner_test.go
git commit -m "test(suites): add oidc conformance plan runner"
```

---

### Task 7: The suite tests

Wire the runners to testify. The suite creates the seven plans, starts a runner per plan, and each profile method drains its runner's channel and asserts.

**Files:**

- Create: `internal/suites/suite_oidc_conformance_test.go`

**Interfaces:**

- Consumes: everything from Tasks 2 to 6, plus `oidcConformanceReadPlans` and `oidcConformanceBaseURL` from Task 4.
- Produces: `TestOIDCConformanceSuite`, the entry point `authelia-scripts` invokes.

- [ ] **Step 1: Write the suite**

Create `internal/suites/suite_oidc_conformance_test.go`:

```go
// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/oidc/conformance"
)

// conformanceAcceptedResults are the module results which pass. REVIEW is here because filling an image placeholder
// forces it, and this suite fills placeholders for every module which blocks on one, so REVIEW carries no signal. Add
// WARNING here to accept results which the conformance suite flags but does not fail.
var conformanceAcceptedResults = []string{"PASSED", "REVIEW", "SKIPPED"}

// conformancePlanTimeout is the budget for one whole plan.
const conformancePlanTimeout = time.Minute * 75

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
func (s *OIDCConformanceSuite) SetupSuite() {
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

		created, err := s.client.CreatePlan(ctx, &plan.Plan)
		s.Require().NoErrorf(err, "error creating the '%s' plan", plan.Name)

		s.planIDs[plan.Name] = created.ID

		browser, err := NewConformanceBrowser(s.RodSession)
		s.Require().NoError(err)

		s.browsers = append(s.browsers, browser)

		s.outcomes[plan.Name] = NewConformanceRunner(s.client, browser, created.ID, created.Modules).Run(ctx)
	}
}

// TearDownSuite exports each plan's log for the CI artifacts and disposes of the browser contexts.
func (s *OIDCConformanceSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}

	for _, browser := range s.browsers {
		browser.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	for name, id := range s.planIDs {
		path := fmt.Sprintf("../../screenshots/%s/conformance-%s.zip", oidcConformanceSuiteName, name)

		if err := s.client.ExportPlanHTML(ctx, id, path); err != nil {
			s.T().Logf("Error exporting the '%s' plan log: %v", name, err)
		}
	}
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

			if outcome.LogURL != "" {
				t.Logf("Conformance log: %s", outcome.LogURL)
			}

			require.NoError(t, outcome.Err)
			require.Containsf(t, conformanceAcceptedResults, outcome.Result,
				"module '%s' finished with the result '%s' and the status '%s'", outcome.Module, outcome.Result, outcome.Status)
		})
	}

	s.Require().NotZerof(count, "the '%s' plan produced no modules", name)
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
```

- [ ] **Step 2: Verify it compiles**

Run: `go vet ./internal/suites/`
Expected: no output.

- [ ] **Step 3: Verify the suite is addressable**

Run: `go test ./internal/suites/ -run '^(TestOIDCConformanceSuite)$' -short -v`
Expected: the test is skipped with "skipping suite test in short mode", proving the name matches what `authelia-scripts` invokes.

- [ ] **Step 4: Run the suite for real**

```bash
go run ./cmd/authelia-scripts --log-level debug suites test OIDCConformance --headless
```

This is the first full run and it is the point of the whole plan. Expect it to take a long time and expect failures; they are data. For each failing module, open the `Conformance log:` URL the subtest logged and decide which of these it is:

1. A real Authelia conformance regression — report it, do not paper over it.
2. A module which cannot run unattended — add it to `conformanceUnattended` with a one-line justification.
3. A module needing an override — add a `ConformanceOverride` for it.
4. A driver gap, such as a page the classifier returns `ConformancePageUnknown` for — extend `ConformanceClassifyPage` and its table test.

Record the wall clock of the run; Task 8 needs it.

- [ ] **Step 5: Commit**

```bash
golangci-lint run
git add internal/suites/suite_oidc_conformance_test.go internal/suites/conformance_runner.go internal/suites/conformance_browser.go
git commit -m "test(suites): add oidc conformance suite tests"
```

---

### Task 8: CI wiring

`.buildkite/steps/e2etests.sh` enumerates `authelia-scripts suites list`, so the suite already gets a Buildkite step the moment it is registered. All it needs is a timeout other than the 20 minute default.

No change is needed in `.buildkite/pipeline.sh`. The `BUILD_DUO` / `BUILD_HAPROXY` / `BUILD_SAMBA` flags exist to trigger separate pipelines which publish standalone images to `authelia/integration-*`; the conformance server image is built inline by compose instead, exactly as `internal/suites/example/compose/pam/compose.yml` does, and that suite has no such flag either.

**Files:**

- Modify: `.buildkite/steps/e2etests.sh:16-18`

**Interfaces:**

- Consumes: the suite name `OIDCConformance` and the wall clock measured in Task 7.
- Produces: nothing other tasks use.

- [ ] **Step 1: Add the suite timeout**

In `.buildkite/steps/e2etests.sh`, extend `SUITE_TIMEOUTS`:

```bash
declare -A SUITE_TIMEOUTS=(
  [Kubernetes]="30"
  [OIDCConformance]="90"
)
```

Replace `90` with roughly double the wall clock measured in Task 7, rounded up to the nearest ten minutes. Under-setting this is the failure mode that matters: Buildkite kills the step and the artifacts are lost, so the run produces no diagnosis.

- [ ] **Step 2: Verify the pipeline still renders**

```bash
bash -n .buildkite/steps/e2etests.sh
go run ./cmd/authelia-scripts suites list | grep OIDCConformance
SUITE_NAME=OIDCConformance bash -c 'source /dev/stdin <<< "$(sed -n "9,18p" .buildkite/steps/e2etests.sh)"; echo "${SUITE_TIMEOUTS[OIDCConformance]}"'
```

Expected: no syntax errors, the suite is listed, and the timeout prints.

- [ ] **Step 3: Commit**

```bash
git add .buildkite/steps/e2etests.sh
git commit -m "ci: set the oidc conformance suite timeout"
```

---

## Self-Review Notes

Spec sections and the task covering each:

| Spec section                             | Task                                      |
| ---------------------------------------- | ----------------------------------------- |
| Architecture / file layout               | 1, 2, 3, 4, 5, 6, 7                       |
| Environment: containers, PKI, networking | 4                                         |
| Configuration generation                 | 4 (generation) and 7 (plan creation)      |
| Conformance API client                   | 2                                         |
| Test structure and naming                | 3, 7                                      |
| Concurrency                              | 6, 7                                      |
| Module loop                              | 6                                         |
| Browser driver                           | 5, 6                                      |
| Placeholders                             | 2 (image constant, upload), 6 (fill loop) |
| Verdicts                                 | 6 (allowlist), 7 (accepted results)       |
| Diagnostics                              | 7 (log URL, plan export), 4 (`PrintLogs`) |
| CI                                       | 8                                         |
| Testing the suite's own code             | 1, 2, 3, 5, 6                             |
