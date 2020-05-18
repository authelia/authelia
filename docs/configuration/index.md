---
layout: default
title: Configuration
nav_order: 4
has_children: true
---

# Configuration

Authelia uses a YAML file as configuration file. A template with all possible
options can be found [here](https://github.com/authelia/authelia/blob/master/config.template.yml), at the root of the repository.

When running **Authelia**, you can specify your configuration by passing
the file path as shown below.

    $ authelia --config config.custom.yml
 
 
## Validation

Authelia validates the configuration when it starts. This process checks multiple factors including configuration keys
that don't exist, configuration keys that have changed, the values of the keys are valid, and that a configuration
key isn't supplied at the same time as a secret for the same configuration option.

You may also optionally validate your configuration against this validation process manually by using the validate-config
option with the Authelia binary as shown below. Keep in mind if you're using [secrets](./secrets.md) you will have to 
manually provide these if you don't want to get certain validation errors (specifically requesting you provide one of 
the secret values). You can choose to ignore them if you know what you're doing. This command is useful prior to 
upgrading to prevent configuration changes from impacting downtime in an upgrade. This process does not validate
integrations, it only checks that your configuration syntax is valid.

    $ authelia validate-config configuration.yml
    
   
## Duration Notation Format

We have implemented a string based notation for configuration options that take a duration. This section describes its
usage. You can use this implementation in: session for expiration, inactivity, and remember_me_duration; and regulation 
for ban_time, and find_time. This notation also supports just providing the number of seconds instead.
 
The notation is comprised of a number which must be positive and not have leading zeros, followed by a letter
denoting the unit of time measurement. The table below describes the units of time and the associated letter.

|Unit   |Associated Letter|
|:-----:|:---------------:|
|Years  |y                |
|Months |M                |
|Weeks  |w                |
|Days   |d                |
|Hours  |h                |
|Minutes|m                |
|Seconds|s                |

Examples:
* 1 hour and 30 minutes: 90m
* 1 day: 1d
* 10 hours: 10h