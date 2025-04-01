FROM node:23-alpine@sha256:60c1ca4b3ec60fd9c983b54e5f9fbd46748fb07c9e14681d25765443ed30904c

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
