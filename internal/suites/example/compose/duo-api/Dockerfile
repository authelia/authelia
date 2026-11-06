# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM node:26-alpine@sha256:8996bb5821d62a8fd9bb27c1cb07462f3fdce6e47d640c6a881ead0b064ce216

WORKDIR /usr/app/src

ENV \
    SHELL="/bin/sh" \
    PNPM_HOME="/usr/local"

COPY --link package.json pnpm-lock.yaml pnpm-workspace.yaml ./

RUN wget -qO /tmp/pnpm-install.sh https://raw.githubusercontent.com/pnpm/get.pnpm.io/7c26e549143daf5cf072a38c428170ad8eccaf2d/install.sh && \
    ENV="/root/.profile" sh /tmp/pnpm-install.sh && \
    pnpm install --frozen-lockfile --prod

EXPOSE 3000

CMD ["node", "duo_api.js"]
