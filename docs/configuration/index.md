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