Deploying **Authelia** using Docker is fairly simple. Pull [clems4ever/authelia](https://hub.docker.com/r/clems4ever/authelia/), edit the template configuration to fit your environment and run the container with the configuration mounted in /etc/authelia/config.yml, that's it!

If you don't have any environment yet, don't worry! There is all you need in the repository to deploy a new test environment from scratch on your development machine. Follow the steps.

## Requirements

1. **Docker** (>=17.03) and **docker-compose** (>=1.14) are installed on the machine.
2. Port **8080** and **8085** are available on the machine.

## Deployment

1. Add the following lines to your **/etc/hosts** to alias multiple sub-domains so that nginx can redirect request to the correct virtual host.

```
127.0.0.1       home.example.com
127.0.0.1       public.example.com
127.0.0.1       dev.example.com
127.0.0.1       admin.example.com
127.0.0.1       mx1.mail.example.com
127.0.0.1       mx2.mail.example.com
127.0.0.1       single_factor.example.com
127.0.0.1       login.example.com
```

2. Deploy the environment with docker-compose.

```
./scripts/example-dockerhub/deploy-example.sh
```

After few seconds the services should be running and you should be able to visit 
[https://home.example.com:8080/](https://home.example.com:8080/).

**Note:** When accessing the login page, a self-signed certificate exception should appear, 
it has to be trusted before you can get to the target page. The certificate
must also be trusted for each subdomain, therefore it is normal to see the exception
 several times.