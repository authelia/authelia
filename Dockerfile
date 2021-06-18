# =======================================
# ===== Build image for the backend =====
# =======================================
FROM golang:1.16.5-alpine AS builder-backend

WORKDIR /go/src/app

RUN \
echo ">> Downloading required apk's..." && \
apk --no-cache add gcc musl-dev

COPY go.mod go.sum ./

RUN \
echo ">> Downloading go modules..." && \
go mod download

COPY / ./

ARG LDFLAGS_EXTRA
# CGO_ENABLED=1 is required for building go-sqlite3
RUN \
mv public_html internal/server/public_html && \
echo ">> Starting go build..." && \
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -tags netgo \
-ldflags "-s -w -linkmode external ${LDFLAGS_EXTRA} -extldflags -static" -trimpath -o authelia ./cmd/authelia

# ===================================
# ===== Authelia official image =====
# ===================================
FROM alpine:3.14.0

WORKDIR /app

RUN apk --no-cache add ca-certificates su-exec tzdata

COPY --from=builder-backend /go/src/app/authelia /go/src/app/LICENSE /go/src/app/entrypoint.sh /go/src/app/healthcheck.sh ./

EXPOSE 9091

VOLUME /config

# Set environment variables
ENV PATH="/app:${PATH}" \
    PUID=0 \
    PGID=0

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["--config", "/config/configuration.yml"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=1m CMD /app/healthcheck.sh
