FROM node:23-alpine@sha256:d03e75e7ba1385c2944f4cc374eb5abe0715234f87da5121dbd64f7262ad10df

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
