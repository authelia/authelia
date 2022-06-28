---
title: "Files"
description: "Using the YAML File Configuration Method."
lead: "Authelia can be configured via files. This section describes utilizing this method."
date: 2022-06-15T17:51:47+10:00
draft: false
images: []
menu:
  configuration:
    parent: "methods"
weight: 101200
toc: true
---

*Authelia* loads `configuration.yml` as the configuration if you just run it. You can override this behaviour with the
following syntax:

```bash
authelia --config config.custom.yml
```

## Multiple Configuration Files

You can have multiple configuration files which will be merged in the order specified. If duplicate keys are specified
the last one to be specified is the one that takes precedence. Example:

```bash
authelia --config configuration.yml --config config-acl.yml --config config-other.yml
authelia --config configuration.yml,config-acl.yml,config-other.yml
```

Authelia's configuration files use the YAML format. A template with all possible options can be found at the root of the
repository [here](https://github.com/authelia/authelia/blob/master/config.template.yml).

*__Important Note:__ You should not have configuration sections such as Access Control Rules or OpenID Connect clients
configured in multiple files. If you wish to split these into their own files that is fine, but if you have two files that
specify these sections and expect them to merge properly you are asking for trouble.*

### Container

By default, the container looks for a configuration file at `/config/configuration.yml`.

### Docker

This is an example of how to override the configuration files loaded in docker:

```bash
docker run -d --volume /path/to/config:/config authelia:authelia:latest authelia --config=/config/configuration.yaml --config=/config/configuration.acl.yaml
```

See the [Docker Documentation](https://docs.docker.com/engine/reference/commandline/run/) for more information on the
`docker run` command.

### Docker Compose

An excerpt from a docker compose that allows you to specify multiple configuration files is as follows:

```yaml
version: "3.8"
services:
  authelia:
    container_name: authelia
    image: authelia/authelia:latest
    command:
      - "authelia"
      - "--config=/config/configuration.yaml"
      - "--config=/config/configuration.acl.yaml"

```

See the [compose file reference](https://docs.docker.com/compose/compose-file/compose-file-v3/#command) for more
information.

### Kubernetes

An excerpt from a Kubernetes container that allows you to specify multiple configuration files is as follows:

```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  name: authelia
  namespace: authelia
  labels:
    app.kubernetes.io/instance: authelia
    app.kubernetes.io/name: authelia
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/instance: authelia
      app.kubernetes.io/name: authelia
  template:
    metadata:
      labels:
        app.kubernetes.io/instance: authelia
        app.kubernetes.io/name: authelia
    spec:
      enableServiceLinks: false
      containers:
        - name: authelia
          image: docker.io/authelia/authelia:fix-missing-head-handler
          command:
            - authelia
          args:
            - '--config=/configuration.yaml'
            - '--config=/configuration.acl.yaml'
```

See the Kubernetes [workloads documentation](https://kubernetes.io/docs/concepts/workloads/pods/#pod-templates) or the
[Container API docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#container-v1-core) for more
information.
