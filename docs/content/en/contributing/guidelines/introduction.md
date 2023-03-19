---
title: "Guidelines"
description: "An introduction into guidelines for contributing to the Authelia project."
lead: "An introduction into guidelines for contributing to the Authelia project."
date: 2022-10-02T14:32:16+11:00
draft: false
images: []
menu:
  contributing:
    parent: "guidelines"
weight: 310
toc: true
---

The guidelines section contains various guidelines for contributing to Authelia. We implement various guidelines via
automatic processes that will provide feedback in the PR, but this does not cover every situation. You will find both
those which are automated and those which are not in this section.

While it's expected that people aim to follow all of these guidelines we understand that there are logical exceptions to
all guidelines and if it makes sense we're likely to agree with you. So if you find a situation where it doesn't make
sense to follow one just let us know your reasoning when you make a PR if it's not obvious.

## General Guidelines

Some general guidelines include:

- Testing:
  - While we aim for 100% coverage on changes, we do not enforce this where it doesn't make practical sense:
    - A test which just marks a line as tested is not necessarily an effectual test
    - Sometimes there is limited ways in which tests can be performed and the limitation makes the test ineffectual
  - Tests should be named to reflect what they testing for and which part of the code they are testing
  - It's strongly encouraged for bug fixes that contributors create a test that fails prior to fixing the bug and passes
    after fixing the bug and that this test is part of the contribution
  - It's strongly encouraged for features that contributors create have as much testing as is reasonable
- It's recommended people wishing to contribute discuss their intended changes prior to contributing
  - This helps avoid people doubling up on contributions
  - This helps avoid conflicts between contributions
  - This helps avoid contributors wasting their percussion time in a contribution that may not be accepted
