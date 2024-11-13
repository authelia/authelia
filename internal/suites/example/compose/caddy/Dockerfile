FROM caddy:2.8.4-builder@sha256:0c4ea007c72ccb87fad94e7cc404f02ae529eef98ceb488f42a6c87101905ab2 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.8.4@sha256:7d233cf2666671c87563ac17fea15d339c746adf754feb5363ee0bd94d6c0e7b

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
