---
title: "HTTP Archive Files"
description: "This guide describes and helps users create HTTP Archive (HAR) files"
summary: "This guide describes and helps users create HTTP Archive (HAR) files."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 220
toc: true
aliases:
  - /r/har
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Introduction

The HTTP Archive File Format (HAR) is a common developer import/export format which shows web requests that browsers
make including all headers which includes cookies, forms submitted, etc.

This format allows users to open the developer tools, perform several actions, and then export a file with all of the
requests that still exist in the network tab of the developer tools. This file is stored in JSON which makes it easy to
view what information exists before sharing it. Subsequently users may import this file on another browser and see all
of these requests which makes it easier to debug certain situations without having to replicate an environment or be
present in an environment.

## Sanitization

The following section outlines some helpful information if you wish to sanitize your HAR file to share it with others.

For generic sanitization information see the [Troubleshooting Sanitization guide](troubleshooting.md#sanitization).

### Security Sensitive Information

*__Important:__ this file may contain sensitive information which should be sanitized manually before sharing it
anywhere with anyone. Sensitive information can vary wildly but some of the key areas that may be sensitive when
exporting this for troubleshooting with Authelia are:*
- `Cookie` request header
- `Set-Cookie` response header
- Data sent to the following endpoints:
  - `/api/firstfactor`: username / password
  - `/api/*/identity/start`: the token query parameter
  - `/api/secondfactor/*`: the post data

## Instructions

The following are instructions on how to perform valuable HAR exports. The instructions for Chrome / Chromium should be
applicable in all Chromium based browsers, and likewise for Firefox based browsers.

1. Open your browser.
2. Open a blank tab.
3. Press Ctrl + Shift + I to open the browser Developer Tools.
4. Open the `Network` tab.
5. Ensure the browser persists logs:
   1. Firefox:
      1. Select the `Network Settings` cog symbol at the top right of the `Network` tab.
      2. Ensure `Persist Logs` is checked.
   2. Chrome / Chromium:
      1. Ensure `Preserve logs` in the top left of the `Network` tab is checked.
6. Perform your intended requests, or the requests that have been requested.
7. Export the HAR File:
   1. Firefox:
      1. Select the `Network Settings` cog symbol at the top right of the `Network` tab.
      2. Select `Save All AS HAR`.
   2. Chrome / Chromium:
      1. Right click any request in the `Network` tab.
      2. Select `Save all as HAR with content` at the bottom of the dialogue.
