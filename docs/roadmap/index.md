---
layout: default
title: Roadmap
nav_order: 9
has_children: true
---

The Authelia team consists of 3 globally distributed developers working actively on improving Authelia in our spare time
and we define our priorities based on a roadmap that we share here for transparency. We also try to balance features and 
improvements as much as possible with the maintenance tasks we have to perform to keep the backlog of open issues in a
reasonable state. If you're willing to contribute and help us move forward faster, get in touch with us on Matrix. We'll
be glad to share ideas and plans with you.

Below are the prioritised roadmap items:

1. Webauthn needs to be implemented because U2F is being deprecated in the browsers. Chrome displays an annoying popup
advertising the deprecation. This is being implemented [here](https://github.com/authelia/authelia/pull/2707).

2. [Authelia acts as an OpenID Connect Provider](https://github.com/authelia/authelia/issues/189). This is a high
priority because currently the only way to pass authentication information back to the protected app is through the
use of HTTP headers as described
[here](https://www.authelia.com/docs/deployment/supported-proxies/#how-can-the-backend-be-aware-of-the-authenticated-users)
however, many apps either do not support this method or are starting to move away from this in favour of OpenID Connect or OAuth2
internally or via plugins. **[In Preview](./oidc.md)** *this roadmap item is in preview status since information is not 
yet persisted in the database. More information can be found [here](./oidc.md) in the docs*.

3. [Multilingual full support](https://github.com/authelia/authelia/issues/625). Support as been added but we heed to study multiple providers like Crowdin or Weblate
to help us translate in more languages and make Authelia available to even more people around the world! 

4. [Protection of multiple root domains](https://github.com/authelia/authelia/issues/1198). This request has been upvoted many times and we heard you!
Currently, an Authelia setup is only able to protect all subdomains of a given root domain. This situation is challenging for
administrators maintaining services across multiple root domains so we have decided to prioritize this to enable those deployments.

5. [User/Administrator interface](https://github.com/authelia/authelia/issues/303). Many use cases raised on Github relates to
being able to audit, configure and administrate a given account on Authelia. For instance, a user should be able to reset the password
manage MFA hardware devices and personal security policies, etc... An administrator should be able to unban accounts after a regulation ban,
kill sessions to reduce security risk due to compromised accounts and many other things. This item will be decomposed into multiple
items for implementing the features but there is preparatory work to be done on the permissions (likely role-based) we want to
implement.

6. [Facilitate setup on Kubernetes](https://github.com/authelia/authelia/issues/575). There are mainly two objectives
here. First, we need to provide the documentation required to setup Authelia on Kubernetes. Even though, some users
already have it working and the feature is even tested in the project, there is a clear lack of documentation. The
second item is to provide a Helm chart to streamline the setup on Kubernetes.
