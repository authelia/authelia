---
title: "Reference: authelia-scripts"
description: "This section covers the authelia-scripts tool."
summary: "This section covers the authelia-scripts tool."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 290
toc: true
aliases:
  - /docs/contributing/authelia-scripts.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

__Authelia__ comes with a set of dedicated scripts to perform a broad range of operations such as building the
distributed version of Authelia, building the Docker image, running suites, testing the code, etc. This is a small
reference guide for the command, the full guide can be found in the
[CLI Reference](../../reference/cli/authelia-scripts/authelia-scripts.md).

## Examples

Those scripts become available after sourcing the `bootstrap.sh` script with:

```bash
source bootstrap.sh
```

Then, you can access the scripts usage by running the following command:

```bash
authelia-scripts --help
```

For instance, you can build __Authelia__ ([go] binary and [React] frontend) with:

```bash
authelia-scripts build
```

Or build the official [Docker] image with:

```bash
authelia-scripts docker build
```

Or start the *Standalone* suite with:

```bash
authelia-scripts suites setup Standalone
```

## Help

The `authelia-scripts` provides help using the `--help` or `-h` flags. Every command should provide some form of help
when provided with either flag. Examples:

```bash
authelia-scripts --help
```

```bash
authelia-scripts build --help
```
