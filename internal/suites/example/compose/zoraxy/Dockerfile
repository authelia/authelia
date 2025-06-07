FROM golang:1.24.2-bookworm@sha256:00eccd446e023d3cd9566c25a6e6a02b90db3e1e0bbe26a48fc29cd96e800901 AS builder-backend

RUN \
    apt-get update && \
    apt install git && \
    git clone https://github.com/james-d-elliott/zoraxy.git && \
    cd zoraxy && \
    git checkout origin/ci-forward-auth && \
    cd src && \
    GOROOT_FINAL=Git/ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /zoraxy

FROM zoraxydocker/zoraxy:2.3.0

COPY --chmod=755 --from=builder-backend /zoraxy /usr/local/bin/zoraxy
