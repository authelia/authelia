FROM node:24-alpine@sha256:477a2f220673ea81a15c8ff0fdb46ffb4101210596565fdd57714cb9daa7d68e

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
