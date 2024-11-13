FROM node:23-alpine@sha256:340cd3926b5116aae1207d6ca1e444d1009aad687c4ad37aa3afbe8a3db591d5

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
