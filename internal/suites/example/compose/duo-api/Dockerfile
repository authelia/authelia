FROM node:23-alpine@sha256:6732970c50bfe5aaf9ebfb2a6897d610c42d1bf287cd78fb072fd65b24d6164a

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
