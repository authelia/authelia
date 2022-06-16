---
title: "Secrets"
description: "Using the Secrets Configuration Method."
lead: "Authelia allows providing configuration via secrets method. This section describes how to implement this."
date: 2022-03-20T12:52:27+11:00
draft: false
images: []
menu:
  configuration:
    parent: "methods"
weight: 101400
toc: true
aliases:
  - /c/secrets
  - /docs/configuration/secrets.html
---

Configuration of *Authelia* requires several secrets and passwords. Even if they can be set in the configuration file or
standard environment variables, the recommended way to set secrets is to use this configuration method as described below.

See the [security](#security) section for more information.

## Layers

*__Important Note:__* While this method is the third layer of the layered configuration model as described by the
[introduction](introduction.md#layers), this layer is special in as much as *Authelia* will not start if you define
a secret as well as any other configuration method.

For example if you define `jwt_secret` in the [files method](files.md) and/or `AUTHELIA_JWT_SECRET` in the
[environment method](environment.md), as well as the `AUTHELIA_JWT_SECRET_FILE`, this will cause the aforementioned error.

## Security

This method is a slight improvement over the security of the other methods as it allows you to easily separate your
configuration in a logically secure way.

## Environment variables

A secret value can be loaded by *Authelia* when the configuration key ends with one of the following words: `key`,
`secret`, `password`, or `token`.

If you take the expected environment variable for the configuration option with the `_FILE` suffix at the end. The value
of these environment variables must be the path of a file that is readable by the Authelia process, if they are not,
*Authelia* will fail to load. Authelia will automatically remove the newlines from the end of the files contents.

For instance the LDAP password can be defined in the configuration
at the path __authentication_backend.ldap.password__, so this password
could alternatively be set using the environment variable called
__AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE__.

Here is the list of the environment variables which are considered secrets and can be defined. Please note that only
secrets can be loaded into the configuration if they end with one of the suffixes above, you can set the value of any
other configuration using the environment but instead of loading a file the value of the environment variable is used.

|                 Configuration Key                 |                   Environment Variable                   |
|:-------------------------------------------------:|:--------------------------------------------------------:|
|                      tls_key                      |                  AUTHELIA_TLS_KEY_FILE                   |
|                    jwt_secret                     |                 AUTHELIA_JWT_SECRET_FILE                 |
|                duo_api.secret_key                 |             AUTHELIA_DUO_API_SECRET_KEY_FILE             |
|                  session.secret                   |               AUTHELIA_SESSION_SECRET_FILE               |
|              session.redis.password               |           AUTHELIA_SESSION_REDIS_PASSWORD_FILE           |
| session.redis.high_availability.sentinel_password | AUTHELIA_REDIS_HIGH_AVAILABILITY_SENTINEL_PASSWORD_FILE  |
|              storage.encryption_key               |           AUTHELIA_STORAGE_ENCRYPTION_KEY_FILE           |
|              storage.mysql.password               |           AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE           |
|             storage.postgres.password             |         AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE          |
|              notifier.smtp.password               |           AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE           |
|       authentication_backend.ldap.password        |    AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE    |
|    identity_providers.oidc.issuer_private_key     | AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE |
|        identity_providers.oidc.hmac_secret        |    AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET_FILE     |

## Secrets in configuration file

If for some reason you decide on keeping the secrets in the configuration file, it is strongly recommended that you
ensure the permissions of the configuration file are appropriately set so that other users or processes cannot access
this file. Generally the UNIX permissions that are appropriate are 0600.

## Secrets exposed in an environment variable

In all versions 4.30.0+ you can technically set secrets using the environment variables without the `_FILE` suffix by
setting the value to the value you wish to set in configuration, however we strongly urge people not to use this option
and instead use the file-based secrets above.

Prior to implementing file secrets the only way you were able to define secret values was either via configuration or
via environment variables in plain text.

See [this article](https://diogomonica.com/2017/03/27/why-you-shouldnt-use-env-variables-for-secret-data/) for reasons
why setting them via the file counterparts is highly encouraged.

## Docker

Secrets can be provided in a `docker-compose.yml` either with Docker secrets or
bind mounted secret files, examples of these are provided below.

### Compose with Docker secrets

This example assumes secrets are stored in `/path/to/authelia/secrets/{secretname}`
on the host and are exposed with Docker secrets in a `docker-compose.yml` file:

```yaml
version: '3.8'

networks:
  net:
    driver: bridge

secrets:
  jwt:
    file: /path/to/authelia/secrets/jwt
  duo:
    file: /path/to/authelia/secrets/duo
  session:
    file: /path/to/authelia/secrets/session
  redis:
    file: /path/to/authelia/secrets/redis
  mysql:
    file: /path/to/authelia/secrets/mysql
  smtp:
    file: /path/to/authelia/secrets/smtp
  ldap:
    file: /path/to/authelia/secrets/ldap

services:
  authelia:
    image: authelia/authelia
    container_name: authelia
    secrets:
      - jwt
      - duo
      - session
      - redis
      - mysql
      - smtp
      - ldap
    volumes:
      - /path/to/authelia:/config
    networks:
      - net
    expose:
      - 9091
    restart: unless-stopped
    environment:
      - AUTHELIA_JWT_SECRET_FILE=/run/secrets/jwt
      - AUTHELIA_DUO_API_SECRET_KEY_FILE=/run/secrets/duo
      - AUTHELIA_SESSION_SECRET_FILE=/run/secrets/session
      - AUTHELIA_SESSION_REDIS_PASSWORD_FILE=/run/secrets/redis
      - AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE=/run/secrets/mysql
      - AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE=/run/secrets/smtp
      - AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE=/run/secrets/ldap
      - TZ=Australia/Melbourne
```

### Compose with bind mounted secret files

This example assumes secrets are stored in `/path/to/authelia/secrets/{secretname}`
on the host and are exposed with bind mounted secret files in a `docker-compose.yml` file
at `/config/secrets/`:

```yaml
version: '3.8'

networks:
  net:
    driver: bridge

services:
  authelia:
    image: authelia/authelia
    container_name: authelia
    volumes:
      - /path/to/authelia:/config
    networks:
      - net
    expose:
      - 9091
    restart: unless-stopped
    environment:
      - AUTHELIA_JWT_SECRET_FILE=/config/secrets/jwt
      - AUTHELIA_DUO_API_SECRET_KEY_FILE=/config/secrets/duo
      - AUTHELIA_SESSION_SECRET_FILE=/config/secrets/session
      - AUTHELIA_SESSION_REDIS_PASSWORD_FILE=/config/secrets/redis
      - AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE=/config/secrets/mysql
      - AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE=/config/secrets/smtp
      - AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE=/config/secrets/ldap
      - TZ=Australia/Melbourne
```

## Kubernetes

Secrets can be mounted as files using the following sample manifests.

To create a secret, the following manifest can be used

```yaml
---
kind: Secret
apiVersion: v1

metadata:
  name: a-nice-name
  namespace: your-authelia-namespace

data:
  duo_key: >-
    UXE1WmM4S0pldnl6eHRwQ3psTGpDbFplOXFueUVyWEZhYjE0Z01IRHN0RT0K

  jwt_secret: >-
    anotherBase64EncodedSecret

...
```

where `UXE1WmM4S0pldnl6eHRwQ3psTGpDbFplOXFueUVyWEZhYjE0Z01IRHN0RT0K` is Base64 encoded for
`Qq5Zc8KJevyzxtpCzlLjClZe9qnyErXFab14gMHDstE`, the actual content of the secret. You can generate these contents with

```console
LENGTH=64
tr -cd '[:alnum:]' < /dev/urandom \
  | fold -w "${LENGTH}"           \
  | head -n 1                     \
  | tr -d '\n'                    \
  | tee actualSecretContent.txt   \
  | base64 --wrap 0               \
  ; echo
```

which writes the secret's content to the `actualSecretContent.txt` file and print the Base64 encoded version on `stdout`. `${LENGTH}` is the length in characters of the secret content generated by this pipe. If you don't want the contents to be written to `actualSecretContent.txt`, just delete the line with the `tee` command.

### Kustomization

* __Filename:__ ./kustomization.yaml
* __Command:__ kubectl apply -k
* __Notes:__ this kustomization expects the Authelia configuration.yml in the same directory. You will need to edit the
  kustomization.yaml with your desired secrets after the equal signs. If you change the value before the equal sign
  you'll have to adjust the volumes section of the daemonset template (or deployment template if you're using it).

```yaml
#filename: ./kustomization.yaml
generatorOptions:
  disableNameSuffixHash: true
  labels:
    type: generated
    app: authelia
configMapGenerator:
  - name: authelia
    files:
      - configuration.yml
secretGenerator:
  - name: authelia
    literals:
      - jwt_secret=myverysecuresecret
      - session_secret=mysessionsecret
      - redis_password=myredispassword
      - sql_password=mysqlpassword
      - ldap_password=myldappassword
      - duo_secret=myduosecretkey
      - smtp_password=mysmtppassword
```

### DaemonSet

* __Filename:__ ./daemonset.yaml
* __Command:__ kubectl apply -f ./daemonset.yaml
* __Notes:__ assumes Kubernetes API 1.16 or greater

```yaml
#filename: daemonset.yaml
#command: kubectl apply -f daemonset.yaml
#notes: assumes kubernetes api 1.16+
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: authelia
  namespace: authelia
  labels:
    app: authelia
spec:
  selector:
    matchLabels:
      app: authelia
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: authelia
    spec:
      containers:
        - name: authelia
          image: authelia/authelia:latest
          imagePullPolicy: IfNotPresent
          env:
            - name: AUTHELIA_JWT_SECRET_FILE
              value: /app/secrets/jwt
            - name: AUTHELIA_DUO_API_SECRET_KEY_FILE
              value: /app/secrets/duo
            - name: AUTHELIA_SESSION_SECRET_FILE
              value: /app/secrets/session
            - name: AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE
              value: /app/secrets/ldap_password
            - name: AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE
              value: /app/secrets/smtp_password
            - name: AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE
              value: /app/secrets/sql_password
            - name: AUTHELIA_SESSION_REDIS_PASSWORD_FILE
              value: /app/secrets/redis_password
            - name: TZ
              value: America/Toronto
          ports:
            - name: authelia-port
              containerPort: 9091
          startupProbe:
            httpGet:
              path: /api/state
              port: authelia-port
            initialDelaySeconds: 15
            timeoutSeconds: 5
            periodSeconds: 5
            failureThreshold: 4
          livenessProbe:
            httpGet:
              path: /api/state
              port: authelia-port
            initialDelaySeconds: 60
            timeoutSeconds: 5
            periodSeconds: 30
            failureThreshold: 2
          readinessProbe:
            httpGet:
              path: /api/state
              port: authelia-port
            initialDelaySeconds: 15
            timeoutSeconds: 5
            periodSeconds: 5
            failureThreshold: 5
          volumeMounts:
            - mountPath: /config
              name: config-volume
            - mountPath: /app/secrets
              name: secrets
              readOnly: true
      volumes:
        - name: config-volume
          configMap:
            name: authelia
            items:
              - key: configuration.yml
                path: configuration.yml
        - name: secrets
          secret:
            secretName: authelia
            items:
              - key: jwt_secret
                path: jwt
              - key: duo_secret
                path: duo
              - key: session_secret
                path: session
              - key: redis_password
                path: redis_password
              - key: sql_password
                path: sql_password
              - key: ldap_password
                path: ldap_password
              - key: smtp_password
                path: smtp_password
```
