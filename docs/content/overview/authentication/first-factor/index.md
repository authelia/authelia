---
title: "First Factor"
description: "Authelia utilizes the standard username and password combination for first factor authentication."
summary: "Authelia utilizes the standard username and password combination for first factor authentication."
date: 2024-03-14T06:00:14+11:00
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

{{< picture src="1FA.png" caption="An example of the first factor sign in portal" alt="First Factor Authentication View" process="resize 400x" >}}

*__IMPORTANT:__ This is currently the only method available for first factor authentication.*

Authelia supports several kind of user databases:

* An LDAP server like OpenLDAP, OpenAM, Active Directory etc.
* A YAML file
