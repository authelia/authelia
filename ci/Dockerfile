FROM docker:dind

# set labels
LABEL maintainer="Nightah"

# set application versions
ARG ARCH="amd64"
ARG BUILDKITE_VERSION="3.17.0"
ARG OVERLAY_VERSION="v1.22.1.0"

# environment variables
ENV PS1="$(whoami)@$(hostname):$(pwd)$ " \
HOME="/buildkite" \
TERM="xterm"

# set runtime variables
ENV BUILDKITE_AGENT_CONFIG=/buildkite/buildkite-agent.cfg \
GOPATH="/buildkite/.go"

# add local files
COPY root/ /

# modifications
RUN \
 echo "**** Install Authelia CI pre-requisites ****" && \
   echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
   echo "@edget http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories && \
   apk add --no-cache \
     bash \
     ca-certificates \
     coreutils \
     chromium \
     chromium-chromedriver \
     curl \
     docker-compose \
     git \
     hub@edget \
     go@edge \
     jq \
     libc6-compat \
     libstdc++ \
     nodejs \
     npm \
     openssh-client \
     perl \
     rsync \
     shadow \
     sudo \
     tzdata \
     yarn@edge && \
 echo "**** Add s6 overlay ****" && \
   cd /tmp && \
   curl -Lfs -o s6-overlay.tar.gz "https://github.com/just-containers/s6-overlay/releases/download/${OVERLAY_VERSION}/s6-overlay-${ARCH}.tar.gz" && \
   tar xfz s6-overlay.tar.gz -C / && \
 echo "**** Patch CVE-2019-5021 ****" && \
   sed -i -e 's/^root::/root:!:/' /etc/shadow && \
 echo "**** Create buildkite user and make our folders ****" && \
   useradd -u 911 -U -d /buildkite -s /bin/false buildkite && \
   usermod -G wheel buildkite && \
   sed -i 's/# %wheel/%wheel/g' /etc/sudoers && \
 echo "**** Install Buildkite ****" && \
   mkdir -p /buildkite/builds /buildkite/hooks /buildkite/plugins && \
   curl -Lfs -o /usr/local/bin/ssh-env-config.sh https://raw.githubusercontent.com/buildkite/docker-ssh-env-config/master/ssh-env-config.sh && \
   chmod +x /usr/local/bin/ssh-env-config.sh && \
   curl -Lfs -o buildkite-agent.tar.gz https://github.com/buildkite/agent/releases/download/v${BUILDKITE_VERSION}/buildkite-agent-linux-${ARCH}-${BUILDKITE_VERSION}.tar.gz && \
   tar xf buildkite-agent.tar.gz && \
   sed -i 's/token=/#token=/g' buildkite-agent.cfg && \
   sed -i 's/\$HOME\/.buildkite-agent/\/buildkite/g' buildkite-agent.cfg && \
   mv buildkite-agent.cfg /buildkite/buildkite-agent.cfg && \
   mv buildkite-agent /usr/local/bin/buildkite-agent && \
 echo "**** Cleanup ****" && \
   rm -rf /tmp/*

# ports and volumes
VOLUME /buildkite

ENTRYPOINT ["/init"]