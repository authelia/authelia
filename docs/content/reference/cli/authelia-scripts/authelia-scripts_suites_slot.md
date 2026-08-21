---
title: "authelia-scripts suites slot"
description: "Reference for the authelia-scripts suites slot command."
lead: ""
date: 2026-08-20T08:52:40+10:00
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

## authelia-scripts suites slot

Show the suite slot allocated to this working tree

### Synopsis

Show the suite slot allocated to this working tree, allocating one if it does not have it yet.

The slot is the number bootstrap.sh derives the compose project, the network subnet, the debug ports and the temporary
directory from, so that several working trees on one machine can run suites at the same time without colliding.

```
authelia-scripts suites slot [flags]
```

### Examples

```
authelia-scripts suites slot
authelia-scripts suites slot --list
authelia-scripts suites slot --release
```

### Options

```
  -h, --help      help for slot
      --list      Lists every allocated slot instead of allocating one
      --release   Releases the slot allocated to this working tree
```

### Options inherited from parent commands

```
      --buildkite          Set CI flag for Buildkite
      --log-level string   Set the log level for the command (default "info")
```

### SEE ALSO

* [authelia-scripts suites](authelia-scripts_suites.md)	 - Commands related to suites management

