---
title: "authelia-scripts docker build"
description: "Reference for the authelia-scripts docker build command."
lead: ""
date: 2025-08-01T16:23:47+10:00
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

## authelia-scripts docker build

Build the docker image of Authelia

### Synopsis

Build the docker image of Authelia.

```
authelia-scripts docker build [flags]
```

### Examples

```
authelia-scripts docker build
```

### Options

```
      --container string   target container among: dev, coverage (default "dev")
  -h, --help               help for build
```

### Options inherited from parent commands

```
      --buildkite          Set CI flag for Buildkite
      --log-level string   Set the log level for the command (default "info")
```

### SEE ALSO

* [authelia-scripts docker](authelia-scripts_docker.md)	 - Commands related to building and publishing docker image

