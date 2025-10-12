---
title: "Testing"
description: "A description of testing procedures for Authelia"
summary: "This section contains reference documentation for Authelia."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 151
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

This section documents Authelia's testing processes and procedures to ensure code quality and reliability.

## Types of Testing

Authelia performs multiple types of testing:
- Linting to enforce code quality and uniform style.
- Unit testing to test code in isolation from other code to protect against regressions and unintended behavior.
- Integration testing to test the application in combination with various external dependencies including:
  - Databases
  - Reverse Proxies
  - Memory Caching
  - Identity Backends (LDAP, AD, File)
- End-to-End testing to test application features work properly under all supported conditions.
- Automated Secrets Scanning using [GitGuardian].
- GitHub Runner Monitoring using [StepSecurity Harden-Runner].

## Automated Testing

### Continuous Integration

Testing is automatically executed on:
- Every push to a branch.
- Every pull request submission and revision, subject to approval for 3rd party pull requests.
- Merges to the main branch.
- Documentation-only changes do not trigger the full test suite.


### Test Execution Pipeline

We make use of [Buildkite] as our CI/CD platform. All automated testing for public code is visible from the [Authelia Buildkite Dashboard](https://buildkite.com/authelia).
A complete testing suite run averages ~9 minutes, subject to fluctuation.
Testing agents are run on our hosted infrastructure using our [custom runner image.](https://github.com/authelia/buildkite)

### Test Coverage

Authelia maintains comprehensive code coverage metrics to ensure thorough testing of the codebase:
- Coverage reports are automatically generated for every test run.
- Current coverage statistics and trends are available on our [Codecov dashboard](https://app.codecov.io/gh/authelia/authelia).
- Coverage is tracked separately for:
  - **Backend**: Go code in `cmd/authelia/` and `internal/` directories
  - **Frontend**: JS/TS code in `web/` directory
- Coverage targets: 70-100% with precision to 2 decimals.
- Pull requests must not decrease coverage by more than 0.15% for either component.
- Coverage is tracked for unit and integration testing.
- Coverage excludes test mocks, test suites, and generated files.
- Pull requests display coverage changes to maintain visibility during development.

## Merge and Release Requirements

### Pull Request Testing

- All testing suites must pass before merge approval.
- Tests run automatically via [Buildkite].

### Release Testing

- Full test suite execution before any release.
- For additional security measures related to releases see our [security measures](../../overview/security).

### Test Validation

- Test results are automatically reported in pull requests.
- Release candidates require 100% pass rate.



[Buildkite]: https://buildkite.com
[Codecov]: https://app.codecov.io/gh/authelia/authelia
[GitGuardian]: https://www.gitguardian.com/
[StepSecurity Harden-Runner]: https://github.com/step-security/harden-runner
