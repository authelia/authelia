---
title: "First Factor"
description: "Authelia utilizes the standard username and password combination for first factor authentication."
summary: "Authelia utilizes the standard username and password combination for first factor authentication."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
weight: 220
toc: true
aliases:
  - /docs/features/first-factor.html
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

{{< picture src="1FA.png" caption="An example of the first factor sign in portal" alt="First Factor Authentication View" width=400 >}}

*__IMPORTANT:__ This is currently the only method available for first factor authentication.*

Authelia supports several kind of user databases:

* An LDAP server like OpenLDAP, OpenAM, Active Directory etc.
* A YAML file
