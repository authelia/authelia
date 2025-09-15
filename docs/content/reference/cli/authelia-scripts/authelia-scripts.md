---
title: "authelia-scripts"
description: "Reference for the authelia-scripts command."
lead: ""
date: 2025-08-01T16:23:47+10:00
draft: false
images: []
weight: 920
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## authelia-scripts

A utility used in the Authelia development process.

### Synopsis

The authelia-scripts utility is utilized by developers and the CI/CD pipeline for configuring
doing various development and testing tasks.

It can be used to automate or manually run unit testing, integration testing, etc. It allows setting up a concept of
suites which include various applications which replicate a functional Authelia environment using docker compose with
the ability for Authelia to perform hot reload for the purpose of development.

This can either be ran directly via go or you can leverage the development environment context by executing
'source bootstrap.sh' from the root of the repository.

Commonly used commands are as follows:

authelia-scripts build - builds the authelia go binary and react frontend
authelia-scripts docker build - builds the authelia docker image
authelia-scripts suites setup Standalone - sets up the Standalone suite (there are many other suites)


### Examples

```
authelia-scripts help
```

### Options

```
      --buildkite          Set CI flag for Buildkite
  -h, --help               help for authelia-scripts
      --log-level string   Set the log level for the command (default "info")
```

### SEE ALSO

* [authelia-scripts bootstrap](authelia-scripts_bootstrap.md)	 - Prepare environment for development and testing
* [authelia-scripts build](authelia-scripts_build.md)	 - Build Authelia binary and static assets
* [authelia-scripts ci](authelia-scripts_ci.md)	 - Run the continuous integration script
* [authelia-scripts clean](authelia-scripts_clean.md)	 - Clean build artifacts
* [authelia-scripts docker](authelia-scripts_docker.md)	 - Commands related to building and publishing docker image
* [authelia-scripts serve](authelia-scripts_serve.md)	 - Serve compiled version of Authelia
* [authelia-scripts suites](authelia-scripts_suites.md)	 - Commands related to suites management
* [authelia-scripts unittest](authelia-scripts_unittest.md)	 - Run unit tests
* [authelia-scripts xflags](authelia-scripts_xflags.md)	 - Generate X LDFlags for building Authelia

