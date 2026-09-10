// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

	// conformanceWaitStateRetryInterval is the pause between retries of a wait-state poll which failed transiently
	// (a 5xx response or a network error, as opposed to the documented 404-on-eviction), so that a server which is
	// hard-failing does not have the client spin against it.
	conformanceWaitStateRetryInterval = time.Millisecond * 500

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
//
// The identifier arrives as "id", not "_id". The conformance suite hand-builds this response as {name, id, modules}
// rather than serializing a persisted document, and only the endpoints that return a stored document — /api/info/{id}
// among them — carry Mongo's "_id". Reading the wrong one decodes silently to an empty string, which surfaces much
// later as a 400 from /api/runner?plan= and a 404 from the plan export.
type ConformanceCreatedPlan struct {
	ID      string                  `json:"id"`
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
//
// Only "urls" is decoded, and every entry in it is reached by navigation. The suite's own log-detail UI reads a
// "urlsWithMethod" list first and falls back to "urls" treating each entry as a GET, but getBrowserStatus emits only
// id, show_qr_code, urls, visited and runners -- there is no urlsWithMethod on the wire in the pinned release. Should a
// later version add one, this degrades exactly as the suite's own UI does, to navigating every entry.
type ConformanceBrowserStatus struct {
	URLs    []string `json:"urls"`
	Visited []string `json:"visited"`
}

// ConformanceLogEntry is one entry of a module's log. An entry with a non-empty Upload is an image placeholder which
// has not been filled, and which is holding the module in WAITING.
//
// Fields holds the entry exactly as it arrived. A failing condition records its own arguments alongside the message —
// the scope tests, for instance, log expected_scope_items, actual_scope_items and missing_items — and those arguments
// are usually the whole diagnosis. They cannot be typed ahead of time because every condition logs something
// different, so they are kept verbatim and rendered on demand.
type ConformanceLogEntry struct {
	Msg    string `json:"msg"`
	Result string `json:"result"`
	Src    string `json:"src"`
	Upload string `json:"upload"`

	Fields map[string]any `json:"-"`
}

// UnmarshalJSON decodes the typed fields and retains the entry whole.
func (e *ConformanceLogEntry) UnmarshalJSON(data []byte) (err error) {
	type entry ConformanceLogEntry

	if err = json.Unmarshal(data, (*entry)(e)); err != nil {
		return err
	}

	return json.Unmarshal(data, &e.Fields)
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

// conformanceUnexpectedStatusError is returned by ConformanceClient's internal request helper when a response's
// status code does not match what was expected. It carries the status code so that callers such as WaitState can
// branch on it - for example to distinguish a genuine 404 eviction from a transient 5xx - without parsing the error
// string, while do() keeps producing the same diagnostic message it always has.
type conformanceUnexpectedStatusError struct {
	Method   string
	URI      string
	Body     []byte
	Expected int
	Status   int
}

// Error implements the error interface.
func (e *conformanceUnexpectedStatusError) Error() string {
	return fmt.Sprintf("%s %s: expected status %d but got %d: %s", e.Method, e.URI, e.Expected, e.Status, e.Body)
}

func (c *ConformanceClient) do(ctx context.Context, method, uri string, body io.Reader, contentType string, expected int, out any) (err error) {
	data, err := c.doRaw(ctx, method, uri, body, contentType, expected)
	if err != nil {
		return err
	}

	if out == nil {
		return nil
	}

	return json.Unmarshal(data, out)
}

// doRaw performs a request and returns the response body without decoding it, for the one caller which has to
// distinguish a body it can decode but does not recognize from a body it cannot decode at all.
func (c *ConformanceClient) doRaw(ctx context.Context, method, uri string, body io.Reader, contentType string, expected int) (data []byte, err error) {
	var req *http.Request

	if req, err = http.NewRequestWithContext(ctx, method, uri, body); err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	var resp *http.Response

	if resp, err = c.client.Do(req); err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if data, err = io.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	if resp.StatusCode != expected {
		return nil, &conformanceUnexpectedStatusError{Method: method, URI: uri, Body: data, Expected: expected, Status: resp.StatusCode}
	}

	return data, nil
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

// CreatePlan creates a conformance test plan and returns its id and module list. The plan name and the variant are
// parameters rather than being read off plan because [conformance.Plan] deliberately excludes both from its JSON
// representation: they belong in the query string, not the body, and the caller is the one which knows them.
func (c *ConformanceClient) CreatePlan(ctx context.Context, name string, variant *conformance.PlanVariant, plan *conformance.Plan) (created *ConformanceCreatedPlan, err error) {
	if name == "" {
		return nil, fmt.Errorf("error creating plan '%s': the conformance plan name is empty", plan.Alias)
	}

	body, err := json.Marshal(plan)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("planName", name)

	if variant != nil {
		var data []byte

		if data, err = json.Marshal(variant); err != nil {
			return nil, err
		}

		query.Set("variant", string(data))
	}

	created = &ConformanceCreatedPlan{}

	if err = c.do(ctx, http.MethodPost, c.uri(query, "plan"), bytes.NewReader(body), "application/json", http.StatusCreated, created); err != nil {
		return nil, err
	}

	if created.ID == "" {
		return nil, fmt.Errorf("error creating plan '%s': the conformance suite returned no plan identifier", plan.Alias)
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
// registry and the endpoint answers 404; the persisted status from Info is authoritative at that point. A transport
// level failure - a 5xx response, a network error - is treated as transient and retried until ctx expires, so a
// single blip during a long unattended run does not fail the poll outright.
//
// A 200 whose body is not the documented shape is not retried. A body which does not decode at all, and a body which
// decodes but carries neither a state nor the timeout flag - an empty body, a proxy's own 200 page, a schema change
// on a suite version bump - are both permanent for as long as the run lasts, so retrying either would only spin
// against the server for the rest of the module budget from every plan at once. Both fail immediately with the body
// echoed, which is also the only way the reason for a schema change ever reaches the run's output.
func (c *ConformanceClient) WaitState(ctx context.Context, id string, states ...string) (state string, err error) {
	query := url.Values{}
	query.Set("states", strings.Join(states, ","))
	query.Set("timeoutMs", strconv.FormatInt(conformanceWaitStateTimeout.Milliseconds(), 10))

	for {
		data, err := c.doRaw(ctx, http.MethodGet, c.uri(query, "runner", id, "wait-state"), nil, "", http.StatusOK)
		if err != nil {
			if ctx.Err() != nil {
				return "", fmt.Errorf("module '%s' did not reach one of %v: %w", id, states, ctx.Err())
			}

			var statusErr *conformanceUnexpectedStatusError

			if errors.As(err, &statusErr) && statusErr.Status == http.StatusNotFound {
				return c.waitStateFromInfo(ctx, id, states)
			}

			// Anything else - a 5xx response, a network error, an unexpected status other than the documented 404
			// eviction - is treated as transient. Retry after a short pause so a hard-failing server doesn't spin,
			// and give up only once ctx expires.
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("module '%s' did not reach one of %v: %w", id, states, ctx.Err())
			case <-time.After(conformanceWaitStateRetryInterval):
			}

			continue
		}

		out := struct {
			State   string `json:"state"`
			Timeout bool   `json:"timeout"`
		}{}

		if err = json.Unmarshal(data, &out); err != nil {
			return "", fmt.Errorf("module '%s' wait-state poll answered with a body which is not JSON: %w: %s", id, err, data)
		}

		switch {
		case out.State != "":
			return out.State, nil
		case out.Timeout:
			// The poll expired without a transition. Fall through to the context check and poll again.
		default:
			return "", fmt.Errorf("module '%s' wait-state poll answered with neither a state nor a timeout: %s", id, data)
		}

		if ctx.Err() != nil {
			return "", fmt.Errorf("module '%s' did not reach one of %v: %w", id, states, ctx.Err())
		}
	}
}

// waitStateFromInfo resolves WaitState's outcome from Info once wait-state has 404'd because the module has left
// the server's running registry.
func (c *ConformanceClient) waitStateFromInfo(ctx context.Context, id string, states []string) (state string, err error) {
	var info *ConformanceTestInfo

	if info, err = c.Info(ctx, id); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("module '%s' did not reach one of %v: %w", id, states, ctx.Err())
		}

		return "", err
	}

	if utils.IsStringInSlice(info.Status, states) {
		return info.Status, nil
	}

	return "", fmt.Errorf("module '%s' is no longer running and its persisted status is '%s', which is not one of %v", id, info.Status, states)
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

// UploadImage adds an image to a module's log without a placeholder, as the log page's own upload does. The server
// marks the module for REVIEW, but does not otherwise release it.
func (c *ConformanceClient) UploadImage(ctx context.Context, id, description, dataURI string) error {
	query := url.Values{}
	query.Set("description", description)

	return c.do(ctx, http.MethodPost, c.uri(query, "log", id, "images"), bytes.NewReader([]byte(dataURI)), "text/plain", http.StatusOK, nil)
}

// StopTest stops a running module, which ends it as INTERRUPTED without changing its result. A module which is no
// longer running is already stopped.
func (c *ConformanceClient) StopTest(ctx context.Context, id string) error {
	err := c.do(ctx, http.MethodDelete, c.uri(nil, "runner", id), nil, "", http.StatusOK, nil)

	var statusErr *conformanceUnexpectedStatusError

	if errors.As(err, &statusErr) && statusErr.Status == http.StatusNotFound {
		return nil
	}

	return err
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

	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, resp.Body)

	return err
}
