FROM caddy:2.10.2-builder@sha256:b6424b4a90e25fde5cb9fd8e1da716159a313869ac3ba1c34b11c50781acab81 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:2adb640cdc0ce1d8870887c30af1e21edfb3cdfd8433431b2a15f40119a7d654

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
