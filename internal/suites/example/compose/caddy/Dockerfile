FROM caddy:2.11.4-builder@sha256:ea6e54f62d2033b80747b022923ae0dd4f817ec4eefa2ca3a34cbebf16b6468c AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.4@sha256:1557580dffd3f309e30bd76b23a26aa3970d98bff11313cfa3a63b735ac8d04a

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
