FROM caddy:2.11.4-builder@sha256:5f5958fecaf8a897ca5ba5951ab96adddda3128db3c45640839d19b41cb09bdc AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:df7f1c2fb114453b951de51a98efc010db1655a92c2e86be6706714e2417a78d

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
