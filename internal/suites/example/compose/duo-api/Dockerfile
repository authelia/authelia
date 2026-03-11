FROM node:25-alpine@sha256:b087258bb49712939096cc774651f50df2f580acc5821968731ae46d4e7c200a

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
