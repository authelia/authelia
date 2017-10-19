FROM node:8.7.0-alpine

WORKDIR /usr/src

COPY package.json /usr/src/package.json

RUN apk update
RUN apk add make g++ python

RUN npm install --production
RUN apk del python make g++ && rm -f /var/cache/apk/*

COPY dist/server /usr/src/server
COPY dist/shared /usr/src/shared

ENV PORT=80
EXPOSE 80

VOLUME /etc/authelia
VOLUME /var/lib/authelia

CMD ["node", "server/src/index.js", "/etc/authelia/config.yml"]
