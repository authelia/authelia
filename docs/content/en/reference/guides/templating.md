---
title: "Templating"
description: "A reference guide on the templates system"
lead: "This section contains reference documentation for Authelia's templating capabilities."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  reference:
    parent: "guides"
weight: 220
toc: true
---

Authelia has several methods where users can interact with templates.

## Functions

Functions can be used to perform specific actions when executing templates. The following is a simple guide on which
functions exist.

### Standard Functions

Go has a set of standard functions which can be used. See the [Go Documentation](https://pkg.go.dev/text/template#hdr-Functions)
for more information.

### Helm-like Functions

The following functions which mimic the behaviour of helm exist in most templating areas:

- env
- expandenv
- split
- splitList
- join
- contains
- hasPrefix
- hasSuffix
- lower
- upper
- title
- trim
- trimAll
- trimSuffix
- trimPrefix
- replace
- quote
- sha1sum
- sha256sum
- sha512sum
- squote
- now

See the [Helm Documentation](https://helm.sh/docs/chart_template_guide/function_list/) for more information. Please
note that only the functions listed above are supported.

__*Special Note:* The `env` and `expandenv` function automatically excludes environment variables that start with
`AUTHELIA_` or `X_AUTHELIA_` and end with one of `KEY`, `SECRET`, `PASSWORD`, `TOKEN`, or `CERTIFICATE_CHAIN`.__

### Special Functions

The following is a list of special functions and their syntax.

#### iterate

Input is a single uint. Returns a slice of uints from 0 to the provided uint.
