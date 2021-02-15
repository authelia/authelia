---
layout: default
title: Style Guide
parent: Contributing
nav_order: 4
---

# Style Guide

This is a general guide to the code style of our commits. This is by no means an exhaustive list and we're constantly
changing and improving it. This is also a work in progress document.

For our commit messages please see our [Commit Message Guidelines](./commitmsg-guidelines.md).

## Tools

We implement the following tools that help us abide by our style guide and include the configuration for them inside
our repository:
- [golangci-lint](https://github.com/golangci/golangci-lint)
- [yamllint](https://yamllint.readthedocs.io/en/stable/)

## Exceptions

This is a style **guide**, there are always going to be exceptions to these guidelines when it makes sense not to follow
them. One notable exception is the README.md for the repository. The line length of the 
[All Contributors](https://allcontributors.org/) individual sections are longer than 120 characters and it doesn't make
sense to apply the [line length](#line-length) guidelines.

## Specific Guidelines

### Line Length

We aim to keep all files to a maximum line length of 120 characters. This allows for most modern computer systems to
display two files side by side (vertically split). This includes but is not limited to the following file types:
- Go (*.go)
- YAML (*.yml, *.yaml)
- Markdown (*.md)
- TypeScript (*.ts, *.tsx)