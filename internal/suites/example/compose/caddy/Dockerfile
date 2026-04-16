FROM caddy:2.11.2-builder@sha256:930cbe20ab83d080fc0f6f0de7ddc90840ee94f9dd1ec1725e70f938ffa715dc AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.2@sha256:22e1d921a7dd98ea722ebd6819de785fd71abdab7f7fed8a2378e96d29bb923a

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
