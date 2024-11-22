FROM node:23-alpine@sha256:419d2ede9d0fe0f30e88a12449ce62dccd54a200b49b73642ca38c0422f60c63

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
