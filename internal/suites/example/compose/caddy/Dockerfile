FROM caddy:2.11.4-builder@sha256:c085abfb0fc68e5cbde7dcbf2cbf5ca416a75a6a3343372586c1851e960885cc AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:df7f1c2fb114453b951de51a98efc010db1655a92c2e86be6706714e2417a78d

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
