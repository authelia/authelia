---
layout: default
title: Authelia Scripts
parent: Contributing
nav_order: 2
---

# Authelia Scripts

Authelia comes with a set of dedicated scripts to perform a broad range of operations such as building the distributed 
version of Authelia, building the Docker image, running suites, testing the code, etc...

Those scripts become available after sourcing the bootstrap.sh script with

```console
$ source bootstrap.sh
```

Then, you can access the scripts usage by running the following command:

```console
$ authelia-scripts --help
```

For instance, you can build Authelia (Go binary and frontend) with:

```console
$ authelia-scripts build
```
    

Or build the official Docker image with:

```console
$ authelia-scripts docker build
```

Or start the *Standalone* suite with:

```console
$ authelia-scripts suites setup Standalone
```

You will find more information in the scripts usage helpers.
