# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM node:26-alpine@sha256:ef24c5053d50fdc3e4e56eb4e7ddb7861874ab0fdc797046ba897581deb8e868

WORKDIR /usr/app/src

ENV \
    SHELL="/bin/sh" \
    PNPM_HOME="/usr/local"

COPY --link package.json pnpm-lock.yaml pnpm-workspace.yaml ./

RUN wget -qO /tmp/pnpm-install.sh https://get.pnpm.io/install.sh && \
    echo '3a6b531bfa164f5cac7c8db261cb2eb7bf929d7b8efb3a4eae82425507050e91 /tmp/pnpm-install.sh' | sha256sum -c - && \
    ENV="/root/.profile" sh /tmp/pnpm-install.sh && \
    pnpm install --frozen-lockfile --prod

EXPOSE 3000

CMD ["node", "duo_api.js"]
