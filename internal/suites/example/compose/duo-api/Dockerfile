FROM node:24-alpine@sha256:49e45bf002728e35c3a466737d8bcfe12c29731c7c2f3e223f9a7c794fff19a4

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
