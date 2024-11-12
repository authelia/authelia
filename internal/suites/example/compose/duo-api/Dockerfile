FROM node:23-alpine@sha256:f3246662f4deec464c1b1376f08f031a2b20705451d655a2174be0f8d4f6f307

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
