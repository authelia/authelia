# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM node:26-alpine@sha256:ae8e1c19d56886270987e99b3cfcef13a919f0655b376ae322bc64ada76f88cd

WORKDIR /usr/app/src

ENV \
    SHELL="/bin/sh" \
    PNPM_HOME="/usr/local"

COPY --link package.json pnpm-lock.yaml pnpm-workspace.yaml ./

RUN wget -qO /tmp/pnpm-install.sh https://raw.githubusercontent.com/pnpm/get.pnpm.io/b7cf47fb3ca4411d8bae39a49f059f830582d3bc/install.sh && \
    ENV="/root/.profile" sh /tmp/pnpm-install.sh && \
    pnpm install --frozen-lockfile --prod

EXPOSE 3000

CMD ["node", "duo_api.js"]
