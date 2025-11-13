FROM node:25-alpine@sha256:ffe8cb42ec0fc97c2a2ddb36482d3c3cb2fcfabdc05f3cf1088303015d553847

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
