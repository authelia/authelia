FROM golang:1.26.2-alpine AS go-builder

WORKDIR /src

COPY --link go.mod go.sum ./
RUN go mod download

COPY --link cmd/authelia-pam/ cmd/authelia-pam/
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /authelia-pam ./cmd/authelia-pam/

FROM alpine:latest AS c-builder

RUN apk --no-cache add gcc musl-dev linux-pam-dev make

COPY --link extras/pam/src/ /pam-src/src/
COPY --link extras/pam/Makefile /pam-src/Makefile
RUN make -C /pam-src

FROM alpine:latest

RUN apk --no-cache add linux-pam openssh-server-pam ca-certificates && \
    mkdir -p /var/run/sshd && \
    ssh-keygen -A && \
    adduser -D -s /bin/sh john && \
    echo "john:password" | chpasswd && \
    sed -i 's/^#*KbdInteractiveAuthentication.*/KbdInteractiveAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/^#*UsePAM.*/UsePAM yes/' /etc/ssh/sshd_config && \
    sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config && \
    echo "AuthenticationMethods keyboard-interactive" >> /etc/ssh/sshd_config

COPY --from=go-builder /authelia-pam /usr/local/bin/authelia-pam
COPY --from=c-builder /pam-src/pam_authelia.so /tmp/pam_authelia.so
RUN PAM_DIR="$(dirname "$(find /lib /usr/lib -name 'pam_unix.so' -print -quit 2>/dev/null)")" && \
    cp /tmp/pam_authelia.so "$PAM_DIR/pam_authelia.so" && \
    rm /tmp/pam_authelia.so
COPY --chmod=755 internal/suites/example/compose/pam/entrypoint.sh /entrypoint.sh

EXPOSE 22

ENTRYPOINT ["/entrypoint.sh"]
