FROM caddy:2.10.2-builder@sha256:5ffdd1eb7c262d118977dfd8c4a0c34c7781366fbadef3b4daeac459cb30e99e AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:61a257991a7cca3d843218fb4a4eaf8b3241443fed67f43b5a2b56848237eba0

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
