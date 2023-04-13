# Security Policy

## Prologue

The __Authelia__ team takes security very seriously. Because __Authelia__ is intended as a security product a lot of
decisions are made with security being the priority and we always aim to implement security by design.

## Coordinated vulnerability disclosure

__Authelia__ follows the [coordinated vulnerability disclosure] model when dealing with security vulnerabilities. This
was previously known as responsible disclosure. We strongly urge anyone reporting vulnerabilities to __Authelia__ or any
other project to follow this model as it is considered as a best practice by many in the security industry.

If you believe you have identified a security vulnerability or security related bug with __Authelia__ please make every
effort to contact us privately using one of the [contact options](#contact-options) below. Please do not open an issue,
do not notify us in public, and do not disclose this issue to third parties.

Using this process helps ensure that users affected have an avenue to fixing the issue as close to the issue being
made public as possible. This mitigates the increasing the attack surface (via improving attacker knowledge) for
diligent administrators simply via the act of disclosing the security issue.

For more information about [security](https://www.authelia.com/security/) related matters, please read
[the documentation](https://www.authelia.com/security/).

## Contact Options

Several contact options exist however it's important you specifically use a security contact method when reporting a
security vulnerability or security related bug. These methods are clearly documented below.

### GitHub Security

Users can utilize GitHub's security vulnerability system to privately [report a vulnerability]. This is an easy method
for users who have a GitHub account.

### Email

Users can utilize the [security@authelia.com](mailto:security@authelia.com) email address to privately report a
vulnerability. This is an easy method of users who do not have a GitHub account.

This email address is only accessible by members of the [core team] for the purpose of disclosing security
vulnerabilities and issues within the __Authelia__ code base.

### Chat

If you wish to chat directly instead of sending an email please use either [Matrix](README.md#matrix) or
[Discord](README.md#discord) to direct / private message one of the [core team] members.

Please avoid this method unless absolutely necessary. We generally prefer that users use either the
[GitHub Security](#github-security) or [Email](#email) option rather than this option as it both allows multiple team
members to deal with the report and prevents mistakes when contacting a [core team] member.

The [core team] members are identified in [Matrix](README.md#matrix) as room admins, and in [Discord](README.md#discord)
with the `Core Team` role.

## Process

1. The user privately reports a potential vulnerability.
2. The report is acknowledged as received.
3. The report is reviewed to ascertain if additional information is required. If it is required:
   1. The user is informed that the additional information is required.
   2. The user privately adds the additional information.
   3. The process begins at step 3 again, proceeding to step 4 if the additional information provided is sufficient.
4. The vulnerability is reproduced.
5. The vulnerability is patched, and if possible the user reporting the bug is given access to a fixed binary, docker
   image, and git patch.
6. The patch is confirmed to resolve the vulnerability.
7. The fix is released and users are notified that they should update urgently.
8. The [security advisory] is published when (whichever happens sooner):
   - The CVE details are published by [MITRE], [NIST], etc.
   - Roughly 7 days after users have been notified the update is available.

[MITRE]: https://www.mitre.org/
[NIST]: https://www.nist.gov/

## Credit

Users who report bugs will at their discretion (i.e. they do not have to be if they wish to remain anonymous) be
credited for the discovery. Both in the [security advisory] and in our [all contributors](README.md#contribute)
documentation.

[coordinated vulnerability disclosure]: https://en.wikipedia.org/wiki/Coordinated_vulnerability_disclosure
[security advisory]: https://github.com/authelia/authelia/security/advisories
[report a vulnerability]: https://github.com/authelia/authelia/security/advisories/new
[core team]: https://www.authelia.com/information/about/#core-team
