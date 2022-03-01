---
layout: default
title: OpenID Connect
parent: Roadmap
nav_order: 1
---

We have decided to implement [OpenID Connect] as a beta feature, it's suggested you only utilize it for testing and
providing feedback, and should take caution in relying on it in production as of now. [OpenID Connect] and it's related
endpoints are not enabled by default unless you specifically configure the [OpenID Connect] section.

As [OpenID Connect] is fairly complex (the [OpenID Connect] Provider role especially so) it's intentional that it is
both a beta and that the implemented features are part of a thoughtful roadmap. Items that are not immediately obvious
as required (i.e. bug fixes or spec features), will likely be discussed in team meetings or on GitHub issues before being
added to the list. We want to implement this feature in a very thoughtful way in order to avoid security issues.

The beta will be broken up into stages. Each stage will bring additional features. The following table is a *rough* plan
for which stage will have each feature, and may evolve over time:

<table>
    <thead>
      <tr>
        <th class="tbl-header">Stage</th>
        <th class="tbl-header">Feature Description</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td rowspan="8" class="tbl-header tbl-beta-stage">beta1 (4.29.0)</td>
        <td><a href="https://openid.net/specs/openid-connect-core-1_0.html#Consent" target="_blank" rel="noopener noreferrer">User Consent</a></td>
      </tr>
      <tr>
        <td><a href="https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowSteps" target="_blank" rel="noopener noreferrer">Authorization Code Flow</a></td>
      </tr>
      <tr>
        <td><a href="https://openid.net/specs/openid-connect-discovery-1_0.html" target="_blank" rel="noopener noreferrer">OpenID Connect Discovery</a></td>
      </tr>
      <tr>
        <td>RS256 Signature Strategy</td>
      </tr>
      <tr>
        <td>Per Client Scope/Grant Type/Response Type Restriction</td>
      </tr>
      <tr>
        <td>Per Client Authorization Policy (1FA/2FA)</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Per Client List of Valid Redirection URI's</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage"><a href="https://datatracker.ietf.org/doc/html/rfc6749#section-2.1" target="_blank" rel="noopener noreferrer">Confidential Client Type</a></td>
      </tr>
      <tr>
        <td rowspan="6" class="tbl-header tbl-beta-stage">beta2 (4.30.0)</td>
        <td class="tbl-beta-stage"><a href="https://openid.net/specs/openid-connect-core-1_0.html#UserInfo" target="_blank" rel="noopener noreferrer">Userinfo Endpoint</a> (missed in beta1)</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Parameter Entropy Configuration</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Token/Code Lifespan Configuration</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Client Debug Messages</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Client Audience</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage"><a href="https://datatracker.ietf.org/doc/html/rfc6749#section-2.1" target="_blank" rel="noopener noreferrer">Public Client Type</a></td>
      </tr>
      <tr>
        <td rowspan="1" class="tbl-header tbl-beta-stage">beta3 (4.34.0)</td>
        <td>Proof Key for Code Exchange (PKCE) for Authorization Code Flow</td>
      </tr>
      <tr>
        <td rowspan="2" class="tbl-header tbl-beta-stage">beta4 <sup>1</sup></td>
        <td>Token Storage</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Audit Storage</td>
      </tr>
      <tr>
        <td rowspan="2" class="tbl-header tbl-beta-stage">beta5 <sup>1</sup></td>
        <td class="tbl-beta-stage">Prompt Handling</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Display Handling</td>
      </tr>
      <tr>
        <td rowspan="5" class="tbl-header tbl-beta-stage">beta6 <sup>1</sup></td>
        <td><a href="https://openid.net/specs/openid-connect-backchannel-1_0.html" target="_blank" rel="noopener noreferrer">Back-Channel Logout</a></td>
      </tr>
      <tr>
        <td>Deny Refresh on Session Expiration</td>
      </tr>
      <tr>
        <td><a href="https://openid.net/specs/openid-connect-messages-1_0-20.html#rotate.sig.keys" target="_blank" rel="noopener noreferrer">Signing Key Rotation Policy</a></td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Client Secrets Hashed in Configuration</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">UUID or Random String for <code>sub</code> claim</td>
      </tr>
      <tr>
        <td class="tbl-header tbl-beta-stage">GA <sup>1</sup></td>
        <td class="tbl-beta-stage">General Availability after previous stages are vetted for bug fixes</td>
      </tr>
      <tr>
        <td rowspan="7" class="tbl-header">misc</td>
        <td>List of other features that may be implemented</td>
      </tr>
      <tr>
        <td class="tbl-beta-stage"><a href="https://openid.net/specs/openid-connect-frontchannel-1_0.html" target="_blank" rel="noopener noreferrer">Front-Channel Logout</a> <sup>2</sup></td>
      </tr>
      <tr>
        <td class="tbl-beta-stage"><a href="https://datatracker.ietf.org/doc/html/rfc8414" target="_blank" rel="noopener noreferrer">OAuth 2.0 Authorization Server Metadata</a> <sup>2</sup></td>
      </tr>
      <tr>
        <td class="tbl-beta-stage"><a href="https://openid.net/specs/openid-connect-session-1_0-17.html" target="_blank" rel="noopener noreferrer">OpenID Connect Session Management</a> <sup>2</sup></td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">End-User Scope Grants <sup>2</sup></td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Client RBAC <sup>2</sup></td>
      </tr>
      <tr>
        <td class="tbl-beta-stage">Add <code>preferred_username</code> claim (4.33.2)</td>
      </tr>
    </tbody>
</table>

¹ _This stage has not been implemented as of yet_.

² _This individual feature has not been implemented as of yet_.

[OpenID Connect]: https://openid.net/connect/
