FROM node:24-alpine@sha256:2e6c7937cb36d1e4af3c261b29e862205beb7a409de01f12b6df34800cc108ec

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
