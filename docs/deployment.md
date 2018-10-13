# Deployment

**Authelia** can be deployed in two different ways: npm and docker.

Here are the available steps to deploy **Authelia** on your machine given 
your configuration file is **/path/to/your/config.yml**. Note that you can
create your own the configuration file from [config.template.yml] at the root
of the repo.

## Standalone

**Authelia** has been designed to be a proxy companion handling the SSO.
Therefore, deploying it in production means having an LDAP, a Redis, a
MongoDB and one or more nginx running and configured to be used with
Authelia.

If you don't have all of this, don't worry, there is a way to deploy
**Authelia** with only an nginx. To do so, please refer to the
[Getting Started]. Otherwise here are the command to run Authelia in your
environment.

### With NPM

    npm install -g authelia
    authelia /path/to/your/config.yml

### With Docker

    docker pull clems4ever/authelia
    docker run -v /path/to/your/config.yml:/etc/authelia/config.yml clems4ever/authelia

## Kubernetes

<img src="/images/kube-logo.png" width="24" align="left">

**Authelia** can also be used on top of Kubernetes using the nginx ingress
controller.

Please refer to the following [README](../example/kube/README.md) for more
information.

[config.template.yml]: ../config.template.yml
[Getting Started]: ./getting-started.md
