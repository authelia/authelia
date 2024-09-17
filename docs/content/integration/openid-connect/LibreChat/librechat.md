---
title: "LibreChat"
description: "Integrating LibreChat with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-09-17T09:54:41+10:00
draft: true
images: []
weight: 620
toc: true
support:
  level: community
  versions: true
  integration: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
    * [v4.38.8](https://github.com/authelia/authelia/releases/tag/v4.38.8)
* [LibreChat]
    * [v0.7.5](https://www.librechat.ai/changelog)

# Authelia

- Generate a client secret using:
  ```
  docker run authelia/authelia:latest authelia crypto hash generate pbkdf2 --variant sha512 --random --random.length 72 --random.charset rfc3986
  ```
- Then in your `configuration.yml` add the following in the oidc section:
  ```bash filename="configuration.yml"
    - id: librechat
      description: LibreChat
      secret: '$pbkdf2-GENERATED_SECRET_KEY_HERE'
      public: false
      authorization_policy: two_factor
      redirect_uris:
        - 'https://LIBRECHAT.URL/oauth/openid/callback'
      scopes:
        - openid
        - profile
        - email
      userinfo_signing_algorithm: none
  ```
- Then restart Authelia

# LibreChat

- Open the `.env` file in your project folder and add the following variables:
  ```bash filename=".env"
  ALLOW_SOCIAL_LOGIN=true
  OPENID_BUTTON_LABEL='Log in with Authelia'
  OPENID_ISSUER=https://auth.example.com
  OPENID_CLIENT_ID=librechat
  OPENID_CLIENT_SECRET=ACTUAL_GENERATED_SECRET_HERE
  OPENID_SESSION_SECRET=ANY_RANDOM_STRING
  OPENID_CALLBACK_URL=https://auth.example.com/api/oidc/authorization
  OPENID_SCOPE="openid profile email"
  ```
