FROM node:23-alpine@sha256:c7f1c738bb26d4d800054899d25f14941422f85b9220dc08851e70fbbff66739

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
