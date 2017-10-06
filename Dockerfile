FROM node:7-alpine

WORKDIR /usr/src

COPY package.json /usr/src/package.json
RUN npm install --production

COPY dist/server /usr/src/server
COPY dist/shared /usr/src/shared

ENV PORT=80
EXPOSE 80

VOLUME /etc/authelia
VOLUME /var/lib/authelia

CMD ["node", "server/src/index.js", "/etc/authelia/config.yml"]
