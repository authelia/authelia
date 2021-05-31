# Overview

This example uses minikube in VirtualBox and deploys Traefik, Authelia and Traefik's whoami app. It creates two users, listed below. It uses the Authelia Custom Resource Definitions to configure access control.

* Authelia User - `authelia:authelia` (`users` group)
* Admin User - `admin:admin` (`admins` group)

# Install minikube and create a cluster

1. Install minikube
2. Create a cluster
  * `minikube start`

# Deploy Traefik

1. Install the Traefik Custom Resource Definitions
  * `kubectl apply -f examples/kubernetes/local/traefik/crds.yml`
2. Deploy Traefik
  * `kubectl apply -f examples/kubernetes/local/traefik/traefik.yml`
3. Access the service in a browser using the IP provided by minikube, over HTTPS. You may have to trust the self-signed certificate.
  * `minikube ip`
4. See that `404 page not found` is returned

# Deploy Authelia

1. Install the Authelia Custom Resource Definitions
  * `kubectl apply -f internal/kubernetes/custom-resource-definitions.yml`
2. Deploy Authelia
  * `kubectl apply -f examples/kubernetes/local/authelia/authelia.yml`
3. Create a host entry for `auth.example.org` to the IP given by minikube, over HTTPS. You may have to trust the self-signed certificate.
  * `minikube ip`
  * `nano /etc/hosts`
  * Example entry: `192.168.99.115 auth.example.org`
4. Navigate to `auth.example.org` and see that Authelia is available
5. Authenticate using any of the existing users
6. Logout

# Deploy an example app

1. Deploy the whoami example app
  * `kubectl apply -f examples/kubernetes/local/whoami/whoami.yml`
2. Create a host entry for `whoami.example.org` to the IP given by minikube, over HTTPS. You may have to trust the self-signed certificate.
  * `minikube ip`
  * `nano /etc/hosts`
  * Example entry: `192.168.99.115 whoami.example.org`
3. Navigate to `whoami.example.org` and see that Authelia is available
4. Authenticate as `authelia:authelia`

# Deploy an Access Control Rule

1. Deploy the example access control rule to only allow admins to access `whoami.example.org`
  * `kubectl apply -f examples/kubernetes/local/whoami/access-control.yml`
2. Navigate to `whoami.example.org` and see that the `authelia` user is no longer allowed
3. Logout at `auth.example.org` and try again with the `admin` user, which works

You may also try to remove a rule set and update it.

1. Remove the rule set
  * `kubectl delete accesscontrolrule only-allow-admins`
2. Navigate to `whoami.example.org` and see that no user is allowed
3. Deploy the rule set again
  * `kubectl apply -f examples/kubernetes/local/whoami/access-control.yml`
4. Update the rule set by changing the `access-control.yml` to allow the `authelia` user as well and deploy it
  * Add `- [group:users]` under `subjects`
  * `kubectl apply -f examples/kubernetes/local/whoami/access-control.yml`

# Deploy a local build of Authelia

1. Build the Authelia frontend
  * `cd web`
  * `npm install` or `yarn install`
  * `INLINE_RUNTIME_CHUNK=false npm run build` or `INLINE_RUNTIME_CHUNK=false yarn build`
2. Move or copy the `web/build` directory to the root project directory and change its name to `public_html`
  * `cp -r web/build public_html`
3. Copy the `api` directory to the `public_html` directory
  * `cp -r api public_html/api`
4. Build the Authelia image
  * `docker build . -t authelia/authelia:dev`
5. Send the image to minikube
  * `minikube image load authelia/authelia:dev`
6. You may now use `localhost/authelia/authelia:dev`

# Troubleshooting

If you're having issues with minikube and DNS, you may want to change the DNS provider used. To do so, run `kubectl -n kube-system edit configmap/coredns` to change the configuration of CoreDNS using vim. Then, remove the forward block and change it to your provider of choice. The `forward` line should look like follows: `forward: . 1.1.1.1`. Symptoms of this issue are pods being unable to access the internet, update packages etc.

To access the Kubernets API to perform requests on the API, proxy it using `kubectl` like so: `kubectl proxy kubernetes`.

If you are having issues that an image deployed to minikube doesn't seem to change, try to change the tag to `dev1`, `dev2` and so on. It doesn't seem to respect updated images, despite specifying the `Always` pull policy and replacing the image on minikube.

# Disclaimer

This is an _example_ not a production-ready deployment. For example, the TLS certificates are self-signed. Furthermore, the user database is not mounted into `/var/run/` nor stored as a secret instead of a `ConfigMap`. The deployment is not highly available nor stateful. Lastly, all of the secrets such as the JWT secret are stored in the config instead of in actual secrets.

The Traefik CRDs are taken from here: https://doc.traefik.io/traefik/reference/dynamic-configuration/kubernetes-crd/#definitions.

The Traefik configuration is based on the one deployed with mikrok8s: https://github.com/ubuntu/microk8s/blob/c9d10eaa89c014cd3a159f1829a65f5125ee754c/microk8s-resources/actions/traefik.yaml.
