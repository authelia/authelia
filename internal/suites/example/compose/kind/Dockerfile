FROM alpine:3.16.3

WORKDIR /kind

RUN apk add --no-cache bash curl docker && \
    curl -Lo kind https://github.com/kubernetes-sigs/kind/releases/download/v0.10.0/kind-linux-amd64 && chmod +x kind && \
    curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.20.2/bin/linux/amd64/kubectl && chmod +x kubectl

ADD entrypoint.sh entrypoint.sh
ADD patch-kubeconfig.sh patch-kubeconfig.sh

ENV HOME=/kind/config
ENV KUBECONFIG=/kind/config/.kube/kind-config-kind

VOLUME /kind/config

ENTRYPOINT ["./entrypoint.sh"]