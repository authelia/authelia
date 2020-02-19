---
layout: default
title: Configuration
nav_order: 4
has_children: true
---

# Configuration

Authelia uses a YAML file as configuration file. A template with all possible
options can be found [here](../config.template.yml), at the root of the repository.

When running **Authelia**, you can specify your configuration by passing
the file path as shown below.

    $ authelia --config config.custom.yml
