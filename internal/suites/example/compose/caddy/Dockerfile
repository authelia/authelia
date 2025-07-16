FROM caddy:2.10.0-builder@sha256:c1e0d8e7fe650596356c46319c8c47890045bf3cd4c3d2fc6463a7ff541a99c8 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.10.0@sha256:5ce77de8c70cbdbd17ea69162b5a0539503fb9e1a9329cce42dc0f54821cc2af

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
