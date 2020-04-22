---
layout: default
title: Secrets
parent: Configuration
nav_order: 8
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


## Secrets exposed in an environment variable

Prior to implementing file secrets you were able to define the
values of secrets in the environment variables themselves
in plain text instead of referencing a file. This is still
supported but discouraged. If you still want to do this
just remove _FILE from the environment variable name
and define the value in insecure plain text. See 
[this article](https://diogomonica.com/2017/03/27/why-you-shouldnt-use-env-variables-for-secret-data/)
for reasons why this is considered insecure and is discouraged.

**DEPRECATION NOTICE:** This backwards compatibility feature will be
**removed** in 4.18.0+. 


## Secrets in configuration file

If for some reason you prefer keeping the secrets in the configuration
file, be sure to apply the right permissions to the file in order to
prevent secret leaks if an another application gets compromised on your
server. The UNIX permissions should probably be something like 600.


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