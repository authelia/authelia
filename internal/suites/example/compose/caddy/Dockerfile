# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM caddy:2.11.6-builder@sha256:9fa405e7bb28572759e7c75bea68ed39c4266a509f0c0237c464c6e13ed4dc25 AS builder

RUN xcaddy build fix-empty-copy-headers

FROM caddy:2.11.6@sha256:907efba736324e43f891ccb9d760fe5abe545e313419b3d18d63d4ec670dad8d

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
