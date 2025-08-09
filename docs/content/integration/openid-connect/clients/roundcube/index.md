---
title: "Roundcube"
description: "Integrating Roundcube and Dovecot with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2024-03-14T06:00:14+11:00
draft: false
images: []
weight: 620
toc: true
aliases:
  - '/integration/openid-connect/roundcube/'
support:
  level: community
  versions: true
  integration: true
seo:
  title: "Roundcube | OpenID Connect 1.0 | Integration"
  description: "Step-by-step guide to configuring Roundcube with OpenID Connect 1.0 for secure SSO. Enhance your login flow using Authelia’s modern identity management."
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

- [Authelia]
  - [4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
- [Roundcube]
  - [v1.6.5](https://github.com/roundcube/roundcubemail/releases/tag/1.6.4)
- [Dovecot]
  - [v2.3.20](https://dovecot.org/doc/NEWS)
- [Postfix]
  - [v3.7.6](https://www.postfix.org/announcements/postfix-3.8.1.html)

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

- __Application Root URL:__ `https://roundcube.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Authelia Root URL:__ `https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/`
- __Client ID:__ `roundcube`
- __Client Secret:__ `insecure_secret`

Some of the values presented in this guide can automatically be replaced with documentation variables.

{{< sitevar-preferences >}}

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client configuration] for use with [Roundcube] which will
operate with the application example:

```yaml {title="configuration.yml"}
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
      - client_id: 'roundcube'
        client_name: 'Roundcube'
        client_secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
        public: false
        authorization_policy: 'two_factor'
        require_pkce: false
        pkce_challenge_method: ''
        redirect_uris:
          - 'https://roundcube.{{< sitevar name="domain" nojs="example.com" >}}/oauth/callback/'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        response_types:
          - 'code'
        grant_types:
          - 'authorization_code'
        access_token_signed_response_alg: 'none'
        userinfo_signed_response_alg: 'none'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

To configure [Roundcube OAuth2] there is one method, using the [Configuration File](#configuration-file).

##### Configuration File

##### Roundcube

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `/etc/roundcube/config.inc.php`.
{{< /callout >}}

To configure [Roundcube OAuth2] to utilize Authelia as an [OpenID Connect 1.0] Provider, use the following configuration:

```php {title="/etc/roundcube/config.inc.php"}
$config['use_https'] = true;

$config['oauth_provider'] = 'generic';
$config['oauth_provider_name'] = 'Authelia OIDC';
$config['oauth_client_id'] = 'roundcube';
$config['oauth_client_secret'] = 'insecure_secret';
$config['oauth_auth_uri'] = 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/authorization';
$config['oauth_token_uri'] = 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/token';
$config['oauth_identity_uri'] = 'https://{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/userinfo';
$config['oauth_identity_fields'] = ['email'];
$config['oauth_scope'] = 'email openid profile';
// Optionally, skip Roundcube's login page
// $config['oauth_login_redirect'] = true;
```

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
Roundcube's redirect URI is not configurable, but is dynamically built with bits coming from the
FCGI environment: `<scheme>://<fqdn>[:<port>]/...`. Specifically, the FQDN comes from the `HTTP_HOST` header. With
Authelia, non-localhost HTTP redirection is not allowed, thus you might want to force HTTPS via Roundcube's conf flag
`use_https`. However, the redirection breaks when the upstream application is listening on a explicit port, because the
resulting redirect URI would be something like `https://<fqdn>:<port>/...`. Thus, to obtain the correct redirect URI
`https://<fqdn>/...`, your reverse proxy's `fastcgi` parameter `SERVER_PORT` should be unset.
{{< /callout >}}

IMAP and SMTP backend configuration:
- For an IMAP instance on localhost, the default conf should be enough. Otherwise, set the corresponding SSL/TLS options
  via 'imap_host' and 'imap_conn_options';
- For a SMTP instance on localhost, no auth would be required. However
  [Roundcube OAuth enforces](https://github.com/roundcube/roundcubemail/issues/9183) 'smtp_auth_type' = 'XOAUTH2' plus
  credentials, thus you *must* use TLS or SSL via `smtp_host` and `smtp_conn_options`!


##### Dovecot Common

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `/etc/dovecot/dovecot.conf` or is one of the ancillary files in
`/etc/dovecot/conf.d/`.
{{< /callout >}}

```ext {title="/etc/dovecot/dovecot.conf"}
auth_mechanisms = $auth_mechanisms oauthbearer xoauth2

passdb {
  args = /etc/dovecot/dovecot-oauth2.conf.ext
  driver = oauth2
  mechanisms = xoauth2 oauthbearer
}

# Optional for Postfix SASL on smtpd/submission
service auth {
  unix_listener /var/spool/postfix/private/auth {
    group = postfix
    mode = 0666
    user = postfix
  }
}
```

##### Dovecot Backend

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `/etc/dovecot/dovecot-oauth2.conf.ext`.
{{< /callout >}}

```ext {title="/etc/dovecot/dovecot-oauth2.conf.ext"}
introspection_mode = post
introspection_url = https://roundcube:insecure_secret@{{< sitevar name="subdomain-authelia" nojs="auth" >}}.{{< sitevar name="domain" nojs="example.com" >}}/api/oidc/introspection
username_attribute = username
```

{{< callout context="caution" title="Important Note" icon="outline/alert-triangle" >}}
The client ID and secret must figure as credentials in
the `introspection_url`.
{{< /callout >}}

##### Postfix

{{< callout context="tip" title="Did you know?" icon="outline/rocket" >}}
Generally the configuration file is named `/etc/postfix/main.cf`.
{{< /callout >}}

Even though no authentication would be required when your Postfix instance is on the same host, Roundcube OAuth2
[enforces 'XOAUTH2' auth type plus credentials](https://github.com/roundcube/roundcubemail/issues/9183) and gives up the
SMTP + SSL/TLS handshaking as no auth options would be offered from Postfix. Thus, Postfix must be configured with
(Dovecot-type) [SASL](https://www.postfix.org/SASL_README.html) on port 25 (smtpd) or 587 (submission), with the following minimum set of options:

```cf {title="/etc/postfix/main.cf"}
smtpd_sasl_auth_enable = yes
smtpd_sasl_path = private/auth
smtpd_sasl_security_options = noanonymous, noplaintext
smtpd_sasl_tls_security_options = noanonymous
smtpd_sasl_type = dovecot
```

## See Also

- [Roundcube OAuth2]
- [Dovecot OAuth2]
- [Postfix SASL]

[Authelia]: https://www.authelia.com
[Roundcube]: https://roundcube.net/
[Roundcube OAuth2]: https://github.com/roundcube/roundcubemail/wiki/Configuration:-OAuth2
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[Dovecot]: https://dovecot.org/
[Dovecot OAuth2]: https://doc.dovecot.org/main/core/config/auth/databases/oauth2.html
[Postfix]: https://www.postfix.org/
[Postfix SASL]: https://www.postfix.org/SASL_README.html
[client configuration]: ../../../configuration/identity-providers/openid-connect/clients.md
