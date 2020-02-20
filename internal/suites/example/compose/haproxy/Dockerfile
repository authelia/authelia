FROM haproxy:2.1-alpine

RUN \
apk add --no-cache \
  curl \
  lua5.3-socket \
  openssl && \
curl -Lfs -o /usr/local/etc/haproxy/auth-request.lua "https://raw.githubusercontent.com/TimWolla/haproxy-auth-request/master/auth-request.lua" && \
sed -i 's/HEAD/GET/g' /usr/local/etc/haproxy/auth-request.lua && \
openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 -subj "/C=AU/ST=Victoria/L=Melbourne/O=Authelia/CN=*.example.com" -keyout haproxy.key -out haproxy.crt && \
cat haproxy.key haproxy.crt > /usr/local/etc/haproxy/haproxy.pem