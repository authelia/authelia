# High-availability suite

This suite is made to test Authelia in a *complete* environment, that is, with
all components making Authelia highly available.

## Components

This suite will spawn nginx as the edge reverse proxy, redis and mongo for storing
user sessions and configurations, LDAP for storing user accounts and authenticating,
as well as a few helpers such as a fake webmail to receive e-mails sent by Authelia
and httpbin to check headers forwarded by Authelia.

## Tests

There is broad range of tests in this suite. Check out in the *scenarii* directory.