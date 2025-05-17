FROM caddy:2.10.0-builder@sha256:c9fd63066cb907c05488886402ed9869ba8b90741a6af2d85b33ab43ba94f342 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:30ccf0cb027e1d06cd6e453c04fc1c8eec665629b22ed69602c14c8a0512ead0

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
