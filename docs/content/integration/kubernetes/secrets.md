---
title: "Secrets"
description: "A guide to using secrets when integrating Authelia with Kubernetes."
summary: "A guide to using secrets when integrating Authelia with Kubernetes."
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 530
toc: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

The following serve as examples of how to inject secrets into the Authelia container on [Kubernetes].

## Get started

It's __*strongly recommended*__ that users setting up *Authelia* for the first time take a look at our
[Get started](../prologue/get-started.md) guide. This takes you through various steps which are essential to
bootstrapping *Authelia*.

## Creation

The following section covers creating example secrets. See [Secret Usage](#usage) for usage details. These examples are
not intended to be used as is, you should only include secrets that you're actively using and some secrets may be
missing from these examples. You need to see the [secrets documentation](../../configuration/methods/secrets.md) and
appropriately adapt these examples to your use case.

### Helm Chart

The Helm [Chart](chart.md) automatically generates and injects secrets into an Authelia deployment.

### Manifest

The following manifest is an example which all of the other examples attempt to facilitate as closely as possible. You
can manually create a secret like this with `kubectl apply -f`.

##### String Data Example

##### secret.yml

```yaml {title="secret.yml"}
---
kind: Secret
apiVersion: v1
metadata:
  name: authelia
stringData:
  JWT_SECRET: >-
    NwsVsXv4YCAF9suxWZmT7N6PSzmouCDHqVpzbS5niBKo49b7rTREmwFe6roKswf4
  SESSION_SECRET: >-
    DkezH5zcMQsvaU38YVu673i6JDH4VPiik9xPmYsTN3KPNkxSiiyZ8ASFTdcBcu8q
  REDIS_PASSWORD: >-
    VfhdNhgFG5mLU9s3cjQn9im6dkiWNu3FEUPJRi9bqGm3UV6xzGBZgvdCJhoy26d9
  REDIS_SENTINEL_PASSWORD: >-
    sSJMfX9A6Q6vTpD6rHXcLn2j5kN557RwuohAeyZuGqH9P9LGfuSMnzi9woYZuNqU
  LDAP_PASSWORD: >-
    zafcAShEBfgc48DihdRnnb6UJEGKqzg3FdeZXZ3rhrg6tu2oDoYSBA88w9NPvDhZ
  STORAGE_PASSWORD: >-
    NMHf9Z7C5UQYuKKgh9BJTKeccoZt6c647FQqsEHhkapkkndPkPw3d8bnvkqLgiZ5
  STORAGE_ENCRYPTION_KEY: >-
    rH87rjVMQBvzVgj8vVGSxhop2PPwddrJ7B6oSkGcmoganMf4wqANp9AJwaMHt8RA
  SMTP_PASSWORD: >-
    oi4Yag5HX8Bhc5JTr49nRkdPEr4JcPMfLAPvXxNpHtHqiHXfx3isdWXuTg7yCtjk
  DUO_SECRET_KEY: >-
    d4ypk2UQXxuo86s7vJ2rYWPa5KoxDfU9JQWgEqtANiBaJVQSG8PJbD9U24eiVuPC
  OIDC_HMAC_SECRET: >-
    eSopMjbiuCMhEbXGFsm5B8KWKszxV3CJWSLYrWnBJja4rFNvDxti388WyBjdrsHb
  OIDC_ISSUER_PRIVATE_KEY: |
    -----BEGIN PRIVATE KEY-----
    ...
    -----END PRIVATE KEY-----
...
```

##### Base64 Data Example

This is the same manifest as above but encoded in base64.

```yaml {title="secret.yml"}
---
kind: Secret
apiVersion: v1
type: Opaque
metadata:
  name: authelia
data:
  DUO_SECRET_KEY: ZDR5cGsyVVFYeHVvODZzN3ZKMnJZV1BhNUtveERmVTlKUVdnRXF0QU5pQmFKVlFTRzhQSmJEOVUyNGVpVnVQQw==
  JWT_SECRET: TndzVnNYdjRZQ0FGOXN1eFdabVQ3TjZQU3ptb3VDREhxVnB6YlM1bmlCS280OWI3clRSRW13RmU2cm9Lc3dmNA==
  LDAP_PASSWORD: emFmY0FTaEVCZmdjNDhEaWhkUm5uYjZVSkVHS3F6ZzNGZGVaWFozcmhyZzZ0dTJvRG9ZU0JBODh3OU5QdkRoWg==
  OIDC_HMAC_SECRET: ZVNvcE1qYml1Q01oRWJYR0ZzbTVCOEtXS3N6eFYzQ0pXU0xZclduQkpqYTRyRk52RHh0aTM4OFd5QmpkcnNIYg==
  OIDC_ISSUER_PRIVATE_KEY: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLSBNWElFb2dJQiRBS0NBUUVBeFpWSlAzV0YvL1BHMmZMUW9FQzlEdGRpRkcvKzAwdnFsYlZ6ejQ3bnl4S09OSVBJIGxtTDNVZG1xcEdUS01lLzVCcnFzZTRaQUtsUUhpRGJ3eks5eXBuZmlndEh1dmgvSk8wUzdDaFA3MFJDNjdlZDEgSFYxbnlmejVlVzNsbGJ0R0pQcmxZTHFJVE5nY3RIcDZ6bVJVRnRTelBqOXFGdm96STkzTEppNDkyeUwxK3Z1OCBVbjNEbTgrUXE2WE0ydFBkRWNsZEIvZHRCd09Xb0YrOGVPT1ZzdTBURHVCNWJ3bGhCVkdKdVNBdXpCUFJTMmJGIEdhNHVrMEpEZGtET01DRVF4QzV1V0RGeGdmRVJTTUZ5ZkxWV0Q0N3dvRGJ1V0VCcTEwYzB6K2RwV1BNcDdBaW4gWW5ua3FpY3dDTjg4WjB6aWQ2TW1NUTY1RjQrOUhjK3FDL3A2eHdJREFRQUJBb0lCQUdsaGFBSEtvcitTdTNvLyBBWHFYVEw1L3JiWU16YkxRaUx0MFhlSlQ2OWpwZXFNVHJvWlhIbVd2WEUzMTI4bXFuZjB5encvSzJLbzZ5eEdoIGkrai9vbnlhOEZxcHNWWUNDZ2ZzYm4yL2pzMUF5UkplSXA2WTFPUnNZbnFiWEpueG1rWGE4MEFWL09CUFcyLysgNjBUdFNkUXJlYlkzaUZQYytpMmsrOWJQVHZweXlETEtsejhVd2RaRytrNXV5WU5JeVFUY2N6K1Bqd3NJdkRpaiA3dEtZYW1oaExOM1FYdDMvYVpURnBqVGdlelA0V3lyaVp4aldyZGRIb3djNDdxMnJ3TlM5NU5EMzlKY3lzSkFjIDBQY2J1OEE1bFZhN0Z4MzN1T3R6RGZLV0lXN3hWRU4rT3RQZ04rRmJUalhjWGs1SVplZGwrcFc1bFU1UCsrRy8gWlB2eitXRUNnWUVBOWc2SHdkT0RXM2U2OGJPcXNGb0tnMzUrdmZVRk16bHlNRjhIRnlsTlZmbkxwVEVEcjYzNyBvd3pNRnZjVXhWZDcxYitnVjVubm5iSStyaVVGSWd5Ujh2aENqaHk0bW9vcERQYWhDNC9Ld040Tkc2dXoraTFoIEFCNkQ1K3puMkJqbk8vNXhNTUZHbEFwV3RSTm1KVkdZbE5EajNiWEtoMlZYenp5MDNWTmVEOGtDZ1lFQXpaRkwgT2x6b1JCMUhLcFRXSUVDY3V2eG9mTXhMT0xiM3pzMGsydC9GWU5ZSXBvdm1HV0NDQVVMejEzeTUzZTUrLys1bSA3STlWVVpKRmFJaGFaMzZxVkJBcENLZHJ1NjlwWk1rV0NjUU85akVMRmN4NTFFejdPZ0pXenU3R1MxUUpDUEtDIGZFRHhJMHJaSzIxajkzL1NsL25VbkVpcjdDWXBRK3d2Q2FHdUhnOENnWUFYZ2JuY2ZZMStEb2t3a0I2TmJIeTIgcFQ0TWZiejZjTkdFNTM4dzZrUTJJNEFlRHZtd0xlbnRZTXFhb3c0NzhDaW5lZ0FpZmxTUFR6a0h3QWVtZ2hiciBaR1pQVjFVWGhuMTNmSlJVRzIrZVQxaG5QVmNiWG54MjIzTjBrOEJ1ZDZxWG82NUNueVJUL2t6Y1RiY2pkNUVoIEhuZTJkYWljbU1UenluUG85UTcyYVFLQmdCbW9iTzlYOFZXdklkYmF4Tzg1b1ZabGN0VkEycEsxbzdDWVFtVmYgVU0rSlo0TUNLekkzcllKaXpQUzBpSzUrdWpOUG1tRWtjczIvcUJJb0VzQ2dPcnBMV2hQT2NjLzNVUHhYYlB6RCBEK3NDckJPSWRoeGRqMjNxSk5PblVmRE5DR09wZ1VmcEF6QVlnNHE4R0tJbnZpMWg3WHVrUm5FdlFpOU1KNExZIFAxZFpBb0dBU0djR25UTWttZVNYUDh1eCtkdlFKQWlKc2tuL3NKSWdCWjV1cTVHUkNlTEJVb3NSU1Z4TTc1VUsgdkFoL2MvUkJqK3BZWFZLdVB1SEdaQ1FKeHNkY1JYelhOR291VXRnYmFZTUw1TWUvSGFndDIwUXpEUkJmdUdCZyBxZVpCSmFYaGpFbHZ3NlBVV3RnNHgrTFlSQ0JwcS9iUzNMSzNvelpyU1R1a1ZrS0RlZ3c9IC0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0t
  REDIS_PASSWORD: VmZoZE5oZ0ZHNW1MVTlzM2NqUW45aW02ZGtpV051M0ZFVVBKUmk5YnFHbTNVVjZ4ekdCWmd2ZENKaG95MjZkOQ==
  REDIS_SENTINEL_PASSWORD: c1NKTWZYOUE2UTZ2VHBENnJIWGNMbjJqNWtONTU3Und1b2hBZXladUdxSDlQOUxHZnVTTW56aTl3b1ladU5xVQ==
  SESSION_SECRET: RGtlekg1emNNUXN2YVUzOFlWdTY3M2k2SkRINFZQaWlrOXhQbVlzVE4zS1BOa3hTaWl5WjhBU0ZUZGNCY3U4cQ==
  SMTP_PASSWORD: b2k0WWFnNUhYOEJoYzVKVHI0OW5Sa2RQRXI0SmNQTWZMQVB2WHhOcEh0SHFpSFhmeDNpc2RXWHVUZzd5Q3Rqaw==
  STORAGE_ENCRYPTION_KEY: ckg4N3JqVk1RQnZ6VmdqOHZWR1N4aG9wMlBQd2Rkcko3QjZvU2tHY21vZ2FuTWY0d3FBTnA5QUp3YU1IdDhSQQ==
  STORAGE_PASSWORD: Tk1IZjlaN0M1VVFZdUtLZ2g5QkpUS2VjY29adDZjNjQ3RlFxc0VIaGthcGtrbmRQa1B3M2Q4Ym52a3FMZ2laNQ==
...
```

### Kustomize

The following example is a [Kustomize](https://kustomize.io/) example which can be utilized with `kubectl apply -k`. The
files listed in the `secretGenerator` section  of the `kustomization.yaml` must exist and contain the contents of your
desired secret value.

```yaml {title="kustomization.yaml"}
---
generatorOptions:
  disableNameSuffixHash: true
  labels:
    type: 'generated'
    app: 'authelia'
secretGenerator:
  - name: 'authelia'
    files:
      - 'DUO_SECRET_KEY'
      - 'JWT_SECRET'
      - 'LDAP_PASSWORD'
      - 'OIDC_HMAC_SECRET'
      - 'OIDC_ISSUER_PRIVATE_KEY'
      - 'REDIS_PASSWORD'
      - 'REDIS_SENTINEL_PASSWORD'
      - 'SESSION_SECRET'
      - 'SMTP_PASSWORD'
      - 'STORAGE_ENCRYPTION_KEY'
      - 'STORAGE_PASSWORD'
...
```

## Usage

The following section covers using the created example secrets. See [Creation](#creation) for creation
details.

The example is an excerpt for a manifest which can mount volumes. Examples of these are the [Pod], [Deployment],
[StatefulSet], and [DaemonSet].

```yaml {title="deployment.yml"}
---
spec:
  containers:
    - name: 'authelia'
      env:
        - name: 'AUTHELIA_DUO_API_SECRET_KEY_FILE'
          value: '/app/secrets/DUO_SECRET_KEY'
        - name: 'AUTHELIA_JWT_SECRET_FILE'
          value: '/app/secrets/JWT_SECRET'
        - name: 'AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE'
          value: '/app/secrets/LDAP_PASSWORD'
        - name: 'AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE'
          value: '/app/secrets/OIDC_HMAC_SECRET'
        - name: 'AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE'
          value: '/app/secrets/OIDC_ISSUER_PRIVATE_KEY'
        - name: 'AUTHELIA_SESSION_REDIS_PASSWORD_FILE'
          value: '/app/secrets/REDIS_PASSWORD'
        - name: 'AUTHELIA_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE'
          value: '/app/secrets/REDIS_SENTINEL_PASSWORD'
        - name: 'AUTHELIA_SESSION_SECRET_FILE'
          value: '/app/secrets/SESSION_SECRET'
        - name: 'AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE'
          value: '/app/secrets/SMTP_PASSWORD'
        - name: 'AUTHELIA_STORAGE_ENCRYPTION_KEY_FILE'
          value: '/app/secrets/STORAGE_ENCRYPTION_KEY'
        - name: 'AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE'
          value: '/app/secrets/STORAGE_ENCRYPTION_KEY'
      volumeMounts:
        - mountPath: '/app/secrets'
          name: 'secrets'
          readOnly: true
  volumes:
    - name: 'secrets'
      secret:
        secretName: 'authelia'
        items:
          - key: 'DUO_SECRET_KEY'
            path: 'DUO_SECRET_KEY'
          - key: 'JWT_SECRET'
            path: 'JWT_SECRET'
          - key: 'OIDC_HMAC_SECRET'
            path: 'OIDC_HMAC_SECRET'
          - key: 'OIDC_ISSUER_PRIVATE_KEY'
            path: 'OIDC_ISSUER_PRIVATE_KEY'
          - key: 'REDIS_PASSWORD'
            path: 'REDIS_PASSWORD'
          - key: 'REDIS_SENTINEL_PASSWORD'
            path: 'REDIS_SENTINEL_PASSWORD'
          - key: 'SESSION_SECRET'
            path: 'SESSION_SECRET'
          - key: 'SMTP_PASSWORD'
            path: 'SMTP_PASSWORD'
          - key: 'STORAGE_ENCRYPTION_KEY'
            path: 'STORAGE_ENCRYPTION_KEY'
          - key: 'STORAGE_PASSWORD'
            path: 'STORAGE_PASSWORD'
...
```

[Kubernetes]: https://kubernetes.io/
[Pod]: https://kubernetes.io/docs/concepts/workloads/pods/
[DaemonSet]: https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/
[StatefulSet]: https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/
[Deployment]: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
