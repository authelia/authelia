---
layout: default
title: Community-Tested OIDC Integrations
parent: Community
nav_order: 5
has_children: true
has_toc: false
---

# OIDC Integrations

**Note** This is community-based content for which the core-maintainers cannot guarantee correctness. The parameters may change over time. If a parameter does not work as documented, please submit a PR to update the list.

## Currently Tested Applications

|   Application    |        Minimal Version         |                                                    Notes                                                    |
|:----------------:|:------------------------------:|:-----------------------------------------------------------------------------------------------------------:|
|      Gitea       |            `1.14.6`            |                                                                                                             |
|      GitLab      |            `13.0.0`            |                                                                                                             |
|     Grafana      |            `8.0.5`             |                                                                                                             |
| Hashicorp Vault  |            `1.8.1`             |                                                                                                             |
|      MinIO       | `RELEASE.2021-11-09T03-21-45Z` | must set `MINIO_IDENTITY_OPENID_CLAIM_NAME: groups` in MinIO and set [MinIO policies] as groups in Authelia |
|    Nextcloud     |            `22.1.0`            |   Tested using the `nextcloud-oidc-login` app - [Link](https://github.com/pulsejet/nextcloud-oidc-login)    |
|      Wekan       |             `5.41`             |                                                                                                             |
|   Portainer CE   |            `2.6.1`             |   Settings to use username as ID: set `Scopes` to `openid` and `User Identifier` to `preferred_username`    |
| Bookstack        | `21.10`                        |                                                                                                             |
| Harbor        |                `1.10`             |   It works on >v2.1 also, but not sure if there is OIDC support on v2.0|
| Verdaccio        |              `5`               |   Depends on this fork of verdaccio-github-oauth-ui: [Link](https://github.com/OnekO/verdaccio-github-oauth-ui)
|
[MinIO policies]: https://docs.min.io/minio/baremetal/security/minio-identity-management/policy-based-access-control.html#minio-policy

## Known Callback URLs

If you do not find the application in the list below, you will need to search for yourself - and maybe come back to open a PR to add your application to this list so others won't have to search for them.

`<DOMAIN>` needs to be substituted with the full URL on which the application runs on. If GitLab, as an example, was reachable under `https://gitlab.example.com`, `<DOMAIN>` would be exactly the same.

|   Application   |                Version                |                               Callback URL                               |                                                                                                                                              Notes                                                                                                                                               |
|:---------------:|:-------------------------------------:|:------------------------------------------------------------------------:|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
|      Gitea      |               `1.14.6`                |                 `<DOMAIN>/user/oauth2/authelia/callback`                 | `ROOT_URL` in `[server]` section of `app.ini` must be configured correctly. Typically it is `<DOMAIN>/`. The string `authelia` in the callback url is the `Authentication Name` of the configured Authentication Source in Gitea (Authentication Type: OAuth2, OAuth2 Provider: OpenID Connect). |
|     GitLab      |               `14.0.1`                |              `<DOMAIN>/users/auth/openid_connect/callback`               |                                                                                                                                                                                                                                                                                                  |
| Hasicorp Vault  |               `14.0.1`                | `<DOMAIN>/oidc/callback` and `<DOMAIN>/ui/vault/auth/oidc/oidc/callback` |                                                                                                                                                                                                                                                                                                  |
|      MinIO      |    `RELEASE.2021-07-12T02-44-53Z`     |                        `<DOMAIN>/oauth_callback`                         |                                                                                                                                                                                                                                                                                                  |
|    Nextcloud    | `22.1.0` + `nextcloud-oidc-login` app |                     `<DOMAIN>/apps/oidc_login/oidc`                      |                                                                                                                                                                                                                                                                                                  |
|      Wekan      |                `5.41`                 |                          `<DOMAIN>/_oauth_oidc`                          |                                                                                                                                                                                                                                                                                                  |
|  Portainer CE   |                `2.6.1`                |                                `<DOMAIN>`                                |                                                                                                                                                                                                                                                                                                  |
| Bookstack       | `21.10`                               |        `<DOMAIN>/oidc/callback`                                          |                                                                                                                                                                                                                                                                                                  |
| Harbor          | `1.10`                                |        `<DOMAIN>/-/oauth/callback`                                       |                                                                                                                                                                                                                                                                                                  |
| Verdaccio       | `5`                                   |        `<DOMAIN>/oidc/callback`                                          |                                                                                                                                                                                                                                                                                                  |
