FROM node:23-alpine@sha256:c61b6b12a3c96373673cd52d7ecee2314e82bca5d541eecf0bc6aee870c8c6f7

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
