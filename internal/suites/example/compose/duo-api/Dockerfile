FROM node:26-alpine@sha256:2d984a15c9b54fd0aeb608b8e0d0d83529eb34d2966db27a1fb4f1edc3d298a3

WORKDIR /usr/app/src

ENV \
    SHELL="/bin/sh" \
    PNPM_HOME="/usr/local"

COPY --link package.json pnpm-lock.yaml ./

RUN ENV="/root/.profile" sh -c 'wget -qO- https://get.pnpm.io/install.sh | sh -' && pnpm install --frozen-lockfile --prod

EXPOSE 3000

CMD ["node", "duo_api.js"]
