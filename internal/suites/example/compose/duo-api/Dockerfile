FROM node:23-alpine@sha256:d319827b3b99cca0153f6049fb584a5a4a0ae49252b6dda2314ef564f9857cf2

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
