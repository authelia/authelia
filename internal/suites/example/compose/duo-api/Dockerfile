# SPDX-FileCopyrightText: 2026 Authelia
#
# SPDX-License-Identifier: Apache-2.0

FROM node:26-alpine@sha256:0b36e8c136b94cd4fcf02188228e76c31ad5872eef3fec8cbd2eee500cfd9e80

WORKDIR /usr/app/src

ENV \
    SHELL="/bin/sh" \
    PNPM_HOME="/usr/local"

COPY --link package.json pnpm-lock.yaml pnpm-workspace.yaml ./

RUN wget -qO /tmp/pnpm-install.sh https://raw.githubusercontent.com/pnpm/get.pnpm.io/916ce9e8412d2ce06bb7106a53ce24690b9001dd/install.sh && \
    ENV="/root/.profile" sh /tmp/pnpm-install.sh && \
    pnpm install --frozen-lockfile --prod

EXPOSE 3000

CMD ["node", "duo_api.js"]
