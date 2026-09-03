FROM caddy:2.11.4-builder@sha256:e04bd0aadab74d4ff446980f956fe3c63be24678f1838906962b56a7cab2d028 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:df7f1c2fb114453b951de51a98efc010db1655a92c2e86be6706714e2417a78d

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
