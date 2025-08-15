FROM caddy:2.10.0-builder@sha256:fe8d0d6cdbfd382197c4d715ffd5b41210165e2f64a81ce546ae63eeef35873a AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:133b5eb7ef9d42e34756ba206b06d84f4e3eb308044e268e182c2747083f09de

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
