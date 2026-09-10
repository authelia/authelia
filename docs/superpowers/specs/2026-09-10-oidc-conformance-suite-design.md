<!--
SPDX-FileCopyrightText: 2026 Authelia

SPDX-License-Identifier: Apache-2.0
-->

# OpenID Connect 1.0 Conformance Integration Suite

Design document. 2026-09-10.

## Summary

A new Authelia integration suite, `OIDCConformance`, that stands up the OpenID Foundation conformance
suite in Docker alongside Authelia, creates the seven certification test plans Authelia is certified
for over the conformance HTTP API, drives every module in every plan with go-rod, and asserts each
module's result as a Go subtest.

Authelia is OpenID Certified™ for the Basic OP, Implicit OP, Hybrid OP, Form Post OP and Config OP
profiles. Those map exactly onto the seven plans `cmd/authelia-gen` already knows how to build:
`config`, `basic`, `basic-form-post`, `hybrid`, `hybrid-form-post`, `implicit` and
`implicit-form-post`.

## Goals

- Catch conformance regressions on every pull request, not only at certification time.
- Configure the plans exclusively through the conformance suite's HTTP API.
- Verify browser-facing steps with go-rod assertions rather than screenshot capture and comparison.
- Report one named test per profile, with one named subtest per module, both in PascalCase.
- Keep a single source of truth for the plan configuration and the matching Authelia clients.

## Non-goals

- Producing a submittable certification package. The runs are regression checks; the certification
  submission stays a deliberate manual act using `authelia-gen misc oidc conformance` against the
  hosted conformance suite.
- Covering profiles Authelia is not certified for (Dynamic OP, RP-Initiated Logout, and so on).
- Replacing the existing `OIDC` and `OIDCTraefik` suites, which cover Authelia's own behavior
  against a simple relying party.

## Decisions

| Decision            | Choice                                                                                                                                                                    |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CI posture          | Registered like any other suite, running on every pull request, with a long `SUITE_TIMEOUTS` entry                                                                        |
| Config source       | `OpenIDConnectConformanceSuiteBuilder` extracted to `internal/oidc/conformance`, shared by `authelia-gen` and the suite                                                   |
| Placeholder modules | the driver records what it observed on each leg and the override asserts against that record; a small generated PNG is uploaded purely to release `waitForPlaceholders()` |
| Naming              | `TestBasic` / `Server`; the `oidcc-` prefix is stripped, a variant suffix disambiguates repeated modules                                                                  |
| Accepted results    | `PASSED`, `REVIEW`, `SKIPPED` pass; `WARNING` and `FAILED` fail; the accepted set is a single configurable value                                                          |
| Concurrency         | Plans run concurrently, modules within a plan run strictly serially                                                                                                       |

## Architecture

```
internal/oidc/conformance/
    builder.go                      # OpenIDConnectConformanceSuiteBuilder, moved from cmd/authelia-gen
    types.go                        # plan, variant, server, client and resource types
    const.go                        # plan names, module names, profile names
internal/suites/OIDCConformance/
    compose.yml                     # mongodb, server, nginx, plus Authelia overrides
    Dockerfile.server               # published image + dev CA in the JVM truststore
    configuration.yml               # Authelia configuration for the suite
    users.yml                       # suite users
internal/suites/
    suite_oidc_conformance.go       # registration, setup, teardown
    suite_oidc_conformance_test.go  # OIDCConformanceSuite, one TestX method per profile
    conformance_api.go              # typed client for the conformance HTTP API
    conformance_runner.go           # per-plan runner: module loop, outcomes channel
    conformance_browser.go          # reactive page driver and the per-module override table
    conformance_names.go            # kebab-case module ids to PascalCase subtest names
```

`cmd/authelia-gen` keeps its `conformance` command and its output byte-for-byte, but imports the
builder rather than owning it. `MustHash` moves with the builder; the command-level flag plumbing in
`cmd_misc.go` stays where it is.

## Environment

### Containers

Four containers join the existing suite stack on `authelianet`:

- `conformance-mongodb` — `mongo:6.0.13`, no volume, state is disposable.
- `conformance-server` — built from `Dockerfile.server`. The published image
  `registry.gitlab.com/openid/conformance-suite:release-v5.2.4` is multi-arch (amd64 and arm64), so
  no Maven build is needed. The Dockerfile adds one layer:

  ```dockerfile
  FROM registry.gitlab.com/openid/conformance-suite:release-v5.2.4
  COPY ca.public.crt /tmp/authelia-ca.crt
  RUN keytool -importcert -noprompt -alias authelia-dev-ca -file /tmp/authelia-ca.crt -cacerts -storepass changeit
  ```

  This is required because the JVM fetches Authelia's discovery document, JWKS, token endpoint and
  userinfo endpoint over TLS signed by the suite development CA. Importing into `-cacerts` keeps the
  public roots intact, which a `-Djavax.net.ssl.trustStore` override would not.

  The image's `ENTRYPOINT` is shell-form, so a compose `command:` cannot override it. All
  configuration therefore goes through environment variables:

  | Variable          | Value                                  |
  | ----------------- | -------------------------------------- |
  | `BASE_URL`        | `https://conformance.example.com:8443` |
  | `MONGODB_HOST`    | `conformance-mongodb`                  |
  | `JAVA_EXTRA_ARGS` | `-Dfintechlabs.devmode=true`           |

  `JAVA_EXTRA_ARGS` is interpolated before `-jar`, so it becomes a JVM system property, which Spring
  Boot resolves the same as the upstream compose file's `--fintechlabs.devmode=true` application
  argument. Devmode disables the suite's Google/GitLab OAuth login, which is what makes the API
  reachable without a bearer token.

  The container takes the compose network alias `server`, because the upstream nginx configuration
  proxies to `http://server:8080` and that configuration is baked into the published nginx image.

- `conformance-nginx` — `registry.gitlab.com/openid/conformance-suite/nginx:release-v5.2.4`, pinned
  to `${SUITE_SUBNET:-192.168.240}.140` with the network alias `conformance.example.com`. Its
  self-signed certificate pair at `/etc/ssl/certs/nginx-selfsigned.crt` and
  `/etc/ssl/private/nginx-selfsigned.key` is bind-mounted over with the development CA's
  `*.example.com` pair from `internal/suites/common/pki`. Only port 8443 is used; the mutual-TLS
  listeners on 8444 and 8445 are irrelevant to these profiles.

- `authelia-backend` gains `X_AUTHELIA_CONFIG_FILTERS: 'template'`, a
  `${SUITE_TMP:-/tmp}:/tmp` bind, and
  `X_AUTHELIA_CONFIG: '/config/configuration.yml,/tmp/conformance-clients.yml'`.

### Networking

`HostEntries()` in `internal/suites/hosts.go` gains
`{Domain: "conformance.example.com", IP: SuiteAddress(140)}`, so the host-side browser reaches the
conformance suite by the same mechanism it already uses for the portal. The Go test process reaches
the API at `https://conformance.example.com:8443/api/` through an `http.Client` whose TLS config
trusts `internal/suites/common/pki/ca.public.crt`.

Both sides of every flow therefore resolve `conformance.example.com` to the same address: the
browser from the host via `/etc/hosts`, and Authelia from inside the network via the compose alias.
This matters because the redirect URIs registered with Authelia and the `BASE_URL` the conformance
server builds its authorization requests from must be string-identical.

## Configuration generation

`Suite.SetUp` runs inside the `authelia-scripts` process; the testify suite runs in a separate
`go test` process that `authelia-scripts` spawns afterwards. Nothing can be handed between them in
memory, so the generated configuration crosses the boundary as files in `SuiteTmpPath()`, which the
containers see at `/tmp` and which the test process reads back by absolute path.

`Suite.SetUp`, before `docker compose up`:

1. Build the seven suites from `internal/oidc/conformance` with `suiteURL` set to
   `https://conformance.example.com:8443` and `autheliaURL` set to `https://login.example.com:8080`.
2. Marshal the aggregated `Clients` into `SuiteTmpPath("conformance-clients.yml")`, which Authelia
   loads at `/tmp/conformance-clients.yml`.
3. Marshal the seven plans into `SuiteTmpPath("conformance-plans.json")` for the test process.

`Suite.SetUp`, after `docker compose up`:

4. Wait for Authelia via the existing `waitUntilAutheliaIsReady`.
5. Wait for the conformance server by polling `GET /api/plan?length=1` until it answers 200. The JVM
   takes appreciably longer to become ready than it takes to bind its port, so a port check is not
   sufficient.

`SetupSuite`, in the test process:

6. Read `conformance-plans.json` and, for each profile,
   `POST /api/plan?planName=<planName>&variant=<variant>` with the plan JSON as the body, expecting 201. The response carries `_id` and `modules`, so the module list needs no follow-up request.

Creating the plans in the test process rather than in setup also means a `--test-pattern` run against
an already-provisioned environment creates exactly the plans it is about to use.

Secrets stay randomly generated per run, exactly as `authelia-gen` generates them today. Because both
the plan JSON and the Authelia client definitions come from one builder invocation in one process,
they cannot drift.

The builder derives a plan's alias, and therefore its redirect URIs and client ids, from the `brand`
and `version` inputs. The suite passes a short random token as `version`, which makes every alias
unique per run without touching the builder, so repeated local runs against a container that was not
torn down cannot collide on the alias namespace.

## Conformance API client

`conformance_api.go` wraps the endpoints the runner needs. All of them are unauthenticated under
devmode.

| Method | Endpoint                           | Purpose                                                               |
| ------ | ---------------------------------- | --------------------------------------------------------------------- |
| `POST` | `/api/plan?planName=&variant=`     | Create a plan; 201; returns `_id` and `modules`                       |
| `GET`  | `/api/plan/{id}`                   | Plan detail, including `modules[].testModule` and `modules[].variant` |
| `POST` | `/api/runner?test=&plan=&variant=` | Create a module instance from a plan; 201; returns `id`               |

| `POST` | `/api/runner/{id}` | Start a module that is sitting in `CONFIGURED` |
| `GET` | `/api/runner/{id}/wait-state?states=&timeoutMs=` | Long-poll for a status transition |
| `GET` | `/api/info/{id}` | Persisted `status` and `result` |
| `GET` | `/api/runner/browser/{id}` | `urls` still to visit, and `visited` |
| `POST` | `/api/runner/browser/{id}/visit?url=` | Mark one URL visited; 204 |
| `GET` | `/api/log/{id}` | Log entries; unfilled placeholders carry an `upload` field |
| `POST` | `/api/log/{id}/images/{placeholder}` | Fill a placeholder with a data URI |
| `GET` | `/api/plan/exporthtml/{id}` | Plan HTML export, collected as a CI artifact |

`POST /api/runner` carries the plan's per-module `variant` alongside `plan`. The endpoint's own
documentation describes `variant` as applying to standalone tests, but the upstream
`scripts/conformance.py` sends both together for plan modules and the server honors it; that is the
behavior relied on here, and it is what makes the repeated-module plans work.

Two behaviors the client must handle explicitly:

- `wait-state` resolves against the _running_ test registry. Once a module finishes and is evicted it
  returns 404, and `timeoutMs` is clamped to 30 seconds server-side. The client therefore loops the
  long poll against an outer deadline and, on 404, falls back to `GET /api/info/{id}` for the
  persisted status before deciding the module is genuinely gone.
- `POST /api/runner` returns 409 when alias creation fails, which is the signal that two plans
  collided on an alias.

## Test structure

```
TestOIDCConformanceSuite
    TestConfig
        Discovery
        ...
    TestBasic
        Server
        ScopeAddress
        PromptLogin
        EnsureRequestWithoutNonceSucceedsForCodeFlow
        ...
    TestBasicFormPost
    TestHybrid
    TestHybridFormPost
    TestImplicit
    TestImplicitFormPost
```

`OIDCConformanceSuite` embeds `*RodSuite` and holds, per profile, a plan id and a channel of module
outcomes. Each profile method is a testify suite method, so testify names the subtest after the
method, giving `TestBasic` rather than a nested `t.Run` level.

### Naming

`conformance_names.go` converts a module id to a subtest name:

1. Strip a leading `oidcc-`.
2. Split on `-`, title-case each segment, concatenate.
3. If the plan lists the same `testModule` more than once, append the distinguishing variant: take
   the module's `variant` map in sorted key order, strip non-alphanumeric characters from each value,
   title-case, and concatenate.

So `oidcc-server` becomes `Server`, `oidcc-ensure-request-without-nonce-succeeds-for-code-flow`
becomes `EnsureRequestWithoutNonceSucceedsForCodeFlow`, and a hybrid-plan module repeated across
`response_type` becomes `ResponseTypeMissingCodeIdToken` and so on. The function is pure and gets a
table-driven unit test, including the collision case.

### Concurrency

testify's `suite.Suite` carries a single `*testing.T` that `suite.Run` reassigns per method, so
calling `t.Parallel()` inside a suite method races. Execution concurrency is therefore separated from
test structure:

- Suite setup starts one goroutine per profile. Each owns a `ConformanceRunner`: its own plan, its
  own rod incognito browser context, and its own ordered outcome channel.
- Each runner walks its plan's modules strictly in order, driving each to completion, and emits a
  `ConformanceModuleOutcome{Name, Variant, Result, Status, LogURL, Err}` per module.
- Each profile method ranges over its channel and asserts each outcome as `s.Run(name, ...)`.

Subtests therefore appear as modules complete rather than in one batch at the end, seven plans
progress at once, and no assertion ever runs off the test goroutine. Every go-rod assertion the
driver performs is captured into `Outcome.Err` and asserted inside the subtest, so a browser-side
failure is attributed to the module that caused it.

Using incognito contexts of one browser rather than seven browsers keeps the resource cost down and
gives each plan an isolated cookie jar, which several modules depend on.

## Module loop

For each module in plan order:

1. `POST /api/runner?test=<module>&plan=<planId>&variant=<variant>` to get a module id.
2. Wait for `CONFIGURED`, `WAITING` or `FINISHED`.
3. On `CONFIGURED`, `POST /api/runner/{id}` to start it, then wait for `WAITING` or `FINISHED`.
   `oidcc-server-rotate-keys` is the module that lands here.
4. On `WAITING`, run the browser driver until the module leaves `WAITING`. Modules such as
   `oidcc-prompt-login` require two full authorization round trips, so the driver loops on
   `GET /api/runner/browser/{id}` rather than assuming a single URL.
5. Poll `GET /api/log/{id}` for entries carrying an `upload` field. For each, run that module's
   assertion if the override table has one, then `POST` the stub PNG to
   `/api/log/{id}/images/{placeholder}`.
6. Wait for `FINISHED`, then read `result` from `GET /api/info/{id}`.

## Browser driver

The default path is reactive and module-agnostic. Given a URL from `urls`:

1. Navigate to it in the plan's browser context.
2. Classify the resulting page and act, repeating until the browser reaches the conformance suite's
   callback path or a terminal state:
   - Authelia first-factor form — sign in with the suite credentials via the existing
     `doLoginOneFactor` helper.
   - Authelia consent page — accept via the existing consent helpers.
   - Conformance callback page — done.
   - Authelia error page — record the error text and stop; whether that is a failure is the module's
     business, not the driver's.
3. `POST /api/runner/browser/{id}/visit?url=<url>` for the URL just handled.

The existing `verify_is_first_factor_page.go`, `verify_is_consent_page.go` and `action_login.go`
helpers are reused rather than reimplemented.

A small override table keyed by module id supplies only what differs from that default:

```go
type ConformanceOverride struct {
    ClearCookies bool
    Assert       func(leg ConformanceLeg) error
}
```

The assertion is handed what the driver recorded over one browser leg, not the live page. By the
time a leg ends the browser has left Authelia for the conformance suite's callback, so a probe of
the live DOM at that point answers a question about the callback page. `ConformanceLeg` carries the
leg's zero-based index alongside what the driver saw, which is also what makes these assertions
testable with no browser at all.

Entries are expected for the placeholder modules and the handful that expect a specific page. The
re-authentication expectations apply to the second authorization round trip only: the first merely
establishes the session, and the plan's browser context may already hold one from an earlier module.

- `oidcc-prompt-login` — assert the first-factor form is presented on the second authorization
  despite an existing session.
- `oidcc-max-age-1` — assert re-authentication is prompted on the second authorization.
- `oidcc-max-age-10000` — assert re-authentication is _not_ prompted on the second authorization.
- `oidcc-display-page`, `oidcc-display-popup`, `oidcc-ui-locales`, `oidcc-claims-locales` — clear
  cookies first, assert a normal login page renders.

Anything absent from the table falls through to the reactive driver, so a module added by a future
conformance suite release runs rather than panicking.

## Placeholders

`OIDCCPromptLogin`, `OIDCCMaxAge1` and `OIDCCMaxAge10000` call `waitForPlaceholders()` and hold the
module in `WAITING` until an image is uploaded. `POST /api/log/{id}/images/{placeholder}` accepts a
plain-text body that must begin `data:image/png;base64,` or `data:image/jpeg;base64,`, caps the
decoded size at 500KB and allows at most two images per test, and sets the module's result to
`REVIEW` on success.

The suite therefore treats the upload as protocol plumbing, not evidence. A single package-level
1×1 PNG data URI constant is posted to release the wait. The real verdict for those modules is the
assertion from the override table, which runs against the recorded leg before the upload and lands
in `Outcome.Err`.

This is why `REVIEW` is in the accepted result set: the upload forces it, so it carries no signal.

## Verdicts

```go
var conformanceAcceptedResults = []string{"PASSED", "REVIEW", "SKIPPED"}
```

A subtest fails when the module's result is outside that set, or when `Outcome.Err` is non-nil.
`WARNING` is one entry away from being accepted, which is the configurability asked for.

Modules that cannot run unattended sit in a documented allowlist and report as skipped with a
reason rather than failing. `oidcc-server-rotate-keys` is the one known entry: it requires the OP's
signing keys to be rotated by hand while the module waits. The allowlist is deliberately minimal;
the full set is discovered from the first real runs rather than guessed, and every entry carries a
comment explaining why it cannot run. A module leaving the allowlist becomes a real failure.

## Diagnostics

On a failing subtest the suite logs the module's `log-detail.html?log=<id>` URL and dumps its
`/api/log/{id}` entries. Teardown pulls `GET /api/plan/exporthtml/{id}` for each plan into the
`screenshots/` tree that `e2etests.sh` already declares in `artifact_paths`, and prints the
conformance server container logs alongside Authelia's through the existing `PrintLogs` hook.

## CI

`.buildkite/steps/e2etests.sh` gains a `SUITE_TIMEOUTS[OIDCConformance]` entry of 120 minutes. That
number contains the suite's own Go-side budget rather than guessing at a wall clock: `SetUpTimeout`
15m + `TestTimeout` 85m + `TearDownTimeout` 5m is 105 minutes, and `TestTimeout` in turn contains
`conformancePlanTimeout` (75m) plus teardown's export loop and runner drain. The ordering matters
more than the values — the plan context must be what ends an overrun, because that is the path that
drains, exports and reports. All of it remains an estimate until a real run produces a wall clock.

`.buildkite/pipeline.sh` needs no change. Its `BUILD_DUO` / `BUILD_HAPROXY` / `BUILD_SAMBA` flags trigger separate
pipelines that publish standalone images to `authelia/integration-*`; the conformance server image is built inline by
compose instead, the way `internal/suites/example/compose/pam/compose.yml` already does, and that suite carries no such
flag either.

## Testing the suite's own code

The suite drives external infrastructure, so the parts that can be tested in isolation are tested in
isolation, following the repository's existing table-driven style:

- `conformance_names.go` — pure, table-driven tests including repeated-module variant suffixes.
- `internal/oidc/conformance` — the builder's existing behavior is pinned by moving
  `cmd/authelia-gen/openid_conformance_test.go` with it, so the extraction is provably
  behavior-preserving.
- `conformance_api.go` — `httptest` server tests covering the 404-after-finish fallback, the
  `wait-state` timeout loop and the 409 alias collision.
- The reactive driver's page classification is a pure function over an observed page state, tested
  without a browser.

## Risks and open items

- **Wall-clock.** Seven concurrent plans of roughly 30 modules each, every one a browser round trip.
  the timeouts above are estimates. If the real number is materially worse, the reduced-plan-set fallback
  discussed during design remains available.
- **Allowlist size.** Only `oidcc-server-rotate-keys` is confidently known to be unrunnable
  unattended. Others may surface on first run and will be added with justifications.
- **POST-method browser URLs.** The conformance UI distinguishes URLs that must be reached by form
  POST from ones reachable by GET. That distinction is not expected to arise for OP profiles, where
  the URLs are authorization requests, but the driver should fail loudly rather than silently if a
  URL it cannot handle by navigation appears.
- **Image pinning.** `release-v5.2.4` is pinned deliberately. Conformance suite releases add and
  rename modules, so the pin is the thing that keeps the suite reproducible; bumping it is a
  deliberate change with an expected diff in subtest names.
- **Suite runtime on constrained agents.** Seven browser contexts plus a JVM plus MongoDB on top of
  the existing stack is a heavier agent footprint than any current suite.
