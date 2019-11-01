# =======================================
# ===== Build image for the backend =====
# =======================================
FROM golang:1.13-alpine AS builder-backend

# gcc and musl-dev are required for building go-sqlite3
RUN apk --no-cache add gcc musl-dev

WORKDIR /go/src/app
COPY . .

# CGO_ENABLED=1 is mandatory for building go-sqlite3
RUN cd cmd/authelia && GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o authelia


# ========================================
# ===== Build image for the frontend =====
# ========================================
FROM node:11-alpine AS builder-frontend

WORKDIR /node/src/app
COPY client .

# Install the dependencies and build
RUN npm ci && npm run build

# ===================================
# ===== Authelia official image =====
# ===================================
FROM alpine:3.10.3

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /usr/app

COPY --from=builder-backend /go/src/app/cmd/authelia/authelia authelia
COPY --from=builder-frontend /node/src/app/build public_html

EXPOSE 9091

VOLUME /etc/authelia
VOLUME /var/lib/authelia

CMD ["./authelia", "-config", "/etc/authelia/config.yml"]
