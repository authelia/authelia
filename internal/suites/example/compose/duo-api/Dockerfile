FROM node:24-alpine@sha256:22b3c1a1171c798c0429f36272922dbb356bbab8a6d11b3b095a143d3321262a

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
