---
layout: default
title: First Factor
parent: Features
nav_order: 1
---

# First Factor

2-Factor authentication is a method in which a user is granted access by presenting
two pieces of evidence that she is who she claims to be.

**Authelia** requires usual username and password as a first factor.

<p align="center">
  <img src="../images/1FA.png" width="400">
</p>

*IMPORTANT: This is the only method available as first factor.*

Authelia supports several kind of users databases:

* An LDAP server like OpenLDAP or OpenAM.
* An Active Directory.
* A YAML file
