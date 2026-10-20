# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.7-builder@sha256:a2535ba3f1878ec781a58cf2d6edfecc47633d22e2a92447fa8ca14b3a6bfaff AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.7@sha256:685ae3d7c34d9b5f8caeddc5d14051db7bbd88ff1f5a2fa0c16bd8ca328fb7af

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
