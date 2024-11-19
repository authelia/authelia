FROM node:23-alpine@sha256:ecefaffd4706c5879af52e022fdb8ea30cbd6590e2a30d05347790d690727c6c

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
