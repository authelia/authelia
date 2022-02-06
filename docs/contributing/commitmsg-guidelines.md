---
layout: default
title: Commit Message Guidelines
parent: Contributing
nav_order: 3
---

# Commit Message Guidelines

## The reasons for these conventions:

- simple navigation though and easier to read git history

## Format of the commit message:

Each commit message consists of a **header**, a **body**, and a **footer**.

```bash
<header>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

The `header` is mandatory and must conform to the [Commit Message Header](#commit-message-header) format.
The header cannot be longer than 72 characters.

The `body` is mandatory for all commits except for those of type "docs".
When the body is present it must be at least 20 characters long and must conform to the 
[Commit Message Body](#commit-message-body) format.

The `footer` is optional. The [Commit Message Footer](#commit-message-footer) format describes what the footer is used 
for, and the structure it must have.

### Commit Message Header

```
<type>(<scope>): <summary>
  │       │             │
  │       │             └─⫸ Summary in present tense. Not capitalized. No period at the end.
  │       │
  │       └─⫸ Commit Scope: api|authentication|authorization|cmd|commands|configuration|duo|
  │                          handlers|logging|middlewares|mocks|models|notification|oidc|
  │                          regulation|server|session|storage|suites|templates|utils|web
  │
  └─⫸ Commit Type: build|ci|docs|feat|fix|perf|refactor|release|test
```

The `<type>` and `<summary>` fields are mandatory, the `(<scope>)` field is optional.

#### Allowed `<type>` values:

* **build** Changes that affect the build system or external dependencies 
  (example scopes: bundler, deps, docker, go, npm)
* **ci** Changes to our CI configuration files and scripts 
  (example scopes: autheliabot, buildkite, codecov, golangci-lint, renovate, reviewdog)
* **docs** Documentation only changes
* **feat** A new feature
* **fix** A bug fix
* **perf** A code change that improves performance
* **refactor** A code change that neither fixes a bug nor adds a feature
* **release** Releasing a new version of Authelia
* **test** Adding missing tests or correcting existing tests

#### Allowed `<scope>` values:

The scope should be the name of the package affected 
(as perceived by the person reading the changelog generated from commit messages).

* authentication
* authorization
* commands
* configuration
* duo
* handlers
* logging
* middlewares
* mocks
* models
* notification
* ntp
* oidc
* regulation
* server
* session
* storage
* suites
* templates
* totp
* utils

There are currently a few exceptions to the "use package name" rule:

* `api`: used for changes that change the openapi specification

* `cmd`: used for changes to the `authelia|authelia-scripts|authelia-suites` top level binaries

* `web`: used for changes to the React based frontend

* none/empty string: useful for `test`, `refactor` and changes that are done across multiple packages 
  (e.g. `test: add missing unit tests`) and for docs changes that are not related to a 
  specific package (e.g. `docs: fix typo in tutorial`).

#### Summary

Use the summary field to provide a succinct description of the change:

* use the imperative, present tense: "change" not "changed" nor "changes"
* don't capitalize the first letter
* no dot (.) at the end


### Commit Message Body

Just as in the summary, use the imperative, present tense: "fix" not "fixed" nor "fixes".

Explain the motivation for the change in the commit message body. This commit message should explain _why_ you are 
making the change. You can include a comparison of the previous behavior with the new behavior in order to illustrate 
the impact of the change.


### Commit Message Footer

The footer can contain information about breaking changes and is also the place to reference GitHub issues and other PRs 
that this commit closes or is related to.

```
BREAKING CHANGE: <breaking change summary>
<BLANK LINE>
<breaking change description + migration instructions>
<BLANK LINE>
<BLANK LINE>
Fixes #<issue number>
```

Breaking Change section should start with the phrase "BREAKING CHANGE: " followed by a summary of the breaking change, a 
blank line, and a detailed description of the breaking change that also includes migration instructions.


### Revert commits

If the commit reverts a previous commit, it should begin with `revert: `, followed by the header of the reverted commit.

The content of the commit message body should contain:

- information about the SHA of the commit being reverted in the following format: `This reverts commit <SHA>`,
- a clear description of the reason for reverting the commit message.


## Example commit message:

```bash
fix(logging): disabled colored logging outputs when file is specified

In some scenarios if a user has a log_file_path specified and a TTY seems to be detected this causes terminal coloring outputs to be written to the file.
This in turn will cause issues when attempting to utilise the log with the provided fail2ban regexes.

We now override any TTY detection/logging treatments and disable coloring/removal of the timestamp when a user is utilising the text based logger to a file.

Fixes #1480.
```

This document is based on [AngularJS Git Commit Message Format].

[AngularJS Git Commit Message Format]: https://github.com/angular/angular/blob/master/CONTRIBUTING.md#commit