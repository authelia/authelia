---
layout: default
title: Authelia Scripts
parent: Contributing
---

# Authelia Scripts

Authelia comes with a set of dedicated scripts doing a broad range of operations such as
building the distributed version of Authelia, building the Docker image, running suites,
testing the code, etc...

Those scripts becomes available after sourcing the bootstrap.sh script with

    source bootstrap.sh

Then, you can access the scripts usage by running the following command:

    authelia-scripts --help

For instance, you can build Authelia (Go binary and frontend) with:

    authelia-scripts build

Or build the official Docker image with:

    authelia-scripts docker build

Or start the *Standalone* suite with:

    authelia-scripts suites setup Standalone

You will find more information in the scripts usage helpers.