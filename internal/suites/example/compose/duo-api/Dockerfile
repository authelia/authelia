FROM node:25-alpine@sha256:9a09e15ff72c69c0b9c978329a5db9df7fdeafce3f0475ac60d64e9f5d587f9e

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
