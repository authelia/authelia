---
title: "Testing"
description: "Authelia Development Testing Guidelines"
summary: "This section covers the testing guidelines."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 320
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The following outlines the specific requirements we have for testing the Authelia code contributions.

- While we aim for 100% coverage on changes and additions, we do not enforce this where it doesn't make practical sense:
  - A test which just marks a line as tested is not necessarily an effectual test
  - Sometimes there is limited ways in which tests can be performed and the limitation makes the test ineffectual
- Tests should be named to reflect what they testing for and which part of the code they are testing
- It's required for bug fixes that contributors create a test that fails prior to and passes
  subsequent to the fix being applied, this test must be included in the contribution, excluding this test will likely
  result in the fix being rejected unless explicitly agreed and advised otherwise by the
  [core team](../../information/about.md#core-team)
- It's strongly encouraged for features that contributors create have as much testing as is reasonable i.e. any line
  that can be tested should be tested, if the line can't be tested generally this is an indication a refactor may be
  required

## Testing Methodology and Frequency

We run tests using several test framework and platforms and aim to ensure both SAST and DAST scanning tools are happy
with the code. The rational for this is to ensure that the code is not vulnerable to any known security issues, and
while having more tools increases the noise we can adequately make a better judgment call with more information at our
disposal.

|                                          Tool                                          |                  Purpose                   |                                                              Notes                                                               |
|:--------------------------------------------------------------------------------------:|:------------------------------------------:|:--------------------------------------------------------------------------------------------------------------------------------:|
|                         [Go Test](https://pkg.go.dev/testing)                          | Coverage, Static and Dynamic Code Analysis | Analysis of Go Code, Executed with `go test -cover`, `go test -race`, and `go test -fuzz` before and on every commit to `msster` |
| [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/) | Coverage, Static and Dynamic Code Analysis |                                  Analysis of React Code before and on every commit to `msster`                                   |
|                        [SonarQube](https://www.sonarqube.org/)                         |            Static Code Analysis            |                                   Analysis of All Code before and on every commit to `msster`                                    |
|                          [CodeQL](https://codeql.github.com/)                          |            Static Code Analysis            |                          Analysis of All Code before and on every commit to `msster`, and on a schedule                          |
|                             [Codecov](https://codecov.io/)                             |            Coverage Statistics             |                         Produces Statistics for Go and TypeScript before and on every commit to `msster`                         |
|                       [Grype](https://github.com/anchore/grype)                        |          Vulnerability Management          |                                    SBOM Scanning Only before and on every commit to `msster`                                     |
|                          [Renovate](https://renovatebot.com/)                          |  Vulnerability and Dependency Management   |                                                          On a schedule                                                           |
|                      [golangci-Lint](https://golangci-lint.run/)                       |            Static Code Analysis            |                                    Analysis of Go Code before and on every commit to `msster`                                    |
|                      [GitGuardian](https://www.gitguardian.com/)                       |             Secrets Management             |                                    Analysis of Secrets before and on every commit to `msster`                                    |
|                       [Code Rabbit](https://www.coderabbit.ai/)                        |      Quality and Security Assessment       |                                Analysis of General Pull Requests before every commit to `master`                                 |
|              [OpenSSF Scorecard](https://openssf.org/projects/scorecard/)              |       Security Practices Assessment        |                                            Automated on every new commit to `master`                                             |
|      [OpenSSF Best Practices](https://openssf.org/projects/best-practices-badge/)      |       Security Practices Assessment        |                                   Manual Assessment for Security Practice Posture Improvements                                   |
|        [StepSecurity Harden-Runner](https://docs.stepsecurity.io/harden-runner)        |             CI Agent Security              |                                       As Part of any Job Running in GitHub CI Job Runners                                        |
