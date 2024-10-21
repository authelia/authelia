---
title: "WebAuthn"
description: "A reference guide on various WebAuthn features and topics"
summary: "This section contains reference documentation for Authelia's WebAuthn implementation and capabilities."
date: 2024-10-06T10:34:59+11:00
draft: false
images: []
weight: 220
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Recommended Configurations

When we introduced WebAuthn the configuration was incredibly basic. As time has progressed we've added a lot of
security and trust focused options so we can leverage the technology available more. This section contains various
recommended configurations.

### Passkeys

The following is a configuration that's relatively compliant with the NIST

```yaml
webauthn:
  enable_passkey_login: true
  attestation_conveyance_preference: 'direct'
  filtering:
    prohibit_backup_eligible: true
  metadata:
    enabled: true
    validate_trust_anchor: true
    validate_entry: true
    validate_status: true
    validate_entry_permit_zero_aaguid: false
```

## Metadata Status

Some areas of the configuration allow filtering devices based on the metadata status. This serves as a list of these
status values.

|             Value              |                                                                                                                                                                                                                Description                                                                                                                                                                                                                |
|:------------------------------:|:-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|      `NOT_FIDO_CERTIFIED`      |                                                                                                                                                                                                 This authenticator is not FIDO certified.                                                                                                                                                                                                 |
|        `FIDO_CERTIFIED`        |                                                                                                                                              This authenticator has passed FIDO functional certification. This certification scheme is phased out and will be replaced by FIDO_CERTIFIED_L1.                                                                                                                                              |
|      `FIDO_CERTIFIED_L1`       |                                                                                                                                                   The authenticator has passed FIDO Authenticator certification at level 1. This level is the more strict successor of FIDO_CERTIFIED.                                                                                                                                                    |
|    `FIDO_CERTIFIED_L1plus`     |                                                                                                                                                              The authenticator has passed FIDO Authenticator certification at level 1+. This level is the more than level 1.                                                                                                                                                              |
|      `FIDO_CERTIFIED_L2`       |                                                                                                                                                            The authenticator has passed FIDO Authenticator certification at level 2. This level is more strict than level 1+.                                                                                                                                                             |
|    `FIDO_CERTIFIED_L2plus`     |                                                                                                                                                            The authenticator has passed FIDO Authenticator certification at level 2+. This level is more strict than level 2.                                                                                                                                                             |
|      `FIDO_CERTIFIED_L3`       |                                                                                                                                                            The authenticator has passed FIDO Authenticator certification at level 3. This level is more strict than level 2+.                                                                                                                                                             |
|    `FIDO_CERTIFIED_L3plus`     |                                                                                                                                                            The authenticator has passed FIDO Authenticator certification at level 3+. This level is more strict than level 3.                                                                                                                                                             |
|   `USER_VERIFICATION_BYPASS`   |                                                                                                                  Security: Indicates that malware is able to bypass the user verification. This means that the authenticator could be used without the user’s consent and potentially even without the user’s knowledge.                                                                                                                  |
|  `ATTESTATION_KEY_COMPROMISE`  | Security: Indicates that an attestation key for this authenticator is known to be compromised. The relying party SHOULD check the certificate field and use it to identify the compromised authenticator batch. If the certificate field is not set, the relying party should reject all new registrations of the compromised authenticator. The Authenticator manufacturer should set the date to the date when compromise has occurred. |
|  `USER_KEY_REMOTE_COMPROMISE`  |                                                                 Security: This authenticator has identified weaknesses that allow registered keys to be compromised and should not be trusted. This would include both, e.g. weak entropy that causes predictable keys to be generated or side channels that allow keys or signatures to be forged, guessed or extracted.                                                                 |
| `USER_KEY_PHYSICAL_COMPROMISE` |                                                                 Security: This authenticator has identified weaknesses that allow registered keys to be compromised and should not be trusted. This would include both, e.g. weak entropy that causes predictable keys to be generated or side channels that allow keys or signatures to be forged, guessed or extracted.                                                                 |
|       `UPDATE_AVAILABLE`       |                                                                                                                                                                                        A software or firmware update is available for the device.                                                                                                                                                                                         |
|           `REVOKED`            |                                                                                 The FIDO Alliance has determined that this authenticator should not be trusted for any reason. For example if it is known to be a fraudulent product or contain a deliberate backdoor. Relying parties SHOULD reject any future registration of this authenticator model.                                                                                 |
|   `SELF_ASSERTION_SUBMITTED`   |                                                                                                                     The authenticator vendor has completed and submitted the self-certification checklist to the FIDO Alliance. If this completed checklist is publicly available, the URL will be specified in url.                                                                                                                      |
