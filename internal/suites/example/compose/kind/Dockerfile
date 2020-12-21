FROM alpine:3.12.3

WORKDIR /kind

RUN apk add --no-cache bash curl docker && \
    curl -Lo kind https://github.com/clems4ever/kind/releases/download/docker-network/kind && chmod +x kind && \
    curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.13.0/bin/linux/amd64/kubectl && chmod +x kubectl

ADD entrypoint.sh entrypoint.sh
ADD patch-kubeconfig.sh patch-kubeconfig.sh

ENV HOME=/kind/config
ENV KUBECONFIG=/kind/config/.kube/kind-config-kind

VOLUME /kind/config

ENTRYPOINT ["./entrypoint.sh"]