FROM caddy:2.10.2-builder@sha256:00dfa5f39333243282c82d73de2acfe0609cb4a7879eb96946dcdfbf3a27ff87 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.2@sha256:2adb640cdc0ce1d8870887c30af1e21edfb3cdfd8433431b2a15f40119a7d654

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
