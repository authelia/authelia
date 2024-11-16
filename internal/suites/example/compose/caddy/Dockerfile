FROM caddy:2.8.4-builder@sha256:f21e5c514adb7a81bb3e22f0a22d6a2bd51186291753581ad991ff033be9fc3e AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:07417b42b03b8929ec80240d1a83e27a9c1346ddbff65988608fd1757269c1eb

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
