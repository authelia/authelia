FROM node:22-alpine@sha256:9fcc1a6da2b9eee38638df75c5f826e06e9c79f6a0f97f16ed98fe0ebb0725c0

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
