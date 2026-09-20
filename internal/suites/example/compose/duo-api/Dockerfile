# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM node:26-alpine@sha256:dbaa92e5758cbbcf85d65d5403fdb530fe3442cbe8c6dbfb7ef23365450d5070

WORKDIR /usr/app/src

ENV \
    SHELL="/bin/sh" \
    PNPM_HOME="/usr/local"

COPY --link package.json pnpm-lock.yaml pnpm-workspace.yaml ./

RUN wget -qO /tmp/pnpm-install.sh https://get.pnpm.io/install.sh && \
    echo '44dfbba11a70a9751090894a07f8d64c9f7954cb782dfe39a1f2b2753e1d5eea /tmp/pnpm-install.sh' | sha256sum -c - && \
    ENV="/root/.profile" sh /tmp/pnpm-install.sh && \
    pnpm install --frozen-lockfile --prod

EXPOSE 3000

CMD ["node", "duo_api.js"]
