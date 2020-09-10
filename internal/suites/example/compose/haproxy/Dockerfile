FROM haproxy:2.2.3-alpine

RUN \
apk add --no-cache \
  curl \
  lua-json4 \
  openssl && \
openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 -subj "/C=AU/ST=Victoria/L=Melbourne/O=Authelia/CN=*.example.com" -keyout haproxy.key -out haproxy.crt && \
cat haproxy.key haproxy.crt > /usr/local/etc/haproxy/haproxy.pem