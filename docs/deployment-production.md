# Deployment for Production

**Authelia** can be deployed on bare metal or on Kubernetes with two
different kind of artifacts: an npm package or a Docker image.

**NOTE:** If not done already, we highly recommend you first follow the
[Getting Started] documentation.

## On Bare Metal

**Authelia** has been designed to be a proxy companion handling the 
authentication and authorization requests for your entire infrastructure.

As **Authelia** will be key in your architecture, it requires several
components to make it highly-available. Deploying it in production means having an LDAP server for storing the information about the users, a Redis cache to store the user sessions in a distributed manner, a
MongoDB to persist user configurations and one or more nginx reverse proxies configured to be used with Authelia. With such a setup **Authelia** can easily be scaled to multiple instances to evenly handle the traffic.

**NOTE:** If you don't have all those components, don't worry, there is a way to deploy **Authelia** with only nginx. This is described in [Deployment for Devs].

Here are the available steps to deploy **Authelia** given 
the configuration file is **/path/to/your/config.yml**. Note that you can
create your own configuration file from [config.template.yml] located at
the root of the repo.

### Deploy With NPM

    npm install -g authelia
    authelia /path/to/your/config.yml

### Deploy With Docker

    docker run -v /path/to/your/config.yml:/etc/authelia/config.yml clems4ever/authelia


## On top of Kubernetes

<img src="../images/kubernetes.logo.png" width="50" style="padding-right: 10px" align="left">

**Authelia** can also be installed on top of [Kubernetes] using
[nginx ingress controller](https://github.com/kubernetes/ingress-nginx).

Please refer to the following [documentation](../example/kube/README.md)
for more information.

## FAQ

### Why is this not automated?

Ansible would be a very good candidate to automate the installation of such
an infrastructure on bare metal. We would be more than happy to review any PR on that matter.

Regarding Kubernetes, the right way to go would be to write a helm recipe.
Again, we would be glad to review any PR implementing this.



[config.template.yml]: ../config.template.yml
[Getting Started]: ./getting-started.md
[Deployment for Devs]: ./deployment-dev.md
[Kubernetes]: https://kubernetes.io/
