---
title: "authelia-scripts suites external test"
description: "Reference for the authelia-scripts suites external test command."
lead: ""
date: 2026-04-30T17:13:00+10:00
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

## authelia-scripts suites external test

Run an external test suite

### Synopsis

Run an external test suite.

External suites can be listed with the authelia-scripts suites external list command.

```
authelia-scripts suites external test [suite] [flags]
```

### Examples

```
authelia-scripts suites external test docs
```

### Options

```
      --failfast           Stops tests on first failure
      --headless           Run tests in headless mode
  -h, --help               help for test
      --update-snapshots   Overwrite visual snapshot baselines with the output of the current run
```

### Options inherited from parent commands

```
      --buildkite          Set CI flag for Buildkite
      --log-level string   Set the log level for the command (default "info")
```

### SEE ALSO

* [authelia-scripts suites external](authelia-scripts_suites_external.md)	 - Commands related to external suites management

