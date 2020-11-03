# ========================================
# ===== Build image for the frontend =====
# ========================================
FROM node:15-alpine AS builder-frontend

WORKDIR /node/src/app
COPY web .

# Install the dependencies and build
RUN yarn install --frozen-lockfile && INLINE_RUNTIME_CHUNK=false yarn build

# =======================================
# ===== Build image for the backend =====
# =======================================
FROM golang:1.15.3-alpine AS builder-backend

ARG BUILD_TAG
ARG BUILD_COMMIT

# gcc and musl-dev are required for building go-sqlite3
RUN apk --no-cache add gcc musl-dev

WORKDIR /go/src/app

COPY go.mod go.sum config.template.yml ./
COPY --from=builder-frontend /node/src/app/build public_html

RUN go mod download

COPY cmd cmd
COPY internal internal

# Prepare static files to be embedded in Go binary
RUN go get -u aletheia.icu/broccoli && \
cd internal/configuration && \
go generate . && \
cd ../server && \
go generate .

# Set the build version and time
RUN echo "Write tag ${BUILD_TAG} and commit ${BUILD_COMMIT} in binary." && \
    sed -i "s/__BUILD_TAG__/${BUILD_TAG}/" cmd/authelia/constants.go && \
    sed -i "s/__BUILD_COMMIT__/${BUILD_COMMIT}/" cmd/authelia/constants.go

# CGO_ENABLED=1 is mandatory for building go-sqlite3
RUN cd cmd/authelia && \
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -tags netgo -ldflags '-s -w -linkmode external -extldflags -static' -trimpath -o authelia

# ===================================
# ===== Authelia official image =====
# ===================================
FROM alpine:3.12.1

WORKDIR /app

RUN apk --no-cache add ca-certificates su-exec tzdata

COPY --from=builder-backend /go/src/app/cmd/authelia/authelia ./
COPY ./entrypoint.sh /usr/local/bin/entrypoint.sh

EXPOSE 9091

VOLUME /config

# Set environment variables
ENV PATH="/app:${PATH}" \
PUID=0 \
PGID=0

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["--config", "/config/configuration.yml"]
HEALTHCHECK --interval=30s --timeout=3s CMD wget --quiet --tries=1 --spider http://localhost:9091/api/state || exit 1
