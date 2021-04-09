---
layout: default
title: Example of authelia lite on docker swarm
parent: Community
nav_order: 3
---

The overlay network for docker swarm can be initialized with:

```
$ docker swarm init
$ docker swarm init && docker network create --driver=overlay traefik-public
$ mkdir ./redis ./letsencrypt
```

The structure of the folder should be like this:

```
├── authelia/
│   ├── configuration.yml
│   └── users_database.yml
├── redis/
├── letsencrypt/
│   └── acme.json
└── traefik-compose.yml
```

The following configuration allows you to deploy authelia to docker swarm with traefik 2.x. Please replace the **example.com** and **your@email.com** with your domain and email respectively.  Please save it as **traefik-compose.yml**.

```
version: '3.3'

services:
  authelia:
    image: authelia/authelia:4
    volumes:
      - ./authelia:/config
    networks:
      - traefik-public
    deploy:
      labels:
        - 'traefik.enable=true'
        - 'traefik.http.routers.authelia.rule=Host(`auth.example.com`)'
        - 'traefik.http.routers.authelia.entrypoints=web'
        - "traefik.http.services.authelia.loadbalancer.server.port=9091"
        # TLS
        - "traefik.http.routers.authelias.rule=Host(`auth.example.com`)"
        - "traefik.http.routers.authelias.entrypoints=websecure"
        - "traefik.http.routers.authelias.tls.certresolver=letsencrypt"
        # Redirect
        - "traefik.http.routers.authelia.middlewares=https_redirect"
        - "traefik.http.middlewares.https_redirect.redirectscheme.scheme=https"
        # Authelia
        - 'traefik.http.middlewares.authelia.forwardauth.address=http://authelia:9091/api/verify?rd=https://auth.example.com'
        - 'traefik.http.middlewares.authelia.forwardauth.trustForwardHeader=true'
        - 'traefik.http.middlewares.authelia.forwardauth.authResponseHeaders=Remote-User, Remote-Groups'
        - "traefik.http.routers.authelia.service=authelia"

  redis:
    image: redis:6-alpine
    volumes:
      - ./redis:/data
    networks:
      - traefik-public

  traefik:
    # The official v2.0 Traefik docker image
    image: traefik:v2.2
    deploy:
      labels:
        - 'traefik.enable=true'
        - 'traefik.http.routers.api.rule=Host(`traefik.example.com`)'
        - 'traefik.http.routers.api.entrypoints=web'
        - 'traefik.http.routers.api.service=api@internal'
        - 'traefik.http.services.traefik.loadbalancer.server.port=80'
        # TLS
        - "traefik.http.routers.apis.rule=Host(`traefik.example.com`)"
        - "traefik.http.routers.apis.entrypoints=websecure"
        - "traefik.http.routers.apis.tls.certresolver=letsencrypt"
        # Redirect
        - "traefik.http.routers.api.middlewares=https_redirect"
        - "traefik.http.middlewares.https_redirect.redirectscheme.scheme=https"
        # Authelia
        - 'traefik.http.routers.apis.service=api@internal'
        - 'traefik.http.routers.apis.middlewares=authelia@docker'
    command: 
      - "--api"
      - "--providers.docker=true"
      - "--providers.docker.swarmMode=true"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:80"
      - "--entryPoints.websecure.address=:443"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.letsencrypt.acme.email=your@email.com"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
    ports:
      # Listen on port 80, default for HTTP, necessary to redirect to HTTPS
      - target: 80
        published: 80
        mode: host
      # Listen on port 443, default for HTTPS
      - target: 443
        published: 443
        mode: host
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./letsencrypt:/letsencrypt
    networks:
      - traefik-public

  secure:
    image: containous/whoami
    networks:
      - traefik-public
    deploy:
      labels:
        - 'traefik.enable=true'
        - 'traefik.http.routers.secure.rule=Host(`secure.example.com`)'
        - 'traefik.http.routers.secure.entrypoints=web'
        - 'traefik.http.services.secure.loadbalancer.server.port=80'
        # TLS
        - "traefik.http.routers.secures.rule=Host(`secure.example.com`)"
        - "traefik.http.routers.secures.entrypoints=websecure"
        - "traefik.http.routers.secures.tls.certresolver=letsencrypt"
        # Redirect
        - "traefik.http.routers.secure.middlewares=https_redirect"
        - "traefik.http.middlewares.https_redirect.redirectscheme.scheme=https"
        # Authelia
        - 'traefik.http.routers.secures.middlewares=authelia@docker'

  public:
    image: containous/whoami
    networks:
      - traefik-public
    deploy:
      labels:
        - 'traefik.enable=true'
        - 'traefik.http.routers.public.rule=Host(`public.example.com`)'
        - 'traefik.http.routers.public.entrypoints=web'
        - 'traefik.http.services.public.loadbalancer.server.port=80'
        # TLS
        - "traefik.http.routers.publics.rule=Host(`public.example.com`)"
        - "traefik.http.routers.publics.entrypoints=websecure"
        - "traefik.http.routers.publics.tls.certresolver=letsencrypt"
        # Redirect
        - "traefik.http.routers.public.middlewares=https_redirect"
        - "traefik.http.middlewares.https_redirect.redirectscheme.scheme=https"
        # Authelia
        - 'traefik.http.routers.publics.middlewares=authelia@docker'

networks:
  traefik-public:
    external: true
```

Finally, the stack is ready to be deployed.

```
$ docker stack deploy -c traefik-compose.yml traefik
```

