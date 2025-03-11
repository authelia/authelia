FROM node:23-alpine@sha256:dc4d20572e425f9d4c68a6f9c382fbcfec3fa2f8ef0b12cb1d96feabdb479a48

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
