FROM node:22-alpine@sha256:ed9736a13b88ba55cbc08c75c9edac8ae7f72840482e40324670b299336680c1

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
