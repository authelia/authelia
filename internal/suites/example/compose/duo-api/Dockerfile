FROM node:23-alpine@sha256:5544d48374d62bc863d047d229da035133942f913e1c9af7bc0a900706326ed2

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
