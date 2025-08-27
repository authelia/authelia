FROM caddy:2.10.2-builder@sha256:b33f07bbf5f8a431f224b3df0f0e709f4c7cb484515df2cdf71218e6e6021205 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:7f2af82ffc1849dff5f4b9d9b0405eae77ed6cebe6740b181ae72a668e3ed525

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
