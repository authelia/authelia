FROM alpine:3.9.4

WORKDIR /usr/app

RUN apk --no-cache add ca-certificates tzdata wget

# Install the libc required by the password hashing compiled with CGO.
RUN wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub
RUN wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.30-r0/glibc-2.30-r0.apk
RUN apk --no-cache add glibc-2.30-r0.apk

ADD dist/authelia authelia
ADD dist/public_html public_html

EXPOSE 9091

VOLUME /etc/authelia
VOLUME /var/lib/authelia

CMD ["./authelia", "-config", "/etc/authelia/config.yml"]
