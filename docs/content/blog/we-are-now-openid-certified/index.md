---
title: "We are now OpenID Certified™"
summary: "This is a very important and exiting milestone for the Authelia project."
date: 2025-05-18T11:24:22+10:00
draft: false
weight: 50
categories: ["News", "Announcements"]
tags: ["announcements"]
contributors: ["James Elliott"]
aliases:
  - '/blog/important-announcement-the-future-of-authelia/'
pinned: true
homepage: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

Authelia is now [OpenID Certified™] for the Basic OP, Implicit OP, Hybrid OP, Form Post OP, and Config OP profiles of the
[OpenID Connect™ protocol]. This means our OpenID Connect 1.0 Provider implementation has officially passed the
certification process and is verified to conform to the specification in all areas that we've implemented and those
that have conformance testing. Many providers don’t reach this level of validation, so I’m especially proud and excited
about this milestone.

I’d like to sincerely thank the OpenID Foundation and its members for being so helpful and welcoming during the
certification process; and for promptly fixing an issue with the conformance suite when it was reported. I was
completely flawed seeing the time between the issue being reported, a pull request being drafted, the fix being
released, and the new release being published; was no more than 24 hours.

{{< figure src="/images/oid-certification.jpg" class="center" process="resize 300x" >}}

I fully intend to pursue conformance for the remaining outstanding profiles; 3rd Party-Init OP, Dynamic OP,
Session OP, Front-Channel OP, Back-Channel OP, and RP-Initiated OP; as soon as we implement the necessary underlying
features.

# The Future

Certification has long been a goal of the team and myself. There are several areas of OpenID Connect 1.0 that we support
(and all of those are certified) but there are a number we do not yet support.

<a href="https://openid.net/developers/how-connect-works/">
{{< figure src="https://openid.net/wp-content/uploads/2023/06/OpenIDConnect-Map-December2023.png" class="center" process="resize 400x" caption="The OpenID Connect 1.0 Protocol Suite, image is a trademark of the OpenID Foundation, click the image for the source." >}}
</a>

The elements we support are Core, Discovery, and the Form Post Response Mode. The two major remaining elements
Dynamic Client Registration and Session Management are obvious goals. While they're not required they are certainly
useful. We're making steps towards both of these in the next release.

While we haven’t finalized the next steps, I believe the path ahead (especially around SSO) is gaining significant
clarity. That said, everything is still subject to change and discussions with the team. I just wanted to make this
announcement a surprise for them as well.

Here are some key areas of focus (specifically surrounding SSO):

1. Finish the OpenID Connect 1.0 implementation. The certification is great but there are a few things on my mind that
   ideally need to be addressed before we remove the experimental / beta status, most of which are traditionally
   breaking changes:
    - Consent Policies need to be reworked. Specifically we should make them reusable like other policies, and we should
      ensure it clearly represents only the default behavior when the client does not request something that explicitly
      requires some behavior. For example the prompt parameter can require the display of a login, account selection,
      consent, or require nothing is shown visually to the user.
    - We need to implement multi-issuer configuration to compliment the multi-domain configuration. Each domain should
      require an explicit OpenID Connect 1.0 configuration if you want to use it on that domain.
    - Database Storage of Issuers and Clients. This has become an obvious requirement. We don't want to remove the
      option for users to configure this via the config file but there are several features that will rely on it being
      an option.
2. Implement High-Impact OpenID Connect 1.0 Specification Extensions:
   - We plan to implement several impactful extensions; many of which should be straightforward, though we can reassess
     if needed (I don't think we want to delay OpenID Connect 1.0 Relying Party support too long):
       - Dynamic Client Registration (Dynamic OP Profile)
       - Session Management (Session OP Profile)
       - Front-Channel Logout (Front-Channel OP Profile)
       - Back-Channel Logout (Back-Channel OP Profile)
       - RP-Initiated Logout (RP-Initiated OP Profile)
       - Client Initiated Backchannel Authentication Flow (3rd Party-Init OP Profile)
       - OAuth 2.0 Token Exchange
3. In No Particular Order:
   - Fully implement authentication method references:
       - By allowing customized authorization policies using authentication method references we unlock a large future
         potential for Authelia and allow administrators fine-grained control over authorization.
   - WebFinger
   - [Federated Credential Management (FedCM)](https://www.w3.org/TR/fedcm/)
   - Implement the OpenID Connect 1.0 Relying Party role:
      - Allow users to link their social accounts to other OpenID Connect 1.0 Providers, and subsequently sign in with
        them.
      - Allow administrators to configure trusting the authentication method references from these providers allowing
        seamless SSO, and for those that are untrusted only assume the password was provided.
   - [SAML 2.0](https://docs.oasis-open.org/security/saml/Post2.0/sstc-saml-tech-overview-2.0.html):
      - This is a widely requested feature and we will absolutely implement it. We just wanted to ensure we had a strong
        foundation before we do so.

# Specification Support

I have updated the [OpenID Connect 1.0 Integration](../../integration/openid-connect/introduction.md) with a
[Support Chart](../../integration/openid-connect/introduction.md#support-chart) which lists a majority of the OpenID
Connect 1.0 and OAuth 2.0 specifications that are somewhat relevant and are more likely to have a future within
Authelia. This combined with the [Roadmap](../../roadmap/active/openid-connect-1.0-provider.md) serve as documentation
for our future developments within OpenID Connect 1.0.

This should give you a decent comparison for any other project that wishes to be transparent about its support level by
including a similar chart.

# Join the Discussion and Show Your Support

Feel free to discuss this awesome news in our [Discussion Forum](https://github.com/authelia/authelia/discussions/9525),
or in one of our many [Chat Methods](../../information/contact.md#chat).

You can show your support for the Authelia project by giving us a star on [GitHub](https://www.github.com/authelia/authelia).

[OpenID Certified™]: https://openid.net/certification/
[OpenID Connect™ protocol]: https://openid.net/developers/how-connect-works/
