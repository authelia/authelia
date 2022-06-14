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

|   Application    |        Minimal Version         |                                                    Notes                                                      |
|:----------------:|:------------------------------:|:-------------------------------------------------------------------------------------------------------------:|
| Bookstack        | `21.10`                        |                                                                                                               |
| Gitea            | `1.14.6`                       |                                                                                                               |
| GitLab           | `13.0.0`                       |                                                                                                               |
| Grafana          | `8.0.5`                        |                                                                                                               |
| Harbor           | `1.10`                         | It works on >v2.1 also, but not sure if there is OIDC support on v2.0                                         |
| Hashicorp Vault  | `1.8.1`                        |                                                                                                               |
| Miniflux         | `2.0.21`                       |                                                                                                               |
| MinIO            | `RELEASE.2021-11-09T03-21-45Z` | must set `MINIO_IDENTITY_OPENID_CLAIM_NAME: groups` in MinIO and set [MinIO policies](https://docs.min.io/minio/baremetal/security/minio-identity-management/policy-based-access-control.html#minio-policy) as groups in Authelia |
| Nextcloud        | `22.1.0`                       | Tested using the `nextcloud-oidc-login` app - [Link](https://github.com/pulsejet/nextcloud-oidc-login)        |
| Portainer CE     | `2.6.1`                        | Settings to use username as ID: set `Scopes` to `openid` and `User Identifier` to `preferred_username`        |
| Seafile          | `9.0.4`                        | Requires `OAUTH_ATTRIBUTE_MAP` to contain the mapping of the `id` field even if not present in Authelia, e.g. `'id': (False, "unused")` (see [seahub#5162](https://github.com/haiwen/seahub/issues/5162)) |
| Verdaccio        | `5`                            | Depends on this fork of verdaccio-github-oauth-ui: [Link](https://github.com/OnekO/verdaccio-github-oauth-ui) |
| Wekan            | `5.41`                         |                                                                                                               |

## Known Callback URLs

If you do not find the application in the list below, you will need to search for yourself - and maybe come back to open a PR to add your application to this list so others won't have to search for them.

`<DOMAIN>` needs to be substituted with the full URL on which the application runs on. If GitLab, as an example, was reachable under `https://gitlab.example.com`, `<DOMAIN>` would be exactly the same.

|   Application   |                Version                |                               Callback URL                               |                                                                                                                                              Notes                                                                                                                                               |
|:---------------:|:-------------------------------------:|:------------------------------------------------------------------------:|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|
| Bookstack       | `21.10`                               | `<DOMAIN>/oidc/callback`                                                 |                                                                                                                                                                                                                                                                                                  |
| Gitea           | `1.14.6`                              | `<DOMAIN>/user/oauth2/authelia/callback`                                 | `ROOT_URL` in `[server]` section of `app.ini` must be configured correctly. Typically it is `<DOMAIN>/`. The string `authelia` in the callback url is the `Authentication Name` of the configured Authentication Source in Gitea (Authentication Type: OAuth2, OAuth2 Provider: OpenID Connect). |
| GitLab          | `14.0.1`                              | `<DOMAIN>/users/auth/openid_connect/callback`                            |                                                                                                                                                                                                                                                                                                  |
| Harbor          | `1.10`                                | `<DOMAIN>/-/oauth/callback`                                              |                                                                                                                                                                                                                                                                                                  |
| Hashicorp Vault  | `14.0.1`                              | `<DOMAIN>/oidc/callback` and `<DOMAIN>/ui/vault/auth/oidc/oidc/callback` |                                                                                                                                                                                                                                                                                                  |
| Miniflux        | `2.0.21`                              | `<DOMAIN>/oauth2/oidc/callback`                                          | Set via Miniflux `OAUTH2_REDIRECT_URL` [configuration parameter](https://miniflux.app/docs/configuration.html#oauth2-redirect-url). Example value follows this format                                                                                                                                            |
| MinIO           | `RELEASE.2021-07-12T02-44-53Z`        | `<DOMAIN>/oauth_callback`                                                |                                                                                                                                                                                                                                                                                                  |
| Nextcloud       | `22.1.0` + `nextcloud-oidc-login` app | `<DOMAIN>/apps/oidc_login/oidc`                                          |                                                                                                                                                                                                                                                                                                  |
| Portainer CE    | `2.6.1`                               | `<DOMAIN>`                                                               |                                                                                                                                                                                                                                                                                                  |
| Seafile         | `9.0.4`                               | `<DOMAIN>/oauth/callback/`                                               | Must exactly match `OAUTH_REDIRECT_URL` value as set in `seahub_settings.py`                                                                                                                                                                                                                                          |
| Verdaccio       | `5`                                   | `<DOMAIN>/oidc/callback`                                                 |                                                                                                                                                                                                                                                                                                  |
| Wekan           | `5.41`                                | `<DOMAIN>/_oauth_oidc`                                                   |                                                                                                                                                                                                                                                                                                  |
