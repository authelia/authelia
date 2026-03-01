---
title: "LDAP Integrations"
description: "A LDAP integration reference guide"
summary: "This section contains a LDAP integration reference guide for Authelia."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 320
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

This page details the elements required for LDAP integrations to be considered supported.

## Anonymous Discovery via the RootDSE Search

LDAP servers must permit the query of the RootDSE object by Authelia prior to authentication. This is required to ensure
that Authelia has every opportunity to automatically step-up the security of the authenticated bind. In addition it's
important to ensure that the LDAP server supports LDAPv3.

The two most critical elements for pre-bind discovery is the ability to perform StarTLS and the supported SASL
mechanisms which both assist in preventing the password from being intercepted by a malicious actor.

The LDAP server can either be setup to allow all users to anonymously perform the RootDSE search, or as per our current
recommendation if the LDAP server supports it they can just allow the Authelia IP to anonymously perform the RootDSE
search. This recommendation is purely precautionary based on certain LDAP servers discouraging all anonymous searches
for unqualified security reasons as it pertains to the RootDSE which is specifically intened for the reason Authelia
uses it; discovery of how to interact with an LDAP server.

It's speculation however it's likely that this recommendation is normally made without consideration for this specific
legitimate and intended use case. Ultimately the LDAP server is responsible for ensuring that the RootDSE search is
available anonymously and that it does not include any overly specific version information even to authenticated users.

### Attributes

The attributes that are requested during the RootDSE search are:

|         Attribute         | Importance |                                              Notes                                              |
|:-------------------------:|:----------:|:-----------------------------------------------------------------------------------------------:|
|       `objectClass`       |    Low     |                           Used to determine if the Vendor is OpenLDAP                           |
|  `supportedLDAPVersion`   |    High    |                        Used to check the LDAP protocol version supported                        |
|    `supportedControl`     |    High    |                      Used to convey the controls the LDAP server supports                       |
|   `supportedExtension`    |  Security  |       Used to convey the extensions the LDAP server supports including TLS and PwdModify        |
| `supportedSASLMechanisms` |  Security  |       Used to convey the SASL bind mechanisms the server supports instead of Simple Binds       |
|       `vendorName`        |    Low     |             Used to convey the Name of the Vendor who made the LDAP server software             |
|      `vendorVersion`      |    Low     | Used to convey inspecific Version information from the Vendor who made the LDAP server software |
|   `domainFunctionality`   |    Low     |             Used by Microsoft Corporation to convey the domain functionality level              |
|   `forestFunctionality`   |    Low     |             Used by Microsoft Corporation to convey the forest funcationaliy level              |

## LDAP Version

Authelia only supports LDAPv3, all other versions including when the server does not return a version, are
**_not supported_**.
