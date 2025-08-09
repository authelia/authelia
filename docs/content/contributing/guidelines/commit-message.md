---
title: "Commit Message"
description: "Authelia Development Commit Message Guidelines"
lead: "This section covers the git commit message guidelines we use for development."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 320
toc: true
aliases:
  - /docs/contributing/commitmsg-guidelines.html
  - /contributing/development/guidelines-commit-message/
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The reasons for these conventions are as follows:

* simple navigation though git history
* easier to read git history

## Commit Message Format

Each commit message consists of a __header__, a __body__, and a __footer__.

```bash
<header>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

The `header` is mandatory and must conform to the [Commit Message Header](#commit-message-header) format. The header
cannot be longer than 72 characters.

The `body` is mandatory for all commits except for those of type "docs". When the body is present it must be at least 20
characters long and must conform to the [Commit Message Body](#commit-message-body) format.

The `footer` is optional. The [Commit Message Footer](#commit-message-footer) format describes what the footer is used
for, and the structure it must have.

### Commit Message Header

```text
<type>(<scope>): <summary>
  │       │             │
  │       │             └─⫸ Summary in present tense. Not capitalized. No period at the end.
  │       │
  │       └─⫸ Commit Scope: api|autheliabot|authentication|authorization|buildkite|bundler|clock|
  │                          cmd|codecov|commands|configuration|deps|docker|duo|expression|go|
  │                          golangci-lint|handlers|lefthook|logging|metrics|middlewares|mocks|
  │                          model|notification|npm|ntp|oidc|random|regulation|renovate|reviewdog|
  │                          server|service|session|storage|suites|templates|totp|utils|web|
  │                          webauthn
  │
  └─⫸ Commit Type: build|ci|docs|feat|fix|i18n|perf|refactor|release|revert|test
```

The `<type>` and `<summary>` fields are mandatory, the `(<scope>)` field is optional.

#### Allowed type values:

* __build__ Changes that affect the build system or external dependencies
  (example scopes: bundler, deps, docker, go, npm)
* __ci__ Changes to our CI configuration files and scripts
  (example scopes: autheliabot, buildkite, codecov, lefthook, golangci-lint, renovate, reviewdog)
* __docs__ Documentation only changes
* __feat__ A new feature
* __fix__ A bug fix
* __i18n__ Updating translations or internationalization settings
* __perf__ A code change that improves performance
* __refactor__ A code change that neither fixes a bug nor adds a feature
* __release__ Releasing a new version of Authelia
* __test__ Adding missing tests or correcting existing tests

#### Allowed scope values:

The scope should be the name of the package affected (as perceived by the person reading the changelog generated from
commit messages).

* authentication
* authorization
* clock
* commands
* configuration
* duo
* expression
* handlers
* logging
* metrics
* middlewares
* mocks
* model
* notification
* ntp
* oidc
* random
* regulation
* server
* service
* session
* storage
* suites
* templates
* totp
* utils
* webauthn

There are currently a few exceptions to the "use package name" rule:

* `api`: used for changes that change the openapi specification
* `cmd`: used for changes to the `authelia|authelia-gen|authelia-scripts|authelia-suites` top level binaries
* `web`: used for changes to the React based frontend
* none/empty string: useful for `test`, `refactor` and changes that are done across multiple packages
  (e.g. `test: add missing unit tests`) and for docs changes that are not related to a specific package
  (e.g. `docs: fix typo in tutorial`).

#### Summary

Use the summary field to provide a succinct description of the change:

* use the imperative, present tense: "change" not "changed" nor "changes"
* don't capitalize the first letter
* no dot (.) at the end

### Commit Message Body

Just as in the summary, use the imperative, present tense: "fix" not "fixed" nor "fixes".

Explain the motivation for the change in the commit message body. This commit message should explain *why* you are
making the change. You can include a comparison of the previous behavior with the new behavior in order to illustrate
the impact of the change.

### Commit Message Footer

The footer can contain information about breaking changes and is also the place to reference GitHub issues and other PRs
that this commit closes or is related to.

```text
BREAKING CHANGE: <breaking change summary>
<BLANK LINE>
<breaking change description + migration instructions>
<BLANK LINE>
<BLANK LINE>
Fixes #<issue number>

Signed-off-by: <AUTHOR>
```

Breaking Change section should start with the phrase "BREAKING CHANGE: " followed by a summary of the breaking change, a
blank line, and a detailed description of the breaking change that also includes migration instructions.

### Revert Commits

If the commit reverts a previous commit, it should begin with `revert:`, followed by the header of the reverted commit.

The content of the commit message body should contain:

* information about the SHA of the commit being reverted in the following format: `This reverts commit <SHA>`,
* a clear description of the reason for reverting the commit message.

## Commit Message Examples

```bash
fix(logging): disable colored logging outputs when file is specified

In some scenarios if a user has a log_file_path specified and a TTY seems to be detected this causes terminal coloring outputs to be written to the file.
This in turn will cause issues when attempting to utilise the log with the provided fail2ban regexes.

We now override any TTY detection/logging treatments and disable coloring/removal of the timestamp when a user is utilising the text based logger to a file.

Fixes #1480.

Signed-off-by: John Smith <jsmith@org.com>
```

This document is based on [AngularJS Git Commit Message Format].

[AngularJS Git Commit Message Format]: https://github.com/angular/angular/blob/master/CONTRIBUTING.md#commit
