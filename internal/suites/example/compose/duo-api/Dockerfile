FROM node:24-alpine@sha256:dec6d30010dc33ae4bb341a47b49c7add553b0a1213d9bac6b7fea1e3b5e35a4

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
