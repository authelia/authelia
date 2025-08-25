---
title: "Automation"
description: "Automated Deployment of Authelia."
summary: "Authelia has several features which make automation easy."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 220
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

1. The [configuration](../../configuration/prologue/introduction.md) can be defined statically by YAML.
2. Most areas of the configuration can be defined by [environment variables](../../configuration/methods/environment.md).

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Ansible

*Authelia* could theoretically be easily deployed via [Ansible] however we do not have an [Ansible Role] at this time.
It would be a desirable [Contribution](../../contributing/development/introduction.md).

[Ansible]: https://www.ansible.com/
[Ansible Role]: https://docs.ansible.com/ansible/latest/user_guide/playbooks_reuse_roles.html
