# Authelia Helm Chart

Authelia can be deployed to Kubernetes using a Helm Chart in order to protect your most critical applications using 2-factor authentication and Single Sign-On.

This chart includes **phpldapadmin** by default so that editing the contents of the openldap is a breeze. The **openldap** chart also comes pre-populated with values, check the main `values.yaml` which overrides the defaults of the openldap chart. To test the login use the following credentials: `user01:password`. 

All the services are reachable via their service cluster IPs.

The chart also leverages [ingress-nginx](https://github.com/kubernetes/ingress-nginx) to delegate authentication and authorization to Authelia within the cluster.

## Authelia Config

Authelia's config file can be found inside `files/authelia-config.yml`. We treat the config file as a Helm Template which gets rendered with some value overrides present in the chart's values.yaml.

## Getting started

### Set up a Kube Cluster

**NOTE:** If you have your Kubernetes cluster already running, you can skip this step. 

Create a new Kubernetes cluster by using one of several methods: 

* [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
* [microk8s](https://tutorials.ubuntu.com/tutorial/install-a-local-kubernetes-with-microk8s#0) on Ubuntu
* [kind](https://github.com/kubernetes-sigs/kind) (Kubernetes in Docker)

You will have to enable/deploy the following k8s components for the Helm chart to succeed (usually these are available in most dev environments and should be easy to deploy using the abovementioned options): dns, local storage (hostPath) and ingress. If you don't use a cloud-managed solution (like GKE or OpenShift), enabling the k8s standard *dashboard* is a nice addition to better help you debug issues and have a wide overview of the deployment. 

### Deploy the Chart

To install the chart, you first need to obtain a local copy of Authelia's dependencies:

```console
$ helm dependency update authelia
```

(here, *authelia* represents the top folder containing the Authelia chart)

Then finally:

```console
$ helm install -n authelia authelia
```

## Configuration

The following table lists the configurable parameters of the Authelia chart and their default values.

| Parameter                             | Description                                                                                                                                                                                                                                                                                                                 | Default                                                                                                                                                                                                                          |
|---------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| deployment.image.name                 | Authelia's docker image                                                                                                                                                                                                                                                                                                     | clems4ever/authelia                                                                                                                                                                                                              |
| deployment.image.version              | Authelia's docker image version                                                                                                                                                                                                                                                                                             | latest                                                                                                                                                                                                                           |
| service.serviceport                   | Service HTTP port                                                                                                                                                                                                                                                                                                           | 80                                                                                                                                                                                                                               |
| service.containerport                 | Authelia pod port, it will be automatically written to authelia_config.yaml                                                                                                                                                                                                                                                 | 80                                                                                                                                                                                                                               |
| authelia_conf.default_redirection_url | This parameter allows you to specify the default redirection URL Authelia will use in such a case.                                                                                                                                                                                                                          | https://mypage.example.com                                                                                                                                                                                                       |
| authelia_conf.session_name            | The session cookie name to identify the user once logged in                                                                                                                                                                                                                                                                 | my_life_for_aiur                                                                                                                                                                                                                 |
| authelia_conf.session_secret          | The session cookie secret to identify the user once logged in                                                                                                                                                                                                                                                               | RdajeQNYr3bU6DhY8C5Pf                                                                                                                                                                                                            |
| redis.image                           | The Redis instance docker image configuration                                                                                                                                                                                                                                                                               | registry: docker.io, repository: bitnami/redis, tag: 4.0.14, pullPolicy: IfNotPresent                                                                                                                                            |
| redis.cluster                         | The Redis instance cluster config. These values will override the default ones in the Redis helm chart.                                                                                                                                                                                                                     | enabled: false, slaveCount: 0                                                                                                                                                                                                    |
| redis.master.persistence.enabled      | The Redis instance persistent storage. These values will override the default ones in the Redis helm chart.                                                                                                                                                                                                                 | false                                                                                                                                                                                                                            |
| redis.usePassword                     | Whether you want the Redis instance to require authentication or not. These values will override the default ones in the Redis helm chart.                                                                                                                                                                                  | true                                                                                                                                                                                                                             |
| redis.password                        | The Redis instance authentication password. These values will override the default ones in the Redis helm chart.                                                                                                                                                                                                            | defaultpass                                                                                                                                                                                                                      |
| mongodb.image                         | The MongoDB instance docker image configuration. These values will override the default ones in the Redis helm chart.                                                                                                                                                                                                       | registry: docker.io, repository: bitnami/mongodb, tag: 4.0.9, pullPolicy: Always                                                                                                                                                 |
| mongodb.usePassword                   | Whether you want the MongoDB instance to require authentication or not.  These values will override the default ones in the MongoDB helm chart.                                                                                                                                                                             | true                                                                                                                                                                                                                             |
| mongodb.mongodbRootPassword           | The MongoDB instance authentication password. These values will override the default ones in the MongoDB helm chart.                                                                                                                                                                                                        | defaultpass                                                                                                                                                                                                                      |
| mongodb.mongodbSystemLogVerbosity     | The MongoDB instance log verbosity. These values will override the default ones in the MongoDB helm chart.                                                                                                                                                                                                                  | 5 (quite verbose)                                                                                                                                                                                                                |
| mongodb.persistence                   | The MongoDB instance storage persistence config. NOTE: bear in mind the default values for this will only fit a k8s cluster deployed via "microk8s", adjust the values to reflect your own storage class depending on your cluster deployment. These values will override the default ones in the MongoDB helm chart.       | enabled: true, storageClass: "microk8s-hostpath", storageClassName: "microk8s-hostpath", mountPath: /bitnami/mongodb, accessModes: [- ReadWriteOnce], size: 250Mi                                                                |
| openldap.image                        | The OpenLDAP instance docker image configuration. These values will override the default ones in the OpenLDAP helm chart.                                                                                                                                                                                                   | repository: osixia/openldap, tag: 1.2.1, pullPolicy: IfNotPresent                                                                                                                                                                |
| openldap.env                          | The OpenLDAP instance environment variables that will be injected at runtime to create the starting ldap database.                                                                                                                                                                                                          | LDAP_ORGANISATION: "Authelia", LDAP_DOMAIN: "authelia.com", LDAP_BACKEND: "hdb", LDAP_TLS: "true", LDAP_TLS_ENFORCE: "false", LDAP_REMOVE_CONFIG_AFTER_SETUP: "true"                                                             |
| openldap.adminPassword                | The OpenLDAP instance root (admin) password. These values will override the default ones in the OpenLDAP helm chart.                                                                                                                                                                                                        | defaultpass                                                                                                                                                                                                                      |
| openldap.persistence                  | The OpenLDAP instance storage persistence config. NOTE: bear in mind the  default values for this will only fit a k8s cluster deployed via  "microk8s", adjust the values to reflect your own storage class  depending on your cluster deployment. These values will override the  default ones in the OpenLDAP helm chart. | enabled: true, storageClass: "microk8s-hostpath", storageClassName: "microk8s-hostpath", mountPath: /bitnami/mongodb, accessModes: [- ReadWriteOnce], size: 250Mi                                                                |
| openldap.ldifFiles                    | The OpenLDAP instance LDAP injection values. These will be injected to the ldap database upon start.                                                                                                                                                                                                                        | Check values.yaml for the default value. We inject a basic structure with a user for testing purposes. Adapt it to your needs or remove this key altogether if you want to start from scratch. Test user creds: user01@password. |
| resources                             | CPU/Memory resource requests/limits                                                                                                                                                                                                                                                                                         | Memory: 256Mi, CPU: 100m                                                                                                                                                                                                         |

## Walkthrough using microk8s
We will describe a step-by-step way of deploying the chart to a microk8s test instance. Steps to reproduce:

1. Create a fresh Ubuntu 16 or 18 VM
2. Install microk8s: `snap install microk8s --classic`
2. Enable dashboard, dns, and ingress for the microk8s instance: `microk8s.enable storage ingress dns dashboard`. As you will see, the chart currently doesn't deploy an *ingress controller* because you can deploy it natively in microk8s. 
3. Download Helm binary and initialize Tiller.
4. The Authelia chart uses other sub-charts as dependencies (MongoDB, Redis, OpenLDAP), so we need to add this to the local Helm cache before we can install Authelia: `helm dependency update authelia`
5. **IMPORTANT**: adjust the Authelia config in `./files/authelia-config.yml` to match your test ground, then also adjust the chart's `values.yaml` if required to match authelia-config.yml.
6. Finally just deploy the chart: `helm install -n authelia authelia` (**Note:** the first *authelia* is the name of the release, the 2nd is the name of the folder where we keep our raw chart files)

## Todo

The current chart needs some improvements:

* Add the option to install an ingress controller using a switch in values.yaml (something like *install_ingress: true/false*)
* Craft the secrets.yaml file so that instead of having two hard-coded TLS secrets for the two hard-coded ingresses, they can dynamically generate secrets based on the amount of sites (ingressess) we want to protect


## How does it work?

### Authentication via Authelia

In a Kube clusters, the routing logic of requests is handled by ingress
controllers following rules provided by ingress configurations.

In this example, [ingress-nginx](https://github.com/kubernetes/ingress-nginx)
controller has been installed to handle the incoming requests. Some of them
(specified in the ingress configuration) are forwarded to Authelia so that
it can verify whether they are allowed and should reach the protected endpoint.

The authentication is provided at the ingress level by an annotation called
`nginx.ingress.kubernetes.io/auth-url` that is filled with the URL of
Authelia's verification endpoint.
The ingress controller also requires the URL to the
authentication portal so that the user can be redirected if he is not
yet authenticated. This annotation is as follows:
`nginx.ingress.kubernetes.io/auth-signin: "https://login.example.com:8080/#/"`

Those annotations can be seen in `values.yaml` in the `ingress` configuration section.

## Questions

If you have questions about the implementation, please post them on
[![Gitter](https://img.shields.io/gitter/room/badges/shields.svg)](https://gitter.im/authelia/general?utm_source=share-link&utm_medium=link&utm_campaign=share-link)
