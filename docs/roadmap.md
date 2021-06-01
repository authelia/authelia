---
layout: default
title: Roadmap
nav_order: 9
---

# Roadmap

Currently the team consists of 3 globally distributed developers working actively on improving Authelia in our spare time and we define
our priorities based on a roadmap that we share here for transparency. We also try to balance features and improvements as much as possible with
the maintenance tasks we have to perform to keep the backlog of open issues in a reasonable state.
If you're willing to contribute and help us move forward faster, get in touch with us on Matrix. We'll be glad to share
ideas and plans with you.

Below are the prioritised roadmap items:

1. **[In Preview](./configuration/identity-providers/oidc.md)** *this roadmap item is in preview status, more
   information can be found in the docs*. 
   [Authelia acts as an OpenID Connect Provider](https://github.com/authelia/authelia/issues/189). This is a high
priority because currently the only way to pass authentication information back to the protected app is through the
use of HTTP headers as described
[here](https://www.authelia.com/docs/deployment/supported-proxies/#how-can-the-backend-be-aware-of-the-authenticated-users)
however, many apps either do not support this method or are starting to move away from this in favour of OpenID Connect or OAuth2
internally or via plugins.

2. [Administration interface](https://github.com/authelia/authelia/issues/974). This is useful in many cases to
properly manage users and administrate activities like unbanning banned users. In the future we can even think of
adding/removing users from there, request a password reset for a user or all users, request a 2FA enrollment,
temporarily block users, etc...

3. [User interface](https://github.com/authelia/authelia/issues/303). This will help the users manage their 2FA
devices, reset their password, review their authentication activity.
In the future we envisage users will be able to customize their profile with an avatar, set their preferences
regarding 2FA and according to the global policy defined by administrators, etc...

4. [Facilitate setup on Kubernetes](https://github.com/authelia/authelia/issues/575). There are mainly two objectives
here. First, we need to provide the documentation required to setup Authelia on Kubernetes. Even though, some users
already have it working and the feature is even tested in the project, there is a clear lack of documentation. The
second item is to provide a Helm chart to streamline the setup on Kubernetes.
