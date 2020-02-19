---
layout: default
title: Architecture
parent: Home
nav_order: 1
---

# Architecture

**Authelia** is a companion of reverse proxies like Nginx, Traefik and HAProxy.
It can be seen as an extension of those proxies providing authentication functions
and a login portal.

As shown in the following architecture diagram, Authelia is directly connected to
the reverse proxy but never directly connected to application backends.

<p align="center" style="margin:50px">
  <img src="../images/archi.png"/>
</p>

## Workflow

Reverse proxies are configured so that every incoming requests generates an authentication
request sent to Authelia and to which Authelia responds to order the reverse
proxy to let the incoming request pass through or block it because user is not authenticated
or is not sufficiently authorized.

### Step by step

When the first request of an unauthenticated user hits the reverse proxy, Authelia
determines the user is not authenticated because no session cookie has been sent along with
the request. Consequently, Authelia redirects the user to the authentication portal provided
by Authelia itself. The user can then execute the authentication workflow using that portal
to obtain a session cookie valid for all subdomains of the domain protected by Authelia.

When the user visits the initial website again, the query is sent along with the
session cookie which is forwarded in the authentication request to Authelia. This time,
Authelia can verify the user is authenticated and order the reverse proxy to let the query
pass through.

### Sequence Diagram

Here is a description of the complete workflow:

<p align="center">
  <img src="../images/sequence-diagram.png"/>
</p>

