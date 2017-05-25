FROM node:7-alpine

WORKDIR /usr/src

COPY package.json /usr/src/package.json
RUN npm install --production

COPY dist/src/server /usr/src

ENV PORT=80
EXPOSE 80

VOLUME /etc/auth-server
VOLUME /var/lib/auth-server

CMD ["node", "index.js", "/etc/auth-server/config.yml"]
