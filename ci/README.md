[logo]: https://github.com/clems4ever/authelia/raw/master/docs/images/authelia-title.png "Authelia"
[![alt text][logo]](https://github.com/AntilaX-3/)

# authelia/buildkite
[![Docker Pulls](https://img.shields.io/docker/pulls/authelia/buildkite.svg)](https://hub.docker.com/r/authelia/buildkite/) [![Docker Stars](https://img.shields.io/docker/stars/authelia/buildkite.svg)](https://hub.docker.com/r/authelia/buildkite/)

The [buildkite agent](https://buildkite.com/docs/agent/v3) is a small, reliable and cross-platform build runner that makes it easy to run automated builds on your own infrastructure. Its main responsibilities are polling buildkite.com for work, running build jobs, reporting back the status code and output log of the job, and uploading the job's artifacts.

This custom image is based on the `docker:dind` to provide docker-in-docker alongside Buildkite to support the automated integration cases run for Authelia's CI process.
The image will be re-built if any updates are made to the base `docker:dind` image.

This image shamelessly utilises some of the fine work by the team over at [LinuxServer.io](https://www.linuxserver.io/), credits to their [alpine baseimage](https://github.com/linuxserver/docker-baseimage-alpine/).
  
## Usage

Here are some example snippets to help you get started creating a container.

An example `docker-compose.yml` has also been provided in the repo which includes three nodes and a local registry cache.

### docker

```
docker create \
  --name=buildkite1 \
  -e BUILDKITE_AGENT_NAME=named-node-1 \
  -e BUILDKITE_AGENT_TOKEN=tokenhere \
  -e BUILDKITE_AGENT_TAGS=tags=here,moretags=here \
  -e BUILDKITE_AGENT_PRIORITY=priorityhere \
  -e PUID=1000 \
  -e PGID=1000 \
  -e TZ=Australia/Melbourne \
  -v <path to data>/docker:/buildkite/.docker \
  -v <path to data>/ssh:/buildkite/.ssh \
  -v <path to data>/go:/buildkite/.go \
  -v <path to data>/hooks:/buildkite/hooks \
  --restart unless-stopped \
  --privileged \
  authelia/buildkite
```
### docker-compose

Compatible with docker-compose v2 schemas.

```
---
version: "2.1"
services:
  buildkite1:
    image: authelia/buildkite
    container_name: buildkite1
    privileged: true
    volumes:
      - <path to data>/docker:/buildkite/.docker
      - <path to data>/ssh:/buildkite/.ssh
      - <path to data>/go:/buildkite/.go
      - <path to data>/hooks:/buildkite/hooks
    restart: unless-stopped
    environment:
      - BUILDKITE_AGENT_NAME=named-node-1
      - BUILDKITE_AGENT_TOKEN=tokenhere
      - BUILDKITE_AGENT_TAGS=tags=here,moretags=here
      - BUILDKITE_AGENT_PRIORITY=priorityhere
      - PUID=1000
      - PGID=1000
      - TZ=Australia/Melbourne
```
## Parameters

Container images are configured using parameters passed at runtime (such as those above). These parameters are separated by a colon and indicate `<external>:<internal>` respectively. For example, `-p 8080:80` would expose port `80` from inside the container to be accessible from the host's IP on port `8080` outside the container.

| Parameter | Function |
| :----: | --- |
| `-e BUILDKITE_AGENT_NAME=named-node-1` | [agent name](https://buildkite.com/docs/agent/v3/configuration) for buildkite agent on specified node |
| `-e BUILDKITE_AGENT_TOKEN=tokenhere` | [agent token](https://buildkite.com/docs/agent/v3/tokens) for specified pipeline |
| `-e BUILDKITE_AGENT_TAGS=tags=here,moretags=here` | [agent tags](https://buildkite.com/docs/agent/v3/cli-start#setting-tags) on specified node, tag=value comma separated |
| `-e BUILDKITE_AGENT_PRIORITY=1` | [agent priority](https://buildkite.com/docs/agent/v3/prioritization) |
| `-e PUID=1000` | for UserID - see below for explanation |
| `-e PGID=1000` | for GroupID - see below for explanation |
| `-e TZ=Australia/Melbourne` | for setting timezone information, eg Australia/Melbourne |
| `-v /buildkite/.docker` | Docker `config.json` stored here for permissions |
| `-v /buildkite/.ssh` | SSH `id_rsa` and `ida_rsa.pub` stored here for [GitHub cloning](https://buildkite.com/docs/agent/v3/ssh-keys) |
| `-v /buildkite/.go` | $GOPATH, set this location to share cache between multiple node containers |
| `-v /buildkite/hooks` | Used to provide secrets in to Buildkite such as `DOCKER_USERNAME` `DOCKER_PASSWORD` and `GITHUB_TOKEN` for publish and clean up steps |

## User / Group Identifiers

When using volumes (`-v` flags) permissions issues can arise between the host OS and the container, we avoid this issue by allowing you to specify the user `PUID` and group `PGID`.

Ensure any volume directories on the host are owned by the same user you specify and any permissions issues will vanish like magic.

In this instance `PUID=1000` and `PGID=1000`, to find yours use `id user` as below:

```
  $ id username
    uid=1000(dockeruser) gid=1000(dockergroup) groups=1000(dockergroup)
```

## Version
- **19/12/2019:** Initial release