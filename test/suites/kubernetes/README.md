# Kubernetes suite

This suite has been created to test Authelia in Kubernetes with a nginx-ingress-controller.

## Components

This suite spawns nginx-ingress-controller, redis, mongo, ldap and a fake webmail to catch
emails sent by Authelia. The configuration of all those services is located in *example/kube*.

## Tests

This suite tests if single and two-factor is working.
