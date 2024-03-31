---
title: "Roundcube"
description: "Integrating Roundcube and Dovecot with the Authelia OpenID Connect 1.0 Provider."
summary: ""
date: 2023-11-15T08:58:00+11:00
draft: true
images: []
weight: 620
toc: true
community: true
seo:
  title: "" # custom title (optional)
  description: "" # custom description (recommended)
  canonical: "" # custom canonical URL (optional)
  noindex: false # false (default) or true
---

## Tested Versions

* [Authelia]
  * [4.38.0](https://github.com/authelia/authelia/releases/tag/v4.38.0)
* [Roundcube]
  * [1.6.5](https://github.com/roundcube/roundcubemail/releases/tag/1.6.4)
* [Dovecot]
  * [2.3.20](https://dovecot.org/doc/NEWS)
* [Postfix]
  * [3.7.6](https://www.postfix.org/announcements/postfix-3.8.1.html)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://roundcube.example.com/`
* __Authelia Root URL:__ `https://auth.example.com/`
* __Client ID:__ `roundcube`
* __Client Secret:__ `insecure_secret`

## Configuration

### Authelia

The following YAML configuration is an example __Authelia__ [client
configuration](../../../configuration/identity-providers/openid-connect/clients.md)
for use with [Roundcube]:

```yaml
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
        redirect_uris:
          - 'https://roundcube.example.com/oauth/callback/'
        scopes:
          - 'openid'
          - 'profile'
          - 'email'
        token_endpoint_auth_method: 'client_secret_post'
```

### Application

Configure [Roundcube OAuth2] to use Authelia as an [OpenID Connect 1.0] Provider. Edit your [Roundcube]
`/etc/roundcube/config.inc.php` configuration file and add the following:

```php
// Most probably you need this
$config['use_https'] = true;

$config['oauth_provider'] = 'generic';
$config['oauth_provider_name'] = 'Authelia OIDC';
$config['oauth_client_id'] = 'roundcube';
$config['oauth_client_secret'] = 'insecure_secret';
$config['oauth_auth_uri'] = 'https://auth.example.com/api/oidc/authorization';
$config['oauth_token_uri'] = 'https://auth.example.com/api/oidc/token';
$config['oauth_identity_uri'] = 'https://auth.example.com/api/oidc/userinfo';
$config['oauth_identity_fields'] = ['email'];
$config['oauth_scope'] = 'email openid profile';
// Optionally, skip Roundcube's login page
// $config['oauth_login_redirect'] = true;
```

*__Important Note:__ Roundcube's redirect URI is not configurable, but is dynamically built with bits coming from the
FCGI environment: `<scheme>://<fqdn>[:<port>]/...`. Specifically, the FQDN comes from the `HTTP_HOST` header. With
Authelia, non-localhost HTTP redirection is not allowed, thus you might want to force HTTPS via Roundcube's conf flag
`use_https`. However, the redirection breaks when the upstream application is listening on a explicit port, because the
resulting redirect URI would be something like `https://<fqdn>:<port>/...`. Thus, to obtain the correct redirect URI
`https://<fqdn>/...`, your reverse proxy's `fastcgi` parameter `SERVER_PORT` should be unset.*

IMAP and SMTP backend configuration:
- For an IMAP instance on localhost, the default conf should be enough. Otherwise, set the corresponding SSL/TLS options
  via 'imap_host' and 'imap_conn_options';
- For a SMTP instance on localhost, no auth would be required. However
  [Roundcube OAuth enforces](https://github.com/roundcube/roundcubemail/issues/9183) 'smtp_auth_type' = 'XOAUTH2' plus
  credentials, thus you *must* use TLS or SSL via `smtp_host` and `smtp_conn_options`!


### Dovecot

[Dovecot OAuth2] configuration goes into two files.

#### Common configuration

Normally in file `/etc/dovecot/dovecot.conf` or one of its ancillary files in
`/etc/dovecot/conf.d/`:

```bash
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

#### Backend configuration

As defined above, in file,  `/etc/dovecot/dovecot-oauth2.conf.ext`:

```bash
introspection_mode = post
introspection_url = https://roundcube:insecure_secret@auth.example.com/api/oidc/introspection
username_attribute = username
```

*__Important Note:__ The client ID and secret must figure as credentials in
the `introspection_url`.*

### Postfix

Even though no authentication would be required when your Postfix instance is on the same host, Roundcube OAuth2
[enforces 'XOAUTH2' auth type plus credentials](https://github.com/roundcube/roundcubemail/issues/9183) and gives up the SMTP + SSL/TLS handshaking as no auth options
would be offered from Postfix. Thus, Postfix must be configured with (Dovecot-type) [SASL](https://www.postfix.org/SASL_README.html) on port 25 (smtpd) or
587 (submission), with the following minimum set of options:

```bash
smtpd_sasl_auth_enable = yes
smtpd_sasl_path = private/auth
smtpd_sasl_security_options = noanonymous, noplaintext
smtpd_sasl_tls_security_options = noanonymous
smtpd_sasl_type = dovecot
```

## See Also

* [Roundcube OAuth2]
* [Dovecot OAuth2]
* [Postfix SASL]

[Authelia]: https://www.authelia.com
[Roundcube]: https://roundcube.net/
[Roundcube OAuth2]: https://github.com/roundcube/roundcubemail/wiki/Configuration:-OAuth2
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[Dovecot]: https://dovecot.org/
[Dovecot OAuth2]: https://doc.dovecot.org/configuration_manual/authentication/oauth2/
[Postfix]: https://www.postfix.org/
[Postfix SASL]: https://www.postfix.org/SASL_README.html
