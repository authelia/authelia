# =======================================
# ===== Build image for the backend =====
# =======================================
FROM golang:1.16.4-alpine AS builder-backend

ARG BUILD_TAG
ARG BUILD_COMMIT

WORKDIR /go/src/app

COPY / ./

# CGO_ENABLED=1 is required for building go-sqlite3
RUN \
apk --no-cache add gcc musl-dev && \
go mod download && \
mv public_html internal/server/public_html && \
echo "Write tag ${BUILD_TAG} and commit ${BUILD_COMMIT} in binary." && \
sed -i "s/__BUILD_TAG__/${BUILD_TAG}/" cmd/authelia/constants.go && \
sed -i "s/__BUILD_COMMIT__/${BUILD_COMMIT}/" cmd/authelia/constants.go && \
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -tags netgo -ldflags '-s -w -linkmode external -extldflags -static' -trimpath -o authelia ./cmd/authelia

# ===================================
# ===== Authelia official image =====
# ===================================
FROM alpine:3.13.5

WORKDIR /app

RUN apk --no-cache add ca-certificates su-exec tzdata

COPY --from=builder-backend /go/src/app/authelia ./
COPY entrypoint.sh healthcheck.sh /usr/local/bin/

EXPOSE 9091

VOLUME /config /plugins

# Set environment variables
ENV PATH="/app:${PATH}" \
PUID=0 \
PGID=0

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["--config", "/config/configuration.yml"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=1m CMD /usr/local/bin/healthcheck.sh
