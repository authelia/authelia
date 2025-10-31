FROM node:25-alpine@sha256:7e467cc5aa91c87e94f93c4608cf234ca24aac3ec941f7f3db207367ccccdd11

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
