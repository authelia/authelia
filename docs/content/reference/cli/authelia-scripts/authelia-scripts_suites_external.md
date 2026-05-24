---
title: "authelia-scripts suites external"
description: "Reference for the authelia-scripts suites external command."
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

## authelia-scripts suites external

Commands related to external suites management

### Synopsis

Commands related to external suites management.

External suites drive a project-local dev server and use the go-rod browser harness to assert the rendered output is correct.

```
authelia-scripts suites external [flags]
```

### Examples

```
authelia-scripts suites external
```

### Options

```
  -h, --help   help for external
```

### Options inherited from parent commands

```
      --buildkite          Set CI flag for Buildkite
      --log-level string   Set the log level for the command (default "info")
```

### SEE ALSO

* [authelia-scripts suites](authelia-scripts_suites.md)	 - Commands related to suites management
* [authelia-scripts suites external list](authelia-scripts_suites_external_list.md)	 - List available external suites
* [authelia-scripts suites external test](authelia-scripts_suites_external_test.md)	 - Run an external test suite

