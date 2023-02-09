---
title: "docs/content/en/reference/cli/authelia-scripts/authelia-scripts xflags"
description: "Reference for the docs/content/en/reference/cli/authelia-scripts/authelia-scripts xflags command."
lead: ""
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  reference:
    parent: "cli-docs/content/en/reference/cli/authelia-scripts/authelia-scripts"
weight: 995
toc: true
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

