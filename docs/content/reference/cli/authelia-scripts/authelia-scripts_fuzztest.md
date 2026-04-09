---
title: "authelia-scripts fuzztest"
description: "Reference for the authelia-scripts fuzztest command."
lead: ""
date: 2026-04-04T11:57:43+11:00
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

## authelia-scripts fuzztest

Run fuzz tests

### Synopsis

Run fuzz tests.

This command discovers and runs Go fuzz tests with configurable time budgets.

```
authelia-scripts fuzztest [flags]
```

### Examples

```
authelia-scripts fuzztest ./...
authelia-scripts fuzztest --individual-budget 2m --total-budget 8m ./...
```

### Options

```
  -h, --help                         help for fuzztest
      --individual-budget duration   Individual budget for each fuzz test. (default 2m0s)
      --total-budget duration        Total budget for all fuzz tests. (default 8m0s)
```

### Options inherited from parent commands

```
      --buildkite          Set CI flag for Buildkite
      --log-level string   Set the log level for the command (default "info")
```

### SEE ALSO

* [authelia-scripts](authelia-scripts.md)	 - A utility used in the Authelia development process.

