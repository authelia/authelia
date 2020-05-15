---
layout: default
title: Secrets
parent: Configuration
nav_order: 6
---

# Secrets

Configuration of Authelia requires some secrets and passwords.
Even if they can be set in the configuration file, the recommended
way to set secrets is to use environment variables as described
below.

## Environment variables

A secret can be configured using an environment variable with the
prefix AUTHELIA_ followed by the path of the option capitalized
and with dots replaced by underscores followed by the suffix _FILE.

The contents of the environment variable must be a path to a file
containing the secret data. This file must be readable by the
user the Authelia daemon is running as.

For instance the LDAP password can be defined in the configuration
at the path **authentication_backend.ldap.password**, so this password 
could alternatively be set using the environment variable called
**AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE**.

Here is the list of the environment variables which are considered
secrets and can be defined. Any other option defined using an
environment variable will not be replaced.

|Configuration Key                   |Environment Variable                              |
|:----------------------------------:|:------------------------------------------------:|
|jwt_secret                          |AUTHELIA_JWT_SECRET_FILE                          |
|duo_api.secret_key                  |AUTHELIA_DUO_API_SECRET_KEY_FILE                  |
|session.secret                      |AUTHELIA_SESSION_SECRET_FILE                      |
|session.redis.password              |AUTHELIA_SESSION_REDIS_PASSWORD_FILE              |
|storage.mysql.password              |AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE              |
|storage.postgres.password           |AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE           |
|notifier.smtp.password              |AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE              |
|authentication_backend.ldap.password|AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE|

## Secrets in configuration file

If for some reason you prefer keeping the secrets in the configuration
file, be sure to apply the right permissions to the file in order to
prevent secret leaks if an another application gets compromised on your
server. The UNIX permissions should probably be something like 600.

## Secrets exposed in an environment variable

**DEPRECATION NOTICE:** This backwards compatibility feature **has been removed** in 4.18.0+. 

Prior to implementing file secrets you were able to define the
values of secrets in the environment variables themselves
in plain text instead of referencing a file. **This is no longer available
as an option**, please see the table above for the file based replacements. See 
[this article](https://diogomonica.com/2017/03/27/why-you-shouldnt-use-env-variables-for-secret-data/)
for reasons why this was removed.

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
      - /path/to/authelia:/var/lib/authelia
      - /path/to/authelia/configuration.yml:/etc/authelia/configuration.yml:ro
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
at `/etc/authelia/secrets/`:

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
      - /path/to/authelia:/var/lib/authelia
      - /path/to/authelia/configuration.yml:/etc/authelia/configuration.yml:ro
      - /path/to/authelia/secrets:/etc/authelia/secrets
    networks:
      - net
    expose:
      - 9091
    restart: unless-stopped
    environment:
      - AUTHELIA_JWT_SECRET_FILE=/etc/authelia/secrets/jwt
      - AUTHELIA_DUO_API_SECRET_KEY_FILE=/etc/authelia/secrets/duo
      - AUTHELIA_SESSION_SECRET_FILE=/etc/authelia/secrets/session
      - AUTHELIA_SESSION_REDIS_PASSWORD_FILE=/etc/authelia/secrets/redis
      - AUTHELIA_STORAGE_MYSQL_PASSWORD_FILE=/etc/authelia/secrets/mysql
      - AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE=/etc/authelia/secrets/smtp
      - AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE=/etc/authelia/secrets/ldap
      - TZ=Australia/Melbourne
```


## Kubernetes

Secrets can be mounted as files using the following sample manifests.


### Kustomization

- **Filename:** ./kustomization.yaml
- **Command:** kubectl apply -k
- **Notes:** this kustomization expects the Authelia configuration.yml in
the same directory. You will need to edit the kustomization.yaml with your
desired secrets after the equal signs. If you change the value before the
equal sign you'll have to adjust the volumes section of the daemonset
template (or deployment template if you're using it).
 
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

- **Filename:** ./daemonset.yaml
- **Command:** kubectl apply -f ./daemonset.yaml
- **Notes:** assumes Kubernetes API 1.16 or greater
```yaml
#filename: daemonset.yaml
#command: kubectl apply -f daemonset.yaml
#notes: assumes kubernetes api 1.16+
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: authelia
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
              value: /usr/app/secrets/jwt
            - name: AUTHELIA_DUO_API_SECRET_KEY_FILE
              value: /usr/app/secrets/duo
            - name: AUTHELIA_SESSION_SECRET_FILE
              value: /usr/app/secrets/session
            - name: AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD_FILE
              value: /usr/app/secrets/ldap_password
            - name: AUTHELIA_NOTIFIER_SMTP_PASSWORD_FILE
              value: /usr/app/secrets/smtp_password
            - name: AUTHELIA_STORAGE_POSTGRES_PASSWORD_FILE
              value: /usr/app/secrets/sql_password
          ports:
            - name: http
              containerPort: 80
          startupProbe:
            httpGet:
              path: /api/configuration
              port: http
            initialDelaySeconds: 10
            timeoutSeconds: 5
            periodSeconds: 5
            failureThreshold: 4
          livenessProbe:
            httpGet:
              path: /api/configuration
              port: http
            initialDelaySeconds: 60
            timeoutSeconds: 5
            periodSeconds: 30
            failureThreshold: 2
          readinessProbe:
            httpGet:
              path: /api/configuration
              port: http
            initialDelaySeconds: 10
            timeoutSeconds: 5
            periodSeconds: 5
            failureThreshold: 5
          volumeMounts:
            - mountPath: /etc/authelia
              name: config-volume
            - mountPath: /usr/app/secrets
              name: secrets
              readOnly: true
            - mountPath: /etc/localtime
              name: localtime
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
        - name: localtime
          hostPath:
            path: /etc/localtime
```