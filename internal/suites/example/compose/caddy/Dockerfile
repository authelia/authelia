FROM caddy:2.9.1-builder@sha256:4c455f2cf685637594f7112b2526229a58a6039975fd1915281d452b2075bda3 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.9.1@sha256:1c4bc9ead95a0888f1eea3a56ef79f30bd0d271229828fdd25090d898f553571

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
