FROM caddy:2.11.4-builder@sha256:c7ae80243a530d532d20062d56d6198b3ab161eb6971d28716ef7ec55599fea4 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:844f60b64e4724a5aa8245e019dace0d3f199f7433ce6c57676cb30a920dbad9

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
