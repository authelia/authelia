This tutorial is derived from the example provided in the repository under `example/kube`.

## Requirements

* Kubernetes cluster is set up.
* LDAP server is set up.
* Redis cluster is set up.
* Mongo database is set up.
* SMTP server is set up.

## Getting started

1. Install ingress-nginx

Install an [ingress-nginx](https://github.com/kubernetes/ingress-nginx) ingress controller in your cluster with the following kube configuration.

```
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: nginx-ingress-controller
  namespace: authelia
  labels:
    k8s-app: nginx-ingress-controller
spec:
  replicas: 1
  revisionHistoryLimit: 0
  template:
    metadata:
      labels:
        k8s-app: nginx-ingress-controller
        name: nginx-ingress-controller
      annotations:
        prometheus.io/port: '10254'
        prometheus.io/scrape: 'true'
    spec:
      terminationGracePeriodSeconds: 60
      containers:
      - image: quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.13.0
        name: nginx-ingress-controller
        imagePullPolicy: Always
        ports:
        - containerPort: 80
        - containerPort: 443
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        args:
        - /nginx-ingress-controller
        - --ingress-class=nginx
        - --election-id=ingress-controller-leader-external
        - --default-backend-service=$(POD_NAMESPACE)/default-http-backend
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-ingress-controller-service
  namespace: authelia
  labels:
    k8s-app: nginx-ingress-controller
spec:
  selector:
    k8s-app: nginx-ingress-controller
  ports:
    - port: 80
      name: http
    - port: 443
      name: https
  externalIPs:
    - 192.168.39.26  # <------- Replace this IP with your public IP or use a LoadBalancer service type
---
# Below is the definition of the default backend for requests with unknown routes.
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: default-http-backend
  labels:
    app: default-http-backend
  namespace: authelia
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: default-http-backend
    spec:
      terminationGracePeriodSeconds: 60
      containers:
      - name: default-http-backend
        image: gcr.io/google_containers/defaultbackend:1.4
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: default-http-backend
  namespace: authelia
  labels:
    app: default-http-backend
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: default-http-backend
```

2. Add Authelia's configuration as ConfigMap in your cluster.
For that, create a file called config.yml with the following content and update the configuration with your own parameters (such as you own LDAP, redis and mongo service names, secrets, etc...)
```
###############################################################
#                   Authelia configuration                    #
###############################################################

# The port to listen on
port: 80

# Log level
#
# Level of verbosity for logs
logs_level: debug

# Default redirection URL
#
# If user tries to authenticate without any referer, Authelia
# does not know where to redirect the user to at the end of the
# authentication process.
# This parameter allows you to specify the default redirection
# URL Authelia will use in such a case.
#
# Note: this parameter is optional. If not provided, user won't
# be redirected upon successful authentication.
default_redirection_url: https://login.example.com

# LDAP configuration
#
# Example: for user john, the DN will be cn=john,ou=users,dc=example,dc=com
ldap:
  # The url of the ldap server
  url: ldap://ldap-service

  # The base dn for every entries
  base_dn: dc=example,dc=com

  # An additional dn to define the scope to all users
  additional_users_dn: ou=users

  # The users filter used to find the user DN
  # {0} is a matcher replaced by username.
  # 'cn={0}' by default.
  users_filter: cn={0}

  # An additional dn to define the scope of groups
  additional_groups_dn: ou=groups

  # The groups filter used for retrieving groups of a given user.
  # {0} is a matcher replaced by username.
  # {dn} is a matcher replaced by user DN.
  # 'member={dn}' by default.
  groups_filter: (&(member={dn})(objectclass=groupOfNames))

  # The attribute holding the name of the group
  group_name_attribute: cn

  # The attribute holding the mail address of the user
  mail_attribute: mail

  # The username and password of the admin user.
  user: cn=admin,dc=example,dc=com
  password: password


# Authentication methods
#
# Authentication methods can be defined per subdomain.
# There are currently two available methods: "single_factor" and "two_factor"
#
# Note: by default a domain uses "two_factor" method.
#
# Note: 'per_subdomain_methods' is a dictionary where keys must be subdomains and
# values must be one of the two possible methods.
#
# Note: 'per_subdomain_methods' is optional.
#
# Note: authentication_methods is optional. If it is not set all sub-domains
# are protected by two factors.
authentication_methods:
  default_method: two_factor
#  per_subdomain_methods:
#    single_factor.example.com: single_factor

# Access Control
#
# Access control is a set of rules you can use to restrict user access to certain 
# resources.
# Any (apply to anyone), per-user or per-group rules can be defined.
#
# If 'access_control' is not defined, ACL rules are disabled and the `allow` default 
# policy is applied, i.e., access is allowed to anyone. Otherwise restrictions follow 
# the rules defined.
# 
# Note: One can use the wildcard * to match any subdomain. 
# It must stand at the beginning of the pattern. (example: *.mydomain.com)
# 
# Note: You must put the pattern in simple quotes when using the wildcard for the YAML
# to be syntaxically correct.
#
# Definition: A `rule` is an object with the following keys: `domain`, `policy` 
# and `resources`.
# - `domain` defines which domain or set of domains the rule applies to.
# - `policy` is the policy to apply to resources. It must be either `allow` or `deny`.
# - `resources` is a list of regular expressions that matches a set of resources to 
# apply the policy to.
#
# Note: Rules follow an order of priority defined as follows:
# In each category (`any`, `groups`, `users`), the latest rules have the highest 
# priority. In other words, it means that if a given resource matches two rules in the
# same category, the latest one overrides the first one.
# Each category has also its own priority. That is, `users` has the highest priority, then
# `groups` and `any` has the lowest priority. It means if two rules in different categories
# match a given resource, the one in the category with the highest priority overrides the
# other one.
#
access_control:
  # Default policy can either be `allow` or `deny`.
  # It is the policy applied to any resource if it has not been overriden
  # in the `any`, `groups` or `users` category.
  default_policy: deny

  # The rules that apply to anyone.
  # The value is a list of rules.
  any:
    - domain: '*.example.com'
      policy: allow
  
  # Group-based rules. The key is a group name and the value
  # is a list of rules.
  groups: {}
  
  # User-based rules. The key is a user name and the value
  # is a list of rules.
  users: {}


# Configuration of session cookies
# 
# The session cookies identify the user once logged in.
session:
  # The secret to encrypt the session cookie.
  secret: unsecure_password
  
  # The time in ms before the cookie expires and session is reset.
  expiration: 3600000 # 1 hour

  # The inactivity time in ms before the session is reset.
  inactivity: 300000 # 5 minutes

  # The domain to protect.
  # Note: the authenticator must also be in that domain. If empty, the cookie
  # is restricted to the subdomain of the issuer. 
  domain: example.com
  
  # The redis connection details
  redis:
    host: redis-service
    port: 6379

# Configuration of the authentication regulation mechanism.
#
# This mechanism prevents attackers from brute forcing the first factor.
# It bans the user if too many attempts are done in a short period of
# time.
regulation:
  # The number of failed login attempts before user is banned. 
  # Set it to 0 for disabling regulation.
  max_retries: 3

  # The length of time between login attempts before user is banned.
  find_time: 120

  # The length of time before a banned user can login again.
  ban_time: 300

# Configuration of the storage backend used to store data and secrets.
#
# You must use only an available configuration: local, mongo
storage:
  # The directory where the DB files will be saved
  # local: /var/lib/authelia/store
  
  # Settings to connect to mongo server
  mongo:
    url: mongodb://mongo-service
    database: authelia

# Configuration of the notification system.
#
# Notifications are sent to users when they require a password reset, a u2f
# registration or a TOTP registration.
# Use only an available configuration: filesystem, gmail
notifier:
  # For testing purpose, notifications can be sent in a file
  # filesystem:
  #   filename: /tmp/authelia/notification.txt

  # Use your email account to send the notifications. You can use an app password.
  # List of valid services can be found here: https://nodemailer.com/smtp/well-known/  
  # email:
  #   username: authelia@gmail.com
  #   password: password
  #   sender: authelia@example.com
  #   service: gmail
  
  # Use a SMTP server for sending notifications
  smtp:
    username: test
    password: password
    secure: false
    host: 'mailcatcher-service'
    port: 1025
    sender: admin@example.com
```
Once the file is created use the following command to upload it to your cluster:

`kubectl create configmap authelia-config --namespace=authelia --from-file=config.yml`

3. Install Authelia in your cluster with the following kube configuration.
```
---
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: authelia
  namespace: authelia
  labels:
    app: authelia
spec:
  replicas: 1
  selector:
    matchLabels:
      app: authelia
  template:
    metadata:
      labels:
        app: authelia
    spec:
      containers:
      - name: authelia
        image: clems4ever/authelia:latest # <----- You should use an explicit version here
        imagePullPolicy: Always
        ports:
        - containerPort: 80
        volumeMounts:
        - name: config-volume
          mountPath: /etc/authelia
      volumes:
      - name: config-volume
        configMap:
          name: authelia-config
          items:
          - key: config.yml
            path: config.yml
---
apiVersion: v1
kind: Service
metadata:
  name: authelia-service
  namespace: authelia
spec:
  selector:
    app: authelia
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: authelia-ingress
  namespace: authelia
  annotations:
    kubernetes.io/ingress.class: "nginx"
spec:
  tls:
  - secretName: authelia-tls
    hosts:
    - login.example.com # This is the host to reach Authelia login page
  rules:
  rules:
  - host: login.example.com # This is the host to reach Authelia login page
    http:
      paths:
      - path: /
        backend:
          serviceName: authelia-service
          servicePort: 80
```

4. Modify the ingress of your app for nginx to forward authentication requests to Authelia.

```
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: secure-ingress
  namespace: authelia
  annotations:
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/auth-url: "http://authelia-service.authelia.svc.cluster.local/api/verify"
    nginx.ingress.kubernetes.io/auth-signin: "https://login.example.com" # The url the user will be redirected if she is not authenticated
spec:
  tls: # Your app must be served over HTTPS for U2F to work.
  - secretName: app-tls
    hosts:
    - myapp.example.com
  rules:
  - host: myapp.example.com
    http:
      paths:
      - path: /
        backend:
          serviceName: app-service # Your application service
          servicePort: 80 # The port of your application service
```