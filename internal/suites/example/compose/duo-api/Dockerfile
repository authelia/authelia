FROM node:24-alpine@sha256:820e86612c21d0636580206d802a726f2595366e1b867e564cbc652024151e8a

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
