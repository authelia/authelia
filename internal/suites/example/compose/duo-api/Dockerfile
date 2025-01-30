FROM node:23-alpine@sha256:15b885c5e195296ba49bab32681b478e92eb0f255078a84ad29020b87b0bdcc6

WORKDIR /usr/app/src

ADD package.json package.json
RUN yarn install --frozen-lockfile --production --silent

EXPOSE 3000

CMD ["node", "duo_api.js"]
