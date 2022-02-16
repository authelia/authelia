---
layout: default
title: Logout
parent: Features
nav_order: 8
---

# Logout

Authelia is able to log out your users to ensure their account is not exposed anymore when they stop
surfing the web.

When user is logged out, the cookie attached to this user is reset on the backend side. Therefore, any
subsequent request using this old cookie is considered unauthenticated by Authelia. In this case the user
is simply redirected to the login page and has to authenticate again to generate a new session with a new cookie.

## Methods to log out

## Frontend

In most websites and applications, users can usually click on a logout button to be signed out and Authelia
offers the same feature.

Implementing logout is as easy as putting a link or button somewhere on your application or website with
the following href: `https://auth.example.com/logout` where `auth.example.com` is the domain serving Authelia.
By default, this would redirect the user to the login page of Authelia but one can force the redirection to any
domain protected by Authelia by appending the 'rd' query parameter which should be set to the target URL where
the user should be redirected. For instance, `https://auth.example.com/logout?rd=https://homepage.example.com`.

Please note that an attempt of redirection to a domain which is not a subdomain protected by Authelia will be
skipped for security reasons described later in this page.

## Backend

The backend API can also be called directly from your applications if needed. The endpoint is /api/logout which
is taking a POST request with a body like:

    {
        "targetURL": "https://homepage.example.com"
    }

Please note that an attempt of redirection to a domain which is not a subdomain protected by Authelia will be
skipped for security reasons described later in this page.

## Why preventing redirection to some domains?

This is a security feature which is protecting your users against attacks called open redirect. This kind of attack
is described [here](https://cheatsheetseries.owasp.org/cheatsheets/Unvalidated_Redirects_and_Forwards_Cheat_Sheet.html)
by the [OWASP](https://en.wikipedia.org/wiki/OWASP#:~:text=The%20Open%20Web%20Application%20Security,field%20of%20web%20application%20security.&text=It%20is%20led%20by%20a%20non%2Dprofit%20called%20The%20OWASP%20Foundation.).
In a nutshell, hackers can send phishing emails to your users and trick them by making them click on a legit link
eventually redirecting to an infected website.