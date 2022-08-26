---
title: "HTTP Archive Files"
description: "This guide describes and helps users create HTTP Archive (HAR) files"
lead: "This guide describes and helps users create HTTP Archive (HAR) files."
date: 2022-06-20T10:05:55+10:00
draft: false
images: []
menu:
  reference:
    parent: "guides"
weight: 220
toc: true
aliases:
  - /r/har
---

## Introduction

The HTTP Archive File Format (HAR) is a common developer import/export format which shows web requests that browsers
make including all headers which includes cookies, forms submitted, etc.

This format allows users to open the developer tools, perform several actions, and then export a file with all of the
requests that still exist in the network tab of the developer tools. This file is stored in JSON which makes it easy to
view what information exists before sharing it.

## Sanitization

*__Important:__ this file may contain sensitive information which should be sanitized manually before sharing it
anywhere with anyone. Sensitive information can vary wildly but some of the key areas that may be sensitive when
exporting this for troubleshooting with Authelia are:*
- `Cookie` request header
- `Set-Cookie` response header
- Data sent to the following endpoints:
  - `/api/firstfactor`: username / password
  - `/api/*/identity/start`: the token query parameter
  - `/api/secondfactor/*`: the post data

__*Important:*__ In addition to above, some users may wish to hide their domain. It's critical for these purposes that
you hide your domain in a very specific way. Supposing you purchased the domain `abc123.com` and are running services on
`auth.abc123.com`, `app.abc123.com`, and so on; you should replace all instances of `abc123.com` with `example.com`.

In instances where there are multiple domains it's recommended these domains are replaced with `example1.com`,
`example2.com`, etc.

## Instructions

The following are instructions on how to perform valuable HAR exports:

1. Open your browser.
2. Open a blank tab.
3. Press Ctrl + Shift + I to open the browser Developer Tools.
4. Open the Network tab.
5. Ensure the browser persists logs:
   1. Firefox:
      1. Select the `Network Settings` cog symbol at the top right of the `Network` tab.
      2. Ensure `Persist Logs` is checked.
   2. Chrome:
      1. Ensure `Preserve logs` in the top left of the `Network` tab is checked.
6. Perform your intended requests, or the requests that have been requested.
7. Export the HAR File:
   1. Firefox:
      1. Select the `Network Settings` cog symbol at the top right of the `Network` tab.
      2. Select `Save All AS HAR`.
   2. Chrome:
      1. Right click any request in the `Network` tab.
      2. Select `Save all as HAR with content` at the bottom of the dialogue.
