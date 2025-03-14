FROM node:23-alpine@sha256:11ad42ffdb6a531e1bb463752708ac035962d6831053415487bd73a180e84430

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
