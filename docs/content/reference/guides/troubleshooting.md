---
title: "Troubleshooting"
description: "This guide describes and helps users provide information during troubleshooting including sanitization etc"
summary: "This guide describes and helps users provide troubleshooting information."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 220
toc: true
aliases:
  - /r/sanitize
  - /r/troubleshoot
  - /r/troubleshooting
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Frequently Asked Questions

See the [Frequently Asked Questions](frequently-asked-questions.md) reference guide for links to frequently asked
question documentation.

## Logs

It's really important when troubleshooting and even more important when reporting a bug that users provide complete
log files. In addition the [log level](../../configuration/miscellaneous/logging.md#level) should always be set to
`debug` at minimum, if not `trace`.

Complete logs means at minimum from the [log severity line](#log-severity-line) until the actual issue occurs. Though
more than this may be included at the users discretion.

### Output

Logging can be [configured](../../configuration/miscellaneous/logging.md) to output to both a file and stdout / console
/ docker logs.

- If you configure the [file_path] option with the [keep_stdout] configuration option enabled then you will only be
  required to supply the stdout / console / docker logs and should ignore the file logs.
- If you configure the [file_path] option without the [keep_stdout] configuration option enabled then you will be
  required to join the stdout / console / docker logs with the file logs.

[keep_stdout]: ../../configuration/miscellaneous/logging.md#keep_stdout
[file_path]: ../../configuration/miscellaneous/logging.md#file_path

### Large Amount of Logs

In instances where the logs are substantial we will accept truncated logs. There are however some very intentional rules
about truncation of logs.

1. You must show more than 1 minute of logs prior to the issue occurring (i.e. show every log line that is within 1
   minute of the log line where the issue occurred plus one additional line outside the minute).
2. The level rule still applies.
3. You must show:
   1. The [log severity line](#log-severity-line).
   2. The [listening line](#listening-line) and the 10 log lines which follow this line.
   3. The lines between these two lines, which should also include the [starting line](#starting-line).
4. No warning or error that is potentially related to the issue should be truncated.

### Important Log Lines

The following log lines are important and are referenced by other sections of this document.

#### Log Severity Line

The log severity line will have a message which is similar to the following examples:

- `Log severity set to debug`
- `Log severity set to trace`

#### Starting Line

The starting line will have a message which is similar to the following examples:

- `Authelia v4.37.5 is starting`
- `Authelia v4.38.0 is starting`
- `Authelia untagged-v4.38.0 (master, 50d8b4a) is starting`

#### Listening Line

The listening line will have a message which is similar to the following examples:

- `Listening for non-TLS connections on '0.0.0.0:{{< sitevar name="port" nojs="9091" >}}' path '/'`
- `Listening for TLS connections on '0.0.0.0:{{< sitevar name="port" nojs="9091" >}}' path '/'`
- `Listening for non-TLS connections on ':{{< sitevar name="port" nojs="9091" >}}' path '/'`
- `Listening for non-TLS connections on ':{{< sitevar name="port" nojs="9091" >}}' path '/' and '/authelia'`

## Sanitization

Some users may wish to hide their domain in files provided during troubleshooting. While this is discouraged, if a user
decides to perform this action it's critical for these purposes that you hide your domain in a very specific
way. Most editors allow replacing all instances of a value, utilizing this is essential to making troubleshooting
possible.

### General Rules

1. Only replace the purchased portion of domains:
   - For example if you have `auth.abc123.com` and `app.abc123.com` they
   should become `auth.example.com` and `app.example.com`, i.e. replace all instances of `abc123.com` with `example.com`.
2. Make sure value replaced is replaced with a unique value:
   - For example if you replace `abc123.com` with `example.com` DO NOT replace any other value other than `abc123.com` with
   `example.com`. The same rule applies to IP addresses, usernames, and groups.
3. Make sure the value replaced is replaced across logs, configuration, and any references:
   - For example if you replace `abc123.com` with `example.com` in your configuration, make exactly the same replacement
   for the log files.
4. Make sure this consistency is followed for all communication regarding a single issue.

### Multiple Domains

*__Replacement Value:__* `example#.com` (where `#` is a unique number per domain)

In instances where there are multiple domains it's recommended these domains are replaced with `example1.com`,
`example2.com`, etc.
