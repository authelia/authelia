---
layout: default
title: Single Factor
parent: Features
nav_order: 3
---

# Single Factor

**Authelia** supports single factor authentication to let applications
send authenticated requests to other applications.

Single or two-factor authentication can be configured per resource of an
application for flexibility.

For instance, you can configure Authelia to grant access to all resources
matching `app1.example.com/api/(.*)` with only a single factor and all
resources matching `app1.example.com/admin` with two factors.

To know more about the configuration of the feature, please visit the
documentation about the [configuration](../configuration/access-control.md).


## HTTP Basic Auth

Authelia supports two different methods for basic auth.

### Proxy-Authorization header

Authelia reads credentials from the header `Proxy-Authorization` instead of
the usual `Authorization` header. This is because in some circumstances both Authelia
and the application could require authentication in order to provide specific
authorizations at the level of the application.

### API argument

If instead of the `Proxy-Authorization` header you want, or need, to use the more
conventional `Authorization` header, you should then configure your reverse-proxy
to use `/api/verify?auth=basic`.  
When authentication fails and `auth=basic` was set, Authelia's response will include
the `WWW-Authenticate` header. This will cause browsers to prompt for authentication,
and users will not land on the HTML login page.


## Session-Username header

Authelia by default only verifies the cookie and the associated user with that cookie can
access a protected resource. The client browser does not know the username and does not send
this to Authelia, it's stored by Authelia for security reasons.

The Session-Username header has been implemented as a means
to use Authelia with non-web services such as PAM. Basically how it works is if the
Session-Username header is sent in the request to the /api/verify endpoint it will
only respond with a sucess message if the cookie username and the header username
match.

### Example

These examples are for demonstration purposes only, the original use case and full instructions
are described [here](https://github.com/authelia/authelia/issues/1322#issuecomment-729519155).
You will need to adjust the FORWARDED_HOST and VERIFY_URL vars to achieve a functional result.

#### PAM Rule

`auth    [success=1 default=ignore]      pam_exec.so expose_authtok /usr/bin/pam-authelia `

#### PAM Script

```bash
#!/bin/bash
# The password from stdin
PAM_PASSWORD=$(cat -)

# url from which authelia session key was created
FORWARDED_HOST=auth.example.com

# internal path to verify api
VERIFY_URL=http://127.0.0.1:80/api/verify

AUTH_RESULT=$(curl -b "authelia_session=${PAM_PASSWORD}" -H "Session-Username: ${PAM_USER}" -H "X-Forwarded-Host: ${FORWARDED_HOST}" -H "X-Forwarded-Proto: https" -s -o /dev/null -I -w "%{http_code}" -L "${VERIFY_URL}")

if [[ "$AUTH_RESULT" == 200 ]]; then
  echo "Auth verify ok"
  exit 0
else
  echo "Auth verify failed $AUTH_RESULT"
  exit 1
fi
```

