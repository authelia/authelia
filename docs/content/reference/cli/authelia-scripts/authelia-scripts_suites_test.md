---
title: "authelia-scripts suites test"
description: "Reference for the authelia-scripts suites test command."
lead: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 925
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## authelia-scripts suites test

Run a test suite

### Synopsis

Run a test suite.

Suites can be listed with the authelia-scripts suites list command.

```
authelia-scripts suites test [suite] [flags]
```

### Examples

```
authelia-scripts suites test Standalone
```

### Options

```
      --failfast      Stops tests on first failure
      --headless      Run tests in headless mode
  -h, --help          help for test
      --test string   The single test to run
```

### Options inherited from parent commands

```
      --buildkite          Set CI flag for Buildkite
      --log-level string   Set the log level for the command (default "info")
```

### SEE ALSO

* [authelia-scripts suites](authelia-scripts_suites.md)	 - Commands related to suites management

