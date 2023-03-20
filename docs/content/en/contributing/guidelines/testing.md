---
title: "Testing"
description: "Authelia Development Testing Guidelines"
lead: "This section covers the testing guidelines."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  contributing:
    parent: "guidelines"
weight: 320
toc: true
---

The following outlines the specific requirements we have for testing the Authelia code contributions.

- While we aim for 100% coverage on changes and additions, we do not enforce this where it doesn't make practical sense:
  - A test which just marks a line as tested is not necessarily an effectual test
  - Sometimes there is limited ways in which tests can be performed and the limitation makes the test ineffectual
- Tests should be named to reflect what they testing for and which part of the code they are testing
- It's required for bug fixes that contributors create a test that fails prior to and passes
  subsequent to fix being applied unless there is a specific reason to exclude this test, and this test is part of the
  contribution
- It's strongly encouraged for features that contributors create have as much testing as is reasonable i.e. any line
  that can be tested should be tested, if the line can't be tested generally this is an indication a refactor may be
  required
