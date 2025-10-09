FROM caddy:2.10.2-builder@sha256:5ffdd1eb7c262d118977dfd8c4a0c34c7781366fbadef3b4daeac459cb30e99e AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:614bbc6da7ec42f3c76077e86f429297047680f9cb420ad0f07a8fe049193d89

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
