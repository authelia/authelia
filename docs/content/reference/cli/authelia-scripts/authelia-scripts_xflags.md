---
title: "authelia-scripts xflags"
description: "Reference for the authelia-scripts xflags command."
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

## authelia-scripts xflags

Generate X LDFlags for building Authelia

### Synopsis

Generate X LDFlags for building Authelia.

```
authelia-scripts xflags [flags]
```

### Examples

```
authelia-scripts xflags
```

### Options

```
  -b, --build string   Sets the BuildNumber flag value (default "0")
  -e, --extra string   Sets the BuildExtra flag value
  -h, --help           help for xflags
```

### Options inherited from parent commands

```
      --buildkite          Set CI flag for Buildkite
      --log-level string   Set the log level for the command (default "info")
```

### SEE ALSO

* [authelia-scripts](authelia-scripts.md)	 - A utility used in the Authelia development process.

