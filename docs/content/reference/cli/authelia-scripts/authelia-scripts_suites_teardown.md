---
title: "authelia-scripts suites teardown"
description: "Reference for the authelia-scripts suites teardown command."
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

## authelia-scripts suites teardown

Teardown a test suite environment

### Synopsis

Teardown a test suite environment.

Suites can be listed with the authelia-scripts suites list command.

```
authelia-scripts suites teardown [suite] [flags]
```

### Examples

```
authelia-scripts suites setup Standalone
```

### Options

```
  -h, --help   help for teardown
```

### Options inherited from parent commands

```
      --buildkite          Set CI flag for Buildkite
      --log-level string   Set the log level for the command (default "info")
```

### SEE ALSO

* [authelia-scripts suites](authelia-scripts_suites.md)	 - Commands related to suites management

