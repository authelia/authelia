---
title: "Roundcube"
description: "Integrating Roundcube and Dovecot with the Authelia OpenID Connect 1.0 Provider."
lead: ""
date: 2023-10-23T15:16:47+02:00
draft: true
images: []
menu:
  integration:
    parent: "openid-connect"
weight: 620
toc: true
community: true
---

## Tested Versions

All packages are from Debian 12.

* [Authelia]
  * [4.37.5](https://github.com/authelia/authelia/releases/tag/v4.37.5)
* [Roundcube]
  * [1.6.4](https://github.com/roundcube/roundcubemail/releases/tag/1.6.4)
* [Dovecot]
  * [2.3.20](https://dovecot.org/doc/NEWS)
* [Postfix]
  * [3.7.6](https://www.postfix.org/announcements/postfix-3.8.1.html)
* [Nginx]
  * [1.22.1](https://nginx.org/en/CHANGES-1.22)

## Before You Begin

{{% oidc-common %}}

### Assumptions

This example makes the following assumptions:

* __Application Root URL:__ `https://roundcube.example.com`
* __Authelia Root URL:__ `https://auth.example.com`
* __Client ID:__ `roundcube`
* __Client Secret:__ `insecure_secret`

## Configuration

### Application

Configure [Roundcube OAuth2] to use Authelia as an [OpenID Connect 1.0]
Provider. Edit your [Roundcube] `/etc/roundcube/config.inc.php` configuration
file and add the following:

```php
// Authelia doesn't accept non-localhpst HTTP redirection but RC build its
// redirect_uri from HTTP_HOST header which is a FQDN (even if reverse-proxied
// to localhost)
$config['use_https'] = true;

$config['oauth_provider'] = 'generic';
$config['oauth_provider_name'] = 'Authelia OIDC';
$config['oauth_client_id'] = roundcube';
$config['oauth_client_secret'] = 'insecure_secret';
$config['oauth_auth_uri'] = 'https://auth.example.com/api/oidc/authorization';
$config['oauth_token_uri'] = 'https://auth.example.com/api/oidc/token';
$config['oauth_identity_uri'] = 'https://auth.example.com/api/oidc/userinfo';
$config['oauth_identity_fields'] = ['email'];
$config['oauth_scope'] = 'email openid profile';
// Optionally, skip Roundcube's login page
//$config['oauth_login_redirect'] = true;

// Note on IMAP and SMTP backends (cf. Debian's '/etc/roundcube/defaults.inc.php').
// - For an IMAP instance on localhost, the default conf is enough. Otherwise,
//   set the corresponding SSL/TLS options via 'imap_host' and 'imap_conn_options';
// - For an SMTP instance on localhost, no auth would be required. However
//   RC's OAuth enforces 'smtp_auth_type' = 'XOAUTH2' plus credentials (cf.
//   <https://github.com/roundcube/roundcubemail/issues/9183>), thus you *must*
//   use TLS or SSL via 'smtp_host' and 'smtp_conn_options'!
```

### Authelia

The following YAML configuration is an example __Authelia__
[client configuration](../../../configuration/identity-providers/openid-connect/clients.md) for use with [Roundcube]
which will operate with the above example:

```yaml
identity_providers:
  oidc:
    ## The other portions of the mandatory OpenID Connect 1.0 configuration go here.
    ## See: https://www.authelia.com/c/oidc
    clients:
    - id: 'roundcube'
      description: 'Roundcube'
      secret: '$pbkdf2-sha512$310000$c8p78n7pUMln0jzvd4aK4Q$JNRBzwAo0ek5qKn50cFzzvE9RXV88h1wJn5KGiHrD0YKtZaR/nCb2CJPOsKaPK0hjf.9yHxzQGZziziccp6Yng'  # The digest of 'insecure_secret'.
      public: false
      authorization_policy: 'two_factor'
      redirect_uris:
        - 'https://roundcube.example.com/oauth/callback/'
      scopes:
        - 'openid'
        - 'profile'
        - 'email'
# optional
access_control:
  rules:
    - domain: 'webmail.example.com'
      resources:
        - '^/(skins|plugins|jquery|program)([/?].*)?$'
      policy: bypass
```

Also, mind that your reverse proxy might need special arrangements -- see an example for Nginx here below.

### Nginx

The [standard authrequest-based
example](https://www.authelia.com/integration/proxies/nginx/#standard-example)
applies. However, if you want to use the *same* reverse-proxy instance for
your protected application, its corresponding vhost might need the following
adaptations:

```nginx
# Backend proxied app on the same host
server {
    # See N.B. below
    listen 9092;
    server_name roundcube;

    ...

    location ~ [^/]\.php(?:/|$) {
        fastcgi_split_path_info ^(.+?\.php)(/.*)$;
        if (!-f $document_root$fastcgi_script_name) {
            return 404;
        }

        include fastcgi_params;

        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param PATH_INFO       $fastcgi_path_info;

        # See N.B. below
        fastcgi_param  SERVER_PORT '';

        fastcgi_pass $your_php_upstream;
    }
}
```

**N.B.** Roundcube's redirect URI is _not_ configurable, but is
dynamically built with bits coming from the FCGI environment:
`<scheme>://<fqdn>[:<port>]/...`. To force HTTPS, Roundcube's conf flag
`use_https` must be set. However, the redirection breaks when the backend
application is listening on a specific port, because the resulting redirect
URI would be something like 'https://<fqdn>:<port>/...'. Thus, to obtain the
correct redirect URI `'https://<fqdn>/...'`, the fastcgi_param SERVER_PORT
must be unset.


### Dovecot

[Dovecot OAuth2] configuration goes into two files.

#### Common configuration

Normally in file `/etc/dovecot/dovecot.conf` or one of its ancillary files in
`/etc/dovecot/conf.d/`:

```nginx
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

```nginx
introspection_mode = post
introspection_url = https://roundcube:insecure_secret@auth.example.com/api/oidc/introspection
username_attribute = username
```

**N.B.** The client ID and secret must figure as credentials in the `introspection_url`.

### Postfix

Even though no authentication would be required when your Postfix instance is
on the same host, Roundcube OAuth2 [enforces 'XOAUTH2' auth type plus
credentials](https://github.com/roundcube/roundcubemail/issues/9183) and gives
up the SMTP + SSL/TLS handshaking as no auth options would be offered from
Postfix. Thus, Postfix must be configured with (Dovecot-type)
[SASL](https://www.postfix.org/SASL_README.html) on port 25 (smtpd) or 587
(submission), with the following minimum set of options:

```nginx
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
[Roundcube]: https://www.roundcube.com/
[Roundcube OAuth2]: https://github.com/roundcube/roundcubemail/wiki/Configuration:-OAuth2
[OpenID Connect 1.0]: ../../openid-connect/introduction.md
[Dovecot]: https://dovecot.org/
[Dovecot OAuth2]: https://doc.dovecot.org/configuration_manual/authentication/oauth2/
[Postfix]: https://www.postfix.org/
[Postfix SASL]: https://www.postfix.org/SASL_README.html
